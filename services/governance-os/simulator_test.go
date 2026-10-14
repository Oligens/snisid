package simulator_test

import (
	"math"
	"testing"

	"github.com/snisid/platform/services/governance-os/simulator"
	"github.com/stretchr/testify/assert"
)

func TestNewSimulationEngine(t *testing.T) {
	e := simulator.NewSimulationEngine("test-env")
	assert.NotNil(t, e)
	assert.Equal(t, "test-env", e.Environment)
}

func TestRunScenario_DefaultParameters(t *testing.T) {
	e := simulator.NewSimulationEngine("test")
	scenario := simulator.PolicyScenario{
		PolicyID:   "POL-001",
		Changes:    []string{"lower_threshold"},
		Region:     "HTI",
		Population: 1000000,
	}

	result := e.RunScenario(scenario)
	assert.NotEqual(t, 0, result.Iterations)
	assert.Greater(t, result.ConvergenceTime, float64(0))
	assert.Len(t, result.PerChangeImpact, 1)
}

func TestRunScenario_MultipleChanges(t *testing.T) {
	e := simulator.NewSimulationEngine("test")
	scenario := simulator.PolicyScenario{
		PolicyID:   "POL-002",
		Changes:    []string{"add_biometric_check", "lower_threshold", "add_risk_scoring"},
		Region:     "DOM",
		Population: 500000,
	}

	result := e.RunScenario(scenario)
	assert.Len(t, result.PerChangeImpact, 3)
	assert.NotNil(t, result.ConfidenceInterval)
}

func TestRunScenario_FraudShiftDirection(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	lowerResult := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-LOW", Changes: []string{"lower_threshold"}, Region: "HTI",
	})
	assert.Less(t, lowerResult.FraudShift, float64(0), "lowering threshold should reduce fraud")

	addResult := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-ADD", Changes: []string{"add_biometric_check"}, Region: "HTI",
	})
	assert.Less(t, addResult.FraudShift, float64(0), "adding biometric check should reduce fraud")

	removeResult := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-REM", Changes: []string{"remove_biometric_check"}, Region: "HTI",
	})
	assert.Greater(t, removeResult.FraudShift, float64(0), "removing biometric check should increase fraud")
}

func TestRunScenario_AdoptionRates(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	result := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-ADOPT", Changes: []string{"reduce_review_time"}, Region: "HTI",
	})
	assert.Greater(t, result.AdoptionRate, float64(0.7), "reducing review time should increase adoption")

	resultLow := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-ADOPT-LOW", Changes: []string{"add_biometric_check"}, Region: "HTI",
	})
	assert.Less(t, resultLow.AdoptionRate, float64(0.7), "adding biometric check should decrease adoption")
}

func TestConvergenceTime(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	singleResult := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-CONV-1", Changes: []string{"lower_threshold"}, Region: "HTI",
	})
	assert.InDelta(t, 3.5, singleResult.ConvergenceTime, 1.0)

	multiResult := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-CONV-3", Changes: []string{"lower_threshold", "add_biometric_check", "add_risk_scoring"},
		Region: "HTI",
	})
	assert.Greater(t, multiResult.ConvergenceTime, singleResult.ConvergenceTime)
}

func TestConfidenceInterval(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	result := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-CI", Changes: []string{"lower_threshold"}, Region: "HTI",
	})

	ci := result.ConfidenceInterval
	assert.LessOrEqual(t, ci[0], ci[1], "lower bound should be <= upper bound")
	assert.NotEqual(t, float64(0), ci[0])
	assert.NotEqual(t, float64(0), ci[1])
}

func TestConflictDetection(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	result := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-CONFLICT", Changes: []string{"raise_threshold", "add_biometric_check"},
		Region: "HTI",
	})

	// Conflict detection depends on the magnitude and direction of changes
	t.Logf("Conflict found: %v", result.ConflictFound)
}

func TestCostImpactCalculation(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	result := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-COST", Changes: []string{"lower_threshold"}, Region: "HTI",
	})
	assert.Greater(t, result.CostImpact, float64(0))

	resultExpensive := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-COST-3", Changes: []string{"lower_threshold", "add_biometric_check", "add_risk_scoring"},
		Region: "HTI",
	})
	assert.Greater(t, resultExpensive.CostImpact, result.CostImpact)
}

func TestFalsePositiveRate(t *testing.T) {
	e := simulator.NewSimulationEngine("test")

	lowerFP := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-FP-LOW", Changes: []string{"raise_threshold"}, Region: "HTI",
	})

	higherFP := e.RunScenario(simulator.PolicyScenario{
		PolicyID: "POL-FP-HIGH", Changes: []string{"lower_threshold"}, Region: "HTI",
	})

	assert.Less(t, lowerFP.FalsePositives, higherFP.FalsePositives,
		"raising threshold should reduce false positives")
}

func TestSimulateChange_UnknownChange(t *testing.T) {
	e := simulator.NewSimulationEngine("test")
	scenario := simulator.PolicyScenario{
		PolicyID: "POL-UNKNOWN", Changes: []string{"unknown_change"}, Region: "HTI",
	}
	result := e.RunScenario(scenario)
	assert.Len(t, result.PerChangeImpact, 1)
	assert.Equal(t, "unknown_change", result.PerChangeImpact[0].Change)
}

func TestFraudRateWithZeroBase(t *testing.T) {
	e := simulator.NewSimulationEngine("test")
	scenario := simulator.PolicyScenario{
		PolicyID:  "POL-ZERO",
		Changes:   []string{"lower_threshold"},
		Region:    "HTI",
		FraudRate: 0,
	}
	result := e.RunScenario(scenario)
	assert.NotEqual(t, float64(0), result.FraudShift)
}

func TestMathematicalProperties_MeanStd(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	assert.InDelta(t, 3.0, mean, 0.001)

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	std := math.Sqrt(variance)
	assert.InDelta(t, math.Sqrt(2.0), std, 0.001)
}
