package optimizer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicyOptimizer(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	assert.Equal(t, "opt-1", o.ID)
	assert.Empty(t, o.outcomeHistory)
	assert.Equal(t, 100, o.minDataPoints)
	assert.Equal(t, 0.05, o.learningRate)
}

func TestSuggestOptimizations_InsufficientData(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	result := o.SuggestOptimizations(map[string]interface{}{
		"outcomes": []EnforcementOutcome{
			{PolicyID: "p1", Success: true, FraudPrevented: 50, Region: "HTI"},
		},
	})
	assert.Nil(t, result)
}

func TestSuggestOptimizations_HighFPTriggersUpdate(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	o.minDataPoints = 5

	outcomes := make([]EnforcementOutcome, 10)
	for i := range outcomes {
		outcomes[i] = EnforcementOutcome{
			PolicyID:       "p1",
			Action:         "flag",
			Success:        i < 5,
			FraudPrevented: 100,
			FalsePositive:  i >= 5,
			Region:         "HTI",
			Timestamp:      time.Now().Unix(),
		}
	}

	result := o.SuggestOptimizations(map[string]interface{}{
		"outcomes": outcomes,
	})
	require.NotEmpty(t, result)

	foundFP := false
	for _, u := range result {
		if u.Domain == "FRAUD_DETECTION" {
			foundFP = true
			assert.Contains(t, u.SuggestedDelta, "increase fraud detection threshold")
			assert.Equal(t, "MEDIUM", u.RiskLevel)
		}
	}
	assert.True(t, foundFP)
}

func TestSuggestOptimizations_LowSuccessRateTriggersRegionUpdate(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	o.minDataPoints = 5

	outcomes := make([]EnforcementOutcome, 10)
	for i := range outcomes {
		outcomes[i] = EnforcementOutcome{
			PolicyID:       "p1",
			Success:        false,
			FraudPrevented: 200,
			FalsePositive:  false,
			Region:         "DOM",
		}
	}
	// Mix in some from HTI
	for i := 0; i < 5; i++ {
		outcomes = append(outcomes, EnforcementOutcome{
			PolicyID:       "p2",
			Success:        true,
			FraudPrevented: 300,
			FalsePositive:  false,
			Region:         "HTI",
		})
	}

	result := o.SuggestOptimizations(map[string]interface{}{
		"outcomes": outcomes,
	})
	require.NotEmpty(t, result)

	foundRegion := false
	for _, u := range result {
		if u.Domain == "REGION_ENFORCEMENT" {
			foundRegion = true
			assert.Contains(t, u.SuggestedDelta, "DOM")
			assert.Equal(t, "HIGH", u.RiskLevel)
			assert.True(t, u.RequiresHuman)
		}
	}
	assert.True(t, foundRegion)
}

func TestSuggestOptimizations_LowFraudTriggersCostOpt(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	o.minDataPoints = 5

	outcomes := make([]EnforcementOutcome, 10)
	for i := range outcomes {
		outcomes[i] = EnforcementOutcome{
			PolicyID:       "p1",
			Success:        true,
			FraudPrevented: 50,
			Region:         "HTI",
		}
	}

	result := o.SuggestOptimizations(map[string]interface{}{
		"outcomes": outcomes,
	})

	foundCost := false
	for _, u := range result {
		if u.Domain == "COST_OPTIMIZATION" {
			foundCost = true
			assert.Equal(t, "LOW", u.RiskLevel)
			assert.False(t, u.RequiresHuman)
		}
	}
	assert.True(t, foundCost)
}

func TestSuggestOptimizations_EmptyHistory(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	assert.Nil(t, o.SuggestOptimizations(map[string]interface{}{}))
}

func TestValidateAgainstHumanGovernance_AutoApproved(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	update := PolicyUpdate{
		SuggestedDelta: "auto-adjust threshold",
		RequiresHuman:  false,
		RiskLevel:      "LOW",
	}
	assert.True(t, o.ValidateAgainstHumanGovernance(update))
}

func TestValidateAgainstHumanGovernance_RequiresHuman(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	update := PolicyUpdate{
		SuggestedDelta: "reevaluate enforcement",
		RequiresHuman:  true,
		RiskLevel:      "HIGH",
	}
	assert.False(t, o.ValidateAgainstHumanGovernance(update))
}

func TestRecordOutcome(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	o.RecordOutcome(EnforcementOutcome{
		PolicyID: "p1",
		Success:  true,
		Region:   "HTI",
	})
	assert.Equal(t, 1, len(o.outcomeHistory))
}

func TestRecordOutcome_ConcurrentSafe(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	for i := 0; i < 100; i++ {
		o.RecordOutcome(EnforcementOutcome{
			PolicyID: "p1",
			Success:  true,
			Region:   "HTI",
		})
	}
	assert.Equal(t, 100, len(o.outcomeHistory))
}

func TestGetStats(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	o.RecordOutcome(EnforcementOutcome{PolicyID: "p1", Region: "HTI"})

	stats := o.GetStats()
	assert.Equal(t, 1, stats["total_outcomes"])
	assert.Equal(t, 100, stats["min_data_points"])
	assert.Equal(t, 0.05, stats["learning_rate"])
}

func TestFormatRationale(t *testing.T) {
	r := formatRationale("FP", 0.15, 0.1)
	assert.Contains(t, r, "FP")
	assert.Contains(t, r, "15%")
	assert.Contains(t, r, "10%")
}

func TestFormatPercent(t *testing.T) {
	assert.Equal(t, "50%", formatPercent(0.5))
	assert.Equal(t, "100%", formatPercent(1.0))
	assert.Equal(t, "0%", formatPercent(0.0))
	assert.Equal(t, "12.5%", formatPercent(0.125))
}

func TestComputeOptimizations_AllRegions(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	outcomes := []EnforcementOutcome{
		{Region: "HTI", Success: true, FraudPrevented: 500, FalsePositive: false},
		{Region: "HTI", Success: true, FraudPrevented: 300, FalsePositive: false},
		{Region: "DOM", Success: false, FraudPrevented: 0, FalsePositive: true},
		{Region: "DOM", Success: false, FraudPrevented: 10, FalsePositive: true},
	}

	updates := o.computeOptimizations(outcomes)
	assert.NotEmpty(t, updates)
	assert.True(t, len(updates) >= 2)
}

func TestSortByConfidence(t *testing.T) {
	o := NewPolicyOptimizer("opt-1")
	updates := o.computeOptimizations([]EnforcementOutcome{
		{Region: "HTI", Success: true, FraudPrevented: 1000, FalsePositive: false},
		{Region: "DOM", Success: false, FraudPrevented: 0, FalsePositive: false},
	})
	for i := 1; i < len(updates); i++ {
		assert.GreaterOrEqual(t, updates[i-1].Confidence, updates[i].Confidence)
	}
}
