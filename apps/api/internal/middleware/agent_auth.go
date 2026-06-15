package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type agentEnvironmentKey struct{}

// EnvironmentIDFromContext returns the environment id resolved from the agent
// token by AgentAuth, or "" when the request was not agent-authenticated.
func EnvironmentIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(agentEnvironmentKey{}).(string); ok {
		return v
	}
	return ""
}

// AgentAuth authenticates a customer-cluster agent via an opaque bearer token
// (Authorization: Bearer cpagent_...). It derives (tenant_id, environment_id)
// from the token hash, never from a client-supplied header, then opens a
// request-scoped transaction and sets app.current_tenant_id so RLS-protected
// tables stay tenant-isolated for the rest of the request. This mirrors
// TenantContext but the identity is token-derived, which is the cross-tenant
// escalation defense for agent routes.
func AgentAuth(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := bearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing agent token"})
			return
		}

		sum := sha256.Sum256([]byte(token))
		tokenHash := hex.EncodeToString(sum[:])

		ctx := c.Request.Context()
		var tenantID, environmentID string
		err = pool.QueryRow(ctx, `
			SELECT tenant_id::text, environment_id::text
			FROM agent_tokens
			WHERE token_hash = $1 AND revoked_at IS NULL
		`, tokenHash).Scan(&tenantID, &environmentID)
		if err != nil {
			if err != pgx.ErrNoRows {
				slog.Warn("agent token lookup failed", "error", err, "client_ip", c.ClientIP())
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid agent token"})
			return
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			slog.Error("failed to begin agent transaction", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback(context.Background())
			}
		}()

		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID); err != nil {
			slog.Error("failed to set agent tenant context", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		ctx = database.WithTx(ctx, tx)
		ctx = context.WithValue(ctx, agentEnvironmentKey{}, environmentID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("tenant_id", tenantID)
		c.Set("environment_id", environmentID)

		_, _ = tx.Exec(ctx, "UPDATE agent_tokens SET last_used_at = now() WHERE token_hash = $1", tokenHash)

		c.Next()

		if c.IsAborted() || len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			return
		}
		if err := tx.Commit(ctx); err != nil {
			slog.Error("failed to commit agent transaction", "error", err)
			return
		}
		committed = true
	}
}
