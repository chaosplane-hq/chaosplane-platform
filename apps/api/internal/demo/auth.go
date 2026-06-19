package demo

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func DemoLoginHandler(authSvc *service.AuthService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Demo {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		now := time.Now()
		claims := &service.AccessTokenClaims{
			TenantID:  TenantID,
			Email:     UserEmail,
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.NewString(),
				Subject:   UserID,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTokenTTL)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		accessToken, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"accessToken":  accessToken,
			"refreshToken": "",
			"expiresIn":    int64(cfg.AccessTokenTTL.Seconds()),
			"user": gin.H{
				"id":    UserID,
				"email": UserEmail,
				"name":  UserName,
			},
			"tenant": gin.H{
				"id":   TenantID,
				"name": OrgName,
				"slug": "demo",
			},
		})
	}
}
