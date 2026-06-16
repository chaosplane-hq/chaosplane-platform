package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

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
	RDB    *redis.Client
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
	enterprise *handler.EnterpriseHandler,
	gameday *handler.GameDayHandler,
	resilience *handler.ResilienceScoreHandler,
	wfTemplates *handler.WorkflowTemplateHandler,
	cui *handler.CUIHandler,
	marketplace *handler.MarketplaceHandler,
	federation *handler.FederationHandler,
	cicd *handler.CICDHandler,
	predictive *handler.PredictiveHandler,
	experiments *handler.ExperimentHandler,
	policies *handler.PolicyHandler,
	ws *handler.WebSocketHandler,
	agentWork *handler.AgentWorkHandler,
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

	agentV1 := r.Group("/agent/v1")
	agentV1.Use(middleware.AgentAuth(pool.App))
	{
		agentV1.GET("/work", agentWork.ClaimWork)
		agentV1.POST("/experiments/:id/status", agentWork.ReportStatus)
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
		saas.GET("/members", hierarchy.ListMembers)
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
		saas.GET("/sessions", enterprise.ListSessions)
		saas.DELETE("/sessions/:id", enterprise.RevokeSession)
		saas.POST("/sessions/revoke-all", enterprise.RevokeAllSessions)
		saas.GET("/mfa/recovery-codes/count", enterprise.GetRecoveryCodeCount)
		saas.GET("/audit-exports", enterprise.ListAuditExports)
		saas.GET("/gamedays", gameday.List)
		saas.GET("/gamedays/:id", gameday.Get)
		saas.GET("/resilience-score", resilience.Get)
		saas.GET("/workflow-templates", wfTemplates.List)
		saas.GET("/cui-markings", cui.List)
		saas.GET("/marketplace", marketplace.List)
		saas.GET("/federation/clusters", federation.List)
		saas.GET("/cicd-integrations", cicd.List)
		saas.GET("/predictions", predictive.List)

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
			manage.GET("/saml-providers", enterprise.ListSAMLProviders)
			manage.POST("/saml-providers", enterprise.CreateSAMLProvider)
			manage.DELETE("/saml-providers/:id", enterprise.DeleteSAMLProvider)
			manage.GET("/abac-policies", enterprise.ListABACPolicies)
			manage.POST("/abac-policies", enterprise.CreateABACPolicy)
			manage.DELETE("/abac-policies/:id", enterprise.DeleteABACPolicy)
			manage.POST("/abac-policies/evaluate", enterprise.EvaluateABAC)
			manage.POST("/mfa/recovery-codes", enterprise.GenerateRecoveryCodes)
			manage.POST("/account/request-deletion", enterprise.RequestDeletion)
			manage.POST("/account/cancel-deletion", enterprise.CancelDeletion)
			manage.POST("/account/change-email", enterprise.RequestEmailChange)
			manage.POST("/account/confirm-email-change", enterprise.ConfirmEmailChange)
			manage.POST("/audit-exports", enterprise.CreateAuditExport)
			manage.POST("/gamedays", gameday.Create)
			manage.PATCH("/gamedays/:id/status", gameday.UpdateStatus)
			manage.POST("/gamedays/:id/events", gameday.AddEvent)
			manage.POST("/gamedays/:id/postmortem", gameday.CreatePostmortem)
			manage.POST("/resilience-score/calculate", resilience.Calculate)
			manage.POST("/workflow-templates", wfTemplates.Create)
			manage.DELETE("/workflow-templates/:id", wfTemplates.Delete)
			manage.POST("/cui-markings", cui.Apply)
			manage.DELETE("/cui-markings/:id", cui.Remove)
			manage.POST("/marketplace/install", marketplace.Install)
			manage.DELETE("/marketplace/:id", marketplace.Uninstall)
			manage.POST("/federation/clusters", federation.Register)
			manage.DELETE("/federation/clusters/:id", federation.Remove)
			manage.POST("/cicd-integrations", cicd.Create)
			manage.DELETE("/cicd-integrations/:id", cicd.Delete)
			manage.POST("/predictions/run", predictive.Run)
			manage.PATCH("/predictions/:id", predictive.UpdateStatus)
		}
	}

	api := r.Group("/api/v1")
	api.Use(middleware.EitherAuth(authService, pool.App), middleware.TenantContext(pool.App))
	api.Use(middleware.AuditLog(pool.App))
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
		api.POST("/policies", policies.Create)
		api.DELETE("/policies/:name", policies.Delete)
	}

	r.GET("/ws/experiments/:name", ws.ExperimentStatus)

	return &Server{
		Router: r,
		Config: cfg,
		Pool:   pool,
		RDB:    rdb,
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
	opts.DialTimeout = 2 * time.Second
	opts.ReadTimeout = 1 * time.Second
	opts.WriteTimeout = 1 * time.Second
	opts.ContextTimeoutEnabled = true
	opts.PoolSize = 20
	opts.MinIdleConns = 10
	opts.MaxIdleConns = 20
	opts.PoolTimeout = 1 * time.Second
	opts.ConnMaxIdleTime = 5 * time.Minute
	opts.ConnMaxLifetime = 30 * time.Minute
	opts.MaxRetries = 1
	opts.MinRetryBackoff = 8 * time.Millisecond
	opts.MaxRetryBackoff = 50 * time.Millisecond
	return redis.NewClient(opts)
}

// WarmRedis fills the idle connection pool before traffic arrives. go-redis
// opens connections lazily, so without this the first rate-limit checks pay the
// full TLS handshake cost (~seconds against ElastiCache TLS) and fail open.
func WarmRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	n := max(rdb.Options().MinIdleConns, 1)
	var g errgroup.Group
	for i := 0; i < n; i++ {
		g.Go(func() error {
			return rdb.Ping(ctx).Err()
		})
	}
	return g.Wait()
}
