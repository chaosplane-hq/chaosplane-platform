package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type tenantContextKey struct{}

// TenantIDFromContext extracts the tenant ID from the context.
func TenantIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tenantContextKey{}).(string); ok {
		return v
	}
	return ""
}

// TenantContext opens a per-request transaction, sets app.current_tenant_id on
// it via set_config (transaction-scoped, so Postgres resets it on commit/rollback
// and tenant context can never leak across requests), and stores the tx on the
// request context. All handler queries run on that same connection so RLS sees
// the tenant. The tx commits on success and rolls back on error/abort.
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

		ctx := c.Request.Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			slog.Error("failed to begin tenant transaction", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}

		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback(context.Background())
			}
		}()

		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID); err != nil {
			slog.Error("failed to set tenant context", "error", err, "tenant_id", tenantID)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}

		ctx = database.WithTx(ctx, tx)
		ctx = context.WithValue(ctx, tenantContextKey{}, tenantID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("tenant_id_str", tenantID)

		c.Next()

		if c.IsAborted() || len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			return
		}
		if err := tx.Commit(ctx); err != nil {
			slog.Error("failed to commit tenant transaction", "error", err)
			return
		}
		committed = true
	}
}
