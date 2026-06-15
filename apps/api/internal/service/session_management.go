package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type SessionManagementService struct {
	pool *database.Pool
}

func NewSessionManagementService(pool *database.Pool) *SessionManagementService {
	return &SessionManagementService{pool: pool}
}

type ActiveSession struct {
	ID           string    `json:"id"`
	IPAddress    *string   `json:"ipAddress,omitempty"`
	UserAgent    *string   `json:"userAgent,omitempty"`
	LastActivity time.Time `json:"lastActivity"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ActiveSessionListResponse struct {
	Items []ActiveSession `json:"items"`
}

func (s *SessionManagementService) List(ctx context.Context, actor ActorContext) (*ActiveSessionListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, host(ip_address), user_agent, last_activity, created_at
		FROM active_sessions
		WHERE user_id = $1::uuid AND tenant_id = $2::uuid AND revoked_at IS NULL
		ORDER BY last_activity DESC
	`, actor.UserID, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	defer rows.Close()
	items := []ActiveSession{}
	for rows.Next() {
		var s ActiveSession
		if err := rows.Scan(&s.ID, &s.IPAddress, &s.UserAgent, &s.LastActivity, &s.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return &ActiveSessionListResponse{Items: items}, rows.Err()
}

func (s *SessionManagementService) Register(ctx context.Context, userID, tenantID string, ip net.IP, userAgent string) (string, error) {
	var sessionID string
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO active_sessions (user_id, tenant_id, ip_address, user_agent)
		VALUES ($1::uuid, $2::uuid, $3::inet, $4)
		RETURNING id::text
	`, userID, tenantID, ip, userAgent).Scan(&sessionID)
	if err != nil {
		return "", fmt.Errorf("register session: %w", err)
	}
	return sessionID, nil
}

func (s *SessionManagementService) Revoke(ctx context.Context, actor ActorContext, sessionID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE active_sessions SET revoked_at = now()
		WHERE id = $1::uuid AND user_id = $2::uuid AND revoked_at IS NULL
	`, sessionID, actor.UserID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *SessionManagementService) RevokeAll(ctx context.Context, actor ActorContext) (int, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return 0, err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE active_sessions SET revoked_at = now()
		WHERE user_id = $1::uuid AND tenant_id = $2::uuid AND revoked_at IS NULL
	`, actor.UserID, actor.TenantID)
	if err != nil {
		return 0, fmt.Errorf("revoke all sessions: %w", err)
	}
	return int(cmd.RowsAffected()), nil
}

func (s *SessionManagementService) EnforceLimit(ctx context.Context, userID, tenantID string, maxSessions int) error {
	if maxSessions <= 0 {
		return nil
	}
	_, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE active_sessions SET revoked_at = now()
		WHERE id IN (
			SELECT id FROM active_sessions
			WHERE user_id = $1::uuid AND tenant_id = $2::uuid AND revoked_at IS NULL
			ORDER BY last_activity ASC
			OFFSET $3
		)
	`, userID, tenantID, maxSessions)
	if err != nil {
		return fmt.Errorf("enforce session limit: %w", err)
	}
	return nil
}

func (s *SessionManagementService) Touch(ctx context.Context, sessionID string) {
	_, _ = s.pool.Conn(ctx).Exec(ctx, `
		UPDATE active_sessions SET last_activity = now() WHERE id = $1::uuid AND revoked_at IS NULL
	`, sessionID)
}
