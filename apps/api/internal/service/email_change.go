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

type EmailChangeService struct {
	pool  *database.Pool
	email *EmailService
}

func NewEmailChangeService(pool *database.Pool, email *EmailService) *EmailChangeService {
	return &EmailChangeService{pool: pool, email: email}
}

type EmailChangeRequest struct {
	NewEmail string `json:"newEmail" binding:"required,email"`
}

type EmailChangeResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

func (s *EmailChangeService) Request(ctx context.Context, actor ActorContext, req *EmailChangeRequest) (*EmailChangeResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	newEmail := strings.ToLower(strings.TrimSpace(req.NewEmail))
	var existingID string
	err := s.pool.App.QueryRow(ctx, `SELECT id::text FROM users WHERE lower(email) = $1 AND deleted_at IS NULL`, newEmail).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("email already in use")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check email availability: %w", err)
	}

	plain, tokenHash, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate email change token: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)

	_, _ = s.pool.App.Exec(ctx, `DELETE FROM email_change_requests WHERE user_id = $1::uuid AND confirmed_at IS NULL`, actor.UserID)

	if _, err := s.pool.App.Exec(ctx, `
		INSERT INTO email_change_requests (user_id, new_email, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3, $4)
	`, actor.UserID, newEmail, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("create email change request: %w", err)
	}

	if s.email != nil {
		var currentEmail string
		_ = s.pool.App.QueryRow(ctx, `SELECT email FROM users WHERE id = $1::uuid`, actor.UserID).Scan(&currentEmail)
		go s.email.SendEmailChangeNotification(context.Background(), currentEmail, newEmail, plain)
	}

	return &EmailChangeResponse{Token: plain, ExpiresAt: expiresAt.Format(time.RFC3339)}, nil
}

func (s *EmailChangeService) Confirm(ctx context.Context, token string) error {
	var userID, newEmail string
	var expiresAt time.Time
	err := s.pool.App.QueryRow(ctx, `
		SELECT user_id::text, new_email, expires_at
		FROM email_change_requests
		WHERE token_hash = $1 AND confirmed_at IS NULL
	`, hashToken(token)).Scan(&userID, &newEmail, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("lookup email change request: %w", err)
	}
	if expiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	if _, err := s.pool.App.Exec(ctx, `UPDATE users SET email = $2 WHERE id = $1::uuid`, userID, newEmail); err != nil {
		return fmt.Errorf("update email: %w", err)
	}
	if _, err := s.pool.App.Exec(ctx, `UPDATE email_change_requests SET confirmed_at = now() WHERE token_hash = $1`, hashToken(token)); err != nil {
		return fmt.Errorf("confirm email change: %w", err)
	}
	return nil
}
