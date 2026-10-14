package fraud

import (
	"context"
	"fmt"
	"math"

	"github.com/snisid/platform/internal/service/router"
	"go.uber.org/zap"
)

type ScoringEngine struct {
	aiClient AIClient
	state    StateStore
	logger   *zap.Logger
	rules    []router.Rule
}

type AIClient interface {
	Predict(ctx context.Context, features FeatureVector) (float64, error)
}

type ScoringResult struct {
	Score     float64  `json:"score"`
	RiskLevel string   `json:"risk_level"`
	Factors   []string `json:"factors"`
}

func NewScoringEngine(aiClient AIClient, state StateStore, logger *zap.Logger) *ScoringEngine {
	return &ScoringEngine{
		aiClient: aiClient,
		state:    state,
		logger:   logger,
	}
}

func (e *ScoringEngine) ScoreTransaction(ctx context.Context, userID string, amount float64, graphRisk ...float64) (*ScoringResult, error) {
	velocity, err := e.state.IncrementVelocity(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("increment velocity: %w", err)
	}

	gr := 0.0
	if len(graphRisk) > 0 {
		gr = graphRisk[0]
	}

	features := FeatureVector{
		UserID:    userID,
		Amount:    amount,
		Velocity:  float64(velocity),
		GraphRisk: gr,
	}

	score, err := e.aiClient.Predict(ctx, features)
	if err != nil {
		return nil, fmt.Errorf("predict: %w", err)
	}

	riskLevel := classifyRisk(score)
	factors := e.explainFactors(features, score)

	return &ScoringResult{
		Score:     score,
		RiskLevel: riskLevel,
		Factors:   factors,
	}, nil
}

func classifyRisk(score float64) string {
	switch {
	case score >= 0.85:
		return "CRITICAL"
	case score >= 0.70:
		return "HIGH"
	case score >= 0.40:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func (e *ScoringEngine) CalculateScore(ctx context.Context, event map[string]interface{}) (int, string, string) {
	userID, _ := event["identityId"].(string)
	amount, _ := event["amount"].(float64)
	graphRisk, _ := event["graph_risk"].(float64)
	result, err := e.ScoreTransaction(ctx, userID, amount, graphRisk)
	if err != nil || result == nil {
		return 0, "error", "LOW"
	}
	score := int(result.Score * 100)
	return score, result.RiskLevel, result.RiskLevel
}

func (e *ScoringEngine) ReloadRules(rules []router.Rule) error {
	e.rules = rules
	return nil
}

func (e *ScoringEngine) Rules() []router.Rule {
	return e.rules
}

func (e *ScoringEngine) explainFactors(features FeatureVector, score float64) []string {
	var factors []string
	if features.Velocity > 5 {
		factors = append(factors, "high_transaction_velocity")
	}
	if features.Amount > 10000 {
		factors = append(factors, "large_transaction_amount")
	}
	if features.GraphRisk > 0.7 {
		factors = append(factors, "high_graph_risk")
	}
	if math.Abs(score-0.5) < 0.1 {
		factors = append(factors, "borderline_score")
	}
	return factors
}
