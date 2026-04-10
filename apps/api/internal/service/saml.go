package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type SAMLService struct {
	pool *database.Pool
}

func NewSAMLService(pool *database.Pool) *SAMLService {
	return &SAMLService{pool: pool}
}

type SAMLProvider struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenantId"`
	Name            string    `json:"name"`
	EntityID        string    `json:"entityId"`
	SSOURL          string    `json:"ssoUrl"`
	MetadataURL     *string   `json:"metadataUrl,omitempty"`
	Enabled         bool      `json:"enabled"`
	JITProvisioning bool      `json:"jitProvisioning"`
	DefaultRole     string    `json:"defaultRole"`
	CreatedAt       time.Time `json:"createdAt"`
}

type SAMLProviderListResponse struct {
	Items []SAMLProvider `json:"items"`
}

type CreateSAMLProviderRequest struct {
	Name            string  `json:"name" binding:"required"`
	EntityID        string  `json:"entityId" binding:"required"`
	SSOURL          string  `json:"ssoUrl" binding:"required"`
	Certificate     string  `json:"certificate" binding:"required"`
	MetadataURL     *string `json:"metadataUrl,omitempty"`
	JITProvisioning bool    `json:"jitProvisioning,omitempty"`
	DefaultRole     string  `json:"defaultRole,omitempty"`
}

func (s *SAMLService) List(ctx context.Context, actor ActorContext) (*SAMLProviderListResponse, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	rows, err := s.pool.App.Query(ctx, `
		SELECT id::text, tenant_id::text, name, entity_id, sso_url, metadata_url, enabled, jit_provisioning, default_role, created_at
		FROM saml_providers WHERE tenant_id = $1::uuid ORDER BY created_at DESC
	`, actor.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list saml providers: %w", err)
	}
	defer rows.Close()
	items := []SAMLProvider{}
	for rows.Next() {
		var p SAMLProvider
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.EntityID, &p.SSOURL, &p.MetadataURL, &p.Enabled, &p.JITProvisioning, &p.DefaultRole, &p.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return &SAMLProviderListResponse{Items: items}, rows.Err()
}

func (s *SAMLService) Create(ctx context.Context, actor ActorContext, req *CreateSAMLProviderRequest) (*SAMLProvider, error) {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return nil, err
	}
	role := req.DefaultRole
	if role == "" {
		role = "viewer"
	}
	var p SAMLProvider
	err := s.pool.App.QueryRow(ctx, `
		INSERT INTO saml_providers (tenant_id, name, entity_id, sso_url, certificate, metadata_url, jit_provisioning, default_role)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, tenant_id::text, name, entity_id, sso_url, metadata_url, enabled, jit_provisioning, default_role, created_at
	`, actor.TenantID, req.Name, req.EntityID, req.SSOURL, req.Certificate, req.MetadataURL, req.JITProvisioning, role).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.EntityID, &p.SSOURL, &p.MetadataURL, &p.Enabled, &p.JITProvisioning, &p.DefaultRole, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create saml provider: %w", err)
	}
	return &p, nil
}

func (s *SAMLService) Delete(ctx context.Context, actor ActorContext, providerID string) error {
	if err := ensureActorMembership(ctx, s.pool, actor); err != nil {
		return err
	}
	cmd, err := s.pool.App.Exec(ctx, `DELETE FROM saml_providers WHERE id = $1::uuid AND tenant_id = $2::uuid`, providerID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("delete saml provider: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrHierarchyNotFound
	}
	return nil
}

var _ = errors.Is
var _ = pgx.ErrNoRows
