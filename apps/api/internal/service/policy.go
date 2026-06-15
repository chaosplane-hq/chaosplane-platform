package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var ErrPolicyNotFound = errors.New("policy not found")

type PolicyService struct {
	pool *database.Pool
}

func NewPolicyService(pool *database.Pool) *PolicyService {
	return &PolicyService{pool: pool}
}

type PolicyResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       *string   `json:"description,omitempty"`
	Enforcement       string    `json:"enforcement"`
	MaxConcurrent     *int      `json:"maxConcurrent,omitempty"`
	MaxTargets        *int      `json:"maxTargets,omitempty"`
	AllowedNamespaces []string  `json:"allowedNamespaces"`
	BlockedNamespaces []string  `json:"blockedNamespaces"`
	CreatedAt         time.Time `json:"createdAt"`
}

type PolicyListResponse struct {
	Policies []PolicyResponse `json:"policies"`
	Total    int              `json:"total"`
}

type CreatePolicyRequest struct {
	Name              string   `json:"name" binding:"required"`
	Description       *string  `json:"description,omitempty"`
	Enforcement       string   `json:"enforcement" binding:"required,oneof=enforce audit disabled"`
	MaxConcurrent     *int     `json:"maxConcurrent,omitempty"`
	MaxTargets        *int     `json:"maxTargets,omitempty"`
	AllowedNamespaces []string `json:"allowedNamespaces,omitempty"`
	BlockedNamespaces []string `json:"blockedNamespaces,omitempty"`
}

func (s *PolicyService) List(ctx context.Context, actor ActorContext) (*PolicyListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, name, description, enforcement, max_concurrent, max_targets,
		       allowed_namespaces, blocked_namespaces, created_at
		FROM blast_radius_policies
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()

	items := []PolicyResponse{}
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &PolicyListResponse{Policies: items, Total: len(items)}, nil
}

func (s *PolicyService) Get(ctx context.Context, actor ActorContext, id string) (*PolicyResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrPolicyNotFound
	}
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text, name, description, enforcement, max_concurrent, max_targets,
		       allowed_namespaces, blocked_namespaces, created_at
		FROM blast_radius_policies
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
	`, id, actor.TenantID)
	p, err := scanPolicy(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPolicyNotFound
		}
		return nil, fmt.Errorf("get policy: %w", err)
	}
	return p, nil
}

func (s *PolicyService) Create(ctx context.Context, actor ActorContext, req *CreatePolicyRequest) (*PolicyResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	allowed := req.AllowedNamespaces
	if allowed == nil {
		allowed = []string{}
	}
	blocked := req.BlockedNamespaces
	if blocked == nil {
		blocked = []string{}
	}
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO blast_radius_policies
			(tenant_id, name, description, enforcement, max_concurrent, max_targets,
			 allowed_namespaces, blocked_namespaces, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9::uuid)
		RETURNING id::text, name, description, enforcement, max_concurrent, max_targets,
		          allowed_namespaces, blocked_namespaces, created_at
	`, actor.TenantID, strings.TrimSpace(req.Name), req.Description, req.Enforcement,
		req.MaxConcurrent, req.MaxTargets, allowed, blocked, actor.UserID)
	p, err := scanPolicy(row)
	if err != nil {
		return nil, fmt.Errorf("create policy: %w", err)
	}
	return p, nil
}

func (s *PolicyService) Delete(ctx context.Context, actor ActorContext, id string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	if _, err := uuid.Parse(id); err != nil {
		return ErrPolicyNotFound
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE blast_radius_policies SET deleted_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
	`, id, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrPolicyNotFound
	}
	return nil
}

func scanPolicy(row pgx.Row) (*PolicyResponse, error) {
	var p PolicyResponse
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Enforcement, &p.MaxConcurrent,
		&p.MaxTargets, &p.AllowedNamespaces, &p.BlockedNamespaces, &p.CreatedAt); err != nil {
		return nil, err
	}
	if p.AllowedNamespaces == nil {
		p.AllowedNamespaces = []string{}
	}
	if p.BlockedNamespaces == nil {
		p.BlockedNamespaces = []string{}
	}
	return &p, nil
}
