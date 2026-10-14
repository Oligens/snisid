package risk

import (
	"context"
	"fmt"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type RiskResult struct {
	Score  int
	Reason string
}

type RiskModel interface {
	Name() string
	Evaluate(ctx context.Context, data map[string]interface{}) (RiskResult, error)
}

// SanctionsModel checks names against a blocked list
type SanctionsModel struct {
	metric strutil.StringMetric
}

func NewSanctionsModel() *SanctionsModel {
	return &SanctionsModel{
		metric: metrics.NewJaroWinkler(),
	}
}

func (m *SanctionsModel) Name() string { return "sanctions_check" }
func (m *SanctionsModel) Evaluate(ctx context.Context, data map[string]interface{}) (RiskResult, error) {
	name, _ := data["fullName"].(string)
	// Mock: Check against a "Blocked" name
	blockedName := "TEMPORARY_BLOCKED_CITIZEN"
	score := m.metric.Compare(name, blockedName)
	if score > 0.9 {
		return RiskResult{Score: 100, Reason: fmt.Sprintf("High similarity to sanctioned entity: %s", blockedName)}, nil
	}
	return RiskResult{Score: 0, Reason: "No sanctions match found"}, nil
}

// TravelModel detects "Impossible Travel"
type TravelModel struct{}

func (m *TravelModel) Name() string { return "travel_velocity" }
func (m *TravelModel) Evaluate(ctx context.Context, data map[string]interface{}) (RiskResult, error) {
	// Mock logic: if 'location' changed from 'CityA' to 'CityB' in metadata
	// in a very short time (stored in Redis in prod)
	if loc, ok := data["location"].(string); ok && loc == "SUSPICIOUS_REMOTE_LOCATION" {
		return RiskResult{Score: 70, Reason: "Access from high-risk or impossible geographic location"}, nil
	}
	return RiskResult{Score: 0, Reason: "Travel pattern normal"}, nil
}
