package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var ErrOAuthProviderDisabled = errors.New("oauth provider not configured")

type OAuthService struct {
	pool      *database.Pool
	cfg       *config.Config
	rdb       *redis.Client
	auth      *AuthService
	providers map[string]*oauth2.Config
}

func NewOAuthService(pool *database.Pool, cfg *config.Config, rdb *redis.Client, auth *AuthService) *OAuthService {
	svc := &OAuthService{pool: pool, cfg: cfg, rdb: rdb, auth: auth, providers: make(map[string]*oauth2.Config)}

	callbackBase := strings.TrimRight(cfg.FrontendURL, "/")

	if cfg.GoogleClientID != "" {
		svc.providers["google"] = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			RedirectURL: callbackBase + "/auth/callback?provider=google",
			Scopes:      []string{"openid", "email", "profile"},
		}
	}

	if cfg.GitHubClientID != "" {
		svc.providers["github"] = &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
			RedirectURL: callbackBase + "/auth/callback?provider=github",
			Scopes:      []string{"user:email"},
		}
	}

	if cfg.MicrosoftClientID != "" {
		svc.providers["microsoft"] = &oauth2.Config{
			ClientID:     cfg.MicrosoftClientID,
			ClientSecret: cfg.MicrosoftClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
			RedirectURL: callbackBase + "/auth/callback?provider=microsoft",
			Scopes:      []string{"openid", "email", "profile"},
		}
	}

	return svc
}

type OAuthAuthorizeResponse struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

type OAuthCallbackRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

type oauthUserInfo struct {
	Email      string
	Name       string
	AvatarURL  string
	Provider   string
	ProviderID string
}

func (s *OAuthService) Authorize(ctx context.Context, provider string) (*OAuthAuthorizeResponse, error) {
	cfg, ok := s.providers[strings.ToLower(provider)]
	if !ok {
		return nil, ErrOAuthProviderDisabled
	}

	state, err := generateState()
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		s.rdb.Set(ctx, oauthStateKey(state), "1", 10*time.Minute)
	}

	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return &OAuthAuthorizeResponse{AuthURL: url, State: state}, nil
}

func (s *OAuthService) Callback(ctx context.Context, req *OAuthCallbackRequest, ip net.IP, userAgent string) (*AuthResponse, error) {
	providerName := strings.ToLower(req.Provider)
	cfg, ok := s.providers[providerName]
	if !ok {
		return nil, ErrOAuthProviderDisabled
	}

	if s.rdb != nil {
		key := oauthStateKey(req.State)
		res, err := s.rdb.GetDel(ctx, key).Result()
		if err != nil || res == "" {
			return nil, ErrInvalidCredentials
		}
	}

	token, err := cfg.Exchange(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("oauth token exchange: %w", err)
	}

	userInfo, err := fetchOAuthUserInfo(ctx, providerName, token)
	if err != nil {
		return nil, err
	}

	return s.findOrCreateUser(ctx, userInfo, ip, userAgent)
}

func (s *OAuthService) findOrCreateUser(ctx context.Context, info *oauthUserInfo, ip net.IP, userAgent string) (*AuthResponse, error) {
	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin oauth user tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var ident authIdentity
	err = tx.QueryRow(ctx, `
		SELECT
			u.id::text, u.email, u.name, COALESCE(u.password_hash, ''),
			u.email_verified, COALESCE(u.accepted_tos_version, ''), u.status,
			u.failed_login_count, u.locked_until, u.last_login_at,
			t.id::text, t.name, t.slug, ut.default_org_id::text
		FROM users u
		JOIN user_tenants ut ON ut.user_id = u.id
		JOIN tenants t ON t.id = ut.tenant_id
		WHERE lower(u.email) = lower($1) AND u.deleted_at IS NULL
		ORDER BY ut.joined_at ASC
		LIMIT 1
	`, strings.ToLower(info.Email)).Scan(
		&ident.UserID, &ident.Email, &ident.Name, &ident.PasswordHash,
		&ident.EmailVerified, &ident.AcceptedTOSVersion, &ident.Status,
		&ident.FailedLogins, &ident.LockedUntil, &ident.LastLoginAt,
		&ident.TenantID, &ident.TenantName, &ident.TenantSlug, &ident.DefaultOrgID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		ident, err = s.createOAuthUser(ctx, tx, info)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("lookup oauth user: %w", err)
	}

	if !ident.EmailVerified {
		if _, err := tx.Exec(ctx, `
			UPDATE users SET email_verified = true, email_verified_at = now() WHERE id = $1::uuid
		`, ident.UserID); err != nil {
			return nil, fmt.Errorf("mark oauth user email verified: %w", err)
		}
		ident.EmailVerified = true
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users SET last_login_at = now(), last_login_ip = $2 WHERE id = $1::uuid
	`, ident.UserID, ip); err != nil {
		return nil, fmt.Errorf("update oauth login metadata: %w", err)
	}

	resp, err := s.auth.issueSession(ctx, tx, ident, "", ip, userAgent)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit oauth user tx: %w", err)
	}
	return resp, nil
}

func (s *OAuthService) createOAuthUser(ctx context.Context, tx pgx.Tx, info *oauthUserInfo) (authIdentity, error) {
	baseName := strings.TrimSpace(info.Name)
	if baseName == "" {
		baseName = strings.Split(info.Email, "@")[0]
	}

	tenantSlug := uniqueSlug(baseName)
	var ident authIdentity
	var orgID string

	err := tx.QueryRow(ctx, `
		WITH new_user AS (
			INSERT INTO users (email, name, avatar_url, email_verified, email_verified_at)
			VALUES (lower($1), $2, NULLIF($3, ''), true, now())
			RETURNING id, email, name, email_verified, last_login_at
		), new_tenant AS (
			INSERT INTO tenants (name, slug)
			VALUES ($4, $5)
			RETURNING id, name, slug
		), new_org AS (
			INSERT INTO organizations (tenant_id, name, slug)
			SELECT id, $6, 'default' FROM new_tenant
			RETURNING id, tenant_id
		), link_user AS (
			INSERT INTO user_tenants (user_id, tenant_id, default_org_id)
			SELECT new_user.id, new_tenant.id, new_org.id
			FROM new_user, new_tenant, new_org
		), new_workspace AS (
			INSERT INTO workspaces (tenant_id, organization_id, name, slug)
			SELECT new_org.tenant_id, new_org.id, 'Default Workspace', 'default'
			FROM new_org
			RETURNING id, tenant_id
		), new_team AS (
			INSERT INTO teams (tenant_id, workspace_id, name, slug)
			SELECT new_workspace.tenant_id, new_workspace.id, 'Core Team', 'core-team'
			FROM new_workspace
			RETURNING id, tenant_id
		), new_member AS (
			INSERT INTO team_members (team_id, user_id, role)
			SELECT new_team.id, new_user.id, 'lead'
			FROM new_team, new_user
		), onboarding AS (
			INSERT INTO onboarding_progress (user_id, tenant_id, step_org_created, step_workspace_created, step_team_created)
			SELECT new_user.id, new_tenant.id, true, true, true
			FROM new_user, new_tenant
		)
		SELECT
			new_user.id::text, new_user.email, new_user.name, new_user.email_verified, new_user.last_login_at,
			new_tenant.id::text, new_tenant.name, new_tenant.slug, new_org.id::text
		FROM new_user, new_tenant, new_org
	`, info.Email, baseName, info.AvatarURL,
		fmt.Sprintf("%s Org", baseName), tenantSlug,
		fmt.Sprintf("%s Organization", baseName),
	).Scan(
		&ident.UserID, &ident.Email, &ident.Name, &ident.EmailVerified, &ident.LastLoginAt,
		&ident.TenantID, &ident.TenantName, &ident.TenantSlug, &orgID,
	)
	if err != nil {
		return authIdentity{}, fmt.Errorf("create oauth user with tenant hierarchy: %w", err)
	}
	ident.DefaultOrgID = &orgID
	return ident, nil
}

func fetchOAuthUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*oauthUserInfo, error) {
	switch provider {
	case "google":
		return fetchGoogleUserInfo(ctx, token)
	case "github":
		return fetchGitHubUserInfo(ctx, token)
	case "microsoft":
		return fetchMicrosoftUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported oauth provider: %s", provider)
	}
}

func fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*oauthUserInfo, error) {
	data, err := oauthGet(ctx, token, "https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	return &oauthUserInfo{
		Email:      getString(data, "email"),
		Name:       getString(data, "name"),
		AvatarURL:  getString(data, "picture"),
		Provider:   "google",
		ProviderID: getString(data, "id"),
	}, nil
}

func fetchGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*oauthUserInfo, error) {
	data, err := oauthGet(ctx, token, "https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	email := getString(data, "email")
	if email == "" {
		emails, err := oauthGetArray(ctx, token, "https://api.github.com/user/emails")
		if err == nil {
			for _, e := range emails {
				if m, ok := e.(map[string]interface{}); ok {
					if primary, _ := m["primary"].(bool); primary {
						email, _ = m["email"].(string)
						break
					}
				}
			}
		}
	}
	return &oauthUserInfo{
		Email:      email,
		Name:       getString(data, "name"),
		AvatarURL:  getString(data, "avatar_url"),
		Provider:   "github",
		ProviderID: fmt.Sprintf("%v", data["id"]),
	}, nil
}

func fetchMicrosoftUserInfo(ctx context.Context, token *oauth2.Token) (*oauthUserInfo, error) {
	data, err := oauthGet(ctx, token, "https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, err
	}
	email := getString(data, "mail")
	if email == "" {
		email = getString(data, "userPrincipalName")
	}
	return &oauthUserInfo{
		Email:      email,
		Name:       getString(data, "displayName"),
		Provider:   "microsoft",
		ProviderID: getString(data, "id"),
	}, nil
}

func oauthGet(ctx context.Context, token *oauth2.Token, url string) (map[string]interface{}, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("oauth api request to %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read oauth response from %s: %w", url, err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parse oauth response from %s: %w", url, err)
	}
	return data, nil
}

func oauthGetArray(ctx context.Context, token *oauth2.Token, url string) ([]interface{}, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data []interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func oauthStateKey(state string) string {
	return "oauth:state:" + state
}
