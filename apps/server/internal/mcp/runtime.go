package mcp

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mcp-station/server/internal/storage"
	"github.com/mcp-station/server/internal/tool"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Runtime struct {
	dispatcher *MessageDispatcher
	status     string
	startedAt  *time.Time
	mu         sync.RWMutex
	writeMu    sync.Mutex
}

func NewRuntime(registry *tool.Registry) *Runtime {
	runtime := &Runtime{status: "disconnected"}
	runtime.dispatcher = NewDispatcher(registry, &runtime.writeMu)
	return runtime
}

func (rt *Runtime) Status() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.status
}

func (rt *Runtime) Uptime() int64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	if rt.startedAt != nil {
		return int64(time.Since(*rt.startedAt).Seconds())
	}
	return 0
}

func (rt *Runtime) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("mcp: upgrade error: %v", err)
		return
	}
	defer conn.Close()

	rt.mu.Lock()
	rt.status = "connected"
	now := time.Now()
	rt.startedAt = &now
	rt.mu.Unlock()

	storage.AddLog("connection", "MCP Runtime", "Client connected via WebSocket", "success")

	defer func() {
		rt.mu.Lock()
		rt.status = "disconnected"
		rt.startedAt = nil
		rt.mu.Unlock()
		storage.AddLog("connection", "MCP Runtime", "Client disconnected", "success")
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("mcp: read error: %v", err)
			break
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(msgBytes, &req); err != nil {
			rt.dispatcher.sendError(conn, nil, -32700, "Parse error")
			continue
		}

		rt.dispatcher.Dispatch(conn, &req)
	}
}
