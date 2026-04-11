package service

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/jackc/pgx/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

type SAMLService struct {
	pool *database.Pool
	cfg  *config.Config
}

func NewSAMLService(pool *database.Pool, cfg *config.Config) *SAMLService {
	return &SAMLService{pool: pool, cfg: cfg}
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

type SAMLLoginResponse struct {
	AuthURL string `json:"authUrl"`
}

type SAMLAssertionResult struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	TenantID   string `json:"tenantId"`
	ProviderID string `json:"providerId"`
	Role       string `json:"role"`
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

	if _, err := parseCertificate(req.Certificate); err != nil {
		return nil, fmt.Errorf("invalid certificate: %w", err)
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

func (s *SAMLService) InitiateLogin(ctx context.Context, tenantID, providerID string) (*SAMLLoginResponse, error) {
	var entityID, ssoURL string
	err := s.pool.App.QueryRow(ctx, `
		SELECT entity_id, sso_url
		FROM saml_providers
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND enabled = true
	`, providerID, tenantID).Scan(&entityID, &ssoURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHierarchyNotFound
		}
		return nil, fmt.Errorf("load saml provider: %w", err)
	}

	rootURL, _ := url.Parse(s.cfg.FrontendURL)
	acsURL := *rootURL
	acsURL.Path = "/auth/saml/acs"
	idpSSOURL, _ := url.Parse(ssoURL)

	sp := saml.ServiceProvider{
		EntityID: rootURL.String() + "/saml/metadata",
		AcsURL:   acsURL,
		IDPMetadata: &saml.EntityDescriptor{
			EntityID: entityID,
			IDPSSODescriptors: []saml.IDPSSODescriptor{
				{
					SingleSignOnServices: []saml.Endpoint{
						{Location: idpSSOURL.String(), Binding: saml.HTTPRedirectBinding},
					},
				},
			},
		},
	}

	authReq, _ := sp.MakeAuthenticationRequest(idpSSOURL.String(), saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	redirectURL, _ := authReq.Redirect("", &sp)

	return &SAMLLoginResponse{AuthURL: redirectURL.String()}, nil
}

func (s *SAMLService) ProcessAssertion(ctx context.Context, tenantID, providerID string, r *http.Request) (*SAMLAssertionResult, error) {
	var entityID, ssoURL, certificate, defaultRole string
	err := s.pool.App.QueryRow(ctx, `
		SELECT entity_id, sso_url, certificate, default_role
		FROM saml_providers
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND enabled = true
	`, providerID, tenantID).Scan(&entityID, &ssoURL, &certificate, &defaultRole)
	if err != nil {
		return nil, fmt.Errorf("load saml provider for assertion: %w", err)
	}

	cert, err := parseCertificate(certificate)
	if err != nil {
		return nil, fmt.Errorf("parse idp certificate: %w", err)
	}

	rootURL, _ := url.Parse(s.cfg.FrontendURL)
	acsURL := *rootURL
	acsURL.Path = "/auth/saml/acs"

	idpMetadata := &saml.EntityDescriptor{
		EntityID: entityID,
		IDPSSODescriptors: []saml.IDPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						KeyDescriptors: []saml.KeyDescriptor{
							{
								Use: "signing",
								KeyInfo: saml.KeyInfo{
									X509Data: saml.X509Data{
										X509Certificates: []saml.X509Certificate{
											{Data: certificate},
										},
									},
								},
							},
						},
					},
				},
				SingleSignOnServices: []saml.Endpoint{
					{Location: ssoURL, Binding: saml.HTTPRedirectBinding},
				},
			},
		},
	}

	sp := saml.ServiceProvider{
		EntityID:    rootURL.String() + "/saml/metadata",
		AcsURL:      acsURL,
		IDPMetadata: idpMetadata,
		Certificate: cert,
	}

	assertion, err := sp.ParseResponse(r, []string{})
	if err != nil {
		return nil, fmt.Errorf("parse saml response: %w", err)
	}

	email := ""
	name := ""
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			switch attr.FriendlyName {
			case "email", "mail":
				if len(attr.Values) > 0 {
					email = attr.Values[0].Value
				}
			case "displayName", "name":
				if len(attr.Values) > 0 {
					name = attr.Values[0].Value
				}
			}
			switch attr.Name {
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
				if len(attr.Values) > 0 && email == "" {
					email = attr.Values[0].Value
				}
			case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
				if len(attr.Values) > 0 && name == "" {
					name = attr.Values[0].Value
				}
			}
		}
	}

	if email == "" && assertion.Subject != nil && assertion.Subject.NameID != nil {
		email = assertion.Subject.NameID.Value
	}

	return &SAMLAssertionResult{
		Email:      email,
		Name:       name,
		TenantID:   tenantID,
		ProviderID: providerID,
		Role:       defaultRole,
	}, nil
}

func parseCertificate(certData string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certData))
	if block != nil {
		return x509.ParseCertificate(block.Bytes)
	}
	return x509.ParseCertificate([]byte(certData))
}
