package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mcp-station/server/internal/opencode"
	"github.com/mcp-station/server/internal/storage"
)

type enrichedSession struct {
	SessionID string `json:"sessionId"`
	TrackID   string `json:"trackId"`
	ToolName  string `json:"toolName"`
	Prompt    string `json:"prompt"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	Age       string `json:"age"`
}

func (h *Handler) GetOpenCodeSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tracked, err := storage.ListTrackedSessions()
	if err != nil {
		log.Printf("opencode api: session error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ocSessions, err := h.OCManager.Client().ListSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to query opencode sessions"})
		return
	}

	ocMap := make(map[string]struct {
		title  string
		status string
	})
	for _, s := range ocSessions {
		ocMap[s.ID] = struct {
			title  string
			status string
		}{s.Title, s.Status}
	}

	result := make([]enrichedSession, 0, len(tracked))
	seen := make(map[string]struct{}, len(tracked))
	for _, t := range tracked {
		if _, ok := seen[t.SessionID]; ok {
			continue
		}
		seen[t.SessionID] = struct{}{}

		meta := ocMap[t.SessionID]
		created, _ := time.Parse(time.RFC3339, t.CreatedAt)
		age := time.Since(created).Truncate(time.Second).String()

		result = append(result, enrichedSession{
			SessionID: t.SessionID,
			TrackID:   t.ID,
			ToolName:  t.ToolName,
			Prompt:    t.Prompt,
			Title:     meta.title,
			Status:    meta.status,
			CreatedAt: t.CreatedAt,
			Age:       age,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DeleteOpenCodeSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionID := r.PathValue("id")

	var req struct {
		TrackID string `json:"trackId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if req.TrackID != "" {
		if err := storage.DeleteTrackedSession(req.TrackID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete tracking record"})
			return
		}
	}

	if err := storage.DeleteTrackedSessionsBySessionID(sessionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete session history"})
		return
	}

	if err := h.OCManager.Client().DeleteSession(sessionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete opencode session: " + err.Error()})
		return
	}
	if err := storage.DeleteOpenCodeSessionBindingsBySessionID(sessionID); err != nil {
		log.Printf("opencode api: failed to clear session binding for %s: %v", sessionID, err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"ok": "deleted"})
}

func (h *Handler) CleanExpiredSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ids, err := storage.CleanExpiredSessions(24 * time.Hour)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cleanup failed"})
		return
	}

	for _, sid := range ids {
		h.OCManager.Client().DeleteSession(sid)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cleaned": len(ids),
		"ids":     ids,
	})
}

func (h *Handler) GetOpenCodeLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	logs := h.OCManager.Logs()
	if logs == nil {
		logs = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"ready": h.OCManager.IsRunning(),
	})
}

func (h *Handler) SendOpenCodeCommand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Command == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "command is required"})
		return
	}

	result, err := h.OCManager.SendCommand(req.Command)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": result})
}

func (h *Handler) GetOpenCodeModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !h.OCManager.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "opencode not running"})
		return
	}

	models, err := h.OCManager.Client().ListModels()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to list models: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models)
}

func (h *Handler) CreateOpenCodeSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Prompt    string `json:"prompt"`
		Agent     string `json:"agent"`
		SessionID string `json:"sessionId"`
		Model     string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	if !h.OCManager.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "opencode is not running"})
		return
	}

	var result *opencode.AskResult
	var err error

	if req.SessionID != "" {
		result, err = h.OCManager.Client().ContinueWithModel(req.SessionID, req.Prompt, req.Agent, req.Model)
	} else {
		result, err = h.OCManager.Client().AskWithModel(req.Prompt, req.Agent, req.Model)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	storage.TrackOpenCodeSession(result.SessionID, "ui_prompt", req.Prompt)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sessionId": result.SessionID,
		"messageId": result.MessageID,
		"text":      result.Text,
	})
}

func (h *Handler) GetOpenCodeSessionMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionID := r.PathValue("id")

	if !h.OCManager.IsRunning() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "opencode is not running"})
		return
	}

	messages, err := h.OCManager.Client().GetSessionMessages(sessionID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, messages)
}
