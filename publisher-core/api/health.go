package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    int64             `json:"uptime_seconds"`
	Services  map[string]ServiceHealth `json:"services"`
	System    SystemHealth      `json:"system"`
}

type ServiceHealth struct {
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Latency   int64     `json:"latency_ms,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

type SystemHealth struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
	MemSysMB     uint64 `json:"mem_sys_mb"`
}

type HealthChecker struct {
	startTime time.Time
	version   string
	checks    map[string]HealthCheck
}

type HealthCheck func() ServiceHealth

func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		version:   version,
		checks:    make(map[string]HealthCheck),
	}
}

func (h *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	h.checks[name] = check
	logrus.Infof("Health check registered: %s", name)
}

func (h *HealthChecker) Check() HealthStatus {
	overallStatus := "healthy"
	services := make(map[string]ServiceHealth)

	for name, check := range h.checks {
		serviceHealth := check()
		services[name] = serviceHealth

		if serviceHealth.Status == "unhealthy" {
			overallStatus = "unhealthy"
		} else if serviceHealth.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemHealth := SystemHealth{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAllocMB:   m.Alloc / 1024 / 1024,
		MemTotalMB:   m.TotalAlloc / 1024 / 1024,
		MemSysMB:     m.Sys / 1024 / 1024,
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    int64(time.Since(h.startTime).Seconds()),
		Services:  services,
		System:    systemHealth,
	}
}

func (s *Server) healthCheckHandler(checker *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := checker.Check()

		statusCode := http.StatusOK
		if status.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(status)
	}
}

func (s *Server) livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) readinessProbe(checker *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := checker.Check()

		if status.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}
