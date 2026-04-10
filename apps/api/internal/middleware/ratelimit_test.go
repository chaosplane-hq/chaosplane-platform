package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimit_NoAPIKey_PassesThrough(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(RateLimit(rdb, nil, DefaultRateLimiterConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimit_RedisUnavailable_FailOpen(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.Use(RateLimit(rdb, nil, DefaultRateLimiterConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("X-API-Key", "test-key")
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (fail-open), got %d", w.Code)
	}
}

func TestRateLimit_HeadersSet(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(RateLimit(rdb, nil, DefaultRateLimiterConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	if cfg.RequestsPerSecond != 10 {
		t.Errorf("expected 10 req/s, got %d", cfg.RequestsPerSecond)
	}
}
