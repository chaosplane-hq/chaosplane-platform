package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type BillingHandler struct {
	svc *service.BillingService
}

func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

func (h *BillingHandler) GetStatus(c *gin.Context) {
	resp, err := h.svc.GetStatus(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeBillingError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BillingHandler) Upgrade(c *gin.Context) {
	var req service.UpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Upgrade(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeBillingError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BillingHandler) Cancel(c *gin.Context) {
	resp, err := h.svc.Cancel(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeBillingError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BillingHandler) Reactivate(c *gin.Context) {
	resp, err := h.svc.Reactivate(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeBillingError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BillingHandler) WebhookStripe(c *gin.Context) {
	h.processWebhook(c, "stripe")
}

func (h *BillingHandler) WebhookToss(c *gin.Context) {
	h.processWebhook(c, "toss")
}

func (h *BillingHandler) WebhookDodo(c *gin.Context) {
	h.processWebhook(c, "dodo")
}

func (h *BillingHandler) processWebhook(c *gin.Context, gateway string) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	eventType := c.GetHeader("X-Event-Type")
	if eventType == "" {
		eventType = "unknown"
	}
	if err := h.svc.ProcessWebhookEvent(c.Request.Context(), gateway, eventType, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func writeBillingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUsageLimitExceeded):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
