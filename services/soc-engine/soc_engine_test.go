package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve_HighRisk_TriggersQuarantine(t *testing.T) {
	r := &IncidentResolver{}
	incident := Incident{
		ID:        "inc-001",
		Type:      "MALWARE",
		Risk:      0.95,
		TargetSvc: "identity-api",
	}
	action := r.Resolve(incident)
	assert.Contains(t, action, "ISOLATE_SERVICE")
	assert.Contains(t, action, "identity-api")
}

func TestResolve_DeepFakeAttempt(t *testing.T) {
	r := &IncidentResolver{}
	incident := Incident{
		ID:        "inc-002",
		Type:      "DEEP_FAKE_ATTEMPT",
		Risk:      0.5,
		TargetSvc: "biometrics",
	}
	action := r.Resolve(incident)
	assert.Equal(t, "ACTION: BLOCK_IDENTITY_AUTHENTICATION", action)
}

func TestResolve_HighRiskDeepFake_PrioritizesQuarantine(t *testing.T) {
	r := &IncidentResolver{}
	incident := Incident{
		ID:        "inc-003",
		Type:      "DEEP_FAKE_ATTEMPT",
		Risk:      0.95,
		TargetSvc: "biometrics",
	}
	action := r.Resolve(incident)
	// High risk overrides deep fake check
	assert.Contains(t, action, "ISOLATE_SERVICE")
}

func TestResolve_LowRiskDefault(t *testing.T) {
	r := &IncidentResolver{}
	incident := Incident{
		ID:        "inc-004",
		Type:      "SUSPICIOUS_LOGIN",
		Risk:      0.3,
		TargetSvc: "auth-api",
	}
	action := r.Resolve(incident)
	assert.Equal(t, "ACTION: INCREASED_MONITORING", action)
}

func TestResolve_MediumRisk(t *testing.T) {
	r := &IncidentResolver{}
	incident := Incident{
		ID:        "inc-005",
		Type:      "BRUTE_FORCE",
		Risk:      0.7,
		TargetSvc: "auth-api",
	}
	action := r.Resolve(incident)
	assert.Equal(t, "ACTION: INCREASED_MONITORING", action)
}

func TestResolve_AtRiskThreshold(t *testing.T) {
	r := &IncidentResolver{}

	incident := Incident{
		ID:        "inc-006",
		Type:      "PORT_SCAN",
		Risk:      0.9,
		TargetSvc: "network",
	}
	action := r.Resolve(incident)
	assert.Contains(t, action, "ISOLATE_SERVICE")

	incident.Risk = 0.9000001
	action = r.Resolve(incident)
	assert.Contains(t, action, "ISOLATE_SERVICE")
}

func TestResolve_EmptyTargetService(t *testing.T) {
	r := &IncidentResolver{}
	incident := Incident{
		ID:        "inc-007",
		Type:      "MALWARE",
		Risk:      0.95,
		TargetSvc: "",
	}
	action := r.Resolve(incident)
	assert.Contains(t, action, "ISOLATE_SERVICE_")
}

func TestTriggerQuarantine(t *testing.T) {
	r := &IncidentResolver{}
	action := r.triggerQuarantine("payment-gateway")
	assert.Equal(t, "ACTION: ISOLATE_SERVICE_payment-gateway_VIA_ISTIO_DENY_POLICY", action)
}

func TestTriggerQuarantine_EmptyService(t *testing.T) {
	r := &IncidentResolver{}
	action := r.triggerQuarantine("")
	assert.Contains(t, action, "ISOLATE_SERVICE_")
}

func TestResolve_TableDriven(t *testing.T) {
	r := &IncidentResolver{}

	tests := []struct {
		name     string
		incident Incident
		contains []string
	}{
		{
			name:     "critical risk quarantine",
			incident: Incident{ID: "t1", Type: "INTRUSION", Risk: 1.0, TargetSvc: "core"},
			contains: []string{"ISOLATE_SERVICE"},
		},
		{
			name:     "deep fake response",
			incident: Incident{ID: "t2", Type: "DEEP_FAKE_ATTEMPT", Risk: 0.5, TargetSvc: "biometrics"},
			contains: []string{"BLOCK_IDENTITY_AUTHENTICATION"},
		},
		{
			name:     "default monitoring",
			incident: Incident{ID: "t3", Type: "INFO", Risk: 0.1, TargetSvc: "audit"},
			contains: []string{"INCREASED_MONITORING"},
		},
		{
			name:     "high risk with deep fake type",
			incident: Incident{ID: "t4", Type: "DEEP_FAKE_ATTEMPT", Risk: 0.91, TargetSvc: "face-api"},
			contains: []string{"ISOLATE_SERVICE"},
		},
		{
			name:     "zero risk",
			incident: Incident{ID: "t5", Type: "HEARTBEAT", Risk: 0.0, TargetSvc: "monitor"},
			contains: []string{"INCREASED_MONITORING"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action := r.Resolve(tc.incident)
			for _, s := range tc.contains {
				assert.Contains(t, action, s)
			}
		})
	}
}
