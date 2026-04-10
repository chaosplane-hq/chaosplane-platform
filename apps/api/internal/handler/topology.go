package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type TopologyHandler struct {
	svc *service.TopologyService
}

func NewTopologyHandler(svc *service.TopologyService) *TopologyHandler {
	return &TopologyHandler{svc: svc}
}

func (h *TopologyHandler) Latest(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId query parameter is required"})
		return
	}
	resp, err := h.svc.Latest(c.Request.Context(), actorFromContext(c), envID)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TopologyHandler) List(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId query parameter is required"})
		return
	}
	limit := 10
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c), envID, limit)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TopologyHandler) Submit(c *gin.Context) {
	var req service.SubmitTopologyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID, _ := c.Get("tenant_id")
	resp, err := h.svc.Submit(c.Request.Context(), tenantID.(string), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}
