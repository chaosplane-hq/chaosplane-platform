package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RequireTenantRole(pool *pgxpool.Pool, allowed ...string) gin.HandlerFunc {
	allowedRoles := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		allowedRoles[strings.ToLower(role)] = struct{}{}
	}

	return func(c *gin.Context) {
		userID, userExists := c.Get("user_id")
		tenantID, tenantExists := c.Get("tenant_id")
		if !userExists || !tenantExists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth context"})
			return
		}

		var role string
		err := pool.QueryRow(c.Request.Context(), `
			SELECT role
			FROM team_members tm
			JOIN teams t ON t.id = tm.team_id
			WHERE tm.user_id = $1::uuid AND t.tenant_id = $2::uuid AND t.deleted_at IS NULL
			ORDER BY tm.joined_at ASC
			LIMIT 1
		`, userID.(string), tenantID.(string)).Scan(&role)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}

		normalized := normalizeTenantRole(role)
		if _, ok := allowedRoles[normalized]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}

		c.Set("tenant_role", normalized)
		c.Next()
	}
}

func normalizeTenantRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "lead", "admin":
		return "admin"
	case "member", "editor", "operator":
		return "editor"
	default:
		return "viewer"
	}
}
