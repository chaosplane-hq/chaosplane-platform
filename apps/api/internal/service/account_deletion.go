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

type AccountDeletionService struct {
	pool *database.Pool
}

func NewAccountDeletionService(pool *database.Pool) *AccountDeletionService {
	return &AccountDeletionService{pool: pool}
}

type DeletionRequest struct {
	ID              string     `json:"id"`
	UserID          string     `json:"userId"`
	Reason          *string    `json:"reason,omitempty"`
	GracePeriodEnds time.Time  `json:"gracePeriodEnds"`
	CancelledAt     *time.Time `json:"cancelledAt,omitempty"`
	ExecutedAt      *time.Time `json:"executedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type RequestDeletionRequest struct {
	Reason string `json:"reason,omitempty"`
}

func (s *AccountDeletionService) Request(ctx context.Context, actor ActorContext, req *RequestDeletionRequest) (*DeletionRequest, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	gracePeriodEnds := time.Now().Add(30 * 24 * time.Hour)
	var dr DeletionRequest
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		INSERT INTO account_deletion_requests (user_id, reason, grace_period_ends)
		VALUES ($1::uuid, NULLIF($2, ''), $3)
		RETURNING id::text, user_id::text, reason, grace_period_ends, cancelled_at, executed_at, created_at
	`, actor.UserID, req.Reason, gracePeriodEnds).Scan(
		&dr.ID, &dr.UserID, &dr.Reason, &dr.GracePeriodEnds, &dr.CancelledAt, &dr.ExecutedAt, &dr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("request account deletion: %w", err)
	}

	if _, err := s.pool.Conn(ctx).Exec(ctx, `UPDATE users SET status = 'pending_deletion' WHERE id = $1::uuid`, actor.UserID); err != nil {
		return nil, fmt.Errorf("mark user pending_deletion: %w", err)
	}
	return &dr, nil
}

func (s *AccountDeletionService) Cancel(ctx context.Context, actor ActorContext) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE account_deletion_requests SET cancelled_at = now()
		WHERE user_id = $1::uuid AND executed_at IS NULL AND cancelled_at IS NULL
	`, actor.UserID)
	if err != nil {
		return fmt.Errorf("cancel deletion: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	if _, err := s.pool.Conn(ctx).Exec(ctx, `UPDATE users SET status = 'active', deleted_at = NULL WHERE id = $1::uuid`, actor.UserID); err != nil {
		return fmt.Errorf("restore user status: %w", err)
	}
	return nil
}

func (s *AccountDeletionService) ExecuteExpired(ctx context.Context) (int, error) {
	rows, err := s.pool.Conn(ctx).Query(ctx, `
		SELECT id::text, user_id::text FROM account_deletion_requests
		WHERE grace_period_ends < now() AND executed_at IS NULL AND cancelled_at IS NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("find expired deletion requests: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var reqID, userID string
		if rows.Scan(&reqID, &userID) != nil {
			continue
		}
		if err := s.anonymizeUser(ctx, userID); err != nil {
			continue
		}
		if _, err := s.pool.Conn(ctx).Exec(ctx, `UPDATE account_deletion_requests SET executed_at = now() WHERE id = $1::uuid`, reqID); err != nil {
			continue
		}
		count++
	}
	return count, rows.Err()
}

func (s *AccountDeletionService) anonymizeUser(ctx context.Context, userID string) error {
	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE users SET
			email = 'deleted-' || id::text || '@anonymized.local',
			name = 'Deleted User',
			password_hash = '',
			avatar_url = NULL,
			status = 'deleted',
			deleted_at = COALESCE(deleted_at, now())
		WHERE id = $1::uuid
	`, userID); err != nil {
		return fmt.Errorf("anonymize user: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1::uuid AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke tokens: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE api_keys SET revoked_at = now() WHERE user_id = $1::uuid AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke api keys: %w", err)
	}

	return tx.Commit(ctx)
}

var _ = errors.Is
var _ = strings.TrimSpace
