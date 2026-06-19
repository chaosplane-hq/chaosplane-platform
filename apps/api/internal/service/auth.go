package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
	ErrAccountLocked       = errors.New("account is locked")
	ErrEmailNotVerified    = errors.New("email is not verified")
	ErrTokenExpired        = errors.New("token expired")
	ErrRefreshReuse        = errors.New("refresh token reuse detected")
	ErrCSRFInvalid         = errors.New("invalid csrf token")
)

type AuthService struct {
	pool  *database.Pool
	cfg   *config.Config
	rdb   *redis.Client
	email *EmailService
}

func NewAuthService(pool *database.Pool, cfg *config.Config, rdb *redis.Client, email *EmailService) *AuthService {
	return &AuthService{pool: pool, cfg: cfg, rdb: rdb, email: email}
}

type RegisterRequest struct {
	Email              string `json:"email" binding:"required,email"`
	Password           string `json:"password" binding:"required,min=8"`
	Name               string `json:"name" binding:"required,min=2"`
	AcceptedTOSVersion string `json:"acceptedTosVersion,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type CurrentUserRequest struct {
	UserID   string
	TenantID string
}

type AuthUser struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	EmailVerified bool       `json:"emailVerified"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty"`
}

type AuthTenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AuthResponse struct {
	AccessToken           string     `json:"accessToken"`
	RefreshToken          string     `json:"refreshToken"`
	ExpiresIn             int64      `json:"expiresIn"`
	User                  AuthUser   `json:"user"`
	Tenant                AuthTenant `json:"tenant"`
	VerificationToken     string     `json:"verificationToken,omitempty"`
	VerificationExpiresAt string     `json:"verificationExpiresAt,omitempty"`
}

type CurrentUserResponse struct {
	User      AuthUser   `json:"user"`
	Tenant    AuthTenant `json:"tenant"`
	CSRFToken string     `json:"csrfToken"`
}

type VerificationTokenResponse struct {
	VerificationToken string `json:"verificationToken,omitempty"`
	ExpiresAt         string `json:"expiresAt"`
}

type PasswordResetTokenResponse struct {
	ResetToken string `json:"resetToken,omitempty"`
	ExpiresAt  string `json:"expiresAt"`
}

type AccessTokenClaims struct {
	TenantID  string `json:"tid"`
	Email     string `json:"email"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

type authIdentity struct {
	UserID             string
	Email              string
	Name               string
	PasswordHash       string
	EmailVerified      bool
	AcceptedTOSVersion string
	Status             string
	FailedLogins       int
	LockedUntil        *time.Time
	LastLoginAt        *time.Time
	TenantID           string
	TenantName         string
	TenantSlug         string
	DefaultOrgID       *string
}

type storedRefreshToken struct {
	ID        string
	UserID    string
	TenantID  string
	FamilyID  string
	ExpiresAt time.Time
	RevokedAt *time.Time
	ParentID  *string
	User      authIdentity
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest, ip net.IP, userAgent string) (*AuthResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	baseName := strings.TrimSpace(req.Name)
	if baseName == "" {
		return nil, fmt.Errorf("name is required")
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var existingUserID string
	err = tx.QueryRow(ctx, `
		SELECT id::text
		FROM users
		WHERE lower(email) = lower($1) AND deleted_at IS NULL
	`, email).Scan(&existingUserID)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	tenantName := fmt.Sprintf("%s Org", baseName)
	tenantSlug := uniqueSlug(baseName)
	orgName := fmt.Sprintf("%s Organization", baseName)
	workspaceName := "Default Workspace"
	teamName := "Core Team"

	var ident authIdentity
	var orgID string
	err = tx.QueryRow(ctx, `
		WITH new_user AS (
			INSERT INTO users (email, name, password_hash, accepted_tos_version)
			VALUES ($1, $2, $3, NULLIF($4, ''))
			RETURNING id, email, name, email_verified, last_login_at
		), new_tenant AS (
			INSERT INTO tenants (name, slug)
			VALUES ($5, $6)
			RETURNING id, name, slug
		), new_org AS (
			INSERT INTO organizations (tenant_id, name, slug)
			SELECT id, $7, 'default' FROM new_tenant
			RETURNING id, tenant_id
		), link_user AS (
			INSERT INTO user_tenants (user_id, tenant_id, default_org_id)
			SELECT new_user.id, new_tenant.id, new_org.id
			FROM new_user, new_tenant, new_org
			RETURNING tenant_id, default_org_id
		), new_workspace AS (
			INSERT INTO workspaces (tenant_id, organization_id, name, slug)
			SELECT new_org.tenant_id, new_org.id, $8, 'default'
			FROM new_org
			RETURNING id, tenant_id
		), new_team AS (
			INSERT INTO teams (tenant_id, workspace_id, name, slug)
			SELECT new_workspace.tenant_id, new_workspace.id, $9, 'core-team'
			FROM new_workspace
			RETURNING id, tenant_id
		), new_member AS (
			INSERT INTO team_members (team_id, user_id, role)
			SELECT new_team.id, new_user.id, 'lead'
			FROM new_team, new_user
		), onboarding AS (
			INSERT INTO onboarding_progress (
				user_id, tenant_id, step_org_created, step_workspace_created, step_team_created
			)
			SELECT new_user.id, new_tenant.id, true, true, true
			FROM new_user, new_tenant
		)
		SELECT
			new_user.id::text,
			new_user.email,
			new_user.name,
			new_user.email_verified,
			new_user.last_login_at,
			new_tenant.id::text,
			new_tenant.name,
			new_tenant.slug,
			new_org.id::text
		FROM new_user, new_tenant, new_org
	`, email, baseName, string(hashedPassword), req.AcceptedTOSVersion, tenantName, tenantSlug, orgName, workspaceName, teamName).
		Scan(
			&ident.UserID,
			&ident.Email,
			&ident.Name,
			&ident.EmailVerified,
			&ident.LastLoginAt,
			&ident.TenantID,
			&ident.TenantName,
			&ident.TenantSlug,
			&orgID,
		)
	if err != nil {
		return nil, fmt.Errorf("create user and default tenant hierarchy: %w", err)
	}
	ident.DefaultOrgID = &orgID

	resp, err := s.issueSession(ctx, tx, ident, "", ip, userAgent)
	if err != nil {
		return nil, err
	}

	verificationToken, expiresAt, err := s.createEmailVerificationToken(ctx, tx, ident.UserID)
	if err != nil {
		return nil, err
	}
	resp.VerificationToken = verificationToken
	resp.VerificationExpiresAt = expiresAt.Format(time.RFC3339)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit register tx: %w", err)
	}

	if s.email != nil {
		go s.email.SendVerificationEmail(context.Background(), email, verificationToken)
	}

	return resp, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest, ip net.IP, userAgent string) (*AuthResponse, error) {
	ident, err := s.lookupIdentityByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if ident.LockedUntil != nil && ident.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}
	if !ident.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	if ident.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(ident.PasswordHash), []byte(req.Password)) != nil {
		s.recordFailedLogin(ctx, ident.UserID)
		return nil, ErrInvalidCredentials
	}

	if err := s.clearFailedLogin(ctx, ident.UserID, ip); err != nil {
		return nil, err
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin login tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	resp, err := s.issueSession(ctx, tx, ident, "", ip, userAgent)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit login tx: %w", err)
	}

	return resp, nil
}

func (s *AuthService) Refresh(ctx context.Context, req *RefreshRequest, ip net.IP, userAgent string) (*AuthResponse, error) {
	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin refresh tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	stored, err := s.lookupRefreshToken(ctx, tx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if stored.ExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenInvalid
	}
	if stored.RevokedAt != nil {
		if err := s.revokeRefreshFamily(ctx, tx, stored.FamilyID); err != nil {
			return nil, err
		}
		return nil, ErrRefreshReuse
	}

	if _, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1::uuid`, stored.ID); err != nil {
		return nil, fmt.Errorf("revoke previous refresh token: %w", err)
	}

	resp, err := s.issueSession(ctx, tx, stored.User, stored.FamilyID, ip, userAgent, stored.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit refresh tx: %w", err)
	}

	return resp, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*PasswordResetTokenResponse, error) {
	var userID string
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT id::text
		FROM users
		WHERE lower(email) = lower($1) AND deleted_at IS NULL
	`, normalizeEmail(req.Email)).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &PasswordResetTokenResponse{ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339)}, nil
		}
		return nil, fmt.Errorf("lookup forgot-password user: %w", err)
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin forgot-password tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE user_id = $1::uuid AND used_at IS NULL
	`, userID); err != nil {
		return nil, fmt.Errorf("expire old reset tokens: %w", err)
	}

	plainToken, tokenHash, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate reset token: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if _, err := tx.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
	`, userID, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("store reset token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit forgot-password tx: %w", err)
	}

	if s.email != nil && userID != "" {
		go s.email.SendPasswordResetEmail(context.Background(), normalizeEmail(req.Email), plainToken)
	}

	return &PasswordResetTokenResponse{ResetToken: plainToken, ExpiresAt: expiresAt.Format(time.RFC3339)}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin reset-password tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id::text, expires_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL
	`, hashToken(req.Token)).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("lookup reset token: %w", err)
	}
	if expiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET password_hash = $2, failed_login_count = 0, locked_until = NULL, status = 'active'
		WHERE id = $1::uuid
	`, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE password_reset_tokens SET used_at = now() WHERE token_hash = $1 AND used_at IS NULL
	`, hashToken(req.Token)); err != nil {
		return fmt.Errorf("mark reset token used: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1::uuid AND revoked_at IS NULL
	`, userID); err != nil {
		return fmt.Errorf("revoke refresh tokens after password reset: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit reset-password tx: %w", err)
	}
	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) error {
	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin verify-email tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id::text, expires_at
		FROM email_verification_tokens
		WHERE token_hash = $1 AND used_at IS NULL
	`, hashToken(req.Token)).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("lookup verification token: %w", err)
	}
	if expiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET email_verified = true, email_verified_at = now()
		WHERE id = $1::uuid
	`, userID); err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE email_verification_tokens SET used_at = now() WHERE token_hash = $1 AND used_at IS NULL
	`, hashToken(req.Token)); err != nil {
		return fmt.Errorf("mark verification token used: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit verify-email tx: %w", err)
	}
	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, req *ResendVerificationRequest) (*VerificationTokenResponse, error) {
	ident, err := s.lookupIdentityByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return &VerificationTokenResponse{ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339)}, nil
		}
		return nil, err
	}
	if ident.EmailVerified {
		return &VerificationTokenResponse{ExpiresAt: time.Now().Format(time.RFC3339)}, nil
	}

	tx, err := s.pool.App.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin resend-verification tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	plain, expiresAt, err := s.createEmailVerificationToken(ctx, tx, ident.UserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit resend-verification tx: %w", err)
	}

	if s.email != nil {
		go s.email.SendVerificationEmail(context.Background(), normalizeEmail(req.Email), plain)
	}

	return &VerificationTokenResponse{VerificationToken: plain, ExpiresAt: expiresAt.Format(time.RFC3339)}, nil
}

func (s *AuthService) Logout(ctx context.Context, claims *AccessTokenClaims, req *LogoutRequest) error {
	if claims != nil && claims.ID != "" {
		expiresAt := claims.ExpiresAt.Time
		if expiresAt.After(time.Now()) {
			_, err := s.pool.Conn(ctx).Exec(ctx, `
				INSERT INTO jwt_blacklist (jti, user_id, reason, expires_at)
				VALUES ($1::uuid, $2::uuid, 'logout', $3)
				ON CONFLICT (jti) DO NOTHING
			`, claims.ID, claims.Subject, expiresAt)
			if err != nil {
				return fmt.Errorf("blacklist access token: %w", err)
			}
		}
	}

	if req.RefreshToken != "" {
		if _, err := s.pool.Conn(ctx).Exec(ctx, `
			UPDATE refresh_tokens
			SET revoked_at = now()
			WHERE token_hash = $1 AND revoked_at IS NULL
		`, hashToken(req.RefreshToken)); err != nil {
			return fmt.Errorf("revoke refresh token: %w", err)
		}
	}

	return nil
}

func (s *AuthService) CurrentUser(ctx context.Context, req *CurrentUserRequest) (*CurrentUserResponse, error) {
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT
			u.id::text,
			u.email,
			u.name,
			u.email_verified,
			u.last_login_at,
			t.id::text,
			t.name,
			t.slug
		FROM users u
		JOIN tenants t ON t.id = $2::uuid
		WHERE u.id = $1::uuid AND u.deleted_at IS NULL
	`, req.UserID, req.TenantID)

	var user AuthUser
	var tenant AuthTenant
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.EmailVerified, &user.LastLoginAt, &tenant.ID, &tenant.Name, &tenant.Slug); err != nil {
		return nil, fmt.Errorf("load current user: %w", err)
	}

	csrfToken, err := s.GenerateCSRFToken(user.ID, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &CurrentUserResponse{User: user, Tenant: tenant, CSRFToken: csrfToken}, nil
}

func (s *AuthService) ParseAccessToken(token string) (*AccessTokenClaims, error) {
	claims, err := parseSignedToken(token, s.cfg.JWTSecret)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" || claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}

func (s *AuthService) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	if s.pool == nil || s.pool.App == nil {
		return false, nil
	}
	if s.rdb != nil {
		res, err := s.rdb.Exists(ctx, blacklistRedisKey(jti)).Result()
		if err == nil && res > 0 {
			return true, nil
		}
	}
	var exists bool
	err := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM jwt_blacklist WHERE jti = $1::uuid AND expires_at > now()
		)
	`, jti).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return exists, nil
}

func (s *AuthService) GenerateCSRFToken(userID, tenantID string) (string, error) {
	data := strings.Join([]string{userID, tenantID}, ":")
	mac := hmac.New(sha256.New, []byte(s.cfg.CSRFSecret))
	if _, err := mac.Write([]byte(data)); err != nil {
		return "", fmt.Errorf("sign csrf token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString([]byte(data)) + "." + hex.EncodeToString(mac.Sum(nil)), nil
}

func (s *AuthService) ValidateCSRFToken(userID, tenantID, token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return ErrCSRFInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrCSRFInvalid
	}
	if string(decoded) != strings.Join([]string{userID, tenantID}, ":") {
		return ErrCSRFInvalid
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.CSRFSecret))
	if _, err := mac.Write(decoded); err != nil {
		return fmt.Errorf("validate csrf token: %w", err)
	}
	if !hmac.Equal([]byte(parts[1]), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return ErrCSRFInvalid
	}
	return nil
}

func (s *AuthService) lookupIdentityByEmail(ctx context.Context, email string) (authIdentity, error) {
	row := s.pool.Conn(ctx).QueryRow(ctx, `
		SELECT
			u.id::text,
			u.email,
			u.name,
			COALESCE(u.password_hash, ''),
			u.email_verified,
			COALESCE(u.accepted_tos_version, ''),
			u.status,
			u.failed_login_count,
			u.locked_until,
			u.last_login_at,
			t.id::text,
			t.name,
			t.slug,
			ut.default_org_id::text
		FROM users u
		JOIN user_tenants ut ON ut.user_id = u.id
		JOIN tenants t ON t.id = ut.tenant_id
		WHERE lower(u.email) = lower($1) AND u.deleted_at IS NULL
		ORDER BY ut.joined_at ASC
		LIMIT 1
	`, normalizeEmail(email))

	var ident authIdentity
	err := row.Scan(
		&ident.UserID,
		&ident.Email,
		&ident.Name,
		&ident.PasswordHash,
		&ident.EmailVerified,
		&ident.AcceptedTOSVersion,
		&ident.Status,
		&ident.FailedLogins,
		&ident.LockedUntil,
		&ident.LastLoginAt,
		&ident.TenantID,
		&ident.TenantName,
		&ident.TenantSlug,
		&ident.DefaultOrgID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authIdentity{}, ErrInvalidCredentials
		}
		return authIdentity{}, fmt.Errorf("lookup identity: %w", err)
	}

	return ident, nil
}

func (s *AuthService) issueSession(ctx context.Context, tx pgx.Tx, ident authIdentity, existingFamilyID string, ip net.IP, userAgent string, parentIDs ...string) (*AuthResponse, error) {
	accessToken, claims, err := s.generateAccessToken(ident)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	familyID := existingFamilyID
	if familyID == "" {
		familyID = uuid.NewString()
	}

	var parentID any
	if len(parentIDs) > 0 && parentIDs[0] != "" {
		parentID = parentIDs[0]
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (
			user_id, tenant_id, token_hash, family_id, parent_id, ip_address, user_agent, expires_at
		)
		VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7, $8)
	`, ident.UserID, ident.TenantID, refreshHash, familyID, parentID, ip, userAgent, time.Now().Add(s.cfg.RefreshTokenTTL))
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	user := AuthUser{
		ID:            ident.UserID,
		Email:         ident.Email,
		Name:          ident.Name,
		EmailVerified: ident.EmailVerified,
		LastLoginAt:   ident.LastLoginAt,
	}
	tenant := AuthTenant{ID: ident.TenantID, Name: ident.TenantName, Slug: ident.TenantSlug}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Until(claims.ExpiresAt.Time).Seconds()),
		User:         user,
		Tenant:       tenant,
	}, nil
}

func (s *AuthService) generateAccessToken(ident authIdentity) (string, *AccessTokenClaims, error) {
	now := time.Now()
	claims := &AccessTokenClaims{
		TenantID:  ident.TenantID,
		Email:     ident.Email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   ident.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
		},
	}
	signed, err := signToken(claims, s.cfg.JWTSecret)
	if err != nil {
		return "", nil, fmt.Errorf("sign access token: %w", err)
	}
	return signed, claims, nil
}

func (s *AuthService) lookupRefreshToken(ctx context.Context, tx pgx.Tx, plainToken string) (*storedRefreshToken, error) {
	row := tx.QueryRow(ctx, `
		SELECT
			rt.id::text,
			rt.user_id::text,
			rt.tenant_id::text,
			rt.family_id::text,
			rt.expires_at,
			rt.revoked_at,
			rt.parent_id::text,
			u.email,
			u.name,
			u.email_verified,
			u.last_login_at,
			t.name,
			t.slug
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		JOIN tenants t ON t.id = rt.tenant_id
		WHERE rt.token_hash = $1
	`, hashToken(plainToken))

	var token storedRefreshToken
	var user authIdentity
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TenantID,
		&token.FamilyID,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.ParentID,
		&user.Email,
		&user.Name,
		&user.EmailVerified,
		&user.LastLoginAt,
		&user.TenantName,
		&user.TenantSlug,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, fmt.Errorf("lookup refresh token: %w", err)
	}

	user.UserID = token.UserID
	user.TenantID = token.TenantID
	user.Email = normalizeEmail(user.Email)
	token.User = user
	return &token, nil
}

func (s *AuthService) recordFailedLogin(ctx context.Context, userID string) {
	if _, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE users
		SET
			failed_login_count = failed_login_count + 1,
			locked_until = CASE
				WHEN failed_login_count + 1 >= 20 THEN now() + interval '24 hours'
				WHEN failed_login_count + 1 >= 10 THEN now() + interval '30 minutes'
				WHEN failed_login_count + 1 >= 5 THEN now() + interval '5 minutes'
				ELSE locked_until
			END,
			status = CASE
				WHEN failed_login_count + 1 >= 20 THEN 'locked'
				ELSE status
			END
		WHERE id = $1::uuid
	`, userID); err != nil {
		slog.Error("failed to record login failure", "userId", userID, "error", err)
	}
}

func (s *AuthService) clearFailedLogin(ctx context.Context, userID string, ip net.IP) error {
	_, err := s.pool.Conn(ctx).Exec(ctx, `
		UPDATE users
		SET
			failed_login_count = 0,
			locked_until = NULL,
			status = CASE WHEN status = 'locked' THEN 'active' ELSE status END,
			last_login_at = now(),
			last_login_ip = $2
		WHERE id = $1::uuid
	`, userID, ip)
	if err != nil {
		return fmt.Errorf("clear failed login state: %w", err)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func uniqueSlug(input string) string {
	base := slugify(input)
	if base == "" {
		base = "workspace"
	}
	return fmt.Sprintf("%s-%s", base, strings.ToLower(uuid.NewString()[:8]))
}

func slugify(input string) string {
	v := strings.ToLower(strings.TrimSpace(input))
	v = slugSanitizer.ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")
	return v
}

func generateOpaqueToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	plain := hex.EncodeToString(raw)
	return plain, hashToken(plain), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) createEmailVerificationToken(ctx context.Context, tx pgx.Tx, userID string) (string, time.Time, error) {
	if _, err := tx.Exec(ctx, `
		UPDATE email_verification_tokens
		SET used_at = now()
		WHERE user_id = $1::uuid AND used_at IS NULL
	`, userID); err != nil {
		return "", time.Time{}, fmt.Errorf("expire old verification tokens: %w", err)
	}

	plainToken, tokenHash, err := generateOpaqueToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate verification token: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	if _, err := tx.Exec(ctx, `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
	`, userID, tokenHash, expiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("store verification token: %w", err)
	}
	return plainToken, expiresAt, nil
}

func (s *AuthService) revokeRefreshFamily(ctx context.Context, tx pgx.Tx, familyID string) error {
	if _, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, now())
		WHERE family_id = $1::uuid
	`, familyID); err != nil {
		return fmt.Errorf("revoke refresh family: %w", err)
	}
	return nil
}

func blacklistRedisKey(jti string) string {
	return "jwt:blacklist:" + jti
}

func signToken(claims *AccessTokenClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func parseSignedToken(tokenStr, secret string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}
