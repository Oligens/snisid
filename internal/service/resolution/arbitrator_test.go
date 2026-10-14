package resolution

import (
	"context"
	"testing"
)

func TestNewArbitrator(t *testing.T) {
	a := NewArbitrator()
	if a == nil {
		t.Fatal("NewArbitrator returned nil")
	}
	if len(a.trustMap) != 4 {
		t.Errorf("Trust map size = %d, want 4", len(a.trustMap))
	}
}

func TestResolve_HighestTrust(t *testing.T) {
	a := NewArbitrator()
	values := map[string]interface{}{
		"user_self_service": "123 Main St",
		"national_registry": "456 Oak Ave",
		"passport_office":   "456 Oak Ave",
	}

	value, source := a.Resolve(context.Background(), "address", values)
	if source != "national_registry" {
		t.Errorf("Source = %s, want national_registry (trust=100)", source)
	}
	if value != "456 Oak Ave" {
		t.Errorf("Value = %s, want 456 Oak Ave", value)
	}
}

func TestResolve_SingleSource(t *testing.T) {
	a := NewArbitrator()
	values := map[string]interface{}{
		"national_police": "case-file-123",
	}

	value, source := a.Resolve(context.Background(), "reference", values)
	if source != "national_police" {
		t.Errorf("Source = %s, want national_police", source)
	}
	if value != "case-file-123" {
		t.Errorf("Value = %s, want case-file-123", value)
	}
}

func TestResolve_EmptyValues(t *testing.T) {
	a := NewArbitrator()
	value, source := a.Resolve(context.Background(), "empty", map[string]interface{}{})
	if value != nil {
		t.Errorf("Value = %v, want nil", value)
	}
	if source != "" {
		t.Errorf("Source = %s, want empty", source)
	}
}

func TestResolve_LowestTrustChosen(t *testing.T) {
	a := NewArbitrator()
	values := map[string]interface{}{
		"user_self_service": "self-reported",
	}

	value, source := a.Resolve(context.Background(), "field", values)
	if source != "user_self_service" {
		t.Errorf("Source = %s, want user_self_service", source)
	}
	if value != "self-reported" {
		t.Errorf("Value = %s, want self-reported", value)
	}
}

func TestResolve_UnknownSource(t *testing.T) {
	a := NewArbitrator()
	values := map[string]interface{}{
		"unknown_agency": "some-value",
	}
	// Unknown sources have trust score 0
	_, source := a.Resolve(context.Background(), "test", values)
	if source != "unknown_agency" {
		t.Errorf("Source = %s, want unknown_agency", source)
	}
}
