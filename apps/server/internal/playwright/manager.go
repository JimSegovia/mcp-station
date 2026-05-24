package playwright

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mcp-station/server/internal/storage"
	"github.com/mcp-station/server/internal/tool"
)

const (
	defaultStartupTimeout = 2 * time.Minute
	defaultToolTimeout    = 5 * time.Minute
)

type Manager struct {
	command string
	args    []string

	httpClient *http.Client

	mu              sync.RWMutex
	rpcMu           sync.Mutex
	cmd             *exec.Cmd
	nextID          int
	enabled         bool
	status          string
	tools           []tool.ExternalToolDef
	lastConnectedAt *string
	lastError       string
	generation      int64
	port            int
	fixedPort       int
	endpoint        string
	sessionID       string
	protocolVersion string
}

type jsonrpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Method  string          `json:"method"`
	Result  json.RawMessage `json:"result"`
	Error   *jsonrpcError   `json:"error"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewManager(command string, args []string, fixedPort int) *Manager {
	return &Manager{
		command:   command,
		args:      append([]string(nil), args...),
		fixedPort: fixedPort,
		status:    "disconnected",
		nextID:    1,
		httpClient: &http.Client{
			Timeout: defaultToolTimeout,
		},
	}
}

func (m *Manager) Endpoint() string {
	m.mu.RLock()
	endpoint := m.endpoint
	port := m.port
	m.mu.RUnlock()

	if endpoint != "" {
		return endpoint
	}
	if port > 0 {
		return fmt.Sprintf("http://localhost:%d/mcp", port)
	}
	return "playwright-mcp (managed)"
}

func (m *Manager) Enabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

func (m *Manager) Status() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *Manager) Tools() []tool.ExternalToolDef {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tools := make([]tool.ExternalToolDef, len(m.tools))
	copy(tools, m.tools)
	return tools
}

func (m *Manager) LastConnectedAt() *string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.lastConnectedAt == nil {
		return nil
	}
	value := *m.lastConnectedAt
	return &value
}

func (m *Manager) LastError() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastError
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.cmd != nil && m.status == "connected" {
		m.enabled = true
		m.mu.Unlock()
		return nil
	}
	m.enabled = true
	m.status = "connecting"
	m.lastError = ""
	m.mu.Unlock()

	var port int
	if m.fixedPort > 0 {
		port = m.fixedPort
	} else {
		p, err := reservePort()
		if err != nil {
			m.markError(fmt.Errorf("reserve Playwright MCP port: %w", err))
			return err
		}
		port = p
	}

	command, args := m.commandSpec(port)
	cmd := exec.CommandContext(context.Background(), command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.markError(fmt.Errorf("open stdout: %w", err))
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.markError(fmt.Errorf("open stderr: %w", err))
		return err
	}

	if err := cmd.Start(); err != nil {
		m.markError(fmt.Errorf("start command: %w", err))
		return err
	}

	endpoint := fmt.Sprintf("http://localhost:%d/mcp", port)

	m.mu.Lock()
	m.cmd = cmd
	m.nextID = 1
	m.port = port
	m.endpoint = endpoint
	m.sessionID = ""
	m.protocolVersion = "2025-06-18"
	m.generation++
	generation := m.generation
	m.mu.Unlock()

	go m.drainOutput("playwright-mcp", stdout)
	go m.drainOutput("playwright-mcp", stderr)
	go m.waitForExit(cmd, generation)

	startCtx, cancel := context.WithTimeout(ctx, defaultStartupTimeout)
	defer cancel()

	var initResult struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := m.waitUntilReady(startCtx, &initResult); err != nil {
		m.abortCurrentProcess(fmt.Errorf("initialize Playwright MCP: %w", err), true)
		return err
	}

	if initResult.ProtocolVersion != "" {
		m.mu.Lock()
		m.protocolVersion = initResult.ProtocolVersion
		m.mu.Unlock()
	}

	if err := m.notify("notifications/initialized", nil); err != nil {
		m.abortCurrentProcess(fmt.Errorf("notify initialized: %w", err), true)
		return err
	}

	var toolsResult struct {
		Tools []tool.ExternalToolDef `json:"tools"`
	}
	if err := m.request(startCtx, "tools/list", map[string]interface{}{}, &toolsResult); err != nil {
		m.abortCurrentProcess(fmt.Errorf("discover Playwright tools: %w", err), true)
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	m.mu.Lock()
	if m.cmd == cmd && m.generation == generation {
		m.tools = toolsResult.Tools
		m.status = "connected"
		m.lastConnectedAt = &now
		m.lastError = ""
	}
	m.mu.Unlock()

	storage.AddLog("connection", "Playwright MCP", "Playwright MCP server connected", "success")
	return nil
}

func (m *Manager) Stop() {
	m.abortCurrentProcess(nil, false)
}

func (m *Manager) CallTool(toolName string, args map[string]interface{}) (string, error) {
	m.mu.RLock()
	status := m.status
	enabled := m.enabled
	m.mu.RUnlock()

	if status != "connected" {
		if !enabled {
			return "", fmt.Errorf("playwright server is disabled")
		}
		if err := m.Start(context.Background()); err != nil {
			return "", err
		}
	}

	callCtx, cancel := context.WithTimeout(context.Background(), defaultToolTimeout)
	defer cancel()

	var result struct {
		Content []map[string]interface{} `json:"content"`
		IsError bool                     `json:"isError"`
	}
	if err := m.request(callCtx, "tools/call", map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}, &result); err != nil {
		return "", err
	}

	text := extractToolText(result.Content)
	if result.IsError {
		if text == "" {
			text = "Playwright MCP tool returned an error"
		}
		return "", fmt.Errorf("%s", text)
	}
	if text == "" {
		return "Playwright MCP tool completed without text output.", nil
	}
	return text, nil
}

func extractToolText(content []map[string]interface{}) string {
	parts := make([]string, 0, len(content))
	for _, item := range content {
		itemType, _ := item["type"].(string)
		switch itemType {
		case "text":
			if text, ok := item["text"].(string); ok && strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		default:
			if encoded, err := json.Marshal(item); err == nil {
				parts = append(parts, string(encoded))
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func (m *Manager) waitUntilReady(ctx context.Context, out interface{}) error {
	ticker := time.NewTicker(750 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		err := m.request(ctx, "initialize", map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-station",
				"version": "0.1.0",
			},
		}, out)
		if err == nil {
			return nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return lastErr
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *Manager) request(ctx context.Context, method string, params interface{}, out interface{}) error {
	m.rpcMu.Lock()
	defer m.rpcMu.Unlock()

	reqID := m.nextRequestID()
	payload := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      &reqID,
		Method:  method,
		Params:  params,
	}

	resp, err := m.doPost(ctx, payload, true)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(resp.Result, out); err != nil {
		return fmt.Errorf("decode %s response: %w", method, err)
	}
	return nil
}

func (m *Manager) notify(method string, params interface{}) error {
	m.rpcMu.Lock()
	defer m.rpcMu.Unlock()

	_, err := m.doPost(context.Background(), jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}, false)
	return err
}

func (m *Manager) doPost(ctx context.Context, payload jsonrpcRequest, expectResponse bool) (*jsonrpcResponse, error) {
	m.mu.RLock()
	endpoint := m.endpoint
	sessionID := m.sessionID
	protocolVersion := m.protocolVersion
	cmd := m.cmd
	m.mu.RUnlock()

	if cmd == nil || endpoint == "" {
		return nil, fmt.Errorf("playwright server is not running")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if protocolVersion != "" {
		httpReq.Header.Set("MCP-Protocol-Version", protocolVersion)
	}
	if sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", sessionID)
	}

	httpResp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if sid := httpResp.Header.Get("Mcp-Session-Id"); sid != "" {
		m.mu.Lock()
		m.sessionID = sid
		m.mu.Unlock()
	}

	if !expectResponse {
		if httpResp.StatusCode != http.StatusAccepted && httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
			data, _ := io.ReadAll(httpResp.Body)
			return nil, fmt.Errorf("unexpected HTTP %d: %s", httpResp.StatusCode, strings.TrimSpace(string(data)))
		}
		return nil, nil
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		data, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("unexpected HTTP %d: %s", httpResp.StatusCode, strings.TrimSpace(string(data)))
	}

	contentType := strings.ToLower(httpResp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "text/event-stream") {
		return readSSEResponse(ctx, httpResp.Body, payload.ID)
	}

	var resp jsonrpcResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func readSSEResponse(ctx context.Context, body io.Reader, reqID *int) (*jsonrpcResponse, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var dataLines []string
	flush := func() (*jsonrpcResponse, bool, error) {
		if len(dataLines) == 0 {
			return nil, false, nil
		}
		raw := strings.Join(dataLines, "\n")
		dataLines = nil

		var resp jsonrpcResponse
		if err := json.Unmarshal([]byte(raw), &resp); err != nil {
			return nil, false, err
		}
		if reqID != nil && resp.ID != nil && *resp.ID == *reqID {
			return &resp, true, nil
		}
		return nil, false, nil
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			if resp, done, err := flush(); err != nil {
				return nil, err
			} else if done {
				return resp, nil
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if resp, done, err := flush(); err != nil {
		return nil, err
	} else if done {
		return resp, nil
	}
	return nil, fmt.Errorf("sse stream ended before response")
}

func (m *Manager) nextRequestID() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextID
	m.nextID++
	return id
}

func (m *Manager) waitForExit(cmd *exec.Cmd, generation int64) {
	err := cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != cmd || m.generation != generation {
		return
	}

	m.cmd = nil
	m.port = 0
	m.endpoint = ""
	m.sessionID = ""
	if m.enabled {
		m.status = "error"
		if err != nil {
			m.lastError = err.Error()
		} else {
			m.lastError = "playwright process exited"
		}
	} else {
		m.status = "disconnected"
		m.lastError = ""
	}
}

func (m *Manager) abortCurrentProcess(err error, keepEnabled bool) {
	m.mu.Lock()
	cmd := m.cmd
	m.cmd = nil
	m.port = 0
	m.endpoint = ""
	m.sessionID = ""
	m.generation++
	if keepEnabled {
		m.enabled = true
		m.status = "error"
		if err != nil {
			m.lastError = err.Error()
		}
	} else {
		m.enabled = false
		m.status = "disconnected"
		m.lastError = ""
	}
	m.mu.Unlock()

	if err != nil {
		log.Printf("playwright-mcp: %v", err)
		storage.AddLog("connection", "Playwright MCP", err.Error(), "error")
	}

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (m *Manager) markError(err error) {
	m.mu.Lock()
	m.status = "error"
	m.lastError = err.Error()
	m.mu.Unlock()
	log.Printf("playwright-mcp: %v", err)
	storage.AddLog("connection", "Playwright MCP", err.Error(), "error")
}

func (m *Manager) drainOutput(prefix string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		log.Printf("%s: %s", prefix, line)
	}
}

func (m *Manager) commandSpec(port int) (string, []string) {
	if m.command != "" {
		return m.command, replacePortTokens(m.args, port)
	}

	if runtime.GOOS == "windows" {
		return "wsl", []string{
			"bash",
			"-lc",
			fmt.Sprintf("npx -y @playwright/mcp@latest --headless --port %d --host 0.0.0.0", port),
		}
	}

	return "npx", []string{
		"-y",
		"@playwright/mcp@latest",
		"--headless",
		"--port",
		strconv.Itoa(port),
	}
}

func replacePortTokens(args []string, port int) []string {
	out := make([]string, len(args))
	needle := "{port}"
	portValue := strconv.Itoa(port)
	for i, arg := range args {
		out[i] = strings.ReplaceAll(arg, needle, portValue)
	}
	return out
}

func reservePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected listener addr %T", listener.Addr())
	}
	return addr.Port, nil
}
