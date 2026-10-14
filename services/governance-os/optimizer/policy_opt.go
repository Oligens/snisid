package optimizer

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type PolicyUpdate struct {
	Version        string  `json:"version"`
	SuggestedDelta string  `json:"suggested_delta"`
	Confidence     float64 `json:"confidence"`
	Domain         string  `json:"domain"`
	Rationale      string  `json:"rationale"`
	RiskLevel      string  `json:"risk_level"` // LOW, MEDIUM, HIGH
	RequiresHuman  bool    `json:"requires_human"`
}

type EnforcementOutcome struct {
	PolicyID       string  `json:"policy_id"`
	Action         string  `json:"action"`
	Success        bool    `json:"success"`
	FraudPrevented float64 `json:"fraud_prevented"`
	FalsePositive  bool    `json:"false_positive"`
	Timestamp      int64   `json:"timestamp"`
	Region         string  `json:"region"`
}

type PolicyOptimizer struct {
	ID             string
	mu             sync.Mutex
	outcomeHistory []EnforcementOutcome
	minDataPoints  int
	learningRate   float64
}

func NewPolicyOptimizer(id string) *PolicyOptimizer {
	return &PolicyOptimizer{
		ID:             id,
		outcomeHistory: []EnforcementOutcome{},
		minDataPoints:  100,
		learningRate:   0.05,
	}
}

func (o *PolicyOptimizer) SuggestOptimizations(history map[string]interface{}) []PolicyUpdate {
	logger.Info(context.Background(), "GOS-OPT: analyzing enforcement outcomes for optimization")

	o.mu.Lock()
	if outcomes, ok := history["outcomes"].([]EnforcementOutcome); ok {
		o.outcomeHistory = append(o.outcomeHistory, outcomes...)
	}
	current := make([]EnforcementOutcome, len(o.outcomeHistory))
	copy(current, o.outcomeHistory)
	o.mu.Unlock()

	if len(current) < o.minDataPoints {
		logger.Info(context.Background(), "GOS-OPT: insufficient data points for optimization",
			zap.Int("have", len(current)),
			zap.Int("need", o.minDataPoints),
		)
		return nil
	}

	return o.computeOptimizations(current)
}

func (o *PolicyOptimizer) computeOptimizations(outcomes []EnforcementOutcome) []PolicyUpdate {
	var updates []PolicyUpdate

	regionStats := make(map[string]struct {
		total   int
		success int
		fraud   float64
	})
	for _, oc := range outcomes {
		s := regionStats[oc.Region]
		s.total++
		if oc.Success {
			s.success++
		}
		s.fraud += oc.FraudPrevented
		regionStats[oc.Region] = s
	}

	falsePositives := 0
	for _, oc := range outcomes {
		if oc.FalsePositive {
			falsePositives++
		}
	}
	fpRate := float64(falsePositives) / float64(len(outcomes))

	if fpRate > 0.1 {
		updates = append(updates, PolicyUpdate{
			Version:        "v1." + time.Now().Format("20060102") + ".1",
			SuggestedDelta: "increase fraud detection threshold to reduce false positives",
			Confidence:     math.Round((1-fpRate)*100) / 100,
			Domain:         "FRAUD_DETECTION",
			Rationale:      formatRationale("FP", fpRate, 0.1),
			RiskLevel:      "MEDIUM",
			RequiresHuman:  false,
		})
	}

	for region, stats := range regionStats {
		successRate := float64(stats.success) / float64(stats.total)
		if successRate < 0.7 {
			updates = append(updates, PolicyUpdate{
				Version:        "v1." + time.Now().Format("20060102") + ".2",
				SuggestedDelta: "reevaluate policy enforcement in region " + region,
				Confidence:     math.Round(successRate*100) / 100,
				Domain:         "REGION_ENFORCEMENT",
				Rationale:      formatRationale("ENFORCE", successRate, 0.7),
				RiskLevel:      "HIGH",
				RequiresHuman:  true,
			})
		}
	}

	avgFraud := 0.0
	for _, stats := range regionStats {
		avgFraud += stats.fraud
	}
	avgFraud /= float64(len(regionStats))

	if avgFraud < 100.0 {
		updates = append(updates, PolicyUpdate{
			Version:        "v1." + time.Now().Format("20060102") + ".3",
			SuggestedDelta: "reduce enforcement cost by lowering review frequency",
			Confidence:     0.75,
			Domain:         "COST_OPTIMIZATION",
			Rationale:      "average fraud prevented is low across all regions",
			RiskLevel:      "LOW",
			RequiresHuman:  false,
		})
	}

	sort.Slice(updates, func(i, j int) bool {
		return updates[i].Confidence > updates[j].Confidence
	})

	return updates
}

func (o *PolicyOptimizer) ValidateAgainstHumanGovernance(update PolicyUpdate) bool {
	if !update.RequiresHuman {
		return true
	}

	logger.Info(context.Background(), "GOS-OPT: policy update requires human validation",
		zap.Float64("delta", update.SuggestedDelta),
		zap.String("risk", update.RiskLevel),
	)
	return false
}

func (o *PolicyOptimizer) RecordOutcome(outcome EnforcementOutcome) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.outcomeHistory = append(o.outcomeHistory, outcome)
}

func (o *PolicyOptimizer) GetStats() map[string]interface{} {
	o.mu.Lock()
	defer o.mu.Unlock()

	return map[string]interface{}{
		"total_outcomes":  len(o.outcomeHistory),
		"min_data_points": o.minDataPoints,
		"learning_rate":   o.learningRate,
	}
}

func formatRationale(metric string, current, target float64) string {
	return metric + " rate " + formatPercent(current) + " exceeds threshold " + formatPercent(target)
}

func formatPercent(v float64) string {
	return math.Round(v*1000)/10 + "%"
}
