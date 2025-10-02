package monitoring

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/models"
	"github.com/kweusuf/pocketbase-demo/pkg/service"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/db"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// HealthChecker handles all health check operations
type HealthChecker struct {
	app *pocketbase.PocketBase
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(app *pocketbase.PocketBase) *HealthChecker {
	return &HealthChecker{app: app}
}

// CheckSystemHealth performs a comprehensive health check
func (hc *HealthChecker) CheckSystemHealth(ctx context.Context) (*models.SystemHealth, error) {
	start := time.Now()

	components := []models.ComponentHealth{}
	overallStatus := constants.StatusHealthy

	// Check application component
	appHealth := hc.checkApplicationHealth(ctx)
	components = append(components, appHealth)
	if appHealth.Status != constants.StatusHealthy {
		overallStatus = constants.StatusUnhealthy
	}

	// Check database component
	dbHealth := hc.checkDatabaseHealth(ctx)
	components = append(components, dbHealth)
	if dbHealth.Status != constants.StatusHealthy && overallStatus == constants.StatusHealthy {
		overallStatus = constants.StatusDegraded
	}

	// Check external dependencies
	depsHealth := hc.checkExternalDependencies(ctx)
	components = append(components, depsHealth)
	if depsHealth.Status != constants.StatusHealthy && overallStatus == constants.StatusHealthy {
		overallStatus = constants.StatusDegraded
	}

	// Calculate total response time
	totalTime := time.Since(start)

	// If any component is unhealthy, overall status is unhealthy
	for _, comp := range components {
		if comp.Status == constants.StatusUnhealthy {
			overallStatus = constants.StatusUnhealthy
			break
		}
	}

	health := &models.SystemHealth{
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
func (hc *HealthChecker) checkApplicationHealth(ctx context.Context) models.ComponentHealth {
	start := time.Now()

	// Simple application health check
	message := "Application is running"
	status := constants.StatusHealthy

	// Check if we can access basic application resources
	if hc.app == nil {
		message = "PocketBase app instance is nil"
		status = constants.StatusUnhealthy
	}

	responseTime := time.Since(start)
	return models.ComponentHealth{
		Name:         "application",
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		LastCheck:    time.Now(),
	}
}

// checkDatabaseHealth checks database connectivity and performance
func (hc *HealthChecker) checkDatabaseHealth(ctx context.Context) models.ComponentHealth {
	start := time.Now()

	status := constants.StatusHealthy
	message := "Database connection healthy"

	// Check database connection
	if hc.app.DB() == nil {
		status = constants.StatusUnhealthy
		message = "Database connection is nil"
		responseTime := time.Since(start)
		return models.ComponentHealth{
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
		status = constants.StatusUnhealthy
		message = fmt.Sprintf("Database query failed: %v", err)
	} else {
		// Check database performance with a simple query
		testQuery := "SELECT COUNT(*) as count FROM urls"
		result, err := hc.app.DB().NewQuery(testQuery).Execute()
		if err != nil {
			status = constants.StatusDegraded
			message = fmt.Sprintf("Database performance check failed: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			message = fmt.Sprintf("Database healthy, found %d URL records", rowsAffected)
		}
	}

	responseTime := time.Since(start)
	return models.ComponentHealth{
		Name:         "database",
		Status:       status,
		Message:      message,
		ResponseTime: responseTime,
		LastCheck:    time.Now(),
	}
}

// checkExternalDependencies checks external service dependencies
func (hc *HealthChecker) checkExternalDependencies(ctx context.Context) models.ComponentHealth {
	start := time.Now()

	status := constants.StatusHealthy
	message := "All external dependencies healthy"

	// In a real application, you would check external services here
	// For now, we'll just return healthy
	// Example checks you might add:
	// - Redis connectivity
	// - External API endpoints
	// - Message queue health
	// - CDN availability

	responseTime := time.Since(start)
	return models.ComponentHealth{
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
func (hc *HealthChecker) getSystemInfo(totalTime time.Duration) models.SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get database info
	dbInfo := db.GetDatabaseInfo(hc.app)

	return models.SystemInfo{
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		MemoryUsage: models.MemoryStats{
			AllocatedBytes:      memStats.Alloc,
			TotalAllocatedBytes: memStats.TotalAlloc,
			SystemMemoryBytes:   memStats.Sys,
			GCRuns:              memStats.NumGC,
		},
		DatabaseInfo: dbInfo,
	}
}

// RegisterHealthRoutes registers health check routes with the PocketBase app
func RegisterHealthRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	checker := NewHealthChecker(app)
	healthService := service.NewHealthService(checker)

	// Basic health check endpoint (using different path to avoid conflict)
	e.Router.GET("/api/system/health", func(e *core.RequestEvent) error {
		return healthService.CheckSystemHealth(e)
	})

	// Detailed health check endpoint
	e.Router.GET("/api/system/health/detailed", func(e *core.RequestEvent) error {
		return healthService.CheckSystemHealthDetailed(e)
	})

	// Readiness probe endpoint (for Kubernetes/Docker health checks)
	e.Router.GET("/api/system/ready", func(e *core.RequestEvent) error {
		return healthService.CheckSystemReady(e)
	})

	// Liveness probe endpoint (for Kubernetes/Docker health checks)
	e.Router.GET("/api/system/live", func(e *core.RequestEvent) error {
		// Simple liveness check - just verify the app is responding
		return healthService.CheckSystemLive(e)
	})

	// Metrics endpoint for monitoring systems
	e.Router.GET("/api/system/metrics", func(e *core.RequestEvent) error {
		return healthService.CheckSystemMetrics(e)
	})

	return nil
}
