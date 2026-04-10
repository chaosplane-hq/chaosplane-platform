package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func CSRFSameSite(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		userID, userExists := c.Get("user_id")
		tenantID, tenantExists := c.Get("tenant_id")
		if !userExists || !tenantExists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth context"})
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if err := auth.ValidateCSRFToken(userID.(string), tenantID.(string), token); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.Next()
	}
}
