package api

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/mcp-station/server/internal/model"
)

var startTime = time.Now()
var pid = os.Getpid()

func Uptime() int64 {
	return int64(time.Since(startTime).Seconds())
}

func (h *Handler) SystemStats(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	stats := model.SystemStats{
		PID:        pid,
		Uptime:     int64(time.Since(startTime).Seconds()),
		Goroutines: runtime.NumGoroutine(),
		NumCPU:     runtime.NumCPU(),
		GoVersion:  runtime.Version(),
		Memory: model.MemoryStats{
			Alloc:      mem.Alloc,
			TotalAlloc: mem.TotalAlloc,
			Sys:        mem.Sys,
			HeapAlloc:  mem.HeapAlloc,
			HeapSys:    mem.HeapSys,
			NumGC:      mem.NumGC,
		},
	}

	writeJSON(w, http.StatusOK, stats)
}
