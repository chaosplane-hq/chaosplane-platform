package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func JWT(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		token, err := bearerToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		claims, err := auth.ParseAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		blacklisted, err := auth.IsBlacklisted(c.Request.Context(), claims.JTI)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to validate access token"})
			return
		}
		if blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "access token revoked"})
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("tenant_id", claims.TenantID)
		c.Set("auth_claims", claims)
		c.Next()
	}
}

func bearerToken(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("invalid Authorization header")
	}
	return strings.TrimSpace(parts[1]), nil
}
