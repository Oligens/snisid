package predictiveengine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForecastRisk_NoDrift(t *testing.T) {
	features := map[string]interface{}{
		"transaction_count": 100,
		"avg_amount":        50.0,
	}
	forecast := ForecastRisk("SUBJ-001", features)
	assert.Equal(t, "SUBJ-001", forecast.SubjectID)
	assert.InDelta(t, 0.15, forecast.FraudProbability, 0.001)
	assert.Equal(t, "STABLE", forecast.RiskTrend)
	assert.Equal(t, "7_days", forecast.TimeHorizon)
}

func TestForecastRisk_WithBehaviorDrift(t *testing.T) {
	features := map[string]interface{}{
		"behavior_drift":    0.7,
		"transaction_count": 5,
	}
	forecast := ForecastRisk("SUBJ-002", features)
	assert.InDelta(t, 0.55, forecast.FraudProbability, 0.001)
	assert.Equal(t, "INCREASING", forecast.RiskTrend)
}

func TestForecastRisk_BoundaryDrift(t *testing.T) {
	features := map[string]interface{}{
		"behavior_drift": 0.5,
	}
	forecast := ForecastRisk("SUBJ-003", features)
	assert.InDelta(t, 0.15, forecast.FraudProbability, 0.001)
	assert.Equal(t, "STABLE", forecast.RiskTrend)
}

func TestForecastRisk_JustAboveThreshold(t *testing.T) {
	features := map[string]interface{}{
		"behavior_drift": 0.51,
	}
	forecast := ForecastRisk("SUBJ-004", features)
	assert.InDelta(t, 0.55, forecast.FraudProbability, 0.001)
	assert.Equal(t, "INCREASING", forecast.RiskTrend)
}

func TestForecastRisk_EmptyFeatures(t *testing.T) {
	features := map[string]interface{}{}
	forecast := ForecastRisk("SUBJ-005", features)
	assert.Equal(t, "SUBJ-005", forecast.SubjectID)
	assert.InDelta(t, 0.15, forecast.FraudProbability, 0.001)
	assert.Equal(t, "STABLE", forecast.RiskTrend)
}

func TestForecastRisk_NilFeatures(t *testing.T) {
	forecast := ForecastRisk("SUBJ-006", nil)
	assert.Equal(t, "SUBJ-006", forecast.SubjectID)
	assert.InDelta(t, 0.15, forecast.FraudProbability, 0.001)
}

func TestForecastRisk_DriftNonFloat(t *testing.T) {
	features := map[string]interface{}{
		"behavior_drift": "high",
	}
	forecast := ForecastRisk("SUBJ-007", features)
	assert.Equal(t, 0.15, forecast.FraudProbability)
	assert.Equal(t, "STABLE", forecast.RiskTrend)
}

func TestForecastRisk_DriftIntValue(t *testing.T) {
	features := map[string]interface{}{
		"behavior_drift": 1,
	}
	forecast := ForecastRisk("SUBJ-008", features)
	assert.Equal(t, 0.15, forecast.FraudProbability)
	assert.Equal(t, "STABLE", forecast.RiskTrend)
}

func TestForecastRisk_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		subjectID string
		features  map[string]interface{}
		wantProb  float64
		wantTrend string
	}{
		{"normal activity", "U001", map[string]interface{}{"tx_count": 50}, 0.15, "STABLE"},
		{"behavior drift", "U002", map[string]interface{}{"behavior_drift": 0.8}, 0.55, "INCREASING"},
		{"drift at boundary", "U003", map[string]interface{}{"behavior_drift": 0.5}, 0.15, "STABLE"},
		{"drift just over", "U004", map[string]interface{}{"behavior_drift": 0.5001}, 0.55, "INCREASING"},
		{"max drift", "U005", map[string]interface{}{"behavior_drift": 1.0}, 0.55, "INCREASING"},
		{"string drift value", "U006", map[string]interface{}{"behavior_drift": "0.9"}, 0.15, "STABLE"},
		{"negative drift", "U007", map[string]interface{}{"behavior_drift": -0.1}, 0.15, "STABLE"},
		{"multiple features", "U008", map[string]interface{}{
			"behavior_drift": 0.6,
			"amount":         10000,
			"new_device":     true,
		}, 0.55, "INCREASING"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := ForecastRisk(tc.subjectID, tc.features)
			assert.Equal(t, tc.subjectID, f.SubjectID)
			assert.InDelta(t, tc.wantProb, f.FraudProbability, 0.001)
			assert.Equal(t, tc.wantTrend, f.RiskTrend)
			assert.Equal(t, "7_days", f.TimeHorizon)
		})
	}
}
