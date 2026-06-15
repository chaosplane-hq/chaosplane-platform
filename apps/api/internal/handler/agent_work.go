package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/middleware"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AgentWorkHandler struct {
	svc *service.AgentWorkService
}

func NewAgentWorkHandler(svc *service.AgentWorkService) *AgentWorkHandler {
	return &AgentWorkHandler{svc: svc}
}

func (h *AgentWorkHandler) ClaimWork(c *gin.Context) {
	envID := middleware.EnvironmentIDFromContext(c.Request.Context())
	agentInstance := c.GetHeader("X-Agent-Instance")
	if agentInstance == "" {
		agentInstance = "default"
	}

	work, err := h.svc.ClaimWork(c.Request.Context(), envID, agentInstance)
	if err != nil {
		if errors.Is(err, service.ErrNoWork) {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"experiment": work})
}

func (h *AgentWorkHandler) ReportStatus(c *gin.Context) {
	var req service.AgentStatusReport
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	envID := middleware.EnvironmentIDFromContext(c.Request.Context())
	agentInstance := c.GetHeader("X-Agent-Instance")
	if agentInstance == "" {
		agentInstance = "default"
	}

	ack, err := h.svc.ReportStatus(c.Request.Context(), envID, c.Param("id"), agentInstance, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotClaimable):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, ack)
}
