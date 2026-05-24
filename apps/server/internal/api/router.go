package api

import (
	"net/http"

	"github.com/mcp-station/server/internal/mcp"
	"github.com/mcp-station/server/internal/opencode"
	"github.com/mcp-station/server/internal/tool"
)

type Handler struct {
	Runtime   *mcp.Runtime
	Registry  *tool.Registry
	Client    *mcp.Client
	OCManager *opencode.Manager
}

func NewRouter(runtime *mcp.Runtime, registry *tool.Registry, client *mcp.Client, ocManager *opencode.Manager) http.Handler {
	h := &Handler{Runtime: runtime, Registry: registry, Client: client, OCManager: ocManager}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/dashboard", h.Dashboard)

	mux.HandleFunc("GET /api/integration", h.GetIntegration)
	mux.HandleFunc("GET /api/integration/stream", h.StreamIntegration)
	mux.HandleFunc("POST /api/integration/connect", h.ConnectIntegration)
	mux.HandleFunc("POST /api/integration/disconnect", h.DisconnectIntegration)
	mux.HandleFunc("POST /api/integration/error", h.SetIntegrationError)
	mux.HandleFunc("GET /api/integration/health", h.GetIntegrationHealth)
	mux.HandleFunc("GET /api/integration/tools", h.GetIntegrationTools)
	mux.HandleFunc("PUT /api/integration/tools/{name}", h.ToggleIntegrationTool)

	mux.HandleFunc("GET /api/servers", h.GetServers)
	mux.HandleFunc("POST /api/servers", h.CreateServer)
	mux.HandleFunc("DELETE /api/servers/{id}", h.DeleteServer)
	mux.HandleFunc("PUT /api/servers/{id}", h.ToggleServer)
	mux.HandleFunc("PUT /api/servers/{id}/tools/{name}", h.ToggleServerTool)
	mux.HandleFunc("POST /api/servers/{id}/discover", h.DiscoverServerTools)

	mux.HandleFunc("GET /api/logs", h.GetLogs)
	mux.HandleFunc("DELETE /api/logs", h.DeleteLogs)

	mux.HandleFunc("GET /api/system", h.SystemStats)

	mux.HandleFunc("GET /api/opencode/sessions", h.GetOpenCodeSessions)
	mux.HandleFunc("POST /api/opencode/sessions", h.CreateOpenCodeSession)
	mux.HandleFunc("GET /api/opencode/sessions/{id}/messages", h.GetOpenCodeSessionMessages)
	mux.HandleFunc("DELETE /api/opencode/sessions/{id}", h.DeleteOpenCodeSession)
	mux.HandleFunc("DELETE /api/opencode/sessions", h.CleanExpiredSessions)
	mux.HandleFunc("GET /api/opencode/log", h.GetOpenCodeLogs)
	mux.HandleFunc("GET /api/opencode/models", h.GetOpenCodeModels)
	mux.HandleFunc("POST /api/opencode/command", h.SendOpenCodeCommand)

	mux.HandleFunc("GET /api/tools", h.GetTools)
	mux.HandleFunc("PUT /api/tools/{name}", h.ToggleTool)

	mux.HandleFunc("GET /ws", h.Runtime.ServeWS)

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
