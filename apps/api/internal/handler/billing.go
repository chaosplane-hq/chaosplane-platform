package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type BillingHandler struct {
	svc *service.BillingService
	cfg *config.Config
}

func NewBillingHandler(svc *service.BillingService, cfg *config.Config) *BillingHandler {
	return &BillingHandler{svc: svc, cfg: cfg}
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

	if err := h.verifyWebhookSignature(gateway, c.Request.Header, body); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("signature verification failed: %v", err)})
		return
	}

	eventType := c.GetHeader("X-Event-Type")
	if gateway == "stripe" {
		eventType = extractStripeEventType(body)
	}
	if eventType == "" {
		eventType = "unknown"
	}
	if err := h.svc.ProcessWebhookEvent(c.Request.Context(), gateway, eventType, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *BillingHandler) verifyWebhookSignature(gateway string, headers http.Header, body []byte) error {
	switch gateway {
	case "stripe":
		return h.verifyStripeSignature(headers, body)
	case "toss":
		return h.verifyTossSignature(headers, body)
	case "dodo":
		return h.verifyDodoSignature(headers, body)
	}
	return nil
}

func (h *BillingHandler) verifyStripeSignature(headers http.Header, body []byte) error {
	secret := h.cfg.StripeWebhookSecret
	if secret == "" {
		return fmt.Errorf("stripe webhook secret not configured")
	}
	sigHeader := headers.Get("Stripe-Signature")
	if sigHeader == "" {
		return fmt.Errorf("missing Stripe-Signature header")
	}

	var timestamp, sig string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch strings.TrimSpace(kv[0]) {
		case "t":
			timestamp = kv[1]
		case "v1":
			sig = kv[1]
		}
	}
	if timestamp == "" || sig == "" {
		return fmt.Errorf("invalid Stripe-Signature format")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}
	if time.Since(time.Unix(ts, 0)) > 5*time.Minute {
		return fmt.Errorf("timestamp too old")
	}

	payload := fmt.Sprintf("%s.%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func (h *BillingHandler) verifyTossSignature(headers http.Header, body []byte) error {
	secret := h.cfg.TossWebhookSecret
	if secret == "" {
		return fmt.Errorf("toss webhook secret not configured")
	}
	sig := headers.Get("Toss-Signature")
	if sig == "" {
		return fmt.Errorf("missing Toss-Signature header")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func (h *BillingHandler) verifyDodoSignature(headers http.Header, body []byte) error {
	secret := h.cfg.DodoWebhookSecret
	if secret == "" {
		return fmt.Errorf("dodo webhook secret not configured")
	}
	sig := headers.Get("X-Dodo-Signature")
	if sig == "" {
		return fmt.Errorf("missing X-Dodo-Signature header")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func extractStripeEventType(body []byte) string {
	idx := strings.Index(string(body), `"type"`)
	if idx < 0 {
		return ""
	}
	rest := string(body)[idx+7:]
	start := strings.Index(rest, `"`)
	if start < 0 {
		return ""
	}
	end := strings.Index(rest[start+1:], `"`)
	if end < 0 {
		return ""
	}
	return rest[start+1 : start+1+end]
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
