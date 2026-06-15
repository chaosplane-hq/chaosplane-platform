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

type OnboardingService struct {
	pool *database.Pool
}

func NewOnboardingService(pool *database.Pool) *OnboardingService {
	return &OnboardingService{pool: pool}
}

type OnboardingProgress struct {
	UserID               string     `json:"userId"`
	TenantID             string     `json:"tenantId"`
	StepOrgCreated       bool       `json:"stepOrgCreated"`
	StepWorkspaceCreated bool       `json:"stepWorkspaceCreated"`
	StepTeamCreated      bool       `json:"stepTeamCreated"`
	StepMemberInvited    bool       `json:"stepMemberInvited"`
	StepClusterConnected bool       `json:"stepClusterConnected"`
	StepFirstExperiment  bool       `json:"stepFirstExperiment"`
	StepResultViewed     bool       `json:"stepResultViewed"`
	CompletedAt          *time.Time `json:"completedAt,omitempty"`
	SkippedAt            *time.Time `json:"skippedAt,omitempty"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type UpdateOnboardingRequest struct {
	StepOrgCreated       *bool `json:"stepOrgCreated,omitempty"`
	StepWorkspaceCreated *bool `json:"stepWorkspaceCreated,omitempty"`
	StepTeamCreated      *bool `json:"stepTeamCreated,omitempty"`
	StepMemberInvited    *bool `json:"stepMemberInvited,omitempty"`
	StepClusterConnected *bool `json:"stepClusterConnected,omitempty"`
	StepFirstExperiment  *bool `json:"stepFirstExperiment,omitempty"`
	StepResultViewed     *bool `json:"stepResultViewed,omitempty"`
}

type QuickSetupResponse struct {
	OrganizationID string              `json:"organizationId"`
	WorkspaceID    string              `json:"workspaceId"`
	TeamID         string              `json:"teamId"`
	Progress       *OnboardingProgress `json:"progress"`
}

type TestAgentConnectionRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required,uuid"`
}

type TestAgentConnectionResponse struct {
	EnvironmentID string `json:"environmentId"`
	AgentStatus   string `json:"agentStatus"`
	Connected     bool   `json:"connected"`
}

func (s *OnboardingService) Get(ctx context.Context, actor ActorContext) (*OnboardingProgress, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	return s.getOrCreateProgress(ctx, actor)
}

func (s *OnboardingService) Update(ctx context.Context, actor ActorContext, req *UpdateOnboardingRequest) (*OnboardingProgress, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	progress, err := s.getOrCreateProgress(ctx, actor)
	if err != nil {
		return nil, err
	}

	if req.StepOrgCreated != nil {
		progress.StepOrgCreated = *req.StepOrgCreated
	}
	if req.StepWorkspaceCreated != nil {
		progress.StepWorkspaceCreated = *req.StepWorkspaceCreated
	}
	if req.StepTeamCreated != nil {
		progress.StepTeamCreated = *req.StepTeamCreated
	}
	if req.StepMemberInvited != nil {
		progress.StepMemberInvited = *req.StepMemberInvited
	}
	if req.StepClusterConnected != nil {
		progress.StepClusterConnected = *req.StepClusterConnected
	}
	if req.StepFirstExperiment != nil {
		progress.StepFirstExperiment = *req.StepFirstExperiment
	}
	if req.StepResultViewed != nil {
		progress.StepResultViewed = *req.StepResultViewed
	}

	if progress.StepOrgCreated && progress.StepWorkspaceCreated && progress.StepTeamCreated && progress.StepClusterConnected && progress.StepFirstExperiment && progress.StepResultViewed {
		now := time.Now()
		progress.CompletedAt = &now
		progress.SkippedAt = nil
	}

	return s.persistProgress(ctx, progress)
}

func (s *OnboardingService) Skip(ctx context.Context, actor ActorContext) (*OnboardingProgress, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	progress, err := s.getOrCreateProgress(ctx, actor)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	progress.SkippedAt = &now
	if progress.CompletedAt == nil {
		progress.CompletedAt = &now
	}
	return s.persistProgress(ctx, progress)
}

func (s *OnboardingService) Complete(ctx context.Context, actor ActorContext) (*OnboardingProgress, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	progress, err := s.getOrCreateProgress(ctx, actor)
	if err != nil {
		return nil, err
	}
	progress.StepOrgCreated = true
	progress.StepWorkspaceCreated = true
	progress.StepTeamCreated = true
	progress.StepMemberInvited = true
	progress.StepClusterConnected = true
	progress.StepFirstExperiment = true
	progress.StepResultViewed = true
	now := time.Now()
	progress.CompletedAt = &now
	progress.SkippedAt = nil
	return s.persistProgress(ctx, progress)
}

func (s *OnboardingService) QuickSetup(ctx context.Context, actor ActorContext) (*QuickSetupResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin quick-setup tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	orgID, err := ensureDefaultOrganization(ctx, tx, actor)
	if err != nil {
		return nil, err
	}
	workspaceID, err := ensureDefaultWorkspace(ctx, tx, actor.TenantID, orgID)
	if err != nil {
		return nil, err
	}
	teamID, err := ensureDefaultTeam(ctx, tx, actor, workspaceID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO onboarding_progress (user_id, tenant_id, step_org_created, step_workspace_created, step_team_created)
		VALUES ($1::uuid, $2::uuid, true, true, true)
		ON CONFLICT (user_id, tenant_id)
		DO UPDATE SET
			step_org_created = true,
			step_workspace_created = true,
			step_team_created = true,
			updated_at = now()
	`, actor.UserID, actor.TenantID); err != nil {
		return nil, fmt.Errorf("upsert onboarding quick-setup progress: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit quick-setup tx: %w", err)
	}

	progress, err := s.getOrCreateProgress(ctx, actor)
	if err != nil {
		return nil, err
	}

	return &QuickSetupResponse{OrganizationID: orgID, WorkspaceID: workspaceID, TeamID: teamID, Progress: progress}, nil
}

func (s *OnboardingService) TestAgentConnection(ctx context.Context, actor ActorContext, req *TestAgentConnectionRequest) (*TestAgentConnectionResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	var resp TestAgentConnectionResponse
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text, agent_status
		FROM environments
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL
	`, req.EnvironmentID, actor.TenantID).Scan(&resp.EnvironmentID, &resp.AgentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, fmt.Errorf("lookup environment agent status: %w", err)
	}

	resp.Connected = strings.EqualFold(resp.AgentStatus, "connected") || strings.EqualFold(resp.AgentStatus, "degraded")
	if resp.Connected {
		progress, err := s.getOrCreateProgress(ctx, actor)
		if err == nil {
			progress.StepClusterConnected = true
			_, _ = s.persistProgress(ctx, progress)
		}
	}

	return &resp, nil
}

func (s *OnboardingService) getOrCreateProgress(ctx context.Context, actor ActorContext) (*OnboardingProgress, error) {
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT user_id::text, tenant_id::text, step_org_created, step_workspace_created, step_team_created,
		       step_member_invited, step_cluster_connected, step_first_experiment, step_result_viewed,
		       completed_at, skipped_at, updated_at
		FROM onboarding_progress
		WHERE user_id = $1::uuid AND tenant_id = $2::uuid
	`, actor.UserID, actor.TenantID)

	progress := &OnboardingProgress{}
	if err := row.Scan(
		&progress.UserID,
		&progress.TenantID,
		&progress.StepOrgCreated,
		&progress.StepWorkspaceCreated,
		&progress.StepTeamCreated,
		&progress.StepMemberInvited,
		&progress.StepClusterConnected,
		&progress.StepFirstExperiment,
		&progress.StepResultViewed,
		&progress.CompletedAt,
		&progress.SkippedAt,
		&progress.UpdatedAt,
	); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("load onboarding progress: %w", err)
		}
		if _, insertErr := s.pool.Conn(ctx).Exec(ctx, `
			INSERT INTO onboarding_progress (user_id, tenant_id)
			VALUES ($1::uuid, $2::uuid)
			ON CONFLICT (user_id, tenant_id) DO NOTHING
		`, actor.UserID, actor.TenantID); insertErr != nil {
			return nil, fmt.Errorf("create onboarding progress: %w", insertErr)
		}
		return s.getOrCreateProgress(ctx, actor)
	}
	return progress, nil
}

func (s *OnboardingService) persistProgress(ctx context.Context, progress *OnboardingProgress) (*OnboardingProgress, error) {
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		UPDATE onboarding_progress
		SET step_org_created = $3,
		    step_workspace_created = $4,
		    step_team_created = $5,
		    step_member_invited = $6,
		    step_cluster_connected = $7,
		    step_first_experiment = $8,
		    step_result_viewed = $9,
		    completed_at = $10,
		    skipped_at = $11,
		    updated_at = now()
		WHERE user_id = $1::uuid AND tenant_id = $2::uuid
		RETURNING user_id::text, tenant_id::text, step_org_created, step_workspace_created, step_team_created,
		       step_member_invited, step_cluster_connected, step_first_experiment, step_result_viewed,
		       completed_at, skipped_at, updated_at
	`, progress.UserID, progress.TenantID, progress.StepOrgCreated, progress.StepWorkspaceCreated, progress.StepTeamCreated, progress.StepMemberInvited, progress.StepClusterConnected, progress.StepFirstExperiment, progress.StepResultViewed, progress.CompletedAt, progress.SkippedAt)

	updated := &OnboardingProgress{}
	if err := row.Scan(
		&updated.UserID,
		&updated.TenantID,
		&updated.StepOrgCreated,
		&updated.StepWorkspaceCreated,
		&updated.StepTeamCreated,
		&updated.StepMemberInvited,
		&updated.StepClusterConnected,
		&updated.StepFirstExperiment,
		&updated.StepResultViewed,
		&updated.CompletedAt,
		&updated.SkippedAt,
		&updated.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("persist onboarding progress: %w", err)
	}
	return updated, nil
}

func ensureActorMembership(ctx context.Context, pool *database.Pool, actor ActorContext) error {
	var exists bool
	err := pool.Conn(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_tenants WHERE user_id = $1::uuid AND tenant_id = $2::uuid
		)
	`, actor.UserID, actor.TenantID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check tenant membership: %w", err)
	}
	if !exists {
		return ErrInvalidCredentials
	}
	return nil
}

func ensureDefaultOrganization(ctx context.Context, tx pgx.Tx, actor ActorContext) (string, error) {
	var orgID string
	err := tx.QueryRow(ctx, `
		SELECT id::text FROM organizations
		WHERE tenant_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1
	`, actor.TenantID).Scan(&orgID)
	if err == nil {
		return orgID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("lookup default organization: %w", err)
	}

	var userName string
	if err := tx.QueryRow(ctx, `SELECT name FROM users WHERE id = $1::uuid AND deleted_at IS NULL`, actor.UserID).Scan(&userName); err != nil {
		return "", fmt.Errorf("lookup user for quick setup org: %w", err)
	}
	slug := defaultSlug("", userName+" org")
	err = tx.QueryRow(ctx, `
		INSERT INTO organizations (tenant_id, name, slug)
		VALUES ($1::uuid, $2, $3)
		RETURNING id::text
	`, actor.TenantID, strings.TrimSpace(userName)+" Organization", slug).Scan(&orgID)
	if err != nil {
		return "", fmt.Errorf("create default organization: %w", err)
	}
	return orgID, nil
}

func ensureDefaultWorkspace(ctx context.Context, tx pgx.Tx, tenantID, orgID string) (string, error) {
	var workspaceID string
	err := tx.QueryRow(ctx, `
		SELECT id::text FROM workspaces
		WHERE tenant_id = $1::uuid AND organization_id = $2::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1
	`, tenantID, orgID).Scan(&workspaceID)
	if err == nil {
		return workspaceID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("lookup default workspace: %w", err)
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (tenant_id, organization_id, name, slug)
		VALUES ($1::uuid, $2::uuid, 'Default Workspace', 'default')
		RETURNING id::text
	`, tenantID, orgID).Scan(&workspaceID)
	if err != nil {
		return "", fmt.Errorf("create default workspace: %w", err)
	}
	return workspaceID, nil
}

func ensureDefaultTeam(ctx context.Context, tx pgx.Tx, actor ActorContext, workspaceID string) (string, error) {
	var teamID string
	err := tx.QueryRow(ctx, `
		SELECT id::text FROM teams
		WHERE tenant_id = $1::uuid AND workspace_id = $2::uuid AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1
	`, actor.TenantID, workspaceID).Scan(&teamID)
	if err == nil {
		_, _ = tx.Exec(ctx, `
			INSERT INTO team_members (team_id, user_id, role)
			VALUES ($1::uuid, $2::uuid, 'lead')
			ON CONFLICT (team_id, user_id) DO NOTHING
		`, teamID, actor.UserID)
		return teamID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("lookup default team: %w", err)
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO teams (tenant_id, workspace_id, name, slug)
		VALUES ($1::uuid, $2::uuid, 'Core Team', 'core-team')
		RETURNING id::text
	`, actor.TenantID, workspaceID).Scan(&teamID)
	if err != nil {
		return "", fmt.Errorf("create default team: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, 'lead')
		ON CONFLICT (team_id, user_id) DO NOTHING
	`, teamID, actor.UserID); err != nil {
		return "", fmt.Errorf("create default team membership: %w", err)
	}
	return teamID, nil
}
