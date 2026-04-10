package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type APIKeyService struct {
	pool *database.Pool
}

func NewAPIKeyService(pool *database.Pool) *APIKeyService {
	return &APIKeyService{pool: pool}
}

type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	Plaintext  string     `json:"plaintext,omitempty"`
}

type ListAPIKeysResponse struct {
	Items []APIKey `json:"items"`
}

type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required,min=2"`
	ExpiresIn string `json:"expiresIn,omitempty"`
}

type RotateAPIKeyRequest struct {
	Name      string `json:"name,omitempty"`
	ExpiresIn string `json:"expiresIn,omitempty"`
}

func (s *APIKeyService) List(ctx context.Context, actor ActorContext) (*ListAPIKeysResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, name, last_used_at, expires_at, revoked_at, created_at
		FROM api_keys
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
		ORDER BY created_at DESC
	`, actor.TenantID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	items := []APIKey{}
	for rows.Next() {
		item, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api keys: %w", err)
	}

	return &ListAPIKeysResponse{Items: items}, nil
}

func (s *APIKeyService) Create(ctx context.Context, actor ActorContext, req *CreateAPIKeyRequest) (*APIKey, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	plain, err := s.generateAPIKey()
	if err != nil {
		return nil, err
	}
	expiresAt, err := parseOptionalExpiry(req.ExpiresIn)
	if err != nil {
		return nil, err
	}

	row := s.pool.App.QueryRow(ctx, `
		INSERT INTO api_keys (user_id, tenant_id, name, key_hash, expires_at)
		VALUES ($1::uuid, $2::uuid, $3, digest($4, 'sha256'), $5)
		RETURNING id::text, name, last_used_at, expires_at, revoked_at, created_at
	`, actor.UserID, actor.TenantID, strings.TrimSpace(req.Name), plain, expiresAt)

	item, err := scanAPIKey(row)
	if err != nil {
		return nil, err
	}
	item.Plaintext = plain
	return item, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, actor ActorContext, id string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `
		UPDATE api_keys
		SET revoked_at = now()
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND user_id = $3::uuid AND revoked_at IS NULL
	`, id, actor.TenantID, actor.UserID)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

func (s *APIKeyService) Rotate(ctx context.Context, actor ActorContext, id string, req *RotateAPIKeyRequest) (*APIKey, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}

	plain, err := s.generateAPIKey()
	if err != nil {
		return nil, err
	}
	expiresAt, err := parseOptionalExpiry(req.ExpiresIn)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)

	row := s.pool.App.QueryRow(ctx, `
		UPDATE api_keys
		SET key_hash = digest($4, 'sha256'),
		    name = COALESCE(NULLIF($5, ''), name),
		    expires_at = COALESCE($6, expires_at),
		    revoked_at = NULL
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND user_id = $3::uuid
		RETURNING id::text, name, last_used_at, expires_at, revoked_at, created_at
	`, id, actor.TenantID, actor.UserID, plain, name, expiresAt)

	item, err := scanAPIKey(row)
	if err != nil {
		return nil, err
	}
	item.Plaintext = plain
	return item, nil
}

func (s *APIKeyService) generateAPIKey() (string, error) {
	plain, _, err := generateOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return "cp_" + plain, nil
}

func parseOptionalExpiry(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid expiresIn duration: %w", err)
	}
	t := time.Now().Add(d)
	return &t, nil
}

func scanAPIKey(row interface{ Scan(dest ...any) error }) (*APIKey, error) {
	var item APIKey
	if err := row.Scan(&item.ID, &item.Name, &item.LastUsedAt, &item.ExpiresAt, &item.RevokedAt, &item.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrHierarchyNotFound
		}
		return nil, fmt.Errorf("scan api key: %w", err)
	}
	return &item, nil
}
