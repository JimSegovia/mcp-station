package model

type Server struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Type            string      `json:"type"`
	Endpoint        string      `json:"endpoint"`
	Enabled         bool        `json:"enabled"`
	Status          string      `json:"status"`
	ToolsJSON       string      `json:"-"`
	Tools           []McpTool   `json:"tools"`
	LastConnectedAt *string     `json:"lastConnected"`
	CreatedAt       string      `json:"createdAt"`
	UpdatedAt       string      `json:"updatedAt"`
}
