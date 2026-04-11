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

type FederationService struct {
	pool *database.Pool
}

func NewFederationService(pool *database.Pool) *FederationService {
	return &FederationService{pool: pool}
}

type FederationCluster struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Region      string          `json:"region"`
	Provider    string          `json:"provider"`
	APIEndpoint string          `json:"apiEndpoint"`
	Status      string          `json:"status"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type FederationListResponse struct {
	Items []FederationCluster `json:"items"`
}

type RegisterClusterRequest struct {
	Name        string          `json:"name" binding:"required"`
	Region      string          `json:"region" binding:"required"`
	Provider    string          `json:"provider" binding:"required"`
	APIEndpoint string          `json:"apiEndpoint" binding:"required"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

func (s *FederationService) List(ctx context.Context, actor ActorContext) (*FederationListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, name, region, provider, api_endpoint, status, metadata, created_at
		FROM federation_clusters WHERE tenant_id = $1::uuid ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list federation clusters: %w", err)
	}
	defer rows.Close()
	items := []FederationCluster{}
	for rows.Next() {
		var c FederationCluster
		if err := rows.Scan(&c.ID, &c.Name, &c.Region, &c.Provider, &c.APIEndpoint, &c.Status, &c.Metadata, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return &FederationListResponse{Items: items}, rows.Err()
}

func (s *FederationService) Register(ctx context.Context, actor ActorContext, req *RegisterClusterRequest) (*FederationCluster, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	metadata := "{}"
	if len(req.Metadata) > 0 {
		metadata = string(req.Metadata)
	}
	var c FederationCluster
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO federation_clusters (tenant_id, name, region, provider, api_endpoint, metadata)
		VALUES ($1::uuid, $2, $3, $4, $5, $6::jsonb)
		RETURNING id::text, name, region, provider, api_endpoint, status, metadata, created_at
	`, actor.TenantID, req.Name, req.Region, req.Provider, req.APIEndpoint, metadata).Scan(
		&c.ID, &c.Name, &c.Region, &c.Provider, &c.APIEndpoint, &c.Status, &c.Metadata, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("register federation cluster: %w", err)
	}
	return &c, nil
}

func (s *FederationService) Remove(ctx context.Context, actor ActorContext, clusterID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `DELETE FROM federation_clusters WHERE id = $1::uuid AND tenant_id = $2::uuid`, clusterID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("remove federation cluster: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

var _ = errors.Is
var _ = pgx.ErrNoRows
