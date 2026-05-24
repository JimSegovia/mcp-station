package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mcp-station/server/internal/storage"
)

type Client struct {
	baseURL  string
	username string
	password string
	http     *http.Client
	fastHTTP *http.Client
}

const opencodeRequestTimeout = 10 * time.Minute
const opencodeFastTimeout = 10 * time.Second

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		http:     &http.Client{Timeout: opencodeRequestTimeout},
		fastHTTP: &http.Client{Timeout: opencodeFastTimeout},
	}
}

type HealthResponse struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

func (c *Client) Health() (*HealthResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/global/health", nil)
	if err != nil {
		return nil, fmt.Errorf("health request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health: HTTP %d", resp.StatusCode)
	}

	var h HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, fmt.Errorf("health decode: %w", err)
	}
	return &h, nil
}

type SessionInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type MessagePart struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Reasoning string `json:"reasoning"`
	ToolName  string `json:"tool_name,omitempty"`
	ToolID    string `json:"tool_id,omitempty"`
}

type MessageInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type MessageResponse struct {
	Info  MessageInfo   `json:"info"`
	Parts []MessagePart `json:"parts"`
}

type AskResult struct {
	SessionID string
	MessageID string
	Text      string
}

const XiaoZhiMCPScopeKey = "xiaozhi_mcp"

var XiaoZhiMCPMaxIdle = 15 * time.Minute

func (c *Client) Ask(prompt string, agent string) (*AskResult, error) {
	return c.AskWithModel(prompt, agent, "")
}

func (c *Client) AskWithModel(prompt, agent, model string) (*AskResult, error) {
	sess, err := c.createSession()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return c.sendMessageWithModel(sess.ID, prompt, agent, model)
}

func (c *Client) AskWithScopedSession(scopeKey, prompt, agent, model string, maxIdle time.Duration) (*AskResult, error) {
	now := time.Now().UTC()

	binding, err := storage.GetOpenCodeSessionBinding(scopeKey)
	if err != nil {
		return nil, fmt.Errorf("load scoped session binding: %w", err)
	}

	if storage.IsOpenCodeSessionBindingFresh(binding, maxIdle, now) {
		log.Printf("opencode: reusing scoped session %s for scope=%s", binding.SessionID, scopeKey)
		storage.AddLog("opencode_session", "OpenCode", fmt.Sprintf("Reusing scoped session %s for %s", binding.SessionID, scopeKey), "info")
		result, continueErr := c.ContinueWithModel(binding.SessionID, prompt, agent, model)
		if continueErr == nil {
			if err := storage.TouchOpenCodeSessionBinding(scopeKey, now); err != nil {
				log.Printf("opencode: failed to touch scoped session binding %s: %v", scopeKey, err)
			}
			return result, nil
		}

		if c.isInvalidSessionError(binding.SessionID, continueErr) {
			log.Printf("opencode: scoped session %s became invalid for scope=%s, creating a replacement", binding.SessionID, scopeKey)
			storage.AddLog("opencode_session", "OpenCode", fmt.Sprintf("Scoped session %s became invalid for %s; creating a new one", binding.SessionID, scopeKey), "warning")
			if err := storage.DeleteOpenCodeSessionBinding(scopeKey); err != nil {
				log.Printf("opencode: failed to clear invalid scoped session binding %s: %v", scopeKey, err)
			}
		} else {
			return nil, continueErr
		}
	} else if binding != nil {
		log.Printf("opencode: scoped session %s expired for scope=%s, creating a new one", binding.SessionID, scopeKey)
		storage.AddLog("opencode_session", "OpenCode", fmt.Sprintf("Scoped session %s expired for %s; creating a new one", binding.SessionID, scopeKey), "info")
	}

	result, err := c.AskWithModel(prompt, agent, model)
	if err != nil {
		return nil, err
	}

	if err := storage.UpsertOpenCodeSessionBinding(scopeKey, result.SessionID, now); err != nil {
		log.Printf("opencode: failed to save scoped session binding %s: %v", scopeKey, err)
	}
	log.Printf("opencode: created scoped session %s for scope=%s", result.SessionID, scopeKey)
	storage.AddLog("opencode_session", "OpenCode", fmt.Sprintf("Created scoped session %s for %s", result.SessionID, scopeKey), "success")

	return result, nil
}

func (c *Client) Run(prompt string, agent string) (sessionID string, err error) {
	sess, err := c.createSession()
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	go c.sendMessage(sess.ID, prompt, agent)
	return sess.ID, nil
}

func (c *Client) Continue(sessionID, prompt, agent string) (*AskResult, error) {
	return c.sendMessage(sessionID, prompt, agent)
}

func (c *Client) ContinueWithModel(sessionID, prompt, agent, model string) (*AskResult, error) {
	return c.sendMessageWithModel(sessionID, prompt, agent, model)
}

func (c *Client) createSession() (*SessionInfo, error) {
	body := map[string]string{}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.baseURL+"/session", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("session request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("session create: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("session: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var sess SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, fmt.Errorf("session decode: %w", err)
	}
	return &sess, nil
}

func (c *Client) sendMessage(sessionID, prompt, agent string) (*AskResult, error) {
	return c.sendMessageWithModel(sessionID, prompt, agent, "")
}

func (c *Client) sendMessageWithModel(sessionID, prompt, agent, model string) (*AskResult, error) {
	body := map[string]interface{}{
		"parts": []map[string]interface{}{
			{"type": "text", "text": prompt},
		},
	}
	if agent != "" {
		body["agent"] = agent
	}
	if model != "" {
		body["model"] = model
	}
	b, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/session/%s/message", c.baseURL, sessionID)
	log.Printf("opencode: POST %s (model=%s, agent=%s)", url, model, agent)

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("message request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("message send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("message: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		log.Printf("opencode: %s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	var msgResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("message decode: %w", err)
	}

	log.Printf("opencode: response received (session=%s, role=%s, parts=%d)", sessionID, msgResp.Info.Role, len(msgResp.Parts))

	var result strings.Builder
	for _, part := range msgResp.Parts {
		if part.Type == "text" {
			result.WriteString(part.Text)
			result.WriteString("\n")
		}
	}

	text := result.String()
	return &AskResult{
		SessionID: sessionID,
		MessageID: msgResp.Info.ID,
		Text:      text,
	}, nil
}

type SessionItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Status    string `json:"status"`
}

type OCProvider struct {
	Name   string   `json:"name"`
	Models []string `json:"models"`
}

type OCToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type OCProviderInfo struct {
	ID      string                     `json:"id"`
	Name    string                     `json:"name"`
	Models  map[string]json.RawMessage `json:"models"`
	Default *string                    `json:"default"`
}

type OCModelsResponse struct {
	Providers []OCProviderInfo  `json:"providers"`
	Default   map[string]string `json:"default"`
}

func (c *Client) ListModels() (*OCModelsResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/config/providers", nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list models: HTTP %d", resp.StatusCode)
	}

	var result OCModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("list models decode: %w", err)
	}
	return &result, nil
}

func (c *Client) ListToolIDs() ([]string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/experimental/tool/ids", nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ids struct {
		BuiltIn  []string `json:"builtIn"`
		External []string `json:"external"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, fmt.Errorf("tool ids decode: %w", err)
	}

	all := append(ids.BuiltIn, ids.External...)
	return all, nil
}

func (c *Client) ExecuteTool(toolName string, args map[string]interface{}) (*AskResult, error) {
	argsJSON, _ := json.Marshal(args)
	prompt := fmt.Sprintf("Execute the tool '%s' with these arguments: %s. Return ONLY the result, no explanations.", toolName, string(argsJSON))
	return c.Ask(prompt, "build")
}

func (c *Client) ExecuteToolWithScopedSession(scopeKey, toolName string, args map[string]interface{}, maxIdle time.Duration) (*AskResult, error) {
	argsJSON, _ := json.Marshal(args)
	prompt := fmt.Sprintf("Execute the tool '%s' with these arguments: %s. Return ONLY the result, no explanations.", toolName, string(argsJSON))
	return c.AskWithScopedSession(scopeKey, prompt, "build", "", maxIdle)
}

type MessageItem struct {
	Info  MessageInfo   `json:"info"`
	Parts []MessagePart `json:"parts"`
}

func (c *Client) ListSessions() ([]SessionItem, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/session", nil)
	if err != nil {
		return nil, fmt.Errorf("sessions list request: %w", err)
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sessions list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sessions list: HTTP %d", resp.StatusCode)
	}

	var sessions []SessionItem
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("sessions list decode: %w", err)
	}
	return sessions, nil
}

func (c *Client) GetSession(sessionID string) (*SessionItem, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/session/"+sessionID, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get session: HTTP %d", resp.StatusCode)
	}

	var sess SessionItem
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, fmt.Errorf("get session decode: %w", err)
	}
	return &sess, nil
}

func (c *Client) GetSessionMessages(sessionID string) ([]MessageItem, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/session/"+sessionID+"/message", nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get messages: HTTP %d", resp.StatusCode)
	}

	var messages []MessageItem
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("get messages decode: %w", err)
	}
	return messages, nil
}

func (c *Client) DeleteSession(sessionID string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/session/"+sessionID, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)

	resp, err := c.fastHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete session: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.password != "" {
		user := c.username
		if user == "" {
			user = "opencode"
		}
		req.SetBasicAuth(user, c.password)
	}
}

func (c *Client) isInvalidSessionError(sessionID string, originalErr error) bool {
	if originalErr == nil {
		return false
	}

	if isOpenCodeSessionNotFoundError(originalErr) {
		return true
	}

	_, err := c.GetSession(sessionID)
	return isOpenCodeSessionNotFoundError(err)
}

func isOpenCodeSessionNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "http 404") || strings.Contains(msg, "not found")
}
