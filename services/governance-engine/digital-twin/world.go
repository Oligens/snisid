package digitaltwin

import "fmt"

type WorldState struct {
	FraudRate    float64
	Enforcement  float64
	EconomicLoad float64
}

func Step(state WorldState, action string) WorldState {
	fmt.Printf("SIMULATOR: Applying action '%s' to synthetic world...\n", action)

	switch action {
	case "increase_control":
		state.FraudRate *= 0.9
		state.EconomicLoad *= 1.1
	case "relax_policy":
		state.FraudRate *= 1.2
		state.EconomicLoad *= 0.95
	}

	return state
}

func RunImpactForecast(initial WorldState, actions []string) []WorldState {
	states := []WorldState{initial}
	for _, a := range actions {
		next := Step(states[len(states)-1], a)
		states = append(states, next)
	}
	return states
}
