package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AIChatHandler struct {
	svc *service.AIChatService
}

func NewAIChatHandler(svc *service.AIChatService) *AIChatHandler {
	return &AIChatHandler{svc: svc}
}

func (h *AIChatHandler) ListSessions(c *gin.Context) {
	resp, err := h.svc.ListSessions(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AIChatHandler) CreateSession(c *gin.Context) {
	var req service.CreateChatSessionRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	resp, err := h.svc.CreateSession(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *AIChatHandler) GetMessages(c *gin.Context) {
	resp, err := h.svc.GetMessages(c.Request.Context(), actorFromContext(c), c.Param("id"))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AIChatHandler) SendMessage(c *gin.Context) {
	var req service.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.SendMessage(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AIChatHandler) DeleteSession(c *gin.Context) {
	if err := h.svc.DeleteSession(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
