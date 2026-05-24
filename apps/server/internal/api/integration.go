package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mcp-station/server/internal/model"
	"github.com/mcp-station/server/internal/storage"
)

func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	i, err := h.loadIntegrationSnapshot()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load integration"})
		return
	}

	log.Printf("api: GET /integration status=%s tools=%d", i.Status, len(i.Tools))
	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) StreamIntegration(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	subID, changes := storage.SubscribeIntegrationChanges()
	defer storage.UnsubscribeIntegrationChanges(subID)

	if err := h.writeIntegrationEvent(w, flusher); err != nil {
		return
	}

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-changes:
			if err := h.writeIntegrationEvent(w, flusher); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

type connectRequest struct {
	Endpoint string `json:"endpoint"`
}

func (h *Handler) ConnectIntegration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req connectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "endpoint is required"})
		return
	}

	i, err := storage.ConnectIntegration(req.Endpoint)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to connect"})
		return
	}

	h.Client.Connect(req.Endpoint)
	h.syncTools()

	log.Printf("api: POST /connect status=%s", i.Status)
	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) DisconnectIntegration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.Client.Disconnect()

	i, err := storage.DisconnectIntegration()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to disconnect"})
		return
	}

	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) SetIntegrationError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}

	i, err := storage.SetIntegrationError(req.Message)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to set error"})
		return
	}

	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) GetIntegrationHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	i, err := storage.GetIntegration()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load health"})
		return
	}

	writeJSON(w, http.StatusOK, i.HealthChecks)
}

func (h *Handler) GetIntegrationTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	i, err := storage.GetIntegration()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load tools"})
		return
	}

	i = enrichWithRegistry(i, h)
	writeJSON(w, http.StatusOK, i.Tools)
}

type toggleRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *Handler) ToggleIntegrationTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.PathValue("name")
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if err := h.Registry.SetEnabled(name, req.Enabled); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	i, err := storage.ToggleIntegrationTool(name, req.Enabled)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to toggle tool"})
		return
	}

	h.syncTools()
	i = enrichWithRegistry(i, h)
	writeJSON(w, http.StatusOK, i)
}

func enrichWithRegistry(i *model.Integration, h *Handler) *model.Integration {
	registryTools := h.Registry.List()
	i.Tools = make([]model.McpTool, 0, len(registryTools))
	for _, t := range registryTools {
		i.Tools = append(i.Tools, model.McpTool{
			Name:        t.Name,
			Description: t.Description,
			Enabled:     t.Enabled,
		})
	}
	return i
}

func (h *Handler) loadIntegrationSnapshot() (*model.Integration, error) {
	i, err := storage.GetIntegration()
	if err != nil {
		return nil, err
	}
	return enrichWithRegistry(i, h), nil
}

func (h *Handler) writeIntegrationEvent(w http.ResponseWriter, flusher http.Flusher) error {
	i, err := h.loadIntegrationSnapshot()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(i)
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte("event: integration\n")); err != nil {
		return err
	}
	if _, err := w.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}
