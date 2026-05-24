package mcp

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mcp-station/server/internal/storage"
	"github.com/mcp-station/server/internal/tool"

	_ "modernc.org/sqlite"
)

func TestLongToolCallKeepsSocketResponsive(t *testing.T) {
	setupMCPTestDB(t)

	previousInterval := toolCallProgressInterval
	toolCallProgressInterval = 100 * time.Millisecond
	defer func() {
		toolCallProgressInterval = previousInterval
	}()

	registry := tool.NewRegistry()
	release := make(chan struct{})
	registry.Register(&tool.Tool{
		Name:        "slow_tool",
		Description: "Test tool that blocks until released",
		Origin:      "test",
		Enabled:     true,
		InputSchema: map[string]interface{}{"type": "object"},
		Execute: func(args map[string]interface{}) (string, error) {
			<-release
			return "finished", nil
		},
	})

	runtime := NewRuntime(registry)
	server := httptest.NewServer(http.HandlerFunc(runtime.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	writeJSONRPC(t, conn, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "0.1.0",
			},
		},
	})

	initEnvelope := readEnvelope(t, conn, 2*time.Second)
	if initEnvelope["id"].(float64) != 1 {
		t.Fatalf("expected initialize response id=1, got %#v", initEnvelope)
	}

	writeJSONRPC(t, conn, map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})

	writeJSONRPC(t, conn, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "slow_tool",
			"arguments": map[string]interface{}{},
			"_meta": map[string]interface{}{
				"progressToken": "token-1",
			},
		},
	})

	writeJSONRPC(t, conn, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "ping",
	})

	var sawProgress bool
	var sawPingResponse bool
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && (!sawProgress || !sawPingResponse) {
		envelope := readEnvelope(t, conn, 2*time.Second)
		if method, _ := envelope["method"].(string); method == "notifications/progress" {
			params := envelope["params"].(map[string]interface{})
			if params["progressToken"] == "token-1" {
				sawProgress = true
			}
			continue
		}
		if id, ok := envelope["id"].(float64); ok && int(id) == 3 {
			sawPingResponse = true
		}
	}

	if !sawProgress {
		t.Fatal("expected progress notification while tool call was running")
	}
	if !sawPingResponse {
		t.Fatal("expected ping response while tool call was still running")
	}

	close(release)

	for {
		envelope := readEnvelope(t, conn, 2*time.Second)
		if id, ok := envelope["id"].(float64); !ok || int(id) != 2 {
			continue
		}
		result := envelope["result"].(map[string]interface{})
		content := result["content"].([]interface{})
		firstItem := content[0].(map[string]interface{})
		if firstItem["text"] != "finished" {
			t.Fatalf("expected final tool result, got %#v", envelope)
		}
		return
	}
}

func writeJSONRPC(t *testing.T, conn *websocket.Conn, payload map[string]interface{}) {
	t.Helper()
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := conn.WriteJSON(payload); err != nil {
		t.Fatalf("write jsonrpc: %v", err)
	}
}

func readEnvelope(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}

	var envelope map[string]interface{}
	if err := json.Unmarshal(msg, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return envelope
}

func setupMCPTestDB(t *testing.T) {
	t.Helper()

	oldDB := storage.DB
	dbPath := filepath.Join(t.TempDir(), "mcp-test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)

	storage.DB = db

	if _, err := storage.DB.Exec(`CREATE TABLE logs (
		id TEXT PRIMARY KEY,
		timestamp TEXT NOT NULL,
		type TEXT NOT NULL,
		source TEXT NOT NULL,
		message TEXT NOT NULL,
		result TEXT NOT NULL
	)`); err != nil {
		t.Fatalf("create logs table: %v", err)
	}

	t.Cleanup(func() {
		storage.DB.Close()
		storage.DB = oldDB
	})
}
