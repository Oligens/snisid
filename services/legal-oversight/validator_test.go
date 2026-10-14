package legaloversight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDecision_Approved(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-001",
		RiskScore:  0.5,
		Signals:    []string{"signal1", "signal2"},
		Confidence: 0.85,
	})
	assert.True(t, ok)
	assert.Equal(t, "APPROVED", reason)
}

func TestValidateDecision_LowConfidence(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-002",
		RiskScore:  0.6,
		Signals:    []string{"signal1", "signal2"},
		Confidence: 0.5,
	})
	assert.False(t, ok)
	assert.Equal(t, "LOW_CONFIDENCE", reason)
}

func TestValidateDecision_InsufficientEvidence(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-003",
		RiskScore:  0.6,
		Signals:    []string{"single"},
		Confidence: 0.8,
	})
	assert.False(t, ok)
	assert.Equal(t, "INSUFFICIENT_EVIDENCE", reason)
}

func TestValidateDecision_NoSignals(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-004",
		RiskScore:  0.6,
		Signals:    []string{},
		Confidence: 0.8,
	})
	assert.False(t, ok)
	assert.Equal(t, "INSUFFICIENT_EVIDENCE", reason)
}

func TestValidateDecision_NilSignals(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-005",
		Signals:    nil,
		Confidence: 0.8,
	})
	assert.False(t, ok)
	assert.Equal(t, "INSUFFICIENT_EVIDENCE", reason)
}

func TestValidateDecision_HighRiskLowConfidence(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-006",
		RiskScore:  0.95,
		Signals:    []string{"signal1", "signal2", "signal3"},
		Confidence: 0.8,
	})
	assert.False(t, ok)
	assert.Equal(t, "HIGH_RISK_REQUIRE_HIGHER_CONFIDENCE", reason)
}

func TestValidateDecision_HighRiskHighConfidence(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-007",
		RiskScore:  0.95,
		Signals:    []string{"signal1", "signal2", "signal3"},
		Confidence: 0.86,
	})
	assert.True(t, ok)
	assert.Equal(t, "APPROVED", reason)
}

func TestValidateDecision_BoundaryConfidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
		riskScore  float64
		approved   bool
	}{
		{"low_conf_at_0.69", 0.69, 0.5, false},
		{"low_conf_at_0.7", 0.7, 0.5, true},
		{"high_risk_at_0.84", 0.84, 0.91, false},
		{"high_risk_at_0.85", 0.85, 0.91, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateDecision(Decision{
				SubjectID:  "CIT-TEST",
				RiskScore:  tt.riskScore,
				Signals:    []string{"s1", "s2"},
				Confidence: tt.confidence,
			})
			assert.Equal(t, tt.approved, ok)
		})
	}
}

func TestValidateDecision_ZeroRiskHighConfidence(t *testing.T) {
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-008",
		RiskScore:  0.0,
		Signals:    []string{"signal1", "signal2"},
		Confidence: 0.95,
	})
	assert.True(t, ok)
	assert.Equal(t, "APPROVED", reason)
}

func TestValidateDecision_ManySignals(t *testing.T) {
	signals := make([]string, 100)
	for i := range signals {
		signals[i] = "sig"
	}
	ok, reason := ValidateDecision(Decision{
		SubjectID:  "CIT-009",
		RiskScore:  0.3,
		Signals:    signals,
		Confidence: 0.9,
	})
	assert.True(t, ok)
	assert.Equal(t, "APPROVED", reason)
}
