package learning

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/logger"
)

type ModelUpdate struct {
	Weights []float64
	Country string
	Epoch   int
}

type SecureAggregator struct {
	ActiveModels int
}

func (a *SecureAggregator) Aggregate(updates []ModelUpdate) []float64 {
	logger.Info(context.Background(), fmt.Sprintf("FEDERATION: Aggregating weights from %d sovereign nodes...", len(updates)))

	if len(updates) == 0 {
		return nil
	}

	size := len(updates[0].Weights)
	result := make([]float64, size)

	for _, u := range updates {
		for i, w := range u.Weights {
			result[i] += w
		}
	}

	// Average weights (Simple FedAvg)
	for i := range result {
		result[i] /= float64(len(updates))
	}

	return result
}
