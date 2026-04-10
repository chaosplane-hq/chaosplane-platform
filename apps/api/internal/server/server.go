package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/handler"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/middleware"
)

type Server struct {
	Router *gin.Engine
	Config *config.Config
	Pool   *database.Pool
}

func New(
	cfg *config.Config,
	pool *database.Pool,
	health *handler.HealthHandler,
	experiments *handler.ExperimentHandler,
	policies *handler.PolicyHandler,
	ws *handler.WebSocketHandler,
	rdb *redis.Client,
) *Server {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.GET("/healthz", health.Healthz)
	r.GET("/readyz", health.Readyz)

	api := r.Group("/api/v1")
	api.Use(middleware.APIKey(pool.App))
	if rdb != nil {
		api.Use(middleware.RateLimit(rdb, middleware.DefaultRateLimiterConfig()))
	}
	{
		api.POST("/experiments", experiments.Create)
		api.GET("/experiments", experiments.List)
		api.GET("/experiments/:name", experiments.Get)
		api.DELETE("/experiments/:name", experiments.Delete)
		api.POST("/experiments/:name/abort", experiments.Abort)

		api.GET("/policies", policies.List)
		api.GET("/policies/:name", policies.Get)
	}

	r.GET("/ws/experiments/:name", ws.ExperimentStatus)

	return &Server{
		Router: r,
		Config: cfg,
		Pool:   pool,
	}
}

func NewRedisClient(cfg *config.Config) *redis.Client {
	if cfg.RedisURL == "" {
		slog.Warn("REDIS_URL not set, rate limiting disabled")
		return nil
	}
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse REDIS_URL, rate limiting disabled", "error", err)
		return nil
	}
	return redis.NewClient(opts)
}
