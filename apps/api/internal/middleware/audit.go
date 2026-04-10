package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuditLog(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		tenantID, _ := c.Get("tenant_id")
		userID, _ := c.Get("user_id")

		tenantStr, _ := tenantID.(string)
		userStr, _ := userID.(string)
		if tenantStr == "" {
			return
		}

		action := deriveAction(c.Request.Method, c.FullPath())
		resourceType, resourceID := deriveResource(c)

		go func() {
			_, err := pool.Exec(c.Request.Context(), `
				INSERT INTO audit_logs (tenant_id, user_id, action, resource_type, resource_id, ip_address, user_agent, request_method, request_path, response_status)
				VALUES ($1::uuid, NULLIF($2, '')::uuid, $3, $4, NULLIF($5, ''), $6::inet, $7, $8, $9, $10)
			`, tenantStr, userStr, action, resourceType, resourceID,
				c.ClientIP(), c.Request.UserAgent(), c.Request.Method, c.Request.URL.Path, c.Writer.Status())
			if err != nil {
				slog.Error("audit log write failed", "error", err, "action", action)
			}
		}()

		_ = start
	}
}

func deriveAction(method, path string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func deriveResource(c *gin.Context) (resourceType string, resourceID string) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasPrefix(parts[i], ":") {
			if i > 0 {
				resourceType = parts[i-1]
			}
			resourceID = c.Param(strings.TrimPrefix(parts[i], ":"))
			return
		}
	}

	if len(parts) > 0 {
		resourceType = parts[len(parts)-1]
	}
	return
}
