package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	RequestsPerSecond int
	Window            time.Duration
}

func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 100,
		Window:            time.Second,
	}
}

// RateLimit implements a Redis sliding window rate limiter per API key.
// Fail-open: if Redis is unavailable, requests are allowed through.
func RateLimit(rdb *redis.Client, cfg RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		now := time.Now()
		windowStart := now.Add(-cfg.Window)
		key := fmt.Sprintf("ratelimit:%s", apiKey)

		allowed, remaining, err := checkRateLimit(c.Request.Context(), rdb, key, now, windowStart, cfg.RequestsPerSecond)
		if err != nil {
			slog.Warn("rate limiter redis error, failing open", "error", err)
			c.Next()
			return
		}

		resetAt := now.Add(cfg.Window)
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerSecond))
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
