package api

import (
	"net/http"
)

func (h *Handler) GetTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groups := h.Registry.ListByOrigin()
	writeJSON(w, http.StatusOK, groups)
}

func (h *Handler) ToggleTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := r.PathValue("name")

	enabled, err := h.Registry.Toggle(name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":    name,
		"enabled": enabled,
	})
}
