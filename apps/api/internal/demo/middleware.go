package demo

import (
	"net/http"
	"strings"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/gin-gonic/gin"
)

func WriteProtection(cfg *config.Config) gin.HandlerFunc {
	allowlist := []struct {
		method  string
		pattern string
	}{
		{"POST", "/api/v1/experiments"},
		{"POST", "/api/v1/experiments/scenario"},
		{"POST", "/api/v1/ai/chat/sessions"},
		{"POST", "/api/v1/ai/chat/sessions/*/messages"},
		{"POST", "/api/v1/agents/test-connection"},
		{"POST", "/auth/demo-login"},
		{"POST", "/api/v1/suggestions"},
	}

	return func(c *gin.Context) {
		if !cfg.Demo {
			c.Next()
			return
		}

		tenantID := c.GetString("tenant_id")
		if tenantID != TenantID {
			c.Next()
			return
		}

		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		for _, entry := range allowlist {
			if method != entry.method {
				continue
			}
			if strings.Contains(entry.pattern, "*") {
				parts := strings.Split(entry.pattern, "*")
				if strings.HasPrefix(path, parts[0]) && strings.HasSuffix(path, parts[1]) {
					c.Next()
					return
				}
			} else if path == entry.pattern {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "mutations disabled in demo mode"})
	}
}
