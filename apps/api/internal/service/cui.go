package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type CUIService struct {
	pool *database.Pool
}

func NewCUIService(pool *database.Pool) *CUIService {
	return &CUIService{pool: pool}
}

type CUIMarking struct {
	ID           string    `json:"id"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	CUICategory  string    `json:"cuiCategory"`
	AppliedBy    *string   `json:"appliedBy,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CUIMarkingListResponse struct {
	Items []CUIMarking `json:"items"`
}

type ApplyCUIMarkingRequest struct {
	ResourceType string `json:"resourceType" binding:"required"`
	ResourceID   string `json:"resourceId" binding:"required"`
	CUICategory  string `json:"cuiCategory" binding:"required"`
}

func (s *CUIService) List(ctx context.Context, actor ActorContext) (*CUIMarkingListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, resource_type, resource_id, cui_category, marking_applied_by::text, created_at
		FROM cui_markings WHERE tenant_id = $1::uuid ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list cui markings: %w", err)
	}
	defer rows.Close()
	items := []CUIMarking{}
	for rows.Next() {
		var m CUIMarking
		if err := rows.Scan(&m.ID, &m.ResourceType, &m.ResourceID, &m.CUICategory, &m.AppliedBy, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return &CUIMarkingListResponse{Items: items}, rows.Err()
}

func (s *CUIService) Apply(ctx context.Context, actor ActorContext, req *ApplyCUIMarkingRequest) (*CUIMarking, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	var m CUIMarking
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO cui_markings (tenant_id, resource_type, resource_id, cui_category, marking_applied_by)
		VALUES ($1::uuid, $2, $3, $4, $5::uuid)
		ON CONFLICT (tenant_id, resource_type, resource_id)
		DO UPDATE SET cui_category = $4, marking_applied_by = $5::uuid
		RETURNING id::text, resource_type, resource_id, cui_category, marking_applied_by::text, created_at
	`, actor.TenantID, req.ResourceType, req.ResourceID, req.CUICategory, actor.UserID).Scan(
		&m.ID, &m.ResourceType, &m.ResourceID, &m.CUICategory, &m.AppliedBy, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("apply cui marking: %w", err)
	}
	return &m, nil
}

func (s *CUIService) Remove(ctx context.Context, actor ActorContext, markingID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `DELETE FROM cui_markings WHERE id = $1::uuid AND tenant_id = $2::uuid`, markingID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("remove cui marking: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}
