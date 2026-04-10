package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type ResultAnalysisHandler struct {
	svc *service.ResultAnalysisService
}

func NewResultAnalysisHandler(svc *service.ResultAnalysisService) *ResultAnalysisHandler {
	return &ResultAnalysisHandler{svc: svc}
}

func (h *ResultAnalysisHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c), c.Query("environmentId"))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ResultAnalysisHandler) Get(c *gin.Context) {
	resp, err := h.svc.Get(c.Request.Context(), actorFromContext(c), c.Param("id"))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ResultAnalysisHandler) Analyze(c *gin.Context) {
	var req service.AnalyzeResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID, _ := c.Get("tenant_id")
	resp, err := h.svc.Analyze(c.Request.Context(), tenantID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
