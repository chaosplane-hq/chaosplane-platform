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

var ErrInvitationSignupRequired = errors.New("invitation requires signup")

type InvitationService struct {
	pool *database.Pool
}

func NewInvitationService(pool *database.Pool) *InvitationService {
	return &InvitationService{pool: pool}
}

type Invitation struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenantId"`
	OrganizationID string     `json:"organizationId"`
	TeamID         *string    `json:"teamId,omitempty"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	InviteToken    string     `json:"inviteToken,omitempty"`
}

type ListInvitationsResponse struct {
	Items []Invitation `json:"items"`
}

type CreateInvitationRequest struct {
	Email          string  `json:"email" binding:"required,email"`
	OrganizationID string  `json:"organizationId" binding:"required,uuid"`
	TeamID         *string `json:"teamId,omitempty" binding:"omitempty,uuid"`
	Role           string  `json:"role,omitempty"`
}

type AcceptInvitationRequest struct {
	InvitationID string `json:"invitationId" binding:"required,uuid"`
}

type DeclineInvitationRequest struct {
	InvitationID string `json:"invitationId" binding:"required,uuid"`
}

func (s *InvitationService) List(ctx context.Context, actor ActorContext) (*ListInvitationsResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, tenant_id::text, organization_id::text, team_id::text, email, role, status, expires_at, accepted_at, created_at
		FROM invitations
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	items := []Invitation{}
	for rows.Next() {
		item, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invitations: %w", err)
	}

	return &ListInvitationsResponse{Items: items}, nil
}

func (s *InvitationService) Create(ctx context.Context, actor ActorContext, req *CreateInvitationRequest) (*Invitation, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	if err := ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM organizations WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, req.OrganizationID, actor.TenantID); err != nil {
		return nil, err
	}
	if req.TeamID != nil && *req.TeamID != "" {
		if err := ensureTenantScopedRow(ctx, s.pool, `SELECT 1 FROM teams WHERE id = $1::uuid AND tenant_id = $2::uuid AND deleted_at IS NULL`, *req.TeamID, actor.TenantID); err != nil {
			return nil, err
		}
	}

	role := normalizeInvitationRole(req.Role)
	plainToken, tokenHash, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate invitation token: %w", err)
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	row := s.pool.App.QueryRow(ctx, `
		INSERT INTO invitations (tenant_id, organization_id, team_id, email, role, invited_by, token_hash, expires_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, lower($4), $5, $6::uuid, $7, $8)
		RETURNING id::text, tenant_id::text, organization_id::text, team_id::text, email, role, status, expires_at, accepted_at, created_at
	`, actor.TenantID, req.OrganizationID, req.TeamID, strings.ToLower(strings.TrimSpace(req.Email)), role, actor.UserID, tokenHash, expiresAt)

	invitation, err := scanInvitation(row)
	if err != nil {
		return nil, err
	}
	invitation.InviteToken = plainToken
	return invitation, nil
}

func (s *InvitationService) Resend(ctx context.Context, actor ActorContext, invitationID string) (*Invitation, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	plainToken, tokenHash, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate resend invitation token: %w", err)
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	row := s.pool.App.QueryRow(ctx, `
		UPDATE invitations
		SET token_hash = $3, expires_at = $4, status = 'pending', accepted_at = NULL
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND status IN ('pending','expired','declined')
		RETURNING id::text, tenant_id::text, organization_id::text, team_id::text, email, role, status, expires_at, accepted_at, created_at
	`, invitationID, actor.TenantID, tokenHash, expiresAt)

	invitation, err := scanInvitation(row)
	if err != nil {
		return nil, err
	}
	invitation.InviteToken = plainToken
	return invitation, nil
}

func (s *InvitationService) Revoke(ctx context.Context, actor ActorContext, invitationID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `
		UPDATE invitations
		SET status = 'revoked'
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND status = 'pending'
	`, invitationID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("revoke invitation: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *InvitationService) Accept(ctx context.Context, actor ActorContext, req *AcceptInvitationRequest) (*Invitation, error) {
	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin accept invitation tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	invitation, err := s.loadPendingInvitation(ctx, tx, req.InvitationID)
	if err != nil {
		return nil, err
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		if _, err := tx.Exec(ctx, `UPDATE invitations SET status = 'expired' WHERE id = $1::uuid`, invitation.ID); err == nil {
			_ = tx.Commit(ctx)
		}
		return nil, ErrTokenExpired
	}

	userEmail, err := s.lookupUserEmail(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(userEmail, invitation.Email) {
		return nil, ErrInvalidCredentials
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_tenants (user_id, tenant_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (user_id, tenant_id) DO NOTHING
	`, actor.UserID, invitation.TenantID); err != nil {
		return nil, fmt.Errorf("link invited user to tenant: %w", err)
	}

	if invitation.TeamID != nil {
		role := invitationRoleToTeamRole(invitation.Role)
		if _, err := tx.Exec(ctx, `
			INSERT INTO team_members (team_id, user_id, role)
			VALUES ($1::uuid, $2::uuid, $3)
			ON CONFLICT (team_id, user_id) DO NOTHING
		`, *invitation.TeamID, actor.UserID, role); err != nil {
			return nil, fmt.Errorf("add invited user to team: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		UPDATE invitations
		SET status = 'accepted', accepted_at = now()
		WHERE id = $1::uuid
		RETURNING id::text, tenant_id::text, organization_id::text, team_id::text, email, role, status, expires_at, accepted_at, created_at
	`, invitation.ID)
	accepted, err := scanInvitation(row)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit accept invitation tx: %w", err)
	}
	return accepted, nil
}

func (s *InvitationService) Decline(ctx context.Context, actor ActorContext, req *DeclineInvitationRequest) (*Invitation, error) {
	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin decline invitation tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	invitation, err := s.loadPendingInvitation(ctx, tx, req.InvitationID)
	if err != nil {
		return nil, err
	}
	userEmail, err := s.lookupUserEmail(ctx, tx, actor.UserID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(userEmail, invitation.Email) {
		return nil, ErrInvalidCredentials
	}

	row := tx.QueryRow(ctx, `
		UPDATE invitations
		SET status = 'declined'
		WHERE id = $1::uuid
		RETURNING id::text, tenant_id::text, organization_id::text, team_id::text, email, role, status, expires_at, accepted_at, created_at
	`, invitation.ID)
	declined, err := scanInvitation(row)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit decline invitation tx: %w", err)
	}
	return declined, nil
}

func (s *InvitationService) loadPendingInvitation(ctx context.Context, tx pgx.Tx, invitationID string) (*Invitation, error) {
	row := tx.QueryRow(ctx, `
		SELECT id::text, tenant_id::text, organization_id::text, team_id::text, email, role, status, expires_at, accepted_at, created_at
		FROM invitations
		WHERE id = $1::uuid AND status = 'pending'
	`, invitationID)
	return scanInvitation(row)
}

func (s *InvitationService) lookupUserEmail(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var email string
	err := tx.QueryRow(ctx, `SELECT email FROM users WHERE id = $1::uuid AND deleted_at IS NULL`, userID).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvitationSignupRequired
		}
		return "", fmt.Errorf("lookup invitation user email: %w", err)
	}
	return email, nil
}

func scanInvitation(row interface{ Scan(dest ...any) error }) (*Invitation, error) {
	var item Invitation
	if err := row.Scan(&item.ID, &item.TenantID, &item.OrganizationID, &item.TeamID, &item.Email, &item.Role, &item.Status, &item.ExpiresAt, &item.AcceptedAt, &item.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, fmt.Errorf("scan invitation: %w", err)
	}
	return &item, nil
}

func normalizeInvitationRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "admin"
	case "editor", "member", "operator":
		return "editor"
	default:
		return "viewer"
	}
}

func invitationRoleToTeamRole(role string) string {
	switch normalizeInvitationRole(role) {
	case "admin":
		return "lead"
	case "editor":
		return "member"
	default:
		return "viewer"
	}
}
