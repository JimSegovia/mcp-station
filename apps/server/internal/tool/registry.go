package tool

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mcp-station/server/internal/model"
	"github.com/mcp-station/server/internal/opencode"
	"github.com/mcp-station/server/internal/storage"
)

type Tool struct {
	Name        string                                            `json:"name"`
	Description string                                            `json:"description"`
	InputSchema map[string]interface{}                            `json:"inputSchema"`
	Origin      string                                            `json:"origin"`
	Enabled     bool                                              `json:"enabled"`
	Execute     func(args map[string]interface{}) (string, error) `json:"-"`
}

type ToolGroup struct {
	Origin string  `json:"origin"`
	Tools  []*Tool `json:"tools"`
}

type Registry struct {
	tools          map[string]*Tool
	mu             sync.RWMutex
	opencodeClient *opencode.Client
}

func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]*Tool)}
	r.registerDefaults()
	return r
}

func (r *Registry) SetOpenCodeClient(client *opencode.Client) {
	r.mu.Lock()
	r.opencodeClient = client
	r.mu.Unlock()

	r.registerOpenCodeTools()
	r.loadOpenCodeProxyTools()
}

func (r *Registry) loadOpenCodeProxyTools() {
	r.mu.RLock()
	client := r.opencodeClient
	r.mu.RUnlock()

	if client == nil {
		return
	}

	ids, err := client.ListToolIDs()
	if err != nil {
		return
	}

	for _, id := range ids {
		toolName := "oc_" + id
		r.mu.RLock()
		_, exists := r.tools[toolName]
		r.mu.RUnlock()
		if exists {
			continue
		}
		name := id
		r.Register(&Tool{
			Name:        toolName,
			Description: "Proxy a OpenCode tool '" + id + "'",
			Origin:      "opencode",
			Enabled:     true,
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"args": map[string]interface{}{
						"type":        "object",
						"description": "Argumentos para la tool " + id,
					},
				},
			},
			Execute: func(args map[string]interface{}) (string, error) {
				toolArgs, _ := args["args"].(map[string]interface{})
				if toolArgs == nil {
					toolArgs = map[string]interface{}{}
				}
				result, err := client.ExecuteToolWithScopedSession(opencode.XiaoZhiMCPScopeKey, name, toolArgs, opencode.XiaoZhiMCPMaxIdle)
				if err != nil {
					return "", fmt.Errorf("opencode %s failed: %w", name, err)
				}
				storage.TrackOpenCodeSession(result.SessionID, "oc_"+name, fmt.Sprintf("tool=%s", name))
				return result.Text, nil
			},
		})
	}
}

func (r *Registry) RegisterExternalTools(origin string, tools []ExternalToolDef) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prefix := origin + "_"
	for _, t := range tools {
		toolDef := t
		key := prefix + t.Name
		if _, exists := r.tools[key]; exists {
			continue
		}
		r.tools[key] = &Tool{
			Name:        key,
			Description: toolDef.Description,
			Origin:      origin,
			Enabled:     true,
			InputSchema: toolDef.InputSchema,
			Execute: func(args map[string]interface{}) (string, error) {
				return fmt.Sprintf("External tool '%s' called via %s. Args: %v", toolDef.Name, origin, args), nil
			},
		}
	}
}

func (r *Registry) RegisterManagedTools(origin string, tools []ExternalToolDef, executor func(name string, args map[string]interface{}) (string, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prefix := origin + "_"
	for _, t := range tools {
		toolDef := t
		key := prefix + t.Name
		r.tools[key] = &Tool{
			Name:        key,
			Description: toolDef.Description,
			Origin:      origin,
			Enabled:     true,
			InputSchema: toolDef.InputSchema,
			Execute: func(args map[string]interface{}) (string, error) {
				return executor(toolDef.Name, args)
			},
		}
	}
}

func (r *Registry) RemoveToolsByOrigin(origin string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	prefix := origin + "_"
	for k := range r.tools {
		if strings.HasPrefix(k, prefix) {
			delete(r.tools, k)
		}
	}
}

type ExternalToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

func (r *Registry) registerDefaults() {
	d := func(t *Tool) { t.Origin = "mcp-station"; t.Enabled = true; r.registerLocked(t) }

	d(&Tool{
		Name:        "read_file",
		Description: "Lee el contenido de un archivo de texto",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Ruta del archivo a leer",
				},
			},
			"required": []string{"path"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			if path == "" {
				return "", fmt.Errorf("path is required")
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file: %w", err)
			}
			return string(data), nil
		},
	})

	d(&Tool{
		Name:        "list_directory",
		Description: "Lista el contenido de un directorio",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Ruta del directorio a listar",
				},
			},
			"required": []string{"path"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			if path == "" {
				path = "."
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return "", fmt.Errorf("failed to read directory: %w", err)
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Directory listing for: %s\n", path))
			for _, e := range entries {
				prefix := "  [F] "
				if e.IsDir() {
					prefix = "  [D] "
				}
				info, _ := e.Info()
				size := int64(0)
				if info != nil {
					size = info.Size()
				}
				sb.WriteString(fmt.Sprintf("%s%s (%d bytes)\n", prefix, e.Name(), size))
			}
			return sb.String(), nil
		},
	})

	d(&Tool{
		Name:        "create_file",
		Description: "Crea un archivo en una ruta permitida",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Ruta del archivo a crear",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Contenido del archivo",
				},
			},
			"required": []string{"path", "content"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			content, _ := args["content"].(string)
			if path == "" {
				return "", fmt.Errorf("path is required")
			}
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			return fmt.Sprintf("File created: %s (%d bytes)", path, len(content)), nil
		},
	})

	d(&Tool{
		Name:        "open_application",
		Description: "Abre una aplicacion del sistema",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Nombre o ruta de la aplicacion a abrir",
				},
			},
			"required": []string{"name"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			name, _ := args["name"].(string)
			if name == "" {
				return "", fmt.Errorf("name is required")
			}
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", name)
			case "linux":
				cmd = exec.Command("xdg-open", name)
			case "windows":
				cmd = exec.Command("cmd", "/c", "start", "", name)
			default:
				return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				return "", fmt.Errorf("failed to open application: %w", err)
			}
			go cmd.Wait()
			return fmt.Sprintf("Application '%s' opened", name), nil
		},
	})

	allowedCommands := map[string]bool{
		"echo": true, "pwd": true, "ls": true, "date": true,
		"whoami": true, "uname": true, "uptime": true,
		"which": true, "whereis": true, "find": true,
		"git": true, "go": true, "node": true, "npm": true,
		"python": true, "python3": true, "cargo": true,
	}

	d(&Tool{
		Name:        "execute_command",
		Description: "Ejecuta un comando controlado (whitelist)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Comando a ejecutar",
				},
			},
			"required": []string{"command"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			command, _ := args["command"].(string)
			if command == "" {
				return "", fmt.Errorf("command is required")
			}
			parts := strings.Fields(command)
			if len(parts) == 0 {
				return "", fmt.Errorf("empty command")
			}
			if !allowedCommands[parts[0]] {
				return fmt.Sprintf("Blocked: '%s' is not in the allowed commands whitelist", parts[0]), nil
			}
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", command)
			} else {
				cmd = exec.Command("sh", "-c", command)
			}
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output)), nil
			}
			return string(output), nil
		},
	})
}

func (r *Registry) registerOpenCodeTools() {
	r.mu.RLock()
	client := r.opencodeClient
	r.mu.RUnlock()

	if client == nil {
		return
	}

	r.Register(&Tool{
		Name:        "opencode_ask",
		Description: "Conversa con OpenCode, reutiliza la sesion reciente de XiaoZhi y al terminar devuelve una instruccion explicita de standby",
		Origin:      "opencode",
		Enabled:     true,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "La pregunta o tarea para OpenCode",
				},
				"agent": map[string]interface{}{
					"type":        "string",
					"description": "Modo del agente: build (default), plan, explore",
				},
			},
			"required": []string{"prompt"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			prompt, _ := args["prompt"].(string)
			if prompt == "" {
				return "", fmt.Errorf("prompt is required")
			}
			agent, _ := args["agent"].(string)
			result, err := client.AskWithScopedSessionDeferred(
				opencode.XiaoZhiMCPScopeKey,
				prompt,
				agent,
				"",
				opencode.XiaoZhiMCPMaxIdle,
				opencode.XiaoZhiAskDeferAfter,
			)
			if err != nil {
				return "", fmt.Errorf("opencode ask failed: %w", err)
			}
			if result.Started {
				sessionID := result.SessionID
				if result.Result != nil && result.Result.SessionID != "" {
					sessionID = result.Result.SessionID
				}
				if sessionID != "" {
					storage.TrackOpenCodeSession(sessionID, "opencode_ask", prompt)
				}
			}
			if result.Deferred {
				if result.Started {
					return "OpenCode sigue procesando tu solicitud. Pregunta luego usando opencode_check_last.", nil
				}
				return "OpenCode sigue procesando una solicitud anterior. Usa opencode_check_last para revisar cuando termine.", nil
			}
			if result.Result == nil || result.Result.Text == "" {
				return "OpenCode returned an empty response", nil
			}
			return strings.TrimSpace(result.Result.Text) + "\n\n" + opencode.XiaoZhiStandbyInstruction, nil
		},
	})

	r.Register(&Tool{
		Name:        "opencode_check_last",
		Description: "Lee el ultimo mensaje final del assistant en la sesion activa compartida de XiaoZhi sin crear una sesion nueva",
		Origin:      "opencode",
		Enabled:     true,
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			text, err := client.GetScopedSessionLastAssistantText(opencode.XiaoZhiMCPScopeKey)
			if err != nil {
				if opencode.IsOpenCodeSessionNotFoundError(err) {
					return "La sesion activa de XiaoZhi ya no existe en OpenCode.", nil
				}
				return "", fmt.Errorf("opencode check last failed: %w", err)
			}
			if text == "" {
				binding, bindingErr := storage.GetOpenCodeSessionBinding(opencode.XiaoZhiMCPScopeKey)
				if bindingErr != nil {
					return "", fmt.Errorf("load scoped session binding: %w", bindingErr)
				}
				if binding == nil {
					return "No hay una sesion activa de XiaoZhi en OpenCode.", nil
				}
				return "La sesion existe pero aun no hay respuesta final de OpenCode.", nil
			}
			return text, nil
		},
	})

	r.Register(&Tool{
		Name:        "opencode_run",
		Description: "Despacha una tarea asincrona en una sesion nueva y devuelve solo el ID; no recomendada para XiaoZhi",
		Origin:      "opencode",
		Enabled:     false,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "La tarea a delegar a OpenCode",
				},
				"agent": map[string]interface{}{
					"type":        "string",
					"description": "Modo del agente",
				},
			},
			"required": []string{"prompt"},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			prompt, _ := args["prompt"].(string)
			if prompt == "" {
				return "", fmt.Errorf("prompt is required")
			}
			agent, _ := args["agent"].(string)
			sessionID, err := client.Run(prompt, agent)
			if err != nil {
				return "", fmt.Errorf("opencode run failed: %w", err)
			}
			storage.TrackOpenCodeSession(sessionID, "opencode_run", prompt)
			return fmt.Sprintf("Task dispatched to OpenCode\nSession ID: %s", sessionID), nil
		},
	})
}

func (r *Registry) registerLocked(t *Tool) {
	r.tools[t.Name] = t
}

func (r *Registry) Register(t *Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.Origin == "" {
		t.Origin = "external"
	}
	r.tools[t.Name] = t
}

func (r *Registry) List() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *Registry) ListEnabled() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		if t.Enabled {
			list = append(list, t)
		}
	}
	return list
}

func (r *Registry) ListByOrigin() []ToolGroup {
	r.mu.RLock()
	defer r.mu.RUnlock()
	groups := make(map[string][]*Tool)
	var order []string

	for _, t := range r.tools {
		origin := t.Origin
		if origin == "" {
			origin = "unknown"
		}
		if _, ok := groups[origin]; !ok {
			order = append(order, origin)
		}
		groups[origin] = append(groups[origin], t)
	}

	result := make([]ToolGroup, 0, len(order))
	for _, origin := range order {
		result = append(result, ToolGroup{Origin: origin, Tools: groups[origin]})
	}
	return result
}

func (r *Registry) Toggle(name string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tools[name]
	if !ok {
		return false, fmt.Errorf("tool not found: %s", name)
	}
	t.Enabled = !t.Enabled
	return t.Enabled, nil
}

func (r *Registry) SetEnabled(name string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tools[name]
	if !ok {
		return fmt.Errorf("tool not found: %s", name)
	}
	t.Enabled = enabled
	return nil
}

func (r *Registry) SetOriginEnabled(origin string, enabled bool) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, t := range r.tools {
		if t.Origin == origin {
			t.Enabled = enabled
			count++
		}
	}
	return count
}

func (r *Registry) CountEnabledByOrigin(origin string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, t := range r.tools {
		if t.Origin == origin && t.Enabled {
			count++
		}
	}
	return count
}

func (r *Registry) IsEnabled(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tools[name]
	return ok && t.Enabled
}

func (r *Registry) ToolsByOrigin(origin string) []model.McpTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]model.McpTool, 0)
	for _, t := range r.tools {
		if t.Origin == origin {
			tools = append(tools, model.McpTool{
				Name:        t.Name,
				Description: t.Description,
				Enabled:     t.Enabled,
			})
		}
	}
	return tools
}

func (r *Registry) Call(name string, args map[string]interface{}) (string, error) {
	r.mu.RLock()
	t, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	if !t.Enabled {
		return "", fmt.Errorf("tool is disabled: %s", name)
	}

	return t.Execute(args)
}
