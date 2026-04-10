package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func TestJWT_ValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := service.NewAuthService(&database.Pool{}, &config.Config{
		JWTSecret:      "test-secret",
		AccessTokenTTL: 15 * time.Minute,
	})

	claims := &service.AccessTokenClaims{
		JTI:       "11111111-1111-1111-1111-111111111111",
		Subject:   "user-1",
		TenantID:  "tenant-1",
		Email:     "user@example.com",
		TokenType: "access",
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
	}
	signed, err := service.SignTestAccessToken(claims, "test-secret")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(JWT(auth))
	r.GET("/protected", func(c *gin.Context) {
		if got, _ := c.Get("user_id"); got != "user-1" {
			t.Fatalf("expected user_id to be set")
		}
		if got, _ := c.Get("tenant_id"); got != "tenant-1" {
			t.Fatalf("expected tenant_id to be set")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestJWT_MissingAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := service.NewAuthService(&database.Pool{}, &config.Config{JWTSecret: "test-secret"})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(JWT(auth))
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
