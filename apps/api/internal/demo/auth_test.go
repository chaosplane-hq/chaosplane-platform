package demo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func TestDemoLoginHandler_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Demo:           true,
		JWTSecret:      "test-secret",
		AccessTokenTTL: 15 * time.Minute,
	}

	handler := DemoLoginHandler(nil, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/demo-login", nil)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.NotEmpty(t, resp["accessToken"])
	assert.Equal(t, "", resp["refreshToken"])
	assert.NotZero(t, resp["expiresIn"])

	user := resp["user"].(map[string]interface{})
	assert.Equal(t, UserID, user["id"])
	assert.Equal(t, UserEmail, user["email"])
	assert.Equal(t, UserName, user["name"])

	tenant := resp["tenant"].(map[string]interface{})
	assert.Equal(t, TenantID, tenant["id"])

	tokenStr := resp["accessToken"].(string)
	claims := &service.AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, UserID, claims.Subject)
	assert.Equal(t, TenantID, claims.TenantID)
	assert.Equal(t, UserEmail, claims.Email)
	assert.Equal(t, "access", claims.TokenType)
}

func TestDemoLoginHandler_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Demo:           false,
		JWTSecret:      "test-secret",
		AccessTokenTTL: 15 * time.Minute,
	}

	handler := DemoLoginHandler(nil, cfg)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/demo-login", nil)

	handler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
