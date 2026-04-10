package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AccountHandler struct {
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) AcceptTOS(c *gin.Context) {
	var req service.AcceptTOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AcceptTOS(c.Request.Context(), actorFromContext(c), &req); err != nil {
		writeAccountError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AccountHandler) Export(c *gin.Context) {
	resp, err := h.svc.Export(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	var req service.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), actorFromContext(c), &req); err != nil {
		writeAccountError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AccountHandler) CancelDeletion(c *gin.Context) {
	if err := h.svc.CancelDeletion(c.Request.Context(), actorFromContext(c)); err != nil {
		writeAccountError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeAccountError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrHierarchyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrTOSOutdated):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrAccountDeleted):
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
