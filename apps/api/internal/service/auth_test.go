package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
)

func TestSlugify(t *testing.T) {
	if got := slugify("  Chaos Plane, Inc. "); got != "chaos-plane-inc" {
		t.Fatalf("unexpected slug: %s", got)
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	a := hashToken("token-123")
	b := hashToken("token-123")
	if a == "" || a != b {
		t.Fatalf("expected deterministic token hash")
	}
}

func TestCSRFTokenRoundTrip(t *testing.T) {
	svc := NewAuthService(nil, &config.Config{CSRFSecret: "csrf-secret"}, nil)
	token, err := svc.GenerateCSRFToken("user-1", "tenant-1")
	if err != nil {
		t.Fatalf("generate csrf token: %v", err)
	}
	if err := svc.ValidateCSRFToken("user-1", "tenant-1", token); err != nil {
		t.Fatalf("validate csrf token: %v", err)
	}
}

func TestSignAndParseAccessToken(t *testing.T) {
	claims := &AccessTokenClaims{
		TenantID:  "tenant-1",
		Email:     "user@example.com",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "11111111-1111-1111-1111-111111111111",
			Subject:   "user-1",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token, err := signToken(claims, "jwt-secret")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	parsed, err := parseSignedToken(token, "jwt-secret")
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if parsed.Subject != claims.Subject || parsed.TenantID != claims.TenantID {
		t.Fatalf("unexpected parsed claims: %+v", parsed)
	}
}

func TestDefaultSlug(t *testing.T) {
	if got := defaultSlug("", "My Workspace"); got != "my-workspace" {
		t.Fatalf("unexpected default slug: %s", got)
	}
}
