package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/mcp-station/server/internal/model"
)

func GetServers() ([]model.Server, error) {
	rows, err := DB.Query(`SELECT id, name, type, endpoint, enabled, status, tools_json, last_connected_at, created_at, updated_at FROM servers ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var s model.Server
		var enabled int
		var tp string
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &s.Endpoint, &enabled, &s.Status, &tp, &s.LastConnectedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Enabled = enabled == 1
		json.Unmarshal([]byte(tp), &s.Tools)
		servers = append(servers, s)
	}

	return servers, nil
}

func CreateServer(name, stype, endpoint string) (*model.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	id := "mcp-" + uuid.NewString()[:8]

	DB.Exec(`INSERT INTO servers (id, name, type, endpoint, enabled, status, tools_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, stype, endpoint, true, "disconnected", "[]", now, now)

	return GetServerByID(id)
}

func GetServerByID(id string) (*model.Server, error) {
	var s model.Server
	var enabled int
	var tp string
	row := DB.QueryRow(`SELECT id, name, type, endpoint, enabled, status, tools_json, last_connected_at, created_at, updated_at FROM servers WHERE id = ?`, id)
	if err := row.Scan(&s.ID, &s.Name, &s.Type, &s.Endpoint, &enabled, &s.Status, &tp, &s.LastConnectedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	s.Enabled = enabled == 1
	json.Unmarshal([]byte(tp), &s.Tools)
	return &s, nil
}

func DeleteServer(id string) error {
	_, err := DB.Exec(`DELETE FROM servers WHERE id = ?`, id)
	return err
}

func ToggleServer(id string, enabled bool) (*model.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	DB.Exec(`UPDATE servers SET enabled = ?, updated_at = ? WHERE id = ?`,
		enabled, now, id)

	return GetServerByID(id)
}

func ToggleServerTool(serverID, toolName string, enabled bool) (*model.Server, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	var tp string
	DB.QueryRow(`SELECT tools_json FROM servers WHERE id = ?`, serverID).Scan(&tp)

	var tools []model.McpTool
	json.Unmarshal([]byte(tp), &tools)

	for i, t := range tools {
		if t.Name == toolName {
			tools[i].Enabled = enabled
		}
	}

	tb, _ := json.Marshal(tools)
	DB.Exec(`UPDATE servers SET tools_json = ?, updated_at = ? WHERE id = ?`, string(tb), now, serverID)

	return GetServerByID(serverID)
}
