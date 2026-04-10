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

type WorkflowTemplateService struct {
	pool *database.Pool
}

func NewWorkflowTemplateService(pool *database.Pool) *WorkflowTemplateService {
	return &WorkflowTemplateService{pool: pool}
}

type WorkflowTemplate struct {
	ID          string          `json:"id"`
	TenantID    *string         `json:"tenantId,omitempty"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category"`
	IsPublic    bool            `json:"isPublic"`
	Spec        json.RawMessage `json:"spec"`
	CreatedBy   *string         `json:"createdBy,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type WorkflowTemplateListResponse struct {
	Items []WorkflowTemplate `json:"items"`
}

type CreateWorkflowTemplateRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category,omitempty"`
	Spec        json.RawMessage `json:"spec" binding:"required"`
}

func (s *WorkflowTemplateService) List(ctx context.Context, actor ActorContext) (*WorkflowTemplateListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, tenant_id::text, name, description, category, is_public, spec, created_by::text, created_at
		FROM workflow_templates
		WHERE tenant_id = $1::uuid OR is_public = true
		ORDER BY is_public DESC, created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list workflow templates: %w", err)
	}
	defer rows.Close()
	items := []WorkflowTemplate{}
	for rows.Next() {
		var t WorkflowTemplate
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Description, &t.Category, &t.IsPublic, &t.Spec, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return &WorkflowTemplateListResponse{Items: items}, rows.Err()
}

func (s *WorkflowTemplateService) Create(ctx context.Context, actor ActorContext, req *CreateWorkflowTemplateRequest) (*WorkflowTemplate, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	category := req.Category
	if category == "" {
		category = "custom"
	}
	var t WorkflowTemplate
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO workflow_templates (tenant_id, name, description, category, spec, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5::jsonb, $6::uuid)
		RETURNING id::text, tenant_id::text, name, description, category, is_public, spec, created_by::text, created_at
	`, actor.TenantID, req.Name, req.Description, category, string(req.Spec), actor.UserID).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Description, &t.Category, &t.IsPublic, &t.Spec, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create workflow template: %w", err)
	}
	return &t, nil
}

func (s *WorkflowTemplateService) Delete(ctx context.Context, actor ActorContext, templateID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `DELETE FROM workflow_templates WHERE id = $1::uuid AND tenant_id = $2::uuid`, templateID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete workflow template: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

var _ = errors.Is
var _ = pgx.ErrNoRows
