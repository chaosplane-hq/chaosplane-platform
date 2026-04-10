package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

func (h *AgentHandler) ListTokens(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId query parameter is required"})
		return
	}
	resp, err := h.svc.ListTokens(c.Request.Context(), actorFromContext(c), envID)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) CreateToken(c *gin.Context) {
	var req service.CreateAgentTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreateToken(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *AgentHandler) RevokeToken(c *gin.Context) {
	if err := h.svc.RevokeToken(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeAgentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AgentHandler) Register(c *gin.Context) {
	var req service.RegisterAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) Heartbeat(c *gin.Context) {
	var req service.AgentHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Heartbeat(c.Request.Context(), &req)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func writeAgentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAgentTokenInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrHierarchyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
