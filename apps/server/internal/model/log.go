package model

type LogEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Message   string `json:"message"`
	Result    string `json:"result"`
}

type LogQuery struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Result string `json:"result"`
	Limit  int    `json:"limit"`
}
