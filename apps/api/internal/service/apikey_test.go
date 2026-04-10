package service

import "testing"

func TestParseOptionalExpiry(t *testing.T) {
	if _, err := parseOptionalExpiry("1h"); err != nil {
		t.Fatalf("expected valid duration, got %v", err)
	}
	if _, err := parseOptionalExpiry("nope"); err == nil {
		t.Fatalf("expected invalid duration error")
	}
}
