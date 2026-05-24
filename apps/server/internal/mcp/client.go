package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/mcp-station/server/internal/storage"
	"github.com/mcp-station/server/internal/tool"
)

const mcpKeepaliveInterval = 10 * time.Second

type Client struct {
	dispatcher   *MessageDispatcher
	conn         *websocket.Conn
	endpoint     string
	status       string
	mu           sync.RWMutex
	writeMu      sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	reconnectSeq int
	initID       int
	initialized  bool
}

func NewClient(registry *tool.Registry) *Client {
	client := &Client{status: "disconnected"}
	client.dispatcher = NewDispatcher(registry, &client.writeMu)
	return client
}

func (c *Client) Status() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *Client) Connect(endpoint string) {
	c.mu.Lock()
	if c.status == "connected" || c.status == "connecting" {
		c.mu.Unlock()
		return
	}
	c.endpoint = endpoint
	c.status = "connecting"
	c.reconnectSeq = 0
	c.initID = 0
	c.initialized = false
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	go c.run()
}

func (c *Client) Disconnect() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.status = "disconnected"
	c.initID = 0
	c.initialized = false
	c.mu.Unlock()
}

func backoff(seq int) time.Duration {
	switch {
	case seq <= 1:
		return 2 * time.Second
	case seq <= 3:
		return 5 * time.Second
	case seq <= 6:
		return 15 * time.Second
	default:
		return 30 * time.Second
	}
}

func (c *Client) run() {
	for {
		c.mu.RLock()
		ctx := c.ctx
		endpoint := c.endpoint
		seq := c.reconnectSeq
		c.mu.RUnlock()

		select {
		case <-ctx.Done():
			return
		default:
		}

		if seq > 0 {
			c.mu.Lock()
			c.status = "connecting"
			c.mu.Unlock()
			storage.SetIntegrationConnecting()
		}

		err := c.dialAndServe(ctx, endpoint)
		if err != nil {
			log.Printf("mcp-client: %v", err)
			storage.AddLog("connection", "MCP Client", fmt.Sprintf("Connection lost (attempt %d): %v", c.reconnectSeq+1, err), "error")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff(c.reconnectSeq)):
		}

		c.mu.Lock()
		c.reconnectSeq++
		c.mu.Unlock()
	}
}

func (c *Client) DiscoverTools(ctx context.Context, endpoint string) ([]tool.ExternalToolDef, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	initMsg, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-station",
				"version": "0.1.0",
			},
		},
	})
	if err := conn.WriteMessage(websocket.TextMessage, initMsg); err != nil {
		return nil, fmt.Errorf("write initialize: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, initResp, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read initialize response: %w", err)
	}

	var initEnv jsonrpcEnvelope
	if err := json.Unmarshal(initResp, &initEnv); err != nil {
		return nil, fmt.Errorf("parse initialize response: %w", err)
	}
	if initEnv.Error != nil {
		return nil, fmt.Errorf("initialize error: %s", initEnv.Error.Message)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	notifyMsg, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	conn.WriteMessage(websocket.TextMessage, notifyMsg)

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	listMsg, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})
	if err := conn.WriteMessage(websocket.TextMessage, listMsg); err != nil {
		return nil, fmt.Errorf("write tools/list: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, listResp, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read tools/list response: %w", err)
	}

	var listEnv jsonrpcEnvelope
	if err := json.Unmarshal(listResp, &listEnv); err != nil {
		return nil, fmt.Errorf("parse tools/list response: %w", err)
	}
	if listEnv.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s", listEnv.Error.Message)
	}

	var toolsResult struct {
		Tools []tool.ExternalToolDef `json:"tools"`
	}
	if err := json.Unmarshal(listEnv.Result, &toolsResult); err != nil {
		return nil, fmt.Errorf("parse tools: %w", err)
	}

	return toolsResult.Tools, nil
}

func (c *Client) dialAndServe(ctx context.Context, endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, resp, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		c.mu.Lock()
		c.status = "error"
		c.mu.Unlock()
		storage.SetIntegrationError("Dial failed: " + err.Error())
		return fmt.Errorf("dial: %w", err)
	}
	if resp != nil {
		log.Printf("mcp-client: HTTP %d", resp.StatusCode)
	}

	c.mu.Lock()
	c.conn = conn
	c.status = "connected"
	seq := c.reconnectSeq
	c.reconnectSeq = 0
	c.initialized = false
	c.mu.Unlock()

	log.Printf("mcp-client: connected to %s (attempt %d)", endpoint, seq+1)
	storage.AddLog("connection", "MCP Client", "Connected to XiaoZhi MCP endpoint", "success")
	storage.SetIntegrationConnected()

	readDeadline := 120 * time.Second
	conn.SetReadDeadline(time.Now().Add(readDeadline))

	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("mcp-client: close frame received code=%d reason=%q", code, text)
		storage.AddLog("connection", "MCP Client", fmt.Sprintf("Server closed connection: code=%d reason=%s", code, text), "warning")
		return nil
	})

	keepaliveDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(mcpKeepaliveInterval)
		defer ticker.Stop()
		for {
			select {
			case <-keepaliveDone:
				return
			case <-ticker.C:
				c.mu.RLock()
				c2 := c.conn
				c.mu.RUnlock()
				if c2 == nil {
					return
				}
				pingMsg := jsonrpcRequest{
					JSONRPC: "2.0",
					Method:  "ping",
				}
				if err := c.writeJSON(c2, pingMsg); err != nil {
					return
				}
			}
		}
	}()
	defer close(keepaliveDone)

	defer func() {
		c.mu.Lock()
		if c.conn == conn {
			c.status = "disconnected"
			c.conn = nil
		}
		c.mu.Unlock()
		conn.Close()
		storage.AddLog("connection", "MCP Client", "Disconnected from XiaoZhi MCP endpoint", "info")
		if ctx.Err() == nil {
			storage.SetIntegrationDisconnected("WebSocket disconnected")
		}
	}()

	c.sendInitialize(conn)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(readDeadline))
		msgType, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if ce, ok := err.(*websocket.CloseError); ok {
				return fmt.Errorf("ws close: code=%d reason=%s", ce.Code, ce.Text)
			}
			return fmt.Errorf("read: %w", err)
		}

		if msgType != websocket.TextMessage {
			continue
		}

		var env jsonrpcEnvelope
		if err := json.Unmarshal(msgBytes, &env); err != nil {
			c.dispatcher.sendError(conn, nil, -32700, "Parse error")
			continue
		}

		if env.isRequest() {
			var req jsonrpcRequest
			json.Unmarshal(msgBytes, &req)
			c.dispatcher.Dispatch(conn, &req)
		} else if env.isResponse() {
			c.handleResponse(conn, &env)
		}
	}
}

func (c *Client) sendInitialize(conn *websocket.Conn) {
	now := time.Now()
	initID := int(now.UnixMilli() % 100000)
	log.Printf("mcp-client: sending initialize (id=%d)", initID)

	msgBytes, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      initID,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-station",
				"version": "0.1.0",
			},
		},
	})

	if err := c.writeRawMessage(conn, msgBytes); err != nil {
		log.Printf("mcp-client: failed to send initialize: %v", err)
		storage.AddLog("connection", "MCP Client", "Failed to send initialize: "+err.Error(), "error")
	} else {
		c.mu.Lock()
		c.initID = initID
		c.initialized = false
		c.mu.Unlock()
		storage.AddLog("connection", "MCP Client", "Sent MCP initialize request", "info")
	}
}

func (c *Client) handleResponse(conn *websocket.Conn, env *jsonrpcEnvelope) {
	c.dispatcher.handleResponse(conn, env)

	if env.Error != nil || env.ID == nil {
		return
	}

	shouldNotify := false

	c.mu.Lock()
	if *env.ID == c.initID && !c.initialized {
		c.initialized = true
		shouldNotify = true
	}
	c.mu.Unlock()

	if shouldNotify {
		c.sendInitialized(conn)
	}
}

func (c *Client) sendInitialized(conn *websocket.Conn) {
	msgBytes, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})

	if err := c.writeRawMessage(conn, msgBytes); err != nil {
		log.Printf("mcp-client: failed to send initialized notification: %v", err)
		storage.AddLog("connection", "MCP Client", "Failed to send MCP initialized notification: "+err.Error(), "error")
		return
	}

	log.Printf("mcp-client: initialized notification sent")
	storage.AddLog("connection", "MCP Client", "Sent MCP initialized notification", "info")
}

func (c *Client) writeJSON(conn *websocket.Conn, v interface{}) error {
	msgBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.writeRawMessage(conn, msgBytes)
}

func (c *Client) writeRawMessage(conn *websocket.Conn, msgBytes []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	conn.SetWriteDeadline(time.Now().Add(mcpWriteTimeout))
	return conn.WriteMessage(websocket.TextMessage, msgBytes)
}
