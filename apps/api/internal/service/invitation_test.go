package service

import "testing"

func TestNormalizeInvitationRole(t *testing.T) {
	tests := map[string]string{
		"admin":    "admin",
		"editor":   "editor",
		"member":   "editor",
		"operator": "editor",
		"viewer":   "viewer",
		"":         "viewer",
	}

	for input, want := range tests {
		if got := normalizeInvitationRole(input); got != want {
			t.Fatalf("role %q: got %q want %q", input, got, want)
		}
	}
}

func TestInvitationRoleToTeamRole(t *testing.T) {
	tests := map[string]string{
		"admin":  "lead",
		"editor": "member",
		"viewer": "viewer",
	}

	for input, want := range tests {
		if got := invitationRoleToTeamRole(input); got != want {
			t.Fatalf("invitation role %q: got %q want %q", input, got, want)
		}
	}
}
