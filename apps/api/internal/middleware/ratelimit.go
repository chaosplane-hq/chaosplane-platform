package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	RequestsPerSecond int
	Window            time.Duration
}

func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 30,
		Window:            time.Second,
	}
}

type RateLimitSubject struct {
	Key   string
	Limit int
}

func RateLimit(rdb *redis.Client, pool *pgxpool.Pool, cfg RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok, err := resolveRateLimitSubject(c, pool, cfg)
		if err != nil {
			slog.Warn("failed to resolve rate limit subject, failing open", "error", err)
			c.Next()
			return
		}
		if !ok {
			c.Next()
			return
		}

		now := time.Now()
		windowStart := now.Add(-cfg.Window)

		rlCtx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		allowed, remaining, err := checkRateLimit(rlCtx, rdb, subject.Key, now, windowStart, subject.Limit)
		cancel()
		if err != nil {
			slog.Warn("rate limiter redis error, failing open", "error", err)
			c.Next()
			return
		}

		resetAt := now.Add(cfg.Window)
		c.Header("X-RateLimit-Limit", strconv.Itoa(subject.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			retryAfter := cfg.Window.Seconds()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func resolveRateLimitSubject(c *gin.Context, pool *pgxpool.Pool, cfg RateLimiterConfig) (RateLimitSubject, bool, error) {
	apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
	if apiKey != "" {
		limit := cfg.RequestsPerSecond
		if tenantID, ok := c.Get("tenant_id"); ok && tenantID.(string) != "" && pool != nil {
			resolved, err := tenantRateLimit(c.Request.Context(), pool, tenantID.(string), cfg.RequestsPerSecond)
			if err != nil {
				return RateLimitSubject{}, false, err
			}
			limit = resolved
		}
		return RateLimitSubject{Key: fmt.Sprintf("ratelimit:apikey:%s", apiKey), Limit: limit}, true, nil
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists || strings.TrimSpace(tenantID.(string)) == "" || pool == nil {
		return RateLimitSubject{}, false, nil
	}

	limit, err := tenantRateLimit(c.Request.Context(), pool, tenantID.(string), cfg.RequestsPerSecond)
	if err != nil {
		return RateLimitSubject{}, false, err
	}
	return RateLimitSubject{Key: fmt.Sprintf("ratelimit:tenant:%s", tenantID.(string)), Limit: limit}, true, nil
}

func tenantRateLimit(ctx context.Context, pool *pgxpool.Pool, tenantID string, fallback int) (int, error) {
	if pool == nil {
		return fallback, nil
	}
	var plan string
	err := pool.QueryRow(ctx, `SELECT plan FROM tenants WHERE id = $1::uuid AND deleted_at IS NULL`, tenantID).Scan(&plan)
	if err != nil {
		return fallback, err
	}
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "free":
		return 30, nil
	case "team":
		return 50, nil
	case "business":
		return 200, nil
	case "enterprise", "government":
		return 500, nil
	default:
		return fallback, nil
	}
}

func checkRateLimit(ctx context.Context, rdb *redis.Client, key string, now, windowStart time.Time, limit int) (bool, int, error) {
	pipe := rdb.Pipeline()

	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixMicro(), 10))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMicro()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	pipe.Expire(ctx, key, time.Second*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := int(countCmd.Val())
	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	if count >= limit {
		return false, 0, nil
	}

	return true, remaining, nil
}
