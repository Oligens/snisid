package security

import (
	"context"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type PolicyDecision string

const (
	DecisionAllow     PolicyDecision = "ALLOW"
	DecisionDeny      PolicyDecision = "DENY"
	DecisionChallenge PolicyDecision = "CHALLENGE"
)

type PolicyEnforcer struct {
	scorer *AdaptiveTrustScorer
}

func NewPolicyEnforcer(scorer *AdaptiveTrustScorer) *PolicyEnforcer {
	return &PolicyEnforcer{scorer: scorer}
}

func (e *PolicyEnforcer) Authorize(ctx context.Context, principal string, action string, resource string) (PolicyDecision, error) {
	logger.Info(ctx, "Evaluating Zero Trust policy",
		zap.String("principal", principal),
		zap.String("action", action),
		zap.String("resource", resource),
	)

	// 1. Fetch Adaptive Trust Score
	// (In prod, signals come from device health, MFA status, etc.)
	signals := []TrustSignal{
		{Type: "mfa_strength", Value: 100, Weight: 0.4},
		{Type: "device_health", Value: 90, Weight: 0.3},
		{Type: "anomaly_score", Value: 95, Weight: 0.3},
	}
	trust := e.scorer.CalculateScore(ctx, principal, signals)

	// 2. Continuous Policy Evaluation (Mock: OPA logic)
	if trust.Score < 50 {
		logger.Warn(ctx, "Access DENIED: Insufficient trust score", zap.Int("score", trust.Score))
		return DecisionDeny, nil
	}

	if trust.Score < 80 && action == "WRITE" {
		logger.Info(ctx, "Access CHALLENGE: High-sensitivity write requires elevation")
		return DecisionChallenge, nil
	}

	logger.Info(ctx, "Access GRANTED: Policy evaluation successful")
	return DecisionAllow, nil
}
