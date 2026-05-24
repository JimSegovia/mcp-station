package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mcp-station/server/internal/model"
	"github.com/mcp-station/server/internal/storage"
)

func (h *Handler) GetServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	servers, err := storage.GetServers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load servers"})
		return
	}

	virtual := h.buildVirtualServers()

	all := append(virtual, servers...)
	writeJSON(w, http.StatusOK, all)
}

func (h *Handler) buildVirtualServers() []model.Server {
	virtual := make([]model.Server, 0, 2)

	ocRunning := h.OCManager != nil && h.OCManager.IsRunning()
	ocStatus := "disconnected"
	if ocRunning {
		ocStatus = "connected"
	}
	ocTools := h.Registry.ToolsByOrigin("opencode")
	ocServer := model.Server{
		ID:       "opencode",
		Name:     "OpenCode",
		Type:     "virtual",
		Endpoint: "opencode serve",
		Enabled:  ocRunning,
		Status:   ocStatus,
		Tools:    ocTools,
	}
	virtual = append(virtual, ocServer)

	stationTools := h.Registry.ToolsByOrigin("mcp-station")
	stationServer := model.Server{
		ID:       "mcp-station",
		Name:     "MCP Station",
		Type:     "virtual",
		Endpoint: "localhost (built-in)",
		Enabled:  true,
		Status:   "connected",
		Tools:    stationTools,
	}
	virtual = append(virtual, stationServer)

	return virtual
}

type createServerRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
}

func (h *Handler) CreateServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req createServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	stype := req.Type
	if stype == "" {
		stype = "websocket"
	}

	s, err := storage.CreateServer(req.Name, stype, req.Endpoint)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create server"})
		return
	}

	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")

	if id == "opencode" || id == "mcp-station" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot delete built-in server"})
		return
	}

	s, err := storage.GetServerByID(id)
	if err == nil {
		h.Registry.RemoveToolsByOrigin(s.Name)
	}

	if err := storage.DeleteServer(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (h *Handler) ToggleServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if id == "opencode" {
		if req.Enabled {
			if !h.OCManager.IsRunning() {
				h.OCManager.Start(r.Context())
			}
		}
		h.Registry.SetOriginEnabled("opencode", req.Enabled)
		h.syncTools()

		status := "disconnected"
		if req.Enabled && h.OCManager.IsRunning() {
			status = "connected"
		}
		writeJSON(w, http.StatusOK, model.Server{
			ID:       "opencode",
			Name:     "OpenCode",
			Type:     "virtual",
			Endpoint: "opencode serve",
			Enabled:  req.Enabled,
			Status:   status,
			Tools:    h.Registry.ToolsByOrigin("opencode"),
		})
		return
	}

	if id == "mcp-station" {
		h.Registry.SetOriginEnabled("mcp-station", req.Enabled)
		h.syncTools()
		writeJSON(w, http.StatusOK, model.Server{
			ID:       "mcp-station",
			Name:     "MCP Station",
			Type:     "virtual",
			Endpoint: "localhost (built-in)",
			Enabled:  req.Enabled,
			Status:   "connected",
			Tools:    h.Registry.ToolsByOrigin("mcp-station"),
		})
		return
	}

	s, err := storage.ToggleServer(id, req.Enabled)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}

	h.Registry.SetOriginEnabled(s.Name, req.Enabled)
	h.syncTools()

	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) ToggleServerTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	name := r.PathValue("name")
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if id == "opencode" || id == "mcp-station" {
		h.Registry.SetEnabled(name, req.Enabled)
		h.syncTools()
		writeJSON(w, http.StatusOK, map[string]interface{}{"name": name, "enabled": req.Enabled})
		return
	}

	s, err := storage.ToggleServerTool(id, name, req.Enabled)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server or tool not found"})
		return
	}
	h.syncTools()
	writeJSON(w, http.StatusOK, s)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) syncTools() {
	tools := h.Registry.ListEnabled()
	mcpTools := make([]model.McpTool, 0, len(tools))
	for _, t := range tools {
		mcpTools = append(mcpTools, model.McpTool{
			Name:        t.Name,
			Description: t.Description,
			Enabled:     t.Enabled,
		})
	}
	storage.SyncIntegrationTools(mcpTools)
}

func (h *Handler) DiscoverServerTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")

	s, err := storage.GetServerByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
		return
	}

	if s.Type != "websocket" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only websocket servers support auto-discovery"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	tools, err := h.Client.DiscoverTools(ctx, s.Endpoint)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "discovery failed: " + err.Error()})
		return
	}

	h.Registry.RemoveToolsByOrigin(s.Name)
	h.Registry.RegisterExternalTools(s.Name, tools)

	var mcpTools []model.McpTool
	for _, t := range tools {
		mcpTools = append(mcpTools, model.McpTool{
			Name:        s.Name + "_" + t.Name,
			Description: t.Description,
			Enabled:     true,
		})
	}
	toolsJSON, _ := json.Marshal(mcpTools)
	now := time.Now().UTC().Format(time.RFC3339)
	storage.DB.Exec(`UPDATE servers SET tools_json = ?, updated_at = ? WHERE id = ?`, string(toolsJSON), now, id)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"server":     s.Name,
		"discovered": len(tools),
		"tools":      tools,
	})
}
