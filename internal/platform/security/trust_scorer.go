package security

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type TrustSignal struct {
	Type   string
	Value  float32
	Weight float32
}

type TrustScore struct {
	Score   int
	Level   string
	Signals []string
}

type AdaptiveTrustScorer struct{}

func NewAdaptiveTrustScorer() *AdaptiveTrustScorer {
	return &AdaptiveTrustScorer{}
}

func (s *AdaptiveTrustScorer) CalculateScore(ctx context.Context, principal string, signals []TrustSignal) TrustScore {
	logger.Info(ctx, "Calculating adaptive trust score", zap.String("principal", principal))

	totalScore := 0.0
	explanation := []string{}

	for _, sig := range signals {
		contribution := float64(sig.Value) * float64(sig.Weight)
		totalScore += contribution
		explanation = append(explanation, fmt.Sprintf("[%s]: %.2f (weight: %.2f)", sig.Type, sig.Value, sig.Weight))
	}

	scoreInt := int(totalScore)
	level := "LOW"
	if scoreInt >= 70 {
		level = "HIGH"
	} else if scoreInt >= 50 {
		level = "MEDIUM"
	}

	logger.Info(ctx, "Trust score calculated",
		zap.String("principal", principal),
		zap.Int("score", scoreInt),
		zap.String("level", level),
	)

	return TrustScore{
		Score:   scoreInt,
		Level:   level,
		Signals: explanation,
	}
}
