package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mcp-station/server/internal/api"
	"github.com/mcp-station/server/internal/mcp"
	"github.com/mcp-station/server/internal/model"
	"github.com/mcp-station/server/internal/opencode"
	"github.com/mcp-station/server/internal/playwright"
	"github.com/mcp-station/server/internal/storage"
	"github.com/mcp-station/server/internal/tool"
)

func main() {
	port := flag.String("port", "8080", "HTTP server port")
	opencodePort := flag.Int("opencode-port", 4096, "OpenCode serve port")
	opencodeBin := flag.String("opencode-bin", "", "Path to opencode binary (auto-detect if empty, uses wsl on Windows)")
	playwrightPort := flag.Int("playwright-port", 0, "Fixed port for Playwright MCP (0 for dynamic)")
	flag.Parse()

	storage.Init()

	registry := tool.NewRegistry()

	ocManager := opencode.NewManager(*opencodePort,
		os.Getenv("OPENCODE_SERVER_USERNAME"),
		os.Getenv("OPENCODE_SERVER_PASSWORD"),
		*opencodeBin,
	)
	pwManager := playwright.NewManager("", nil, *playwrightPort)

	syncIntegrationTools(registry)

	runtime := mcp.NewRuntime(registry)
	client := mcp.NewClient(registry)

	handler := api.NewRouter(runtime, registry, client, ocManager, pwManager)

	defer ocManager.Stop()
	defer pwManager.Stop()

	go runSessionCleanup(ocManager)

	go startOpenCodeBackground(ocManager, registry)

	log.Printf("MCP Station listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}

func startOpenCodeBackground(ocManager *opencode.Manager, registry *tool.Registry) {
	time.Sleep(500 * time.Millisecond)

	ctx := context.Background()
	if err := ocManager.Start(ctx); err != nil {
		log.Printf("WARNING: OpenCode serve start failed: %v", err)
		return
	}

	if err := ocManager.WaitReady(90 * time.Second); err != nil {
		log.Printf("WARNING: OpenCode serve not ready: %v (check terminal in /opencode panel)", err)
		return
	}

	ensureOpenCodeConfig(ocManager)

	registry.SetOpenCodeClient(ocManager.Client())
	syncIntegrationTools(registry)
	log.Println("OpenCode integration ready — tools synced")
}

func ensureOpenCodeConfig(ocManager *opencode.Manager) {
	cmdStr := `mkdir -p ~/.config/opencode && echo '{"model":"opencode-go/deepseek-v4-flash"}' > ~/.config/opencode/opencode.json`
	ocManager.RunWSLCommand(cmdStr)
}

func runSessionCleanup(ocManager *opencode.Manager) {
	doClean := func() {
		ids, err := storage.CleanExpiredSessions(24 * time.Hour)
		if err != nil {
			log.Printf("session cleanup: storage error: %v", err)
			return
		}
		for _, sid := range ids {
			if err := ocManager.Client().DeleteSession(sid); err != nil {
				log.Printf("session cleanup: failed to delete %s: %v", sid, err)
			}
		}
		if len(ids) > 0 {
			log.Printf("session cleanup: removed %d expired sessions", len(ids))
		}
	}

	doClean()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		doClean()
	}
}

func syncIntegrationTools(registry *tool.Registry) {
	tools := registry.List()
	mcpTools := make([]model.McpTool, 0, len(tools))
	for _, t := range tools {
		mcpTools = append(mcpTools, model.McpTool{
			Name:        t.Name,
			Description: t.Description,
			Enabled:     true,
		})
	}
	toolsJSON, _ := json.Marshal(mcpTools)
	storage.DB.Exec(`UPDATE integrations SET tools_json = ? WHERE id = 1`, string(toolsJSON))
}
