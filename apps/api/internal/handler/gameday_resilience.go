package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type GameDayHandler struct {
	svc *service.GameDayService
}

func NewGameDayHandler(svc *service.GameDayService) *GameDayHandler {
	return &GameDayHandler{svc: svc}
}

func (h *GameDayHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GameDayHandler) Get(c *gin.Context) {
	resp, err := h.svc.Get(c.Request.Context(), actorFromContext(c), c.Param("id"))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GameDayHandler) Create(c *gin.Context) {
	var req service.CreateGameDayRequest
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

func (h *GameDayHandler) UpdateStatus(c *gin.Context) {
	var req service.UpdateGameDayStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.UpdateStatus(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GameDayHandler) AddEvent(c *gin.Context) {
	var req service.AddGameDayEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.AddEvent(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *GameDayHandler) CreatePostmortem(c *gin.Context) {
	var req service.CreatePostmortemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreatePostmortem(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

type ResilienceScoreHandler struct {
	svc *service.ResilienceScoreService
}

func NewResilienceScoreHandler(svc *service.ResilienceScoreService) *ResilienceScoreHandler {
	return &ResilienceScoreHandler{svc: svc}
}

func (h *ResilienceScoreHandler) Get(c *gin.Context) {
	envID := c.Query("environmentId")
	if envID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environmentId is required"})
		return
	}
	resp, err := h.svc.Get(c.Request.Context(), actorFromContext(c), envID)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ResilienceScoreHandler) Calculate(c *gin.Context) {
	var req service.CalculateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantID, _ := c.Get("tenant_id")
	resp, err := h.svc.Calculate(c.Request.Context(), tenantID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

type WorkflowTemplateHandler struct {
	svc *service.WorkflowTemplateService
}

func NewWorkflowTemplateHandler(svc *service.WorkflowTemplateService) *WorkflowTemplateHandler {
	return &WorkflowTemplateHandler{svc: svc}
}

func (h *WorkflowTemplateHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WorkflowTemplateHandler) Create(c *gin.Context) {
	var req service.CreateWorkflowTemplateRequest
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

func (h *WorkflowTemplateHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
