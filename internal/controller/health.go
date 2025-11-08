package controller

import (
	"cruder/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthController handles health check endpoints for Kubernetes probes
type HealthController struct {
	dbConn *repository.PostgresConnection
}

// NewHealthController creates a new health controller
func NewHealthController(dbConn *repository.PostgresConnection) *HealthController {
	return &HealthController{
		dbConn: dbConn,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// LivenessProbe handles the /health endpoint (liveness probe)
//
// Purpose: Tells Kubernetes if the app is alive and should keep running
// - Returns 200: App is alive, keep it running
// - Returns 500+: App is broken, restart the pod
//
// Use case: Detects deadlocks, infinite loops, or crashed processes
func (h *HealthController) LivenessProbe(ctx *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	}

	ctx.JSON(http.StatusOK, response)
}

// ReadinessProbe handles the /ready endpoint (readiness probe)
//
// Purpose: Tells Kubernetes if the app is ready to receive traffic
// - Returns 200: App is ready, send traffic
// - Returns 500+: App is not ready, remove from load balancer
//
// Use case: Detects when app is starting up or database is unreachable
//
// Difference from Liveness:
// - Liveness: "Is app alive?" → Restart if not
// - Readiness: "Is app ready for traffic?" → Remove from LB if not
func (h *HealthController) ReadinessProbe(ctx *gin.Context) {
	checks := make(map[string]string)

	// Check database connection
	if err := h.dbConn.DB().Ping(); err != nil {
		checks["database"] = "unhealthy: " + err.Error()

		response := HealthResponse{
			Status:    "not_ready",
			Timestamp: time.Now().UTC(),
			Checks:    checks,
		}

		ctx.JSON(http.StatusServiceUnavailable, response)
		return
	}

	checks["database"] = "healthy"

	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	ctx.JSON(http.StatusOK, response)
}
