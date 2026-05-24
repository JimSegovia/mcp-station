package api

import (
	"net/http"

	"github.com/mcp-station/server/internal/model"
	"github.com/mcp-station/server/internal/storage"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	integration, err := storage.GetIntegration()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load integration"})
		return
	}

	servers, err := storage.GetServers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load servers"})
		return
	}

	logs, err := storage.GetLogs(model.LogQuery{Limit: 50})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load logs"})
		return
	}

	connectedServers := 0
	for _, s := range servers {
		if s.Status == "connected" {
			connectedServers++
		}
	}

	virtualCount := 0
	if h.OCManager != nil && h.OCManager.IsRunning() {
		virtualCount++
	}
	virtualCount++

	toolsActive := 0
	for _, t := range integration.Tools {
		if t.Enabled {
			toolsActive++
		}
	}

	logErrors := 0
	for _, l := range logs {
		if l.Result == "error" || l.Result == "blocked" {
			logErrors++
		}
	}

	registryTools := len(h.Registry.ListEnabled())

	d := model.Dashboard{
		Integration: model.DashboardIntegration{
			Status:      integration.Status,
			Uptime:      integration.Uptime,
			Latency:     integration.LatencyMs,
			ToolsActive: toolsActive,
			ToolsTotal:  len(integration.Tools),
		},
		Modules: model.DashboardModules{
			Active: virtualCount + connectedServers,
			Total:  virtualCount + len(servers),
		},
		Servers: model.DashboardServers{
			Connected:  connectedServers,
			Registered: len(servers) + virtualCount,
		},
		Logs: model.DashboardLogs{
			Total:  len(logs),
			Errors: logErrors,
		},
	}

	_ = registryTools

	writeJSON(w, http.StatusOK, d)
}
