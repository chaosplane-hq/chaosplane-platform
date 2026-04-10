package server

import (
	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/handler"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/middleware"
)

// Server holds the Gin engine and dependencies.
type Server struct {
	Router *gin.Engine
	Config *config.Config
	Pool   *database.Pool
}

// New creates a Server with the full middleware chain and routes.
func New(cfg *config.Config, pool *database.Pool, health *handler.HealthHandler) *Server {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware chain: Recovery → Logger → RequestID → CORS
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Public health endpoints
	r.GET("/healthz", health.Healthz)
	r.GET("/readyz", health.Readyz)

	// Protected routes: TenantContext → APIKey
	api := r.Group("/api")
	api.Use(middleware.TenantContext(pool.App))
	api.Use(middleware.APIKey(pool.App))
	{
		// Experiment CRUD endpoints will be added in Phase 1
		_ = api
	}

	return &Server{
		Router: r,
		Config: cfg,
		Pool:   pool,
	}
}
