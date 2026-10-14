package ml

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMLService(t *testing.T) {
	s := NewMLService("v1.0")
	require.NotNil(t, s)
	assert.Equal(t, "v1.0", s.ModelVersion)
	assert.InDelta(t, 0.7, s.threshold, 0.001)
}

func TestPredict_NormalTransaction(t *testing.T) {
	s := NewMLService("v1.0")
	input := ModelInput{
		Amount:           50,
		Velocity:         2,
		HourOfDay:        14,
		DayOfWeek:        3,
		IsWeekend:        false,
		IsNewDevice:      false,
		IsNewLocation:    false,
		FailedAttempts:   0,
		AccountAgeDays:   365,
		TransactionCount: 100,
		AvgTransaction:   45,
	}
	score := s.Predict(input)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestPredict_HighRiskTransaction(t *testing.T) {
	s := NewMLService("v1.0")
	input := ModelInput{
		Amount:           100000,
		Velocity:         200,
		HourOfDay:        3,
		DayOfWeek:        6,
		IsWeekend:        true,
		IsNewDevice:      true,
		IsNewLocation:    true,
		FailedAttempts:   15,
		AccountAgeDays:   0,
		TransactionCount: 1,
		AvgTransaction:   0,
	}
	score := s.Predict(input)
	assert.Greater(t, score, 0.5)
}

func TestPredict_MidRiskTransaction(t *testing.T) {
	s := NewMLService("v1.0")
	input := ModelInput{
		Amount:           8000,
		Velocity:         30,
		HourOfDay:        22,
		IsNewDevice:      false,
		IsNewLocation:    true,
		FailedAttempts:   4,
		AccountAgeDays:   10,
		TransactionCount: 15,
		AvgTransaction:   200,
	}
	score := s.Predict(input)
	assert.Greater(t, score, 0.1)
	assert.Less(t, score, 0.9)
}

func TestPredict_ExtremeAmount(t *testing.T) {
	s := NewMLService("v1.0")
	input := ModelInput{
		Amount:           1e9,
		Velocity:         0,
		AvgTransaction:   100,
		TransactionCount: 50,
	}
	score := s.Predict(input)
	assert.Greater(t, score, 0.2)
}

func TestPredict_ZeroAvgTransaction(t *testing.T) {
	s := NewMLService("v1.0")
	input := ModelInput{
		Amount:           100,
		Velocity:         0,
		AvgTransaction:   0,
		TransactionCount: 0,
	}
	score := s.Predict(input)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestPredict_AccountAgeBands(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name     string
		days     int
		minScore float64
	}{
		{"less than 1 day", 0, 0.0},
		{"1 day", 1, 0.0},
		{"6 days", 6, 0.0},
		{"7 days", 7, 0.0},
		{"29 days", 29, 0.0},
		{"30 days", 30, 0.0},
		{"old account", 1000, 0.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := ModelInput{
				AccountAgeDays: tc.days,
				Amount:         10,
				Velocity:       1,
				AvgTransaction: 10,
			}
			score := s.Predict(input)
			assert.GreaterOrEqual(t, score, tc.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestEvaluateAmount(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name     string
		amount   float64
		avg      float64
		expected float64
	}{
		{"avg zero", 100, 0, 0.1},
		{"normal range", 100, 100, 0.05},
		{"1.5x avg", 160, 100, 0.2},
		{"3x avg", 350, 100, 0.5},
		{"5x avg", 600, 100, 0.8},
		{"10x avg", 1100, 100, 1.0},
		{">10x avg", 5000, 100, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.evaluateAmount(tc.amount, tc.avg)
			assert.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestEvaluateVelocity(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name     string
		velocity float64
		expected float64
	}{
		{"zero velocity", 0, 0},
		{"low velocity", 5, 0.1},
		{"moderate velocity", 15, 0.3},
		{"elevated velocity", 30, 0.5},
		{"high velocity", 75, 0.8},
		{"extreme velocity", 150, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.evaluateVelocity(tc.velocity)
			assert.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestEvaluateHourAnomaly(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name     string
		hour     int
		expected float64
	}{
		{"normal hours 10am", 10, 0.1},
		{"normal hours 2pm", 14, 0.1},
		{"late night 2am", 2, 0.7},
		{"early morning 6am", 6, 0.4},
		{"evening 11pm", 23, 0.4},
		{"midnight boundary", 0, 0.4},
		{"early morning 5am", 5, 0.7},
		{"normal 7am", 7, 0.1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.evaluateHourAnomaly(tc.hour)
			assert.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestEvaluateFailedAttempts(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name     string
		attempts int
		expected float64
	}{
		{"none", 0, 0},
		{"one attempt", 1, 0},
		{"two attempts", 2, 0.2},
		{"four attempts", 4, 0.4},
		{"six attempts", 6, 0.7},
		{"eleven attempts", 11, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.evaluateFailedAttempts(tc.attempts)
			assert.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestEvaluateTransactionDeviation(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name    string
		count   int
		avg     float64
		current float64
		exp     float64
	}{
		{"insufficient history", 3, 100, 500, 0.5},
		{"no deviation", 10, 100, 100, 0},
		{"1 std dev", 10, 100, 130, 0.3},
		{"2 std dev", 10, 100, 160, 0.7},
		{"3 std dev", 10, 100, 200, 1.0},
		{"avg is zero", 10, 0, 100, 0.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.evaluateTransactionDeviation(tc.count, tc.avg, tc.current)
			assert.InDelta(t, tc.exp, result, 0.001)
		})
	}
}

func TestAugmentRisk(t *testing.T) {
	s := NewMLService("v1.0")

	tests := []struct {
		name     string
		rule     float64
		ml       float64
		expected float64
	}{
		{"both zero", 0, 0, 0},
		{"rule only", 1.0, 0, 0.4},
		{"ml only", 0, 1.0, 0.6},
		{"both max", 1.0, 1.0, 1.0},
		{"mixed", 0.5, 0.5, 0.5},
		{"rule high ml low", 0.8, 0.2, 0.44},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.AugmentRisk(tc.rule, tc.ml)
			assert.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestGetPercentile_EmptyHistory(t *testing.T) {
	s := NewMLService("v1.0")
	assert.InDelta(t, 0.5, s.GetPercentile(0.5), 0.001)
}

func TestGetPercentile_WithHistory(t *testing.T) {
	s := NewMLService("v1.0")

	s.mu.Lock()
	s.predictionHistory = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	s.mu.Unlock()

	assert.InDelta(t, 0.0, s.GetPercentile(-0.1), 0.001)
	assert.InDelta(t, 0.2, s.GetPercentile(0.1), 0.001)
	assert.InDelta(t, 0.6, s.GetPercentile(0.3), 0.001)
	assert.InDelta(t, 1.0, s.GetPercentile(1.0), 0.001)
}

func TestGetStats(t *testing.T) {
	s := NewMLService("v1.0")

	s.mu.Lock()
	s.predictionHistory = []float64{0.2, 0.4, 0.6}
	s.mu.Unlock()

	stats := s.GetStats()
	assert.Equal(t, "v1.0", stats["model_version"])
	assert.Equal(t, 3, stats["predictions"])
	assert.InDelta(t, 0.4, stats["avg_score"].(float64), 0.01)
	assert.InDelta(t, 0.7, stats["threshold"].(float64), 0.001)
}

func TestUpdateWeights(t *testing.T) {
	s := NewMLService("v1.0")
	newWeights := FeatureWeights{
		Amount:         0.5,
		Velocity:       0.3,
		HourAnomaly:    0.05,
		NewDevice:      0.05,
		NewLocation:    0.05,
		FailedAttempts: 0.03,
		AccountAge:     0.01,
		TransactionDev: 0.01,
	}
	s.UpdateWeights(newWeights)
	assert.Equal(t, 0.5, s.weights.Amount)
	assert.Equal(t, 0.01, s.weights.TransactionDev)
}

func TestPredict_ConcurrentSafety(t *testing.T) {
	s := NewMLService("v1.0")
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input := ModelInput{
				Amount:   100,
				Velocity: 10,
			}
			score := s.Predict(input)
			assert.GreaterOrEqual(t, score, 0.0)
		}()
	}
	wg.Wait()

	stats := s.GetStats()
	assert.Equal(t, 50, stats["predictions"])
}

func TestHistoryCap(t *testing.T) {
	s := NewMLService("v1.0")

	s.mu.Lock()
	s.predictionHistory = make([]float64, 1000)
	s.mu.Unlock()

	for i := 0; i < 100; i++ {
		s.Predict(ModelInput{Amount: float64(i)})
	}

	s.mu.RLock()
	assert.Equal(t, 1000, len(s.predictionHistory))
	s.mu.RUnlock()
}

func TestBoolToFloat(t *testing.T) {
	assert.Equal(t, 1.0, boolToFloat(true))
	assert.Equal(t, 0.0, boolToFloat(false))
}
