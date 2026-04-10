package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type ExperimentSuggestionHandler struct {
	svc *service.ExperimentSuggestionService
}

func NewExperimentSuggestionHandler(svc *service.ExperimentSuggestionService) *ExperimentSuggestionHandler {
	return &ExperimentSuggestionHandler{svc: svc}
}

func (h *ExperimentSuggestionHandler) List(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId is required"})
		return
	}
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c), envID)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentSuggestionHandler) Generate(c *gin.Context) {
	var req service.GenerateSuggestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID, _ := c.Get("tenant_id")
	resp, err := h.svc.Generate(c.Request.Context(), tenantID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentSuggestionHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
