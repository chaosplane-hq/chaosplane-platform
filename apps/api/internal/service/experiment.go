package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var ErrExperimentNotFound = errors.New("experiment not found")

type ExperimentService struct {
	pool *database.Pool
}

func NewExperimentService(pool *database.Pool) *ExperimentService {
	return &ExperimentService{pool: pool}
}

type CreateExperimentRequest struct {
	Name      string        `json:"name" binding:"required"`
	Namespace string        `json:"namespace"`
	Action    ActionRequest `json:"action" binding:"required"`
	Target    TargetRequest `json:"target" binding:"required"`
	Duration  string        `json:"duration"`
}

type ActionRequest struct {
	Type       string          `json:"type" binding:"required"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

type TargetRequest struct {
	Kind          string            `json:"kind,omitempty"`
	Namespace     string            `json:"namespace,omitempty"`
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Mode          string            `json:"mode,omitempty"`
	Value         string            `json:"value,omitempty"`
	Names         []string          `json:"names,omitempty"`
}

type ExperimentAction struct {
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

type ExperimentTarget struct {
	Namespace     string            `json:"namespace"`
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Mode          string            `json:"mode,omitempty"`
	Value         string            `json:"value,omitempty"`
}

type ExperimentStatus struct {
	Phase          string  `json:"phase"`
	StartTime      *string `json:"startTime,omitempty"`
	CompletionTime *string `json:"completionTime,omitempty"`
	Message        string  `json:"message,omitempty"`
}

type ExperimentResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
	Action    ExperimentAction `json:"action"`
	Target    ExperimentTarget `json:"target"`
	Status    ExperimentStatus `json:"status"`
	Duration  string           `json:"duration,omitempty"`
	CreatedAt string           `json:"createdAt,omitempty"`
}

type ExperimentListResponse struct {
	Experiments []ExperimentResponse `json:"experiments"`
	Total       int                  `json:"total"`
	Limit       int                  `json:"limit"`
	Offset      int                  `json:"offset"`
}

type experimentRow struct {
	id        string
	name      string
	status    string
	target    []byte
	action    []byte
	duration  int
	createdAt time.Time
	startedAt *time.Time
	endedAt   *time.Time
}

func (s *ExperimentService) List(ctx context.Context, actor ActorContext, statusFilter, actionFilter string, limit, offset int) (*ExperimentListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	var total int
	if err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT count(*) FROM experiments WHERE tenant_id = $1::uuid AND deleted_at IS NULL
	`, actor.TenantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count experiments: %w", err)
	}

	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, name, status, target, action, duration_seconds, created_at, run_started_at, run_ended_at
		FROM experiments
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, actor.TenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list experiments: %w", err)
	}
	defer rows.Close()

	items := []ExperimentResponse{}
	for rows.Next() {
		var r experimentRow
		if err := rows.Scan(&r.id, &r.name, &r.status, &r.target, &r.action, &r.duration, &r.createdAt, &r.startedAt, &r.endedAt); err != nil {
			return nil, fmt.Errorf("scan experiment: %w", err)
		}
		resp := toExperimentResponse(&r)
		if statusFilter != "" && !strings.EqualFold(resp.Status.Phase, statusFilter) {
			continue
		}
		if actionFilter != "" && resp.Action.Type != actionFilter {
			continue
		}
		items = append(items, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ExperimentListResponse{
		Experiments: items,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
	}, nil
}

func (s *ExperimentService) Get(ctx context.Context, actor ActorContext, id string) (*ExperimentResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	r, err := s.fetchRow(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	resp := toExperimentResponse(r)
	return &resp, nil
}

func (s *ExperimentService) Create(ctx context.Context, actor ActorContext, req *CreateExperimentRequest) (*ExperimentResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	envID, err := s.resolveDefaultEnvironment(ctx, actor)
	if err != nil {
		return nil, err
	}

	targetSpec := ExperimentTarget{
		Namespace:     req.Target.Namespace,
		LabelSelector: req.Target.LabelSelector,
		Mode:          req.Target.Mode,
		Value:         req.Target.Value,
	}
	targetJSON, err := json.Marshal(targetSpec)
	if err != nil {
		return nil, fmt.Errorf("marshal target: %w", err)
	}

	actionParams := req.Action.Parameters
	if len(actionParams) == 0 {
		actionParams = json.RawMessage(`{}`)
	}
	actionSpec := ExperimentAction{Type: req.Action.Type, Parameters: actionParams}
	actionJSON, err := json.Marshal(actionSpec)
	if err != nil {
		return nil, fmt.Errorf("marshal action: %w", err)
	}

	durationSeconds := 60
	if req.Duration != "" {
		if d, derr := time.ParseDuration(req.Duration); derr == nil && d > 0 {
			durationSeconds = int(d.Seconds())
		}
	}

	var r experimentRow
	err = s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO experiments (tenant_id, environment_id, name, target, action, duration_seconds, status, created_by)
		VALUES ($1::uuid, $2::uuid, $3, $4::jsonb, $5::jsonb, $6, 'scheduled', $7::uuid)
		RETURNING id::text, name, status, target, action, duration_seconds, created_at, run_started_at, run_ended_at
	`, actor.TenantID, envID, req.Name, targetJSON, actionJSON, durationSeconds, actor.UserID).Scan(
		&r.id, &r.name, &r.status, &r.target, &r.action, &r.duration, &r.createdAt, &r.startedAt, &r.endedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create experiment: %w", err)
	}

	resp := toExperimentResponse(&r)
	return &resp, nil
}

func (s *ExperimentService) Delete(ctx context.Context, actor ActorContext, id string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	if _, err := uuid.Parse(id); err != nil {
		return ErrExperimentNotFound
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE experiments SET deleted_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
	`, id, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete experiment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrExperimentNotFound
	}
	return nil
}

func (s *ExperimentService) Abort(ctx context.Context, actor ActorContext, id string) (*ExperimentResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrExperimentNotFound
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE experiments
		SET desired_state = 'abort', generation = generation + 1
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
		  AND status IN ('scheduled','running','paused')
	`, id, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("abort experiment: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, ErrExperimentNotFound
	}
	return s.Get(ctx, actor, id)
}

func (s *ExperimentService) fetchRow(ctx context.Context, actor ActorContext, id string) (*experimentRow, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrExperimentNotFound
	}
	var r experimentRow
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text, name, status, target, action, duration_seconds, created_at, run_started_at, run_ended_at
		FROM experiments
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
	`, id, actor.TenantID).Scan(&r.id, &r.name, &r.status, &r.target, &r.action, &r.duration, &r.createdAt, &r.startedAt, &r.endedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExperimentNotFound
		}
		return nil, fmt.Errorf("get experiment: %w", err)
	}
	return &r, nil
}

// resolveDefaultEnvironment creates a default project and environment on demand
// when the tenant has none, since registration does not provision one.
func (s *ExperimentService) resolveDefaultEnvironment(ctx context.Context, actor ActorContext) (string, error) {
	var envID string
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text FROM environments
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT 1
	`, actor.TenantID).Scan(&envID)
	if err == nil {
		return envID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("lookup default environment: %w", err)
	}

	var workspaceID string
	if err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text FROM workspaces
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT 1
	`, actor.TenantID).Scan(&workspaceID); err != nil {
		return "", fmt.Errorf("lookup default workspace: %w", err)
	}

	var projectID string
	if err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO projects (tenant_id, workspace_id, name, slug)
		VALUES ($1::uuid, $2::uuid, 'Default', 'default')
		ON CONFLICT (workspace_id, slug) DO UPDATE SET updated_at = now()
		RETURNING id::text
	`, actor.TenantID, workspaceID).Scan(&projectID); err != nil {
		return "", fmt.Errorf("create default project: %w", err)
	}

	if err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO environments (tenant_id, project_id, name, slug, type)
		VALUES ($1::uuid, $2::uuid, 'Default', 'default', 'staging')
		ON CONFLICT (project_id, slug) DO UPDATE SET updated_at = now()
		RETURNING id::text
	`, actor.TenantID, projectID).Scan(&envID); err != nil {
		return "", fmt.Errorf("create default environment: %w", err)
	}
	return envID, nil
}

func toExperimentResponse(r *experimentRow) ExperimentResponse {
	var target ExperimentTarget
	if len(r.target) > 0 {
		_ = json.Unmarshal(r.target, &target)
	}
	var action ExperimentAction
	if len(r.action) > 0 {
		_ = json.Unmarshal(r.action, &action)
	}

	resp := ExperimentResponse{
		ID:        r.id,
		Name:      r.name,
		Namespace: target.Namespace,
		Action:    action,
		Target:    target,
		Status:    ExperimentStatus{Phase: phaseFromStatus(r.status)},
		Duration:  fmt.Sprintf("%ds", r.duration),
		CreatedAt: r.createdAt.UTC().Format(time.RFC3339),
	}
	if r.startedAt != nil {
		t := r.startedAt.UTC().Format(time.RFC3339)
		resp.Status.StartTime = &t
	}
	if r.endedAt != nil {
		t := r.endedAt.UTC().Format(time.RFC3339)
		resp.Status.CompletionTime = &t
	}
	return resp
}

// phaseFromStatus maps the DB experiment status enum to the frontend
// ExperimentPhase contract (Pending/Running/Completed/Failed/Aborted).
func phaseFromStatus(status string) string {
	switch status {
	case "running":
		return "Running"
	case "completed":
		return "Completed"
	case "failed", "rejected":
		return "Failed"
	case "aborted", "rolled_back":
		return "Aborted"
	default:
		return "Pending"
	}
}
