package api

import (
	"net/http"
	"strconv"

	"github.com/mcp-station/server/internal/model"
	"github.com/mcp-station/server/internal/storage"
)

func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 50
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		limit = l
	}

	query := model.LogQuery{
		Type:   q.Get("type"),
		Source: q.Get("source"),
		Result: q.Get("result"),
		Limit:  limit,
	}

	logs, err := storage.GetLogs(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load logs"})
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) DeleteLogs(w http.ResponseWriter, r *http.Request) {
	if err := storage.DeleteLogs(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to clear logs"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"cleared": "ok"})
}
