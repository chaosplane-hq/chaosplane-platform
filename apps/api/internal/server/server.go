package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/handler"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/middleware"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
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
	auth *handler.AuthHandler,
	hierarchy *handler.HierarchyHandler,
	onboarding *handler.OnboardingHandler,
	invitations *handler.InvitationHandler,
	apiKeys *handler.APIKeyHandler,
	oauth *handler.OAuthHandler,
	account *handler.AccountHandler,
	agent *handler.AgentHandler,
	billing *handler.BillingHandler,
	notifications *handler.NotificationHandler,
	audit *handler.AuditHandler,
	topology *handler.TopologyHandler,
	topoAnalysis *handler.TopologyAnalysisHandler,
	vulnerability *handler.VulnerabilityHandler,
	suggestions *handler.ExperimentSuggestionHandler,
	resultAnalysis *handler.ResultAnalysisHandler,
	aiChat *handler.AIChatHandler,
	experiments *handler.ExperimentHandler,
	policies *handler.PolicyHandler,
	ws *handler.WebSocketHandler,
	rdb *redis.Client,
	authService *service.AuthService,
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

	authPublic := r.Group("/auth")
	{
		authPublic.POST("/register", auth.Register)
		authPublic.POST("/login", auth.Login)
		authPublic.POST("/refresh", auth.Refresh)
		authPublic.POST("/forgot-password", auth.ForgotPassword)
		authPublic.POST("/reset-password", auth.ResetPassword)
		authPublic.POST("/verify-email", auth.VerifyEmail)
		authPublic.POST("/resend-verification", auth.ResendVerification)
		authPublic.GET("/invitations/lookup", invitations.LookupByToken)
		authPublic.POST("/invitations/accept-by-token", invitations.AcceptByToken)
		authPublic.GET("/oauth/:provider/authorize", oauth.Authorize)
		authPublic.POST("/oauth/callback", oauth.Callback)
	}

	agentPublic := r.Group("/agent")
	{
		agentPublic.POST("/register", agent.Register)
		agentPublic.POST("/heartbeat", agent.Heartbeat)
		agentPublic.POST("/topology", topology.Submit)
		agentPublic.POST("/topology/dependencies", topoAnalysis.SubmitDependencies)
		agentPublic.POST("/topology/metrics", topoAnalysis.SubmitMetrics)
		agentPublic.POST("/topology/drifts", topoAnalysis.SubmitDrift)
	}

	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/stripe", billing.WebhookStripe)
		webhooks.POST("/toss", billing.WebhookToss)
		webhooks.POST("/dodo", billing.WebhookDodo)
	}

	authProtected := r.Group("/auth")
	authProtected.Use(middleware.JWT(authService))
	{
		authProtected.GET("/me", auth.Me)
		authProtected.POST("/quick-setup", onboarding.QuickSetup)
		authProtected.POST("/accept-tos", account.AcceptTOS)
		authProtected.GET("/account/export", account.Export)
		authProtected.DELETE("/account", account.Delete)
		authProtected.POST("/account/cancel-deletion", account.CancelDeletion)
		authProtected.Use(middleware.CSRFSameSite(authService))
		authProtected.POST("/logout", auth.Logout)
	}

	saas := r.Group("/api/v1")
	saas.Use(middleware.JWT(authService), middleware.TenantContext(pool.App))
	saas.Use(middleware.AuditLog(pool.App))
	if rdb != nil {
		saas.Use(middleware.RateLimit(rdb, pool.App, middleware.DefaultRateLimiterConfig()))
	}
	{
		saas.GET("/hierarchy", hierarchy.List)
		saas.GET("/onboarding", onboarding.Get)
		saas.PATCH("/onboarding", onboarding.Update)
		saas.POST("/onboarding/skip", onboarding.Skip)
		saas.POST("/onboarding/complete", onboarding.Complete)
		saas.GET("/invitations", invitations.List)
		saas.POST("/invitations/accept", invitations.Accept)
		saas.POST("/invitations/decline", invitations.Decline)
		saas.GET("/api-keys", apiKeys.List)
		saas.GET("/billing", billing.GetStatus)
		saas.GET("/topology/latest", topology.Latest)
		saas.GET("/topology/history", topology.List)
		saas.GET("/topology/dependencies", topoAnalysis.GetDependencyMap)
		saas.GET("/topology/drifts", topoAnalysis.GetDrifts)
		saas.GET("/topology/metrics", topoAnalysis.GetMetrics)
		saas.GET("/vulnerabilities", vulnerability.List)
		saas.GET("/suggestions", suggestions.List)
		saas.GET("/result-analysis", resultAnalysis.List)
		saas.GET("/result-analysis/:id", resultAnalysis.Get)
		saas.GET("/ai/chat/sessions", aiChat.ListSessions)
		saas.POST("/ai/chat/sessions", aiChat.CreateSession)
		saas.GET("/ai/chat/sessions/:id/messages", aiChat.GetMessages)
		saas.POST("/ai/chat/sessions/:id/messages", aiChat.SendMessage)
		saas.DELETE("/ai/chat/sessions/:id", aiChat.DeleteSession)

		manage := saas.Group("")
		manage.Use(middleware.RequireTenantRole(pool.App, "admin", "editor"))
		{
			manage.POST("/organizations", hierarchy.CreateOrganization)
			manage.PATCH("/organizations/:id", hierarchy.UpdateOrganization)
			manage.POST("/workspaces", hierarchy.CreateWorkspace)
			manage.PATCH("/workspaces/:id", hierarchy.UpdateWorkspace)
			manage.POST("/teams", hierarchy.CreateTeam)
			manage.PATCH("/teams/:id", hierarchy.UpdateTeam)
			manage.POST("/projects", hierarchy.CreateProject)
			manage.PATCH("/projects/:id", hierarchy.UpdateProject)
			manage.POST("/environments", hierarchy.CreateEnvironment)
			manage.PATCH("/environments/:id", hierarchy.UpdateEnvironment)
			manage.POST("/agents/test-connection", onboarding.TestAgentConnection)
			manage.POST("/invitations", invitations.Create)
			manage.POST("/invitations/:id/resend", invitations.Resend)
			manage.DELETE("/invitations/:id", invitations.Revoke)
			manage.POST("/api-keys", apiKeys.Create)
			manage.POST("/api-keys/:id/rotate", apiKeys.Rotate)
			manage.DELETE("/api-keys/:id", apiKeys.Revoke)
			manage.GET("/agent-tokens", agent.ListTokens)
			manage.POST("/agent-tokens", agent.CreateToken)
			manage.DELETE("/agent-tokens/:id", agent.RevokeToken)
			manage.POST("/billing/upgrade", billing.Upgrade)
			manage.POST("/billing/cancel", billing.Cancel)
			manage.POST("/billing/reactivate", billing.Reactivate)
			manage.GET("/notification-channels", notifications.ListChannels)
			manage.POST("/notification-channels", notifications.CreateChannel)
			manage.DELETE("/notification-channels/:id", notifications.DeleteChannel)
			manage.GET("/notification-rules", notifications.ListRules)
			manage.POST("/notification-rules", notifications.CreateRule)
			manage.DELETE("/notification-rules/:id", notifications.DeleteRule)
			manage.GET("/audit-logs", audit.List)
			manage.POST("/topology/drifts/:id/acknowledge", topoAnalysis.AcknowledgeDrift)
			manage.PATCH("/vulnerabilities/:id", vulnerability.UpdateStatus)
			manage.POST("/vulnerabilities/scan", vulnerability.Scan)
			manage.POST("/suggestions/generate", suggestions.Generate)
			manage.DELETE("/suggestions/:id", suggestions.Delete)
			manage.POST("/result-analysis", resultAnalysis.Analyze)
		}
	}

	api := r.Group("/api/v1")
	api.Use(middleware.APIKey(pool.App))
	if rdb != nil {
		api.Use(middleware.RateLimit(rdb, pool.App, middleware.DefaultRateLimiterConfig()))
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
