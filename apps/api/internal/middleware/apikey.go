package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APIKey validates the X-API-Key header against the database and sets user context.
func APIKey(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing API key",
			})
			return
		}

		var userID, tenantID string
		err := pool.QueryRow(
			c.Request.Context(),
			"SELECT user_id, tenant_id FROM api_keys WHERE key_hash = digest($1, 'sha256') AND revoked_at IS NULL",
			key,
		).Scan(&userID, &tenantID)
		if err != nil {
			slog.Warn("invalid API key", "error", err, "client_ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		c.Set("user_id", userID)
		c.Set("tenant_id", tenantID)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), tenantContextKey{}, tenantID),
		)

		c.Next()
	}
}
