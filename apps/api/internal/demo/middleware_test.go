package demo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWriteProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		demo       bool
		tenantID   string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET experiments allowed",
			demo:       true,
			tenantID:   TenantID,
			method:     http.MethodGet,
			path:       "/api/v1/experiments",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST experiments allowed",
			demo:       true,
			tenantID:   TenantID,
			method:     http.MethodPost,
			path:       "/api/v1/experiments",
			wantStatus: http.StatusOK,
		},
		{
			name:       "DELETE experiments blocked",
			demo:       true,
			tenantID:   TenantID,
			method:     http.MethodDelete,
			path:       "/api/v1/experiments/123",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "POST billing blocked",
			demo:       true,
			tenantID:   TenantID,
			method:     http.MethodPost,
			path:       "/api/v1/billing/upgrade",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "POST chat messages allowed",
			demo:       true,
			tenantID:   TenantID,
			method:     http.MethodPost,
			path:       "/api/v1/ai/chat/sessions/abc/messages",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-demo tenant passes through",
			demo:       true,
			tenantID:   "other-tenant-id",
			method:     http.MethodPost,
			path:       "/api/v1/billing/upgrade",
			wantStatus: http.StatusOK,
		},
		{
			name:       "demo disabled passes through",
			demo:       false,
			tenantID:   TenantID,
			method:     http.MethodDelete,
			path:       "/api/v1/experiments/123",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)

			cfg := &config.Config{Demo: tt.demo}

			engine.Use(func(c *gin.Context) {
				c.Set("tenant_id", tt.tenantID)
				c.Next()
			})
			engine.Use(WriteProtection(cfg))
			engine.Handle(tt.method, tt.path, func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			c.Request = httptest.NewRequest(tt.method, tt.path, nil)
			engine.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
