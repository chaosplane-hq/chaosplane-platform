package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type TopologyAnalysisHandler struct {
	svc     *service.TopologyAnalysisService
	topoSvc *service.TopologyService
}

func NewTopologyAnalysisHandler(svc *service.TopologyAnalysisService, topoSvc *service.TopologyService) *TopologyAnalysisHandler {
	return &TopologyAnalysisHandler{svc: svc, topoSvc: topoSvc}
}

func (h *TopologyAnalysisHandler) GetDependencyMap(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId is required"})
		return
	}
	resp, err := h.svc.GetDependencyMap(c.Request.Context(), actorFromContext(c), envID)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TopologyAnalysisHandler) GetDrifts(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId is required"})
		return
	}
	unackedOnly := c.Query("unackedOnly") == "true"
	limit := 50
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	resp, err := h.svc.GetDrifts(c.Request.Context(), actorFromContext(c), envID, unackedOnly, limit)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TopologyAnalysisHandler) AcknowledgeDrift(c *gin.Context) {
	if err := h.svc.AcknowledgeDrift(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TopologyAnalysisHandler) GetMetrics(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId is required"})
		return
	}
	limit := 100
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	resp, err := h.svc.GetMetrics(c.Request.Context(), actorFromContext(c), envID, c.Query("metricName"), limit)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TopologyAnalysisHandler) SubmitDependencies(c *gin.Context) {
	var req service.SubmitDependenciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var tid string
	if v, exists := c.Get("tenant_id"); exists && v != nil {
		tid = v.(string)
	}
	if tid == "" {
		resolved, err := h.topoSvc.TenantIDFromEnvironment(c.Request.Context(), req.EnvironmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environmentId"})
			return
		}
		tid = resolved
	}
	count, err := h.svc.SubmitDependencies(c.Request.Context(), tid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *TopologyAnalysisHandler) SubmitMetrics(c *gin.Context) {
	var req service.SubmitMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var tid string
	if v, exists := c.Get("tenant_id"); exists && v != nil {
		tid = v.(string)
	}
	if tid == "" {
		resolved, err := h.topoSvc.TenantIDFromEnvironment(c.Request.Context(), req.EnvironmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environmentId"})
			return
		}
		tid = resolved
	}
	count, err := h.svc.SubmitMetrics(c.Request.Context(), tid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *TopologyAnalysisHandler) SubmitDrift(c *gin.Context) {
	var req service.SubmitDriftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var tid string
	if v, exists := c.Get("tenant_id"); exists && v != nil {
		tid = v.(string)
	}
	if tid == "" {
		resolved, err := h.topoSvc.TenantIDFromEnvironment(c.Request.Context(), req.EnvironmentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environmentId"})
			return
		}
		tid = resolved
	}
	resp, err := h.svc.SubmitDrift(c.Request.Context(), tid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
