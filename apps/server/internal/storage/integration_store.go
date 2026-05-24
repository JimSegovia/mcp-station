package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mcp-station/server/internal/model"
)

func GetIntegration() (*model.Integration, error) {
	row := DB.QueryRow(`SELECT id, endpoint, status, tools_json, last_connected_at, last_error, server_port, protocol_version, uptime_started_at, latency_ms, health_json, created_at, updated_at FROM integrations WHERE id = 1`)

	var i model.Integration
	var tp string
	err := row.Scan(&i.ID, &i.Endpoint, &i.Status, &tp, &i.LastConnectedAt, &i.LastError, &i.ServerPort, &i.ProtocolVersion, &i.UptimeStartedAt, &i.LatencyMs, &i.HealthJSON, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(tp), &i.Tools)
	json.Unmarshal([]byte(i.HealthJSON), &i.HealthChecks)

	if i.Status == "connected" && i.UptimeStartedAt != nil {
		if t, e := time.Parse(time.RFC3339, *i.UptimeStartedAt); e == nil {
			i.Uptime = int64(time.Since(t).Seconds())
		}
	}

	return &i, nil
}

func ConnectIntegration(endpoint string) (*model.Integration, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	health := buildHealth("connecting", "")
	healthJSON, _ := json.Marshal(health)

	DB.Exec(`UPDATE integrations SET endpoint = ?, status = 'connecting', tools_json = '[]', last_error = NULL, uptime_started_at = NULL, latency_ms = 0, health_json = ?, updated_at = ? WHERE id = 1`,
		endpoint, string(healthJSON), now)
	NotifyIntegrationChanged()

	return GetIntegration()
}

func SetIntegrationConnecting() error {
	health := buildHealth("connecting", "")
	healthJSON, _ := json.Marshal(health)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`UPDATE integrations SET status = 'connecting', last_error = NULL, health_json = ?, updated_at = ? WHERE id = 1`,
		string(healthJSON), now)
	if err == nil {
		NotifyIntegrationChanged()
	}
	return err
}

func SetIntegrationConnected() error {
	now := time.Now().UTC().Format(time.RFC3339)
	health := buildHealth("connected", "")
	healthJSON, _ := json.Marshal(health)
	_, err := DB.Exec(`UPDATE integrations SET status = 'connected', last_connected_at = ?, last_error = NULL, uptime_started_at = ?, latency_ms = 0, health_json = ?, updated_at = ? WHERE id = 1`,
		now, now, string(healthJSON), now)
	if err == nil {
		NotifyIntegrationChanged()
	}
	return err
}

func SetIntegrationDisconnected(detail string) error {
	health := buildHealth("disconnected", detail)
	healthJSON, _ := json.Marshal(health)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`UPDATE integrations SET status = 'disconnected', uptime_started_at = NULL, latency_ms = 0, health_json = ?, updated_at = ? WHERE id = 1`,
		string(healthJSON), now)
	if err == nil {
		NotifyIntegrationChanged()
	}
	return err
}

func DisconnectIntegration() (*model.Integration, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	DB.Exec(`UPDATE integrations SET status = 'empty', tools_json = '[]', last_connected_at = NULL, last_error = NULL, uptime_started_at = NULL, latency_ms = 0, health_json = '[]', updated_at = ? WHERE id = 1`, now)
	NotifyIntegrationChanged()
	return GetIntegration()
}

func SetIntegrationError(message string) (*model.Integration, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	health := buildHealth("error", message)
	healthJSON, _ := json.Marshal(health)

	DB.Exec(`UPDATE integrations SET status = 'error', last_error = ?, uptime_started_at = NULL, latency_ms = 0, health_json = ?, updated_at = ? WHERE id = 1`,
		message, string(healthJSON), now)
	NotifyIntegrationChanged()

	return GetIntegration()
}

func UpdateIntegrationStatus(status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`UPDATE integrations SET status = ?, updated_at = ? WHERE id = 1`, status, now)
	if err == nil {
		NotifyIntegrationChanged()
	}
	return err
}

func UpdateIntegrationHealth(status string, detail string) error {
	health := buildHealth(status, detail)
	healthJSON, _ := json.Marshal(health)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`UPDATE integrations SET health_json = ?, updated_at = ? WHERE id = 1`, string(healthJSON), now)
	if err == nil {
		NotifyIntegrationChanged()
	}
	return err
}

func SyncIntegrationTools(tools []model.McpTool) error {
	toolsJSON, _ := json.Marshal(tools)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`UPDATE integrations SET tools_json = ?, updated_at = ? WHERE id = 1`, string(toolsJSON), now)
	if err == nil {
		NotifyIntegrationChanged()
	}
	return err
}

func ToggleIntegrationTool(name string, enabled bool) (*model.Integration, error) {
	i, err := GetIntegration()
	if err != nil {
		return nil, err
	}

	for idx, t := range i.Tools {
		if t.Name == name {
			i.Tools[idx].Enabled = enabled
		}
	}

	toolsJSON, _ := json.Marshal(i.Tools)
	now := time.Now().UTC().Format(time.RFC3339)
	DB.Exec(`UPDATE integrations SET tools_json = ?, updated_at = ? WHERE id = 1`, string(toolsJSON), now)
	NotifyIntegrationChanged()

	return GetIntegration()
}

func buildHealth(status, detail string) []model.HealthCheck {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	endpointLevel := "unknown"
	endpointDetail := "Not checked"
	runtimeLevel := "healthy"
	runtimeDetail := fmt.Sprintf("host=%s pid=%d", hostname, pid)

	switch status {
	case "connected":
		endpointLevel = "healthy"
		endpointDetail = "WebSocket connected"
	case "connecting":
		endpointLevel = "pending"
		endpointDetail = "Connecting..."
	case "disconnected":
		endpointLevel = "degraded"
		endpointDetail = "Disconnected; retrying"
		if detail != "" {
			endpointDetail = detail
		}
	case "error":
		endpointLevel = "unhealthy"
		runtimeLevel = "degraded"
		if detail != "" {
			endpointDetail = detail
		}
	}

	return []model.HealthCheck{
		{Label: "Endpoint", Level: endpointLevel, Detail: endpointDetail},
		{Label: "Runtime", Level: runtimeLevel, Detail: runtimeDetail},
		{Label: "Protocol", Level: "info", Detail: "MCP JSON-RPC 2.0 / WebSocket"},
	}
}
