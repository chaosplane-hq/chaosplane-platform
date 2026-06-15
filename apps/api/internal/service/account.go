package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var (
	ErrTOSOutdated    = errors.New("terms of service acceptance required")
	ErrAccountDeleted = errors.New("account is scheduled for deletion")
)

type AccountService struct {
	pool *database.Pool
	cfg  *config.Config
}

func NewAccountService(pool *database.Pool, cfg *config.Config) *AccountService {
	return &AccountService{pool: pool, cfg: cfg}
}

type AcceptTOSRequest struct {
	Version string `json:"version" binding:"required"`
}

type AccountExportResponse struct {
	User        json.RawMessage `json:"user"`
	Tenants     json.RawMessage `json:"tenants"`
	Teams       json.RawMessage `json:"teams"`
	Invitations json.RawMessage `json:"invitations"`
	APIKeys     json.RawMessage `json:"apiKeys"`
	Onboarding  json.RawMessage `json:"onboarding"`
	ExportedAt  string          `json:"exportedAt"`
}

type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" binding:"required"`
}

func (s *AccountService) AcceptTOS(ctx context.Context, actor ActorContext, req *AcceptTOSRequest) error {
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE users
		SET accepted_tos_version = $2, accepted_tos_at = now()
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, actor.UserID, req.Version)
	if err != nil {
		return fmt.Errorf("accept tos: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *AccountService) Export(ctx context.Context, actor ActorContext) (*AccountExportResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	userJSON, err := queryJSON(ctx, s.pool, `
		SELECT json_build_object(
			'id', id::text, 'email', email, 'name', name,
			'emailVerified', email_verified, 'status', status,
			'acceptedTosVersion', accepted_tos_version,
			'createdAt', created_at, 'lastLoginAt', last_login_at
		) FROM users WHERE id = $1::uuid AND deleted_at IS NULL
	`, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("export user: %w", err)
	}

	tenantsJSON, err := queryJSON(ctx, s.pool, `
		SELECT COALESCE(json_agg(json_build_object(
			'id', t.id::text, 'name', t.name, 'slug', t.slug, 'plan', t.plan, 'joinedAt', ut.joined_at
		)), '[]'::json)
		FROM user_tenants ut
		JOIN tenants t ON t.id = ut.tenant_id
		WHERE ut.user_id = $1::uuid
	`, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("export tenants: %w", err)
	}

	teamsJSON, err := queryJSON(ctx, s.pool, `
		SELECT COALESCE(json_agg(json_build_object(
			'teamId', tm.team_id::text, 'role', tm.role, 'joinedAt', tm.joined_at,
			'teamName', t.name, 'workspaceId', t.workspace_id::text
		)), '[]'::json)
		FROM team_members tm
		JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = $1::uuid AND t.deleted_at IS NULL
	`, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("export teams: %w", err)
	}

	invitationsJSON, err := queryJSON(ctx, s.pool, `
		SELECT COALESCE(json_agg(json_build_object(
			'id', id::text, 'email', email, 'role', role, 'status', status, 'createdAt', created_at
		)), '[]'::json)
		FROM invitations
		WHERE invited_by = $1::uuid
	`, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("export invitations: %w", err)
	}

	apiKeysJSON, err := queryJSON(ctx, s.pool, `
		SELECT COALESCE(json_agg(json_build_object(
			'id', id::text, 'name', name, 'createdAt', created_at, 'lastUsedAt', last_used_at, 'revokedAt', revoked_at
		)), '[]'::json)
		FROM api_keys
		WHERE user_id = $1::uuid
	`, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("export api keys: %w", err)
	}

	onboardingJSON, err := queryJSON(ctx, s.pool, `
		SELECT COALESCE(json_agg(json_build_object(
			'tenantId', tenant_id::text,
			'stepOrgCreated', step_org_created, 'stepWorkspaceCreated', step_workspace_created,
			'stepTeamCreated', step_team_created, 'stepMemberInvited', step_member_invited,
			'stepClusterConnected', step_cluster_connected, 'stepFirstExperiment', step_first_experiment,
			'stepResultViewed', step_result_viewed,
			'completedAt', completed_at, 'skippedAt', skipped_at
		)), '[]'::json)
		FROM onboarding_progress
		WHERE user_id = $1::uuid
	`, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("export onboarding: %w", err)
	}

	return &AccountExportResponse{
		User:        userJSON,
		Tenants:     tenantsJSON,
		Teams:       teamsJSON,
		Invitations: invitationsJSON,
		APIKeys:     apiKeysJSON,
		Onboarding:  onboardingJSON,
		ExportedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

func (s *AccountService) Delete(ctx context.Context, actor ActorContext, req *DeleteAccountRequest) error {
	if req.Confirmation != "DELETE" {
		return fmt.Errorf("confirmation must be 'DELETE'")
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin delete account tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cmd, err := tx.Exec(ctx, `
		UPDATE users
		SET deleted_at = now(), status = 'disabled', email = email || '.deleted.' || id::text
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, actor.UserID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1::uuid AND revoked_at IS NULL
	`, actor.UserID); err != nil {
		return fmt.Errorf("revoke refresh tokens on account delete: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE api_keys SET revoked_at = now() WHERE user_id = $1::uuid AND revoked_at IS NULL
	`, actor.UserID); err != nil {
		return fmt.Errorf("revoke api keys on account delete: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete account tx: %w", err)
	}
	return nil
}

func (s *AccountService) CancelDeletion(ctx context.Context, actor ActorContext) error {
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE users
		SET deleted_at = NULL, status = 'active',
		    email = regexp_replace(email, '\.deleted\.[a-f0-9-]+$', '')
		WHERE id = $1::uuid AND deleted_at IS NOT NULL
		  AND deleted_at > now() - interval '30 days'
	`, actor.UserID)
	if err != nil {
		return fmt.Errorf("cancel account deletion: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func queryJSON(ctx context.Context, pool *database.Pool, query string, args ...any) (json.RawMessage, error) {
	var raw json.RawMessage
	err := pool.Conn(ctx).QueryRow(ctx, query, args...).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return json.RawMessage("null"), nil
		}
		return nil, err
	}
	return raw, nil
}
