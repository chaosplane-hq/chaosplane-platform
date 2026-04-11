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

type CICDService struct {
	pool *database.Pool
}

func NewCICDService(pool *database.Pool) *CICDService {
	return &CICDService{pool: pool}
}

type CICDIntegration struct {
	ID            string          `json:"id"`
	Provider      string          `json:"provider"`
	Name          string          `json:"name"`
	Config        json.RawMessage `json:"config"`
	Enabled       bool            `json:"enabled"`
	LastTriggered *time.Time      `json:"lastTriggered,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type CICDListResponse struct {
	Items []CICDIntegration `json:"items"`
}

type CreateCICDRequest struct {
	Provider string          `json:"provider" binding:"required"`
	Name     string          `json:"name" binding:"required"`
	Config   json.RawMessage `json:"config" binding:"required"`
}

func (s *CICDService) List(ctx context.Context, actor ActorContext) (*CICDListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, provider, name, config, enabled, last_triggered, created_at
		FROM cicd_integrations WHERE tenant_id = $1::uuid ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list cicd integrations: %w", err)
	}
	defer rows.Close()
	items := []CICDIntegration{}
	for rows.Next() {
		var c CICDIntegration
		if err := rows.Scan(&c.ID, &c.Provider, &c.Name, &c.Config, &c.Enabled, &c.LastTriggered, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return &CICDListResponse{Items: items}, rows.Err()
}

func (s *CICDService) Create(ctx context.Context, actor ActorContext, req *CreateCICDRequest) (*CICDIntegration, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	var c CICDIntegration
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO cicd_integrations (tenant_id, provider, name, config)
		VALUES ($1::uuid, $2, $3, $4::jsonb)
		RETURNING id::text, provider, name, config, enabled, last_triggered, created_at
	`, actor.TenantID, req.Provider, req.Name, string(req.Config)).Scan(
		&c.ID, &c.Provider, &c.Name, &c.Config, &c.Enabled, &c.LastTriggered, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create cicd integration: %w", err)
	}
	return &c, nil
}

func (s *CICDService) Delete(ctx context.Context, actor ActorContext, integrationID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `DELETE FROM cicd_integrations WHERE id = $1::uuid AND tenant_id = $2::uuid`, integrationID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete cicd integration: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

var _ = errors.Is
var _ = pgx.ErrNoRows
