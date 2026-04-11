package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type MarketplaceHandler struct {
	svc *service.MarketplaceService
}

func NewMarketplaceHandler(svc *service.MarketplaceService) *MarketplaceHandler {
	return &MarketplaceHandler{svc: svc}
}

func (h *MarketplaceHandler) List(c *gin.Context) {
	limit := 50
	if v := c.Query("limit"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			limit = p
		}
	}
	resp, err := h.svc.List(c.Request.Context(), c.Query("category"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MarketplaceHandler) Install(c *gin.Context) {
	var req service.InstallPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Install(c.Request.Context(), actorFromContext(c), &req); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MarketplaceHandler) Uninstall(c *gin.Context) {
	if err := h.svc.Uninstall(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type FederationHandler struct {
	svc *service.FederationService
}

func NewFederationHandler(svc *service.FederationService) *FederationHandler {
	return &FederationHandler{svc: svc}
}

func (h *FederationHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *FederationHandler) Register(c *gin.Context) {
	var req service.RegisterClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Register(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *FederationHandler) Remove(c *gin.Context) {
	if err := h.svc.Remove(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type CICDHandler struct {
	svc *service.CICDService
}

func NewCICDHandler(svc *service.CICDService) *CICDHandler {
	return &CICDHandler{svc: svc}
}

func (h *CICDHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CICDHandler) Create(c *gin.Context) {
	var req service.CreateCICDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *CICDHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type PredictiveHandler struct {
	svc *service.PredictiveAnalysisService
}

func NewPredictiveHandler(svc *service.PredictiveAnalysisService) *PredictiveHandler {
	return &PredictiveHandler{svc: svc}
}

func (h *PredictiveHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c), c.Query("environmentId"), c.Query("status"))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PredictiveHandler) Run(c *gin.Context) {
	var req service.RunPredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID, _ := c.Get("tenant_id")
	resp, err := h.svc.Run(c.Request.Context(), tenantID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PredictiveHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), actorFromContext(c), c.Param("id"), req.Status); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
