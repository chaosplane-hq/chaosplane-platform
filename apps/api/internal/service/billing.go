package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidTransition    = errors.New("invalid subscription state transition")
	ErrUsageLimitExceeded   = errors.New("usage limit exceeded")
)

type BillingService struct {
	pool     *database.Pool
	cfg      *config.Config
	gateways map[string]PaymentGateway
}

type PaymentGateway interface {
	Name() string
	CreateCheckoutURL(ctx context.Context, tenantID, plan string) (string, error)
	CancelSubscription(ctx context.Context, gatewaySubID string) error
}

func NewBillingService(pool *database.Pool, cfg *config.Config) *BillingService {
	return &BillingService{
		pool:     pool,
		cfg:      cfg,
		gateways: make(map[string]PaymentGateway),
	}
}

func (s *BillingService) RegisterGateway(gw PaymentGateway) {
	s.gateways[strings.ToLower(gw.Name())] = gw
}

type Subscription struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenantId"`
	Plan               string     `json:"plan"`
	Status             string     `json:"status"`
	Gateway            string     `json:"gateway"`
	CurrentPeriodStart *time.Time `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"currentPeriodEnd,omitempty"`
	TrialEndsAt        *time.Time `json:"trialEndsAt,omitempty"`
	CancelledAt        *time.Time `json:"cancelledAt,omitempty"`
	SuspendedAt        *time.Time `json:"suspendedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type UsageStats struct {
	Experiments int64 `json:"experiments"`
	Agents      int64 `json:"agents"`
	APICalls    int64 `json:"apiCalls"`
}

type PlanLimits struct {
	MaxExperiments int64 `json:"maxExperiments"`
	MaxAgents      int64 `json:"maxAgents"`
	MaxAPICalls    int64 `json:"maxApiCalls"`
}

type BillingStatusResponse struct {
	Subscription *Subscription `json:"subscription"`
	Usage        *UsageStats   `json:"usage"`
	Limits       *PlanLimits   `json:"limits"`
}

type UpgradeRequest struct {
	Plan    string `json:"plan" binding:"required"`
	Gateway string `json:"gateway,omitempty"`
}

type UpgradeResponse struct {
	CheckoutURL  string        `json:"checkoutUrl,omitempty"`
	Subscription *Subscription `json:"subscription,omitempty"`
}

func (s *BillingService) GetStatus(ctx context.Context, actor ActorContext) (*BillingStatusResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	sub, err := s.getOrCreateSubscription(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}

	usage, err := s.getCurrentUsage(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}

	limits := planLimitsFor(sub.Plan)

	return &BillingStatusResponse{Subscription: sub, Usage: usage, Limits: limits}, nil
}

func (s *BillingService) Upgrade(ctx context.Context, actor ActorContext, req *UpgradeRequest) (*UpgradeResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	gateway := strings.ToLower(strings.TrimSpace(req.Gateway))
	if gateway == "" {
		gateway = "stripe"
	}

	gw, ok := s.gateways[gateway]
	if ok {
		url, err := gw.CreateCheckoutURL(ctx, actor.TenantID, req.Plan)
		if err != nil {
			return nil, fmt.Errorf("create checkout url: %w", err)
		}
		return &UpgradeResponse{CheckoutURL: url}, nil
	}

	sub, err := s.transitionPlan(ctx, actor.TenantID, req.Plan, gateway)
	if err != nil {
		return nil, err
	}
	return &UpgradeResponse{Subscription: sub}, nil
}

func (s *BillingService) Cancel(ctx context.Context, actor ActorContext) (*Subscription, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	row := s.pool.App.QueryRow(ctx, `
		UPDATE subscriptions
		SET status = 'cancelled', cancelled_at = now()
		WHERE tenant_id = $1::uuid AND status IN ('active','trialing','past_due')
		RETURNING id::text, tenant_id::text, plan, status, gateway,
		          current_period_start, current_period_end, trial_ends_at,
		          cancelled_at, suspended_at, created_at
	`, actor.TenantID)
	return scanSubscription(row)
}

func (s *BillingService) Reactivate(ctx context.Context, actor ActorContext) (*Subscription, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	row := s.pool.App.QueryRow(ctx, `
		UPDATE subscriptions
		SET status = 'active', cancelled_at = NULL, suspended_at = NULL
		WHERE tenant_id = $1::uuid AND status = 'cancelled'
		  AND cancelled_at > now() - interval '30 days'
		RETURNING id::text, tenant_id::text, plan, status, gateway,
		          current_period_start, current_period_end, trial_ends_at,
		          cancelled_at, suspended_at, created_at
	`, actor.TenantID)
	return scanSubscription(row)
}

func (s *BillingService) RecordUsage(ctx context.Context, tenantID, metric string, quantity int64) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	_, err := s.pool.App.Exec(ctx, `
		INSERT INTO usage_records (tenant_id, metric, quantity, period_start, period_end)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, metric, period_start)
		DO UPDATE SET quantity = usage_records.quantity + $3
	`, tenantID, metric, quantity, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("record usage: %w", err)
	}
	return nil
}

func (s *BillingService) CheckUsageLimit(ctx context.Context, tenantID, metric string) error {
	sub, err := s.getOrCreateSubscription(ctx, tenantID)
	if err != nil {
		return err
	}
	limits := planLimitsFor(sub.Plan)
	usage, err := s.getCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}

	var current, limit int64
	switch metric {
	case "experiments":
		current, limit = usage.Experiments, limits.MaxExperiments
	case "agents":
		current, limit = usage.Agents, limits.MaxAgents
	case "api_calls":
		current, limit = usage.APICalls, limits.MaxAPICalls
	}

	if limit > 0 && current >= limit {
		return ErrUsageLimitExceeded
	}
	return nil
}

func (s *BillingService) ProcessWebhookEvent(ctx context.Context, gateway, eventType string, payload []byte) error {
	_, err := s.pool.App.Exec(ctx, `
		INSERT INTO billing_events (tenant_id, event_type, gateway, payload)
		VALUES ('00000000-0000-0000-0000-000000000000'::uuid, $1, $2, $3::jsonb)
	`, eventType, gateway, string(payload))
	if err != nil {
		return fmt.Errorf("store billing event: %w", err)
	}
	return nil
}

func (s *BillingService) getOrCreateSubscription(ctx context.Context, tenantID string) (*Subscription, error) {
	row := s.pool.App.QueryRow(ctx, `
		SELECT id::text, tenant_id::text, plan, status, gateway,
		       current_period_start, current_period_end, trial_ends_at,
		       cancelled_at, suspended_at, created_at
		FROM subscriptions
		WHERE tenant_id = $1::uuid
	`, tenantID)

	sub, err := scanSubscription(row)
	if err == nil {
		return sub, nil
	}
	if !errors.Is(err, ErrSubscriptionNotFound) {
		return nil, err
	}

	trialEnd := time.Now().Add(14 * 24 * time.Hour)
	row = s.pool.App.QueryRow(ctx, `
		INSERT INTO subscriptions (tenant_id, plan, status, trial_ends_at)
		VALUES ($1::uuid, 'free', 'trialing', $2)
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = subscriptions.tenant_id
		RETURNING id::text, tenant_id::text, plan, status, gateway,
		          current_period_start, current_period_end, trial_ends_at,
		          cancelled_at, suspended_at, created_at
	`, tenantID, trialEnd)
	return scanSubscription(row)
}

func (s *BillingService) transitionPlan(ctx context.Context, tenantID, plan, gateway string) (*Subscription, error) {
	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)

	row := s.pool.App.QueryRow(ctx, `
		UPDATE subscriptions
		SET plan = $2, status = 'active', gateway = $3,
		    current_period_start = $4, current_period_end = $5,
		    cancelled_at = NULL, suspended_at = NULL
		WHERE tenant_id = $1::uuid
		RETURNING id::text, tenant_id::text, plan, status, gateway,
		          current_period_start, current_period_end, trial_ends_at,
		          cancelled_at, suspended_at, created_at
	`, tenantID, plan, gateway, now, periodEnd)
	sub, err := scanSubscription(row)
	if err != nil {
		return nil, err
	}

	if _, err := s.pool.App.Exec(ctx, `UPDATE tenants SET plan = $2 WHERE id = $1::uuid`, tenantID, plan); err != nil {
		return nil, fmt.Errorf("sync tenant plan: %w", err)
	}

	return sub, nil
}

func (s *BillingService) getCurrentUsage(ctx context.Context, tenantID string) (*UsageStats, error) {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := s.pool.App.Query(ctx, `
		SELECT metric, quantity
		FROM usage_records
		WHERE tenant_id = $1::uuid AND period_start = $2
	`, tenantID, periodStart)
	if err != nil {
		return nil, fmt.Errorf("get current usage: %w", err)
	}
	defer rows.Close()

	stats := &UsageStats{}
	for rows.Next() {
		var metric string
		var quantity int64
		if err := rows.Scan(&metric, &quantity); err != nil {
			return nil, fmt.Errorf("scan usage record: %w", err)
		}
		switch metric {
		case "experiments":
			stats.Experiments = quantity
		case "agents":
			stats.Agents = quantity
		case "api_calls":
			stats.APICalls = quantity
		}
	}
	return stats, rows.Err()
}

func scanSubscription(row pgx.Row) (*Subscription, error) {
	var sub Subscription
	if err := row.Scan(
		&sub.ID, &sub.TenantID, &sub.Plan, &sub.Status, &sub.Gateway,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.TrialEndsAt,
		&sub.CancelledAt, &sub.SuspendedAt, &sub.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("scan subscription: %w", err)
	}
	return &sub, nil
}

func planLimitsFor(plan string) *PlanLimits {
	switch strings.ToLower(plan) {
	case "free":
		return &PlanLimits{MaxExperiments: 50, MaxAgents: 1, MaxAPICalls: 10000}
	case "team":
		return &PlanLimits{MaxExperiments: 500, MaxAgents: 5, MaxAPICalls: 100000}
	case "business":
		return &PlanLimits{MaxExperiments: 5000, MaxAgents: 25, MaxAPICalls: 1000000}
	case "enterprise", "government":
		return &PlanLimits{MaxExperiments: -1, MaxAgents: -1, MaxAPICalls: -1}
	default:
		return &PlanLimits{MaxExperiments: 50, MaxAgents: 1, MaxAPICalls: 10000}
	}
}
