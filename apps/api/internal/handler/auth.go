package handler

import (
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req, parseClientIP(c.ClientIP()), c.Request.UserAgent())
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), &req, parseClientIP(c.ClientIP()), c.Request.UserAgent())
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Refresh(c.Request.Context(), &req, parseClientIP(c.ClientIP()), c.Request.UserAgent())
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req service.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), &req); err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req service.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.VerifyEmail(c.Request.Context(), &req); err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req service.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.ResendVerification(c.Request.Context(), &req)
	if err != nil {
		h.writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req service.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claimsValue, exists := c.Get("auth_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}
	claims, ok := claimsValue.(*service.AccessTokenClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth claims"})
		return
	}

	if err := h.svc.Logout(c.Request.Context(), claims, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	resp, err := h.svc.CurrentUser(c.Request.Context(), &service.CurrentUserRequest{
		UserID:   userID.(string),
		TenantID: tenantID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
func (h *AuthHandler) writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials), errors.Is(err, service.ErrRefreshTokenInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrEmailNotVerified), errors.Is(err, service.ErrCSRFInvalid):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrTokenExpired):
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrRefreshReuse):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrAccountLocked):
		c.JSON(http.StatusLocked, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseClientIP(raw string) net.IP {
	if raw == "" {
		return nil
	}
	return net.ParseIP(raw)
}
