package handler

import (
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type OAuthHandler struct {
	svc *service.OAuthService
}

func NewOAuthHandler(svc *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{svc: svc}
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	provider := c.Param("provider")
	resp, err := h.svc.Authorize(c.Request.Context(), provider)
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	var req service.OAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := net.ParseIP(c.ClientIP())
	resp, err := h.svc.Callback(c.Request.Context(), &req, ip, c.Request.UserAgent())
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func writeOAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrOAuthProviderDisabled):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
