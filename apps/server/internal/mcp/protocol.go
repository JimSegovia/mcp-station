package mcp

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mcp-station/server/internal/storage"
	"github.com/mcp-station/server/internal/tool"
)

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type jsonrpcResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      *int          `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *jsonrpcError `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type jsonrpcEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id"`
	Method  string          `json:"method"`
	Result  json.RawMessage `json:"result"`
	Error   *jsonrpcError   `json:"error"`
}

func (e *jsonrpcEnvelope) isRequest() bool { return e.Method != "" }
func (e *jsonrpcEnvelope) isResponse() bool {
	return e.Method == "" && (e.Result != nil || e.Error != nil)
}
func (e *jsonrpcEnvelope) isNotification() bool { return e.Method != "" && e.ID == nil }

type MessageDispatcher struct {
	registry *tool.Registry
	writeMu  *sync.Mutex
}

var (
	mcpWriteTimeout          = 10 * time.Second
	toolCallProgressInterval = 3 * time.Second
)

func NewDispatcher(registry *tool.Registry, writeMu *sync.Mutex) *MessageDispatcher {
	return &MessageDispatcher{registry: registry, writeMu: writeMu}
}

func (d *MessageDispatcher) Dispatch(conn *websocket.Conn, req *jsonrpcRequest) {
	switch req.Method {
	case "initialize":
		d.handleInitialize(conn, req)
	case "tools/list":
		d.handleToolsList(conn, req)
	case "tools/call":
		d.handleToolsCall(conn, req)
	case "notifications/initialized":
		log.Println("mcp: peer initialized")
		storage.AddLog("connection", "MCP Client", "MCP session initialized (peer notified)", "info")
	case "ping":
		if req.ID != nil {
			d.sendResult(conn, req.ID, map[string]interface{}{})
		}
	default:
		log.Printf("mcp: unknown method: %s", req.Method)
		d.sendError(conn, req.ID, -32601, "Method not found: "+req.Method)
	}
}

func (d *MessageDispatcher) handleResponse(conn *websocket.Conn, env *jsonrpcEnvelope) {
	if env.Error != nil {
		log.Printf("mcp: jsonrpc error id=%v code=%d message=%s", env.ID, env.Error.Code, env.Error.Message)
		return
	}
	if env.ID != nil {
		log.Printf("mcp: jsonrpc response id=%d result=%s", *env.ID, string(env.Result))
	}
}

func (d *MessageDispatcher) handleInitialize(conn *websocket.Conn, req *jsonrpcRequest) {
	d.sendResult(conn, req.ID, map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]interface{}{
			"name":    "mcp-station",
			"version": "0.1.0",
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
	})
}

func (d *MessageDispatcher) handleToolsList(conn *websocket.Conn, req *jsonrpcRequest) {
	tools := d.registry.ListEnabled()
	toolList := make([]map[string]interface{}, 0, len(tools))
	for _, t := range tools {
		toolList = append(toolList, map[string]interface{}{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	d.sendResult(conn, req.ID, map[string]interface{}{"tools": toolList})
}

func (d *MessageDispatcher) handleToolsCall(conn *websocket.Conn, req *jsonrpcRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
		Meta      map[string]interface{} `json:"_meta"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		d.sendError(conn, req.ID, -32602, "Invalid params")
		return
	}

	id := cloneRequestID(req.ID)
	args := params.Arguments
	if args == nil {
		args = map[string]interface{}{}
	}
	name := params.Name
	progressToken := resolveProgressToken(params.Meta, req.ID)
	startedAt := time.Now()
	storage.AddLog("tool_call", name, "Tool call started", "info")

	go func() {
		resultCh := make(chan struct {
			text string
			err  error
		}, 1)

		go func() {
			result, err := d.registry.Call(name, args)
			resultCh <- struct {
				text string
				err  error
			}{text: result, err: err}
		}()

		d.sendProgress(conn, progressToken, name, 0)

		ticker := time.NewTicker(toolCallProgressInterval)
		defer ticker.Stop()

		for {
			select {
			case result := <-resultCh:
				if result.err != nil {
					storage.AddLog("tool_call", name, result.err.Error(), "error")
					d.sendError(conn, id, -32000, result.err.Error())
					return
				}

				storage.AddLog("tool_call", name, "Tool executed successfully", "success")
				d.sendResult(conn, id, map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": result.text,
						},
					},
				})
				return
			case <-ticker.C:
				d.sendProgress(conn, progressToken, name, time.Since(startedAt))
			}
		}
	}()
}

func (d *MessageDispatcher) sendResult(conn *websocket.Conn, id *int, result interface{}) {
	resp := jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	d.writeJSON(conn, resp)
}

func (d *MessageDispatcher) sendError(conn *websocket.Conn, id *int, code int, message string) {
	resp := jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonrpcError{Code: code, Message: message},
	}
	d.writeJSON(conn, resp)
}

func (d *MessageDispatcher) writeJSON(conn *websocket.Conn, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("mcp: marshal error: %v", err)
		return
	}
	if d.writeMu != nil {
		d.writeMu.Lock()
		defer d.writeMu.Unlock()
	}
	conn.SetWriteDeadline(time.Now().Add(mcpWriteTimeout))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("mcp: write error: %v", err)
	}
}

func (d *MessageDispatcher) sendProgress(conn *websocket.Conn, token interface{}, toolName string, elapsed time.Duration) {
	params := map[string]interface{}{
		"progressToken": token,
		"progress":      int(elapsed.Seconds()),
		"message":       "Waiting for " + toolName + " to finish",
	}

	d.writeJSON(conn, map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/progress",
		"params":  params,
	})
}

func cloneRequestID(id *int) *int {
	if id == nil {
		return nil
	}
	cloned := *id
	return &cloned
}

func resolveProgressToken(meta map[string]interface{}, id *int) interface{} {
	if meta == nil {
		if id == nil {
			return "mcp-station-progress"
		}
		return *id
	}
	token, ok := meta["progressToken"]
	if !ok || token == nil {
		if id == nil {
			return "mcp-station-progress"
		}
		return *id
	}
	return token
}
