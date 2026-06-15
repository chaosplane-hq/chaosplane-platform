package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var ErrAgentTokenInvalid = errors.New("invalid agent token")

type AgentService struct {
	pool *database.Pool
}

func NewAgentService(pool *database.Pool) *AgentService {
	return &AgentService{pool: pool}
}

type AgentToken struct {
	ID            string     `json:"id"`
	EnvironmentID string     `json:"environmentId"`
	Name          string     `json:"name"`
	Plaintext     string     `json:"plaintext,omitempty"`
	LastUsedAt    *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type ListAgentTokensResponse struct {
	Items []AgentToken `json:"items"`
}

type CreateAgentTokenRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required,uuid"`
	Name          string `json:"name" binding:"required,min=2"`
}

type AgentHeartbeatRequest struct {
	Token        string `json:"token" binding:"required"`
	AgentVersion string `json:"agentVersion,omitempty"`
	Status       string `json:"status,omitempty"`
}

type AgentHeartbeatResponse struct {
	EnvironmentID string `json:"environmentId"`
	Acknowledged  bool   `json:"acknowledged"`
}

type RegisterAgentRequest struct {
	Token        string `json:"token" binding:"required"`
	AgentVersion string `json:"agentVersion" binding:"required"`
	ClusterInfo  string `json:"clusterInfo,omitempty"`
}

type RegisterAgentResponse struct {
	EnvironmentID string `json:"environmentId"`
	TenantID      string `json:"tenantId"`
	Registered    bool   `json:"registered"`
}

func (s *AgentService) ListTokens(ctx context.Context, actor ActorContext, environmentID string) (*ListAgentTokensResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM environments WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, environmentID, actor.TenantID); err != nil {
		return nil, err
	}

	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, environment_id::text, name, last_used_at, created_at
		FROM agent_tokens
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, environmentID, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list agent tokens: %w", err)
	}
	defer rows.Close()

	items := []AgentToken{}
	for rows.Next() {
		var item AgentToken
		if err := rows.Scan(&item.ID, &item.EnvironmentID, &item.Name, &item.LastUsedAt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan agent token: %w", err)
		}
		items = append(items, item)
	}
	return &ListAgentTokensResponse{Items: items}, rows.Err()
}

func (s *AgentService) CreateToken(ctx context.Context, actor ActorContext, req *CreateAgentTokenRequest) (*AgentToken, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM environments WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, req.EnvironmentID, actor.TenantID); err != nil {
		return nil, err
	}

	plain, _, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate agent token: %w", err)
	}
	prefixed := "cpagent_" + plain

	var item AgentToken
	err = s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO agent_tokens (tenant_id, environment_id, token_hash, name)
		VALUES ($1::uuid, $2::uuid, $3, $4)
		RETURNING id::text, environment_id::text, name, last_used_at, created_at
	`, actor.TenantID, req.EnvironmentID, hashToken(prefixed), strings.TrimSpace(req.Name)).Scan(
		&item.ID, &item.EnvironmentID, &item.Name, &item.LastUsedAt, &item.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create agent token: %w", err)
	}
	item.Plaintext = prefixed
	return &item, nil
}

func (s *AgentService) RevokeToken(ctx context.Context, actor ActorContext, tokenID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE agent_tokens
		SET revoked_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND revoked_at IS NULL
	`, tokenID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("revoke agent token: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *AgentService) Register(ctx context.Context, req *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	envID, tenantID, err := s.validateAgentToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	clusterInfo := "{}"
	if strings.TrimSpace(req.ClusterInfo) != "" {
		clusterInfo = req.ClusterInfo
	}

	if _, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE environments
		SET agent_status = 'connected', agent_version = $2, cluster_info = $3::jsonb, last_heartbeat = now()
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, envID, req.AgentVersion, clusterInfo); err != nil {
		return nil, fmt.Errorf("register agent: %w", err)
	}

	return &RegisterAgentResponse{EnvironmentID: envID, TenantID: tenantID, Registered: true}, nil
}

func (s *AgentService) Heartbeat(ctx context.Context, req *AgentHeartbeatRequest) (*AgentHeartbeatResponse, error) {
	envID, _, err := s.validateAgentToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	status := "connected"
	if strings.EqualFold(req.Status, "degraded") {
		status = "degraded"
	}

	if _, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE environments
		SET agent_status = $2, last_heartbeat = now(),
		    agent_version = COALESCE(NULLIF($3, ''), agent_version)
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, envID, status, req.AgentVersion); err != nil {
		return nil, fmt.Errorf("update heartbeat: %w", err)
	}

	return &AgentHeartbeatResponse{EnvironmentID: envID, Acknowledged: true}, nil
}

func (s *AgentService) validateAgentToken(ctx context.Context, token string) (envID string, tenantID string, err error) {
	err = s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT at.environment_id::text, at.tenant_id::text
		FROM agent_tokens at
		WHERE at.token_hash = $1 AND at.revoked_at IS NULL
	`, hashToken(token)).Scan(&envID, &tenantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrAgentTokenInvalid
		}
		return "", "", fmt.Errorf("validate agent token: %w", err)
	}

	_, _ = s.pool.Conn(ctx).Exec(ctx, `UPDATE agent_tokens SET last_used_at = now() WHERE token_hash = $1`, hashToken(token))
	return envID, tenantID, nil
}
