package httproutes

import (
	"github.com/kweusuf/pocketbase-demo/pkg/service"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/monitoring"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterHealthRoutes registers health check routes with the PocketBase app
func RegisterHealthRoutes(app *pocketbase.PocketBase, e *core.ServeEvent) error {
	checker := monitoring.NewHealthChecker(app)
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
		return healthService.CheckSystemLive(e)
	})

	// Metrics endpoint for monitoring systems
	e.Router.GET("/api/system/metrics", func(e *core.RequestEvent) error {
		return healthService.CheckSystemMetrics(e)
	})

	return nil
}
