package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type InvitationHandler struct {
	svc *service.InvitationService
}

func NewInvitationHandler(svc *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{svc: svc}
}

func (h *InvitationHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvitationHandler) Create(c *gin.Context) {
	var req service.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *InvitationHandler) Resend(c *gin.Context) {
	resp, err := h.svc.Resend(c.Request.Context(), actorFromContext(c), c.Param("id"))
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvitationHandler) Revoke(c *gin.Context) {
	if err := h.svc.Revoke(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeInvitationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	var req service.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Accept(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvitationHandler) Decline(c *gin.Context) {
	var req service.DeclineInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Decline(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvitationHandler) LookupByToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token query parameter is required"})
		return
	}
	resp, err := h.svc.LookupByToken(c.Request.Context(), token)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvitationHandler) AcceptByToken(c *gin.Context) {
	var req service.AcceptByTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.AcceptByToken(c.Request.Context(), &req)
	if err != nil {
		writeInvitationError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func writeInvitationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrHierarchyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvitationSignupRequired):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrTokenExpired):
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
