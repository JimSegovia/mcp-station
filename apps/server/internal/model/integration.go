package model

type Integration struct {
	ID              int      `json:"id"`
	Endpoint        string   `json:"endpoint"`
	Status          string   `json:"status"`
	ToolsJSON       string   `json:"-"`
	Tools           []McpTool `json:"tools"`
	LastConnectedAt *string  `json:"lastConnected"`
	LastError       *string  `json:"lastError"`
	ServerPort      int      `json:"serverPort"`
	ProtocolVersion string   `json:"protocolVersion"`
	UptimeStartedAt *string  `json:"-"`
	Uptime          int64    `json:"uptime"`
	LatencyMs       int      `json:"latency"`
	HealthJSON      string   `json:"-"`
	HealthChecks    []HealthCheck `json:"healthChecks"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

type McpTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type HealthCheck struct {
	Label  string `json:"label"`
	Level  string `json:"level"`
	Detail string `json:"detail"`
}
