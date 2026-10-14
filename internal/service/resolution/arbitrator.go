package resolution

import (
	"context"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type SourceTrust struct {
	Name  string
	Score int
}

type Arbitrator struct {
	trustMap map[string]int
}

func NewArbitrator() *Arbitrator {
	return &Arbitrator{
		trustMap: map[string]int{
			"national_registry": 100,
			"passport_office":   90,
			"national_police":   80,
			"user_self_service": 20,
		},
	}
}

func (a *Arbitrator) Resolve(ctx context.Context, attrName string, values map[string]interface{}) (interface{}, string) {
	bestScore := -1
	var bestValue interface{}
	bestSource := ""

	for source, value := range values {
		score := a.trustMap[source]
		if score > bestScore {
			bestScore = score
			bestValue = value
			bestSource = source
		}
	}

	logger.Info(ctx, "Attribute conflict resolved",
		zap.String("attribute", attrName),
		zap.String("source", bestSource),
		zap.Int("trust_score", bestScore),
	)

	return bestValue, bestSource
}
