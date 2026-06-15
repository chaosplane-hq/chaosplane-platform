package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

// EitherAuth accepts a JWT Bearer token OR an X-API-Key so a route is reachable
// by both the web dashboard (JWT) and CLI/agents (API key). Both underlying
// middlewares set "tenant_id"/"user_id", so the shared TenantContext that runs
// next opens the RLS transaction regardless of how the caller authenticated.
// Bearer takes precedence when both headers are present.
func EitherAuth(auth *service.AuthService, pool *pgxpool.Pool) gin.HandlerFunc {
	jwt := JWT(auth)
	apiKey := APIKey(pool)
	return func(c *gin.Context) {
		switch {
		case c.GetHeader("Authorization") != "":
			jwt(c)
		case c.GetHeader("X-API-Key") != "":
			apiKey(c)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing credentials: provide a Bearer token or X-API-Key",
			})
		}
	}
}
