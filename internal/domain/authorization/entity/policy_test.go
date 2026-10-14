package entity

import (
	"testing"
	"time"
)

func TestPolicy_Defaults(t *testing.T) {
	p := Policy{
		ID:      "pol-001",
		Name:    "abac_rules",
		Module:  `package snisid.abac\nallow { true }`,
		Enabled: true,
		Version: 1,
	}
	if !p.Enabled {
		t.Error("Policy should be enabled")
	}
	if p.Version != 1 {
		t.Errorf("Version = %d, want 1", p.Version)
	}
}

func TestPolicy_Disabled(t *testing.T) {
	p := Policy{
		ID:      "pol-002",
		Name:    "legacy_rules",
		Enabled: false,
	}
	if p.Enabled {
		t.Error("Policy should be disabled")
	}
}

func TestPolicy_VersionIncrement(t *testing.T) {
	p := Policy{
		ID:      "pol-003",
		Name:    "soc_rules",
		Version: 3,
		Enabled: true,
	}
	p.Version++
	if p.Version != 4 {
		t.Errorf("Version after increment = %d, want 4", p.Version)
	}
}

func TestPolicy_Timestamps(t *testing.T) {
	now := time.Now().UTC()
	p := Policy{
		ID:        "pol-004",
		Name:      "fraud_rules",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if p.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	p.UpdatedAt = time.Now().UTC()
	if !p.UpdatedAt.After(p.CreatedAt) && !p.UpdatedAt.Equal(p.CreatedAt) {
		t.Error("UpdatedAt should be >= CreatedAt")
	}
}

func TestRoleGrant_Values(t *testing.T) {
	g := RoleGrant{
		ID:       "grant-001",
		Role:     "admin",
		Action:   "write",
		Resource: "identities:*",
	}
	if g.Role != "admin" {
		t.Errorf("Role = %s, want admin", g.Role)
	}
	if g.Action != "write" {
		t.Errorf("Action = %s, want write", g.Action)
	}
	if g.Resource != "identities:*" {
		t.Errorf("Resource = %s, want identities:*", g.Resource)
	}
}

func TestAuthorizationRequest_Construction(t *testing.T) {
	req := AuthorizationRequest{
		Subject: SubjectData{
			UserID:    "usr-001",
			Roles:     []string{"officer", "investigator"},
			Agency:    "pnh",
			Clearance: "secret",
		},
		Action:   "read",
		Resource: "identity:NNU-123",
		Attributes: map[string]interface{}{
			"ip": "10.0.0.1",
		},
	}
	if req.Subject.UserID != "usr-001" {
		t.Errorf("Subject.UserID = %s, want usr-001", req.Subject.UserID)
	}
	if len(req.Subject.Roles) != 2 {
		t.Errorf("Roles count = %d, want 2", len(req.Subject.Roles))
	}
	if req.Subject.Clearance != "secret" {
		t.Errorf("Clearance = %s, want secret", req.Subject.Clearance)
	}
}

func TestAuthorizationDecision_Allowed(t *testing.T) {
	d := AuthorizationDecision{
		Allowed:    true,
		PolicyName: "abac_rules",
	}
	if !d.Allowed {
		t.Error("Decision should be allowed")
	}
	if d.Reason != "" {
		t.Errorf("Reason should be empty for allowed, got %s", d.Reason)
	}
}

func TestAuthorizationDecision_Denied(t *testing.T) {
	d := AuthorizationDecision{
		Allowed: false,
		Reason:  "Insufficient clearance",
	}
	if d.Allowed {
		t.Error("Decision should be denied")
	}
	if d.Reason != "Insufficient clearance" {
		t.Errorf("Reason = %s, want 'Insufficient clearance'", d.Reason)
	}
}

func TestSubjectData_EmptyRoles(t *testing.T) {
	s := SubjectData{
		UserID: "usr-002",
	}
	if s.Roles != nil {
		t.Error("Roles should be nil for new subject")
	}
}

func TestAuthorizationRequest_NilAttributes(t *testing.T) {
	req := AuthorizationRequest{
		Subject:  SubjectData{UserID: "usr-003"},
		Action:   "delete",
		Resource: "enrollment:ID-456",
	}
	if req.Attributes != nil {
		t.Error("Attributes should be nil if not set")
	}
}
