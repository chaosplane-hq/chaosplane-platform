package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

// ADR-006: All handlers MUST use Request DTOs, never bind directly to DB models.
// Pattern: c.ShouldBindJSON(&req) → model.Field = req.Field
// FORBIDDEN: c.ShouldBindJSON(&model)

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	pool *database.Pool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *database.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Healthz always returns 200 OK.
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz checks DB connectivity and returns 200 or 503.
func (h *HealthHandler) Readyz(c *gin.Context) {
	if err := h.pool.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
