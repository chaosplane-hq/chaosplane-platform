package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tenantContextKey struct{}

// TenantIDFromContext extracts the tenant ID from the context.
func TenantIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tenantContextKey{}).(string); ok {
		return v
	}
	return ""
}

// TenantContext extracts tenant_id from the request (JWT claims or API key lookup)
// and calls SET LOCAL app.current_tenant_id on the DB connection within the
// request transaction (ADR-001).
func TenantContext(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			if v, exists := c.Get("tenant_id"); exists {
				tenantID, _ = v.(string)
			}
		}

		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "tenant_id is required",
			})
			return
		}

		conn, err := pool.Acquire(c.Request.Context())
		if err != nil {
			slog.Error("failed to acquire connection for tenant context", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}
		defer conn.Release()

		// ADR-001: SET LOCAL scopes the setting to the current transaction
		_, err = conn.Exec(c.Request.Context(), "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			slog.Error("failed to set tenant context", "error", err, "tenant_id", tenantID)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}

		ctx := context.WithValue(c.Request.Context(), tenantContextKey{}, tenantID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("tenant_id_str", tenantID)

		c.Next()
	}
}
