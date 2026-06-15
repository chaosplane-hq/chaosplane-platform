package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type ABACService struct {
	pool *database.Pool
}

func NewABACService(pool *database.Pool) *ABACService {
	return &ABACService{pool: pool}
}

type ABACPolicy struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenantId"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Effect      string          `json:"effect"`
	Subjects    json.RawMessage `json:"subjects"`
	Resources   json.RawMessage `json:"resources"`
	Actions     json.RawMessage `json:"actions"`
	Conditions  json.RawMessage `json:"conditions"`
	Priority    int             `json:"priority"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type ABACPolicyListResponse struct {
	Items []ABACPolicy `json:"items"`
}

type CreateABACPolicyRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description *string         `json:"description,omitempty"`
	Effect      string          `json:"effect" binding:"required"`
	Subjects    json.RawMessage `json:"subjects" binding:"required"`
	Resources   json.RawMessage `json:"resources" binding:"required"`
	Actions     json.RawMessage `json:"actions" binding:"required"`
	Conditions  json.RawMessage `json:"conditions,omitempty"`
	Priority    int             `json:"priority,omitempty"`
}

type EvaluateABACRequest struct {
	Subject  map[string]string `json:"subject" binding:"required"`
	Resource string            `json:"resource" binding:"required"`
	Action   string            `json:"action" binding:"required"`
}

type EvaluateABACResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func (s *ABACService) List(ctx context.Context, actor ActorContext) (*ABACPolicyListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, tenant_id::text, name, description, effect, subjects, resources, actions, conditions, priority, enabled, created_at
		FROM abac_policies WHERE tenant_id = $1::uuid ORDER BY priority DESC, created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list abac policies: %w", err)
	}
	defer rows.Close()
	items := []ABACPolicy{}
	for rows.Next() {
		var p ABACPolicy
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Effect, &p.Subjects, &p.Resources, &p.Actions, &p.Conditions, &p.Priority, &p.Enabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &ABACPolicyListResponse{Items: items}, rows.Err()
}

func (s *ABACService) Create(ctx context.Context, actor ActorContext, req *CreateABACPolicyRequest) (*ABACPolicy, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	conditions := "{}"
	if len(req.Conditions) > 0 {
		conditions = string(req.Conditions)
	}
	var p ABACPolicy
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO abac_policies (tenant_id, name, description, effect, subjects, resources, actions, conditions, priority)
		VALUES ($1::uuid, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb, $9)
		RETURNING id::text, tenant_id::text, name, description, effect, subjects, resources, actions, conditions, priority, enabled, created_at
	`, actor.TenantID, req.Name, req.Description, req.Effect, string(req.Subjects), string(req.Resources), string(req.Actions), conditions, req.Priority).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Description, &p.Effect, &p.Subjects, &p.Resources, &p.Actions, &p.Conditions, &p.Priority, &p.Enabled, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create abac policy: %w", err)
	}
	return &p, nil
}

func (s *ABACService) Delete(ctx context.Context, actor ActorContext, policyID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `DELETE FROM abac_policies WHERE id = $1::uuid AND tenant_id = $2::uuid`, policyID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete abac policy: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *ABACService) Evaluate(ctx context.Context, actor ActorContext, req *EvaluateABACRequest) (*EvaluateABACResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT effect, subjects, resources, actions, conditions, priority
		FROM abac_policies
		WHERE tenant_id = $1::uuid AND enabled = true
		ORDER BY priority DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("load abac policies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var effect string
		var subjectsRaw, resourcesRaw, actionsRaw, conditionsRaw json.RawMessage
		var priority int
		if err := rows.Scan(&effect, &subjectsRaw, &resourcesRaw, &actionsRaw, &conditionsRaw, &priority); err != nil {
			continue
		}

		if !matchSubjects(subjectsRaw, req.Subject) {
			continue
		}
		if !matchResource(resourcesRaw, req.Resource) {
			continue
		}
		if !matchAction(actionsRaw, req.Action) {
			continue
		}

		if effect == "deny" {
			return &EvaluateABACResponse{Allowed: false, Reason: fmt.Sprintf("Denied by policy (priority %d)", priority)}, nil
		}
		return &EvaluateABACResponse{Allowed: true, Reason: fmt.Sprintf("Allowed by policy (priority %d)", priority)}, nil
	}

	return &EvaluateABACResponse{Allowed: false, Reason: "No matching policy found, default deny"}, nil
}

func matchSubjects(raw json.RawMessage, subject map[string]string) bool {
	var patterns map[string]string
	if json.Unmarshal(raw, &patterns) != nil {
		return true
	}
	if len(patterns) == 0 {
		return true
	}
	for k, pattern := range patterns {
		val, ok := subject[k]
		if !ok {
			return false
		}
		if pattern != "*" && pattern != val {
			return false
		}
	}
	return true
}

func matchResource(raw json.RawMessage, resource string) bool {
	var patterns map[string]string
	if json.Unmarshal(raw, &patterns) != nil {
		return true
	}
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if pattern == "*" || pattern == resource {
			return true
		}
	}
	return false
}

func matchAction(raw json.RawMessage, action string) bool {
	var actions []string
	if json.Unmarshal(raw, &actions) != nil {
		return true
	}
	if len(actions) == 0 {
		return true
	}
	for _, a := range actions {
		if a == "*" || a == action {
			return true
		}
	}
	return false
}
