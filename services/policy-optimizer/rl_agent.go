package policyoptimizer

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/logger"
)

type SystemState struct {
	FraudRate      float64
	FalsePositives float64
	Instability    float64
}

type PolicyOptimizer struct {
	LearningRate float64
}

func (o *PolicyOptimizer) CalculateReward(state SystemState, previousFraud float64) float64 {
	fraudReduction := previousFraud - state.FraudRate
	// Penalty for false positives and system instability
	reward := (fraudReduction * 100) - (state.FalsePositives * 50) - (state.Instability * 20)
	return reward
}

func (o *PolicyOptimizer) SuggestAction(state SystemState) string {
	logger.Info(context.Background(), "RL-OPTIMIZER: Analyzing system state for policy suggestion...")

	if state.FraudRate > 0.4 {
		return "SUGGESTION: Tighten transaction verification threshold"
	}

	return "SUGGESTION: Maintain current policy stability"
}
