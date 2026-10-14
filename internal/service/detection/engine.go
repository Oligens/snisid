package detection

import (
	"context"
	"fmt"
)

type DetectionEngine struct {
	strategies []MatchStrategy
}

func NewDetectionEngine() *DetectionEngine {
	return &DetectionEngine{
		strategies: []MatchStrategy{
			NewFuzzyMatcher(),
			NewPhoneticMatcher(),
		},
	}
}

func (e *DetectionEngine) Detect(ctx context.Context, newIdentity map[string]interface{}, candidates []map[string]interface{}) (int, string, map[string]string) {
	maxScore := 0
	bestMatchID := ""
	evidence := make(map[string]string)

	newName := fmt.Sprintf("%v", newIdentity["fullName"])

	for _, cand := range candidates {
		candName := fmt.Sprintf("%v", cand["fullName"])
		candID := fmt.Sprintf("%v", cand["identityId"])

		currentScore := 0
		for _, strategy := range e.strategies {
			s := strategy.Score(newName, candName)
			currentScore += s
			evidence[fmt.Sprintf("%s_%s", candID, strategy.Name())] = fmt.Sprintf("%d", s)
		}

		// Weighted Average (simplified)
		avgScore := currentScore / len(e.strategies)

		// Boost if SSN/TaxID matches exactly
		if newIdentity["taxId"] == cand["taxId"] && newIdentity["taxId"] != "" {
			avgScore = 100
			evidence[fmt.Sprintf("%s_exact_tax_id", candID)] = "100"
		}

		if avgScore > maxScore {
			maxScore = avgScore
			bestMatchID = candID
		}
	}

	return maxScore, bestMatchID, evidence
}
