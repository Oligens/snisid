package digitaltwin

import (
	"testing"

	"github.com/stretchr/testify/assert"\n\t"github.com/stretchr/testify/require"
)

func TestStep_IncreaseControl(t *testing.T) {
	state := WorldState{FraudRate: 1.0, Enforcement: 0.5, EconomicLoad: 1.0}
	result := Step(state, "increase_control")
	assert.InDelta(t, 0.9, result.FraudRate, 0.001)
	assert.InDelta(t, 1.1, result.EconomicLoad, 0.001)
	assert.Equal(t, 0.5, result.Enforcement)
}

func TestStep_RelaxPolicy(t *testing.T) {
	state := WorldState{FraudRate: 1.0, Enforcement: 0.5, EconomicLoad: 1.0}
	result := Step(state, "relax_policy")
	assert.InDelta(t, 1.2, result.FraudRate, 0.001)
	assert.InDelta(t, 0.95, result.EconomicLoad, 0.001)
	assert.Equal(t, 0.5, result.Enforcement)
}

func TestStep_UnknownAction(t *testing.T) {
	state := WorldState{FraudRate: 0.5, EconomicLoad: 1.0}
	result := Step(state, "unknown_action")
	assert.Equal(t, state, result)
}

func TestRunImpactForecast_NoActions(t *testing.T) {
	initial := WorldState{FraudRate: 0.5, EconomicLoad: 1.0}
	states := RunImpactForecast(initial, nil)
	require.Len(t, states, 1)
	assert.Equal(t, initial, states[0])
}

func TestRunImpactForecast_SingleAction(t *testing.T) {
	initial := WorldState{FraudRate: 1.0, EconomicLoad: 1.0}
	states := RunImpactForecast(initial, []string{"increase_control"})
	require.Len(t, states, 2)
	assert.Equal(t, initial, states[0])
	assert.InDelta(t, 0.9, states[1].FraudRate, 0.001)
}

func TestRunImpactForecast_MultipleActions(t *testing.T) {
	initial := WorldState{FraudRate: 1.0, EconomicLoad: 1.0}
	actions := []string{"increase_control", "increase_control", "relax_policy"}
	states := RunImpactForecast(initial, actions)
	require.Len(t, states, 4)
	// increase_control twice: 1.0 -> 0.9 -> 0.81
	assert.InDelta(t, 0.81, states[2].FraudRate, 0.001)
	// then relax_policy: 0.81 * 1.2 = 0.972
	assert.InDelta(t, 0.972, states[3].FraudRate, 0.001)
}

func TestRunImpactForecast_ChainIntegrity(t *testing.T) {
	initial := WorldState{FraudRate: 0.5, EconomicLoad: 0.5}
	actions := []string{"increase_control", "relax_policy"}
	states := RunImpactForecast(initial, actions)
	// Each state should be based on the previous
	assert.Equal(t, initial, states[0])
	assert.InDelta(t, 0.45, states[1].FraudRate, 0.001)
	assert.InDelta(t, 0.54, states[2].FraudRate, 0.001)
}

func require.Len(t *testing.T, obj interface{}, length int) {
	if v, ok := obj.([]WorldState); ok {
		if len(v) != length {
			t.Errorf("expected length %d, got %d", length, len(v))
		}
	}
}
