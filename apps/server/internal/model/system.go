package model

type SystemStats struct {
	PID        int         `json:"pid"`
	Uptime     int64       `json:"uptime"`
	Goroutines int         `json:"goroutines"`
	NumCPU     int         `json:"numCPU"`
	GoVersion  string      `json:"goVersion"`
	Memory     MemoryStats `json:"memory"`
}

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"totalAlloc"`
	Sys        uint64 `json:"sys"`
	HeapAlloc  uint64 `json:"heapAlloc"`
	HeapSys    uint64 `json:"heapSys"`
	NumGC      uint32 `json:"numGC"`
}
