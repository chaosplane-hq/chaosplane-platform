package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type NotificationService struct {
	pool       *database.Pool
	httpClient *http.Client
	email      *EmailService
}

func NewNotificationService(pool *database.Pool, email *EmailService) *NotificationService {
	return &NotificationService{pool: pool, httpClient: &http.Client{Timeout: 10 * time.Second}, email: email}
}

type NotificationChannel struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenantId"`
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	Enabled   bool            `json:"enabled"`
	CreatedAt time.Time       `json:"createdAt"`
}

type NotificationRule struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenantId"`
	ChannelID string          `json:"channelId"`
	EventType string          `json:"eventType"`
	Filters   json.RawMessage `json:"filters"`
	Enabled   bool            `json:"enabled"`
	CreatedAt time.Time       `json:"createdAt"`
}

type ListChannelsResponse struct {
	Items []NotificationChannel `json:"items"`
}

type ListRulesResponse struct {
	Items []NotificationRule `json:"items"`
}

type CreateChannelRequest struct {
	Type   string          `json:"type" binding:"required"`
	Name   string          `json:"name" binding:"required,min=2"`
	Config json.RawMessage `json:"config" binding:"required"`
}

type CreateRuleRequest struct {
	ChannelID string          `json:"channelId" binding:"required,uuid"`
	EventType string          `json:"eventType" binding:"required"`
	Filters   json.RawMessage `json:"filters,omitempty"`
}

func (s *NotificationService) ListChannels(ctx context.Context, actor ActorContext) (*ListChannelsResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, tenant_id::text, type, name, config, enabled, created_at
		FROM notification_channels
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list notification channels: %w", err)
	}
	defer rows.Close()

	items := []NotificationChannel{}
	for rows.Next() {
		var ch NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.TenantID, &ch.Type, &ch.Name, &ch.Config, &ch.Enabled, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification channel: %w", err)
		}
		items = append(items, ch)
	}
	return &ListChannelsResponse{Items: items}, rows.Err()
}

func (s *NotificationService) CreateChannel(ctx context.Context, actor ActorContext, req *CreateChannelRequest) (*NotificationChannel, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	channelType := strings.ToLower(strings.TrimSpace(req.Type))
	var ch NotificationChannel
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO notification_channels (tenant_id, type, name, config)
		VALUES ($1::uuid, $2, $3, $4::jsonb)
		RETURNING id::text, tenant_id::text, type, name, config, enabled, created_at
	`, actor.TenantID, channelType, strings.TrimSpace(req.Name), string(req.Config)).Scan(
		&ch.ID, &ch.TenantID, &ch.Type, &ch.Name, &ch.Config, &ch.Enabled, &ch.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification channel: %w", err)
	}
	return &ch, nil
}

func (s *NotificationService) DeleteChannel(ctx context.Context, actor ActorContext, channelID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `
		DELETE FROM notification_channels
		WHERE id = $1::uuid AND tenant_id = $2::uuid
	`, channelID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete notification channel: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *NotificationService) ListRules(ctx context.Context, actor ActorContext) (*ListRulesResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, tenant_id::text, channel_id::text, event_type, filters, enabled, created_at
		FROM notification_rules
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list notification rules: %w", err)
	}
	defer rows.Close()

	items := []NotificationRule{}
	for rows.Next() {
		var rule NotificationRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.ChannelID, &rule.EventType, &rule.Filters, &rule.Enabled, &rule.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification rule: %w", err)
		}
		items = append(items, rule)
	}
	return &ListRulesResponse{Items: items}, rows.Err()
}

func (s *NotificationService) CreateRule(ctx context.Context, actor ActorContext, req *CreateRuleRequest) (*NotificationRule, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM notification_channels WHERE id = $1::uuid AND tenant_id = $2::uuid`, req.ChannelID, actor.TenantID); err != nil {
		return nil, err
	}

	filters := "{}"
	if len(req.Filters) > 0 {
		filters = string(req.Filters)
	}

	var rule NotificationRule
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO notification_rules (tenant_id, channel_id, event_type, filters)
		VALUES ($1::uuid, $2::uuid, $3, $4::jsonb)
		RETURNING id::text, tenant_id::text, channel_id::text, event_type, filters, enabled, created_at
	`, actor.TenantID, req.ChannelID, req.EventType, filters).Scan(
		&rule.ID, &rule.TenantID, &rule.ChannelID, &rule.EventType, &rule.Filters, &rule.Enabled, &rule.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification rule: %w", err)
	}
	return &rule, nil
}

func (s *NotificationService) DeleteRule(ctx context.Context, actor ActorContext, ruleID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `
		DELETE FROM notification_rules
		WHERE id = $1::uuid AND tenant_id = $2::uuid
	`, ruleID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete notification rule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *NotificationService) Dispatch(ctx context.Context, tenantID, eventType string, payload map[string]interface{}) error {
	rows, err := s.pool.App.Query(ctx, `
		SELECT nc.type, nc.config, nr.id::text
		FROM notification_rules nr
		JOIN notification_channels nc ON nc.id = nr.channel_id
		WHERE nr.tenant_id = $1::uuid AND nr.event_type = $2 AND nr.enabled = true AND nc.enabled = true
	`, tenantID, eventType)
	if err != nil {
		return fmt.Errorf("query notification rules for dispatch: %w", err)
	}
	defer rows.Close()

	payloadJSON, _ := json.Marshal(payload)

	for rows.Next() {
		var channelType string
		var configRaw json.RawMessage
		var ruleID string
		if err := rows.Scan(&channelType, &configRaw, &ruleID); err != nil {
			continue
		}

		var sendErr error
		switch channelType {
		case "slack":
			sendErr = s.sendSlack(ctx, configRaw, payloadJSON)
		case "webhook":
			sendErr = s.sendWebhook(ctx, configRaw, payloadJSON)
		case "email":
			sendErr = s.sendEmail(ctx, configRaw, payloadJSON)
		}

		status := "sent"
		var errMsg *string
		if sendErr != nil {
			status = "failed"
			msg := sendErr.Error()
			errMsg = &msg
		}

		_, _ = s.pool.App.Exec(ctx, `
			INSERT INTO notification_history (tenant_id, channel_id, rule_id, event_type, status, payload, error_message, sent_at)
			SELECT $1::uuid, nr.channel_id, $2::uuid, $3, $4, $5::jsonb, $6,
			       CASE WHEN $4 = 'sent' THEN now() ELSE NULL END
			FROM notification_rules nr WHERE nr.id = $2::uuid
		`, tenantID, ruleID, eventType, status, string(payloadJSON), errMsg)
	}
	return rows.Err()
}

func (s *NotificationService) sendSlack(ctx context.Context, configRaw json.RawMessage, payload []byte) error {
	var cfg struct {
		WebhookURL string `json:"webhookUrl"`
	}
	if err := json.Unmarshal(configRaw, &cfg); err != nil || cfg.WebhookURL == "" {
		return fmt.Errorf("invalid slack config")
	}

	body := map[string]string{"text": string(payload)}
	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}

func (s *NotificationService) sendWebhook(ctx context.Context, configRaw json.RawMessage, payload []byte) error {
	var cfg struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(configRaw, &cfg); err != nil || cfg.URL == "" {
		return fmt.Errorf("invalid webhook config")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if cfg.Secret != "" {
		mac := hmac.New(sha256.New, []byte(cfg.Secret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature-256", "sha256="+sig)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}

func (s *NotificationService) sendEmail(ctx context.Context, configRaw json.RawMessage, payload []byte) error {
	if s.email == nil {
		return nil
	}
	var cfg struct {
		To string `json:"to"`
	}
	if err := json.Unmarshal(configRaw, &cfg); err != nil || cfg.To == "" {
		return fmt.Errorf("invalid email notification config")
	}
	subject := "ChaosPlane Notification"
	body := fmt.Sprintf("<pre>%s</pre>", string(payload))
	return s.email.send(ctx, cfg.To, subject, body)
}

var _ = errors.Is
var _ = pgx.ErrNoRows
