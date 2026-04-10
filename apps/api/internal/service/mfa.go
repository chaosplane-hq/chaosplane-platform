package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type MFAService struct {
	pool *database.Pool
}

func NewMFAService(pool *database.Pool) *MFAService {
	return &MFAService{pool: pool}
}

type RecoveryCode struct {
	Code string `json:"code"`
}

type RecoveryCodesResponse struct {
	Codes     []RecoveryCode `json:"codes"`
	Remaining int            `json:"remaining"`
}

func (s *MFAService) GenerateRecoveryCodes(ctx context.Context, actor ActorContext) (*RecoveryCodesResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	_, _ = s.pool.App.Exec(ctx, `DELETE FROM mfa_recovery_codes WHERE user_id = $1::uuid`, actor.UserID)

	codes := make([]RecoveryCode, 10)
	for i := range codes {
		raw := make([]byte, 6)
		if _, err := rand.Read(raw); err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}
		plain := hex.EncodeToString(raw)
		hash := sha256.Sum256([]byte(plain))
		hashStr := hex.EncodeToString(hash[:])

		if _, err := s.pool.App.Exec(ctx, `
			INSERT INTO mfa_recovery_codes (user_id, code_hash) VALUES ($1::uuid, $2)
		`, actor.UserID, hashStr); err != nil {
			return nil, fmt.Errorf("store recovery code: %w", err)
		}
		codes[i] = RecoveryCode{Code: plain}
	}

	return &RecoveryCodesResponse{Codes: codes, Remaining: 10}, nil
}

func (s *MFAService) VerifyRecoveryCode(ctx context.Context, userID, code string) (bool, int, error) {
	hash := sha256.Sum256([]byte(code))
	hashStr := hex.EncodeToString(hash[:])

	cmd, err := s.pool.App.Exec(ctx, `
		UPDATE mfa_recovery_codes SET used_at = now()
		WHERE user_id = $1::uuid AND code_hash = $2 AND used_at IS NULL
	`, userID, hashStr)
	if err != nil {
		return false, 0, fmt.Errorf("verify recovery code: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return false, 0, nil
	}

	var remaining int
	_ = s.pool.App.QueryRow(ctx, `
		SELECT COUNT(*) FROM mfa_recovery_codes WHERE user_id = $1::uuid AND used_at IS NULL
	`, userID).Scan(&remaining)

	return true, remaining, nil
}

func (s *MFAService) GetRemainingCount(ctx context.Context, actor ActorContext) (int, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return 0, err
	}
	var count int
	err := s.pool.App.QueryRow(ctx, `
		SELECT COUNT(*) FROM mfa_recovery_codes WHERE user_id = $1::uuid AND used_at IS NULL
	`, actor.UserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count recovery codes: %w", err)
	}
	return count, nil
}
