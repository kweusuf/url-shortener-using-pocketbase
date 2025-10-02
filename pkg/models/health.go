package models

import (
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

// ComponentHealth represents the health of an individual component
type ComponentHealth struct {
	Name         string                 `json:"name"`
	Status       constants.HealthStatus `json:"status"`
	Message      string                 `json:"message,omitempty"`
	ResponseTime time.Duration          `json:"response_time_ms"`
	LastCheck    time.Time              `json:"last_check"`
}

// SystemHealth represents the complete system health
type SystemHealth struct {
	Status      constants.HealthStatus `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Uptime      string                 `json:"uptime"`
	Environment string                 `json:"environment"`
	Components  []ComponentHealth      `json:"components"`
	SystemInfo  SystemInfo             `json:"system_info"`
}

// SystemInfo contains system-level information
type SystemInfo struct {
	GoVersion    string      `json:"go_version"`
	Goroutines   int         `json:"goroutines"`
	MemoryUsage  MemoryStats `json:"memory_usage"`
	DatabaseInfo DBInfo      `json:"database_info"`
}

// MemoryStats contains memory usage information
type MemoryStats struct {
	AllocatedBytes      uint64 `json:"allocated_bytes"`
	TotalAllocatedBytes uint64 `json:"total_allocated_bytes"`
	SystemMemoryBytes   uint64 `json:"system_memory_bytes"`
	GCRuns              uint32 `json:"gc_runs"`
}

// DBInfo contains database information
type DBInfo struct {
	Type             string        `json:"type"`
	ConnectionStatus string        `json:"connection_status"`
	ResponseTime     time.Duration `json:"response_time_ms"`
	DatabaseName     string        `json:"database_name,omitempty"`
}

// HealthErrorResponse represents a health check error response
type HealthErrorResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthReadyResponse represents a readiness probe response
type HealthReadyResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthLiveResponse represents a liveness probe response
type HealthLiveResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthMetricsErrorResponse represents a metrics endpoint error response
type HealthMetricsErrorResponse struct {
	Error string `json:"error"`
}

// HealthMetricsResponse represents a metrics endpoint response
type HealthMetricsResponse struct {
	HealthStatus           string    `json:"health_status"`
	Uptime                 string    `json:"uptime"`
	Goroutines             int       `json:"goroutines"`
	MemoryAllocatedMB      float64   `json:"memory_allocated_mb"`
	MemoryTotalAllocatedMB float64   `json:"memory_total_allocated_mb"`
	MemorySystemMB         float64   `json:"memory_system_mb"`
	GCRuns                 uint32    `json:"gc_runs"`
	DatabaseResponseTimeMS float64   `json:"database_response_time_ms"`
	DatabaseStatus         string    `json:"database_status"`
	Timestamp              time.Time `json:"timestamp"`
}
