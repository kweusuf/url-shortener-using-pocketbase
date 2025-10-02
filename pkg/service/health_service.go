package service

import (
	"context"
	"net/http"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/models"
	"github.com/pocketbase/pocketbase/core"
)

// HealthChecker interface for dependency injection
type HealthChecker interface {
	CheckSystemHealth(ctx context.Context) (*models.SystemHealth, error)
}

// HealthService handles health check HTTP responses
type HealthService struct {
	checker HealthChecker
}

// NewHealthService creates a new health service instance
func NewHealthService(checker HealthChecker) *HealthService {
	return &HealthService{
		checker: checker,
	}
}

// CheckSystemMetrics handles metrics endpoint responses
func (hs *HealthService) CheckSystemMetrics(e *core.RequestEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := hs.checker.CheckSystemHealth(ctx)
	if err != nil {
		return e.JSON(constants.HTTPStatusInternalServerError, models.HealthMetricsErrorResponse{
			Error: err.Error(),
		})
	}

	metrics := models.HealthMetricsResponse{
		HealthStatus:           string(health.Status),
		Uptime:                 health.Uptime,
		Goroutines:             health.SystemInfo.Goroutines,
		MemoryAllocatedMB:      float64(health.SystemInfo.MemoryUsage.AllocatedBytes) / 1024 / 1024,
		MemoryTotalAllocatedMB: float64(health.SystemInfo.MemoryUsage.TotalAllocatedBytes) / 1024 / 1024,
		MemorySystemMB:         float64(health.SystemInfo.MemoryUsage.SystemMemoryBytes) / 1024 / 1024,
		GCRuns:                 health.SystemInfo.MemoryUsage.GCRuns,
		DatabaseResponseTimeMS: float64(health.SystemInfo.DatabaseInfo.ResponseTime.Nanoseconds()) / 1e6,
		DatabaseStatus:         health.SystemInfo.DatabaseInfo.ConnectionStatus,
		Timestamp:              time.Now(),
	}

	return e.JSON(http.StatusOK, metrics)
}

// CheckSystemLive handles liveness probe responses
func (hs *HealthService) CheckSystemLive(e *core.RequestEvent) error {
	return e.JSON(http.StatusOK, models.HealthLiveResponse{
		Status:    "alive",
		Timestamp: time.Now(),
	})
}

// CheckSystemReady handles readiness probe responses
func (hs *HealthService) CheckSystemReady(e *core.RequestEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := hs.checker.CheckSystemHealth(ctx)
	if err != nil || health.Status == constants.StatusUnhealthy {
		return e.JSON(http.StatusServiceUnavailable, models.HealthReadyResponse{
			Status:    "not ready",
			Timestamp: time.Now(),
		})
	}

	return e.JSON(http.StatusOK, models.HealthReadyResponse{
		Status:    "ready",
		Timestamp: time.Now(),
	})
}

// CheckSystemHealthDetailed handles detailed health check responses
func (hs *HealthService) CheckSystemHealthDetailed(e *core.RequestEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	health, err := hs.checker.CheckSystemHealth(ctx)
	if err != nil {
		return e.JSON(constants.HTTPStatusInternalServerError, models.HealthErrorResponse{
			Status:    "error",
			Message:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	return e.JSON(http.StatusOK, health)
}

// CheckSystemHealth handles basic health check responses
func (hs *HealthService) CheckSystemHealth(e *core.RequestEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := hs.checker.CheckSystemHealth(ctx)
	if err != nil {
		return e.JSON(constants.HTTPStatusInternalServerError, models.HealthErrorResponse{
			Status:    "error",
			Message:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	statusCode := http.StatusOK
	if health.Status == constants.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == constants.StatusDegraded {
		statusCode = http.StatusPartialContent
	}

	return e.JSON(statusCode, health)
}
