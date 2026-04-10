package service

import "testing"

func TestQuickSetupResponseFields(t *testing.T) {
	resp := &QuickSetupResponse{OrganizationID: "org-1", WorkspaceID: "ws-1", TeamID: "team-1"}
	if resp.OrganizationID == "" || resp.WorkspaceID == "" || resp.TeamID == "" {
		t.Fatalf("expected quick setup ids to be populated")
	}
}
