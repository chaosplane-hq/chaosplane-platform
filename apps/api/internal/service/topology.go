package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type TopologyService struct {
	pool *database.Pool
}

func NewTopologyService(pool *database.Pool) *TopologyService {
	return &TopologyService{pool: pool}
}

type TopologySnapshot struct {
	ID            string          `json:"id"`
	EnvironmentID string          `json:"environmentId"`
	Nodes         json.RawMessage `json:"nodes"`
	Namespaces    json.RawMessage `json:"namespaces"`
	Services      json.RawMessage `json:"services"`
	Deployments   json.RawMessage `json:"deployments"`
	Pods          json.RawMessage `json:"pods"`
	CollectedAt   time.Time       `json:"collectedAt"`
}

type SubmitTopologyRequest struct {
	EnvironmentID string          `json:"environmentId" binding:"required,uuid"`
	Nodes         json.RawMessage `json:"nodes"`
	Namespaces    json.RawMessage `json:"namespaces"`
	Services      json.RawMessage `json:"services"`
	Deployments   json.RawMessage `json:"deployments"`
	Pods          json.RawMessage `json:"pods"`
}

type ListTopologyResponse struct {
	Items []TopologySnapshot `json:"items"`
}

func (s *TopologyService) Latest(ctx context.Context, actor ActorContext, environmentID string) (*TopologySnapshot, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM environments WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, environmentID, actor.TenantID); err != nil {
		return nil, err
	}

	var snap TopologySnapshot
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text, environment_id::text, nodes, namespaces, services, deployments, pods, collected_at
		FROM topology_snapshots
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
		ORDER BY collected_at DESC
		LIMIT 1
	`, environmentID, actor.TenantID).Scan(
		&snap.ID, &snap.EnvironmentID, &snap.Nodes, &snap.Namespaces,
		&snap.Services, &snap.Deployments, &snap.Pods, &snap.CollectedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, fmt.Errorf("get latest topology: %w", err)
	}
	return &snap, nil
}

func (s *TopologyService) List(ctx context.Context, actor ActorContext, environmentID string, limit int) (*ListTopologyResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, environment_id::text, nodes, namespaces, services, deployments, pods, collected_at
		FROM topology_snapshots
		WHERE environment_id = $1::uuid AND tenant_id = $2::uuid
		ORDER BY collected_at DESC
		LIMIT $3
	`, environmentID, actor.TenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list topology snapshots: %w", err)
	}
	defer rows.Close()

	items := []TopologySnapshot{}
	for rows.Next() {
		var snap TopologySnapshot
		if err := rows.Scan(&snap.ID, &snap.EnvironmentID, &snap.Nodes, &snap.Namespaces, &snap.Services, &snap.Deployments, &snap.Pods, &snap.CollectedAt); err != nil {
			return nil, fmt.Errorf("scan topology snapshot: %w", err)
		}
		items = append(items, snap)
	}
	return &ListTopologyResponse{Items: items}, rows.Err()
}

func (s *TopologyService) TenantIDFromEnvironment(ctx context.Context, environmentID string) (string, error) {
	var tenantID string
	err := s.pool.Conn(ctx).QueryRow(ctx, `SELECT tenant_id::text FROM environments WHERE id = $1::uuid AND deleted_at IS NULL`, environmentID).Scan(&tenantID)
	if err != nil {
		return "", fmt.Errorf("resolve tenant from environment: %w", err)
	}
	return tenantID, nil
}

func (s *TopologyService) Submit(ctx context.Context, tenantID string, req *SubmitTopologyRequest) (*TopologySnapshot, error) {
	defaultJSON := func(raw json.RawMessage) string {
		if len(raw) == 0 {
			return "[]"
		}
		return string(raw)
	}

	var snap TopologySnapshot
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO topology_snapshots (tenant_id, environment_id, nodes, namespaces, services, deployments, pods)
		VALUES ($1::uuid, $2::uuid, $3::jsonb, $4::jsonb, $5::jsonb, $6::jsonb, $7::jsonb)
		RETURNING id::text, environment_id::text, nodes, namespaces, services, deployments, pods, collected_at
	`, tenantID, req.EnvironmentID,
		defaultJSON(req.Nodes), defaultJSON(req.Namespaces), defaultJSON(req.Services),
		defaultJSON(req.Deployments), defaultJSON(req.Pods),
	).Scan(&snap.ID, &snap.EnvironmentID, &snap.Nodes, &snap.Namespaces, &snap.Services, &snap.Deployments, &snap.Pods, &snap.CollectedAt)
	if err != nil {
		return nil, fmt.Errorf("submit topology snapshot: %w", err)
	}
	return &snap, nil
}
