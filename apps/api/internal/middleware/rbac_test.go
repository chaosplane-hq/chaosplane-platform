package middleware

import "testing"

func TestNormalizeTenantRole(t *testing.T) {
	tests := map[string]string{
		"lead":     "admin",
		"admin":    "admin",
		"member":   "editor",
		"editor":   "editor",
		"operator": "editor",
		"viewer":   "viewer",
		"unknown":  "viewer",
	}

	for input, want := range tests {
		if got := normalizeTenantRole(input); got != want {
			t.Fatalf("role %q: got %q want %q", input, got, want)
		}
	}
}
