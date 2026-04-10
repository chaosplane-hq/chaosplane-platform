package service

import "testing"

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
