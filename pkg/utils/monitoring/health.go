package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// HealthStatus represents the overall health of the system
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of an individual component
type ComponentHealth struct {
	Name         string        `json:"name"`
	Status       HealthStatus  `json:"status"`
	Message      string        `json:"message,omitempty"`
	ResponseTime time.Duration `json:"response_time_ms"`
	LastCheck    time.Time     `json:"last_check"`
}

// SystemHealth represents the complete system health
type SystemHealth struct {
	Status      HealthStatus      `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	Environment string            `json:"environment"`
	Components  []ComponentHealth `json:"components"`
	SystemInfo  SystemInfo        `json:"system_info"`
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

// HealthChecker handles all health check operations
type HealthChecker struct {
	app *pocketbase.PocketBase
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(app *pocketbase.PocketBase) *HealthChecker {
	return &HealthChecker{app: app}
}

// CheckSystemHealth performs a comprehensive health check
func (hc *HealthChecker) CheckSystemHealth(ctx context.Context) (*SystemHealth, error) {
	start := time.Now()

	components := []ComponentHealth{}
	overallStatus := StatusHealthy

	// Check application component
	appHealth := hc.checkApplicationHealth(ctx)
	components = append(components, appHealth)
	if appHealth.Status != StatusHealthy {
		overallStatus = StatusUnhealthy
	}

	// Check database component
	dbHealth := hc.checkDatabaseHealth(ctx)
	components = append(components, dbHealth)
	if dbHealth.Status != StatusHealthy && overallStatus == StatusHealthy {
		overallStatus = StatusDegraded
	}

	// Check external dependencies
	depsHealth := hc.checkExternalDependencies(ctx)
	components = append(components, depsHealth)
	if depsHealth.Status != StatusHealthy && overallStatus == StatusHealthy {
		overallStatus = StatusDegraded
	}

	// Calculate total response time
	totalTime := time.Since(start)

	// If any component is unhealthy, overall status is unhealthy
	for _, comp := range components {
		if comp.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
	}

	health := &SystemHealth{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     hc.getVersion(),
		Uptime:      hc.getUptime(),
		Environment: hc.getEnvironment(),
		Components:  components,
		SystemInfo:  hc.getSystemInfo(totalTime),
	}

	return health, nil
}

// checkApplicationHealth checks the application itself
func (hc *HealthChecker) checkApplicationHealth(ctx context.Context) ComponentHealth {
	start := time.Now()

	// Simple application health check
	message := "Application is running"
	status := StatusHealthy

	// Check if we can access basic application resources
	if hc.app == nil {
		message = "PocketBase app instance is nil"
		status = StatusUnhealthy
	}

	responseTime := time.Since(start)
	return ComponentHealth{
		Name:         "application",
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		LastCheck:    time.Now(),
	}
}

// checkDatabaseHealth checks database connectivity and performance
func (hc *HealthChecker) checkDatabaseHealth(ctx context.Context) ComponentHealth {
	start := time.Now()

	status := StatusHealthy
	message := "Database connection healthy"

	// Check database connection
	if hc.app.DB() == nil {
		status = StatusUnhealthy
		message = "Database connection is nil"
		responseTime := time.Since(start)
		return ComponentHealth{
			Name:         "database",
			Status:       status,
			Message:      message,
			ResponseTime: responseTime,
			LastCheck:    time.Now(),
		}
	}

	// Test database query
	query := "SELECT 1"
	_, err := hc.app.DB().NewQuery(query).Execute()
	if err != nil {
		status = StatusUnhealthy
		message = fmt.Sprintf("Database query failed: %v", err)
	} else {
		// Check database performance with a simple query
		testQuery := "SELECT COUNT(*) as count FROM urls"
		result, err := hc.app.DB().NewQuery(testQuery).Execute()
		if err != nil {
			status = StatusDegraded
			message = fmt.Sprintf("Database performance check failed: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			message = fmt.Sprintf("Database healthy, found %d URL records", rowsAffected)
		}
	}

	responseTime := time.Since(start)
	return ComponentHealth{
		Name:         "database",
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		LastCheck:    time.Now(),
	}
}

// checkExternalDependencies checks external service dependencies
func (hc *HealthChecker) checkExternalDependencies(ctx context.Context) ComponentHealth {
	start := time.Now()

	status := StatusHealthy
	message := "All external dependencies healthy"

	// In a real application, you would check external services here
	// For now, we'll just return healthy
	// Example checks you might add:
	// - Redis connectivity
	// - External API endpoints
	// - Message queue health
	// - CDN availability

	responseTime := time.Since(start)
	return ComponentHealth{
		Name:         "external_dependencies",
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		LastCheck:    time.Now(),
	}
}

// getVersion returns the application version
func (hc *HealthChecker) getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "1.0.0-dev"
}

// getUptime returns the application uptime
func (hc *HealthChecker) getUptime() string {
	// In a real application, you would track start time
	// For now, return a placeholder
	return "unknown"
}

// getEnvironment returns the current environment
func (hc *HealthChecker) getEnvironment() string {
	if env := os.Getenv("GO_ENV"); env != "" {
		return env
	}
	return "development"
}

// getSystemInfo returns system-level information
func (hc *HealthChecker) getSystemInfo(totalTime time.Duration) SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get database info
	dbInfo := hc.getDatabaseInfo()

	return SystemInfo{
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		MemoryUsage: MemoryStats{
			AllocatedBytes:      memStats.Alloc,
			TotalAllocatedBytes: memStats.TotalAlloc,
			SystemMemoryBytes:   memStats.Sys,
			GCRuns:              memStats.NumGC,
		},
		DatabaseInfo: dbInfo,
	}
}

// getDatabaseInfo returns database information
func (hc *HealthChecker) getDatabaseInfo() DBInfo {
	info := DBInfo{
		Type:             "sqlite",
		ConnectionStatus: "unknown",
	}

	if hc.app.DB() == nil {
		info.ConnectionStatus = "disconnected"
		return info
	}

	start := time.Now()
	_, err := hc.app.DB().NewQuery("SELECT 1").Execute()
	responseTime := time.Since(start)

	if err != nil {
		info.ConnectionStatus = "error"
		info.ResponseTime = responseTime
	} else {
		info.ConnectionStatus = "connected"
		info.ResponseTime = responseTime
	}

	return info
}

// RegisterHealthRoutes registers health check routes with the PocketBase app
func RegisterHealthRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	checker := NewHealthChecker(app)

	// Basic health check endpoint (using different path to avoid conflict)
	e.Router.GET("/api/system/health", func(e *core.RequestEvent) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		health, err := checker.CheckSystemHealth(ctx)
		if err != nil {
			return e.JSON(constants.HTTPStatusInternalServerError, map[string]interface{}{
				"status":    "error",
				"message":   err.Error(),
				"timestamp": time.Now(),
			})
		}

		statusCode := http.StatusOK
		if health.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if health.Status == StatusDegraded {
			statusCode = http.StatusPartialContent
		}

		return e.JSON(statusCode, health)
	})

	// Detailed health check endpoint
	e.Router.GET("/api/system/health/detailed", func(e *core.RequestEvent) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		health, err := checker.CheckSystemHealth(ctx)
		if err != nil {
			return e.JSON(constants.HTTPStatusInternalServerError, map[string]interface{}{
				"status":    "error",
				"message":   err.Error(),
				"timestamp": time.Now(),
			})
		}

		return e.JSON(http.StatusOK, health)
	})

	// Readiness probe endpoint (for Kubernetes/Docker health checks)
	e.Router.GET("/api/system/ready", func(e *core.RequestEvent) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		health, err := checker.CheckSystemHealth(ctx)
		if err != nil || health.Status == StatusUnhealthy {
			return e.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status":    "not ready",
				"timestamp": time.Now(),
			})
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now(),
		})
	})

	// Liveness probe endpoint (for Kubernetes/Docker health checks)
	e.Router.GET("/api/system/live", func(e *core.RequestEvent) error {
		// Simple liveness check - just verify the app is responding
		return e.JSON(http.StatusOK, map[string]interface{}{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	})

	// Metrics endpoint for monitoring systems
	e.Router.GET("/api/system/metrics", func(e *core.RequestEvent) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		health, err := checker.CheckSystemHealth(ctx)
		if err != nil {
			return e.JSON(constants.HTTPStatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			})
		}

		metrics := map[string]interface{}{
			"health_status":             string(health.Status),
			"uptime":                    health.Uptime,
			"goroutines":                health.SystemInfo.Goroutines,
			"memory_allocated_mb":       float64(health.SystemInfo.MemoryUsage.AllocatedBytes) / 1024 / 1024,
			"memory_total_allocated_mb": float64(health.SystemInfo.MemoryUsage.TotalAllocatedBytes) / 1024 / 1024,
			"memory_system_mb":          float64(health.SystemInfo.MemoryUsage.SystemMemoryBytes) / 1024 / 1024,
			"gc_runs":                   health.SystemInfo.MemoryUsage.GCRuns,
			"database_response_time_ms": float64(health.SystemInfo.DatabaseInfo.ResponseTime.Nanoseconds()) / 1e6,
			"database_status":           health.SystemInfo.DatabaseInfo.ConnectionStatus,
			"timestamp":                 time.Now(),
		}

		return e.JSON(http.StatusOK, metrics)
	})

	return nil
}
