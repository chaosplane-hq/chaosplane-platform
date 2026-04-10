package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originsMap := make(map[string]bool, len(allowedOrigins))
	allowAll := false
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAll = true
		}
		originsMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowAll || originsMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Request-ID, X-CSRF-Token")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
