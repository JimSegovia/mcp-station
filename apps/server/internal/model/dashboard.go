package model

type Dashboard struct {
	Integration DashboardIntegration `json:"integration"`
	Modules     DashboardModules     `json:"modules"`
	Servers     DashboardServers     `json:"servers"`
	Logs        DashboardLogs        `json:"logs"`
}

type DashboardIntegration struct {
	Status      string `json:"status"`
	Uptime      int64  `json:"uptime"`
	Latency     int    `json:"latency"`
	ToolsActive int    `json:"toolsActive"`
	ToolsTotal  int    `json:"toolsTotal"`
}

type DashboardModules struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

type DashboardServers struct {
	Connected int `json:"connected"`
	Registered int `json:"registered"`
}

type DashboardLogs struct {
	Total  int `json:"total"`
	Errors int `json:"errors"`
}
