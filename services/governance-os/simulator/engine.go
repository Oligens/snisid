package simulator

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type PolicyScenario struct {
	PolicyID   string   `json:"policy_id"`
	Changes    []string `json:"changes"`
	Region     string   `json:"region"`
	Population int      `json:"population"`
	FraudRate  float64  `json:"fraud_rate"`
}

type SimulationResult struct {
	FraudShift         float64        `json:"fraud_shift"`
	FalsePositives     float64        `json:"false_positives"`
	ConflictFound      bool           `json:"conflict_found"`
	AdoptionRate       float64        `json:"adoption_rate"`
	CostImpact         float64        `json:"cost_impact"`
	Iterations         int            `json:"iterations"`
	ConvergenceTime    float64        `json:"convergence_time_hours"`
	PerChangeImpact    []ChangeImpact `json:"per_change_impact"`
	ConfidenceInterval [2]float64     `json:"confidence_interval"`
}

type ChangeImpact struct {
	Change        string  `json:"change"`
	FraudDelta    float64 `json:"fraud_delta"`
	FPDelta       float64 `json:"fp_delta"`
	AdoptionDelta float64 `json:"adoption_delta"`
}

type SimulationEngine struct {
	Environment   string
	iterations    int
	convergenceAt float64
	rng           *rand.Rand
}

func NewSimulationEngine(environment string) *SimulationEngine {
	return &SimulationEngine{
		Environment:   environment,
		iterations:    1000,
		convergenceAt: 0.01,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (e *SimulationEngine) RunScenario(scenario PolicyScenario) SimulationResult {
	logger.Info(context.Background(), "GOS-SIM: running policy scenario",
		zap.String("policy_id", scenario.PolicyID),
		zap.Int("changes", len(scenario.Changes)),
		zap.String("region", scenario.Region),
	)

	baseFraud := scenario.FraudRate
	if baseFraud == 0 {
		baseFraud = 0.05
	}

	baseFP := 0.03
	baseAdoption := 0.7

	impacts := make([]ChangeImpact, 0, len(scenario.Changes))
	netFraudDelta := 0.0
	netFPDelta := 0.0
	netAdoptionDelta := 0.0

	for _, change := range scenario.Changes {
		impact := e.simulateChange(change, baseFraud)
		impacts = append(impacts, impact)
		netFraudDelta += impact.FraudDelta
		netFPDelta += impact.FPDelta
		netAdoptionDelta += impact.AdoptionDelta
	}

	adoptionRate := math.Min(1.0, math.Max(0.0, baseAdoption+netAdoptionDelta))
	fraudRate := math.Max(0.0, baseFraud+netFraudDelta)
	fpRate := math.Max(0.0, baseFP+netFPDelta)

	fraudShift := (fraudRate - baseFraud) / baseFraud * 100

	costImpact := netFraudDelta*1000 + float64(len(scenario.Changes))*50

	convergence := e.computeConvergence(len(scenario.Changes), netFraudDelta)
	ci := e.computeConfidenceInterval(fraudShift, convergence)

	conflict := e.detectConflict(impacts)

	result := SimulationResult{
		FraudShift:         math.Round(fraudShift*100) / 100,
		FalsePositives:     math.Round(fpRate*100) / 100,
		ConflictFound:      conflict,
		AdoptionRate:       math.Round(adoptionRate*100) / 100,
		CostImpact:         math.Round(costImpact*100) / 100,
		Iterations:         e.iterations,
		ConvergenceTime:    math.Round(convergence*10) / 10,
		PerChangeImpact:    impacts,
		ConfidenceInterval: ci,
	}

	logger.Info(context.Background(), "GOS-SIM: simulation complete",
		zap.Float64("fraud_shift", result.FraudShift),
		zap.Float64("fp_rate", result.FalsePositives),
		zap.Bool("conflict", result.ConflictFound),
	)

	return result
}

func (e *SimulationEngine) simulateChange(change string, baseFraud float64) ChangeImpact {
	noise := e.rng.Float64()*0.1 - 0.05

	var fraudDelta float64
	var fpDelta float64
	var adoptionDelta float64

	switch change {
	case "lower_threshold":
		fraudDelta = -(baseFraud*0.3 + noise)
		fpDelta = 0.05 + noise*0.5
		adoptionDelta = -0.02
	case "raise_threshold":
		fraudDelta = baseFraud*0.2 + noise*0.5
		fpDelta = -(0.03 + noise*0.3)
		adoptionDelta = 0.05
	case "add_biometric_check":
		fraudDelta = -(baseFraud*0.4 + noise)
		fpDelta = 0.08 + noise
		adoptionDelta = -0.1
	case "remove_biometric_check":
		fraudDelta = baseFraud*0.5 + noise
		fpDelta = -(0.04 + noise*0.5)
		adoptionDelta = 0.12
	case "reduce_review_time":
		fraudDelta = baseFraud*0.05 + noise*0.3
		fpDelta = 0.02 + noise*0.3
		adoptionDelta = 0.15
	case "add_risk_scoring":
		fraudDelta = -(baseFraud*0.35 + noise)
		fpDelta = 0.04 + noise*0.4
		adoptionDelta = -0.03
	default:
		fraudDelta = noise * baseFraud
		fpDelta = noise * 0.5
		adoptionDelta = noise * 0.1
	}

	return ChangeImpact{
		Change:        change,
		FraudDelta:    math.Round(fraudDelta*10000) / 10000,
		FPDelta:       math.Round(fpDelta*10000) / 10000,
		AdoptionDelta: math.Round(adoptionDelta*10000) / 10000,
	}
}

func (e *SimulationEngine) detectConflict(impacts []ChangeImpact) bool {
	fraudImpacts := make([]float64, len(impacts))
	fpImpacts := make([]float64, len(impacts))

	for i, imp := range impacts {
		fraudImpacts[i] = imp.FraudDelta
		fpImpacts[i] = imp.FPDelta
	}

	meanF, stdF := meanStd(fraudImpacts)
	meanFP, stdFP := meanStd(fpImpacts)

	if stdF > 0.15 && meanF < 0 {
		return true
	}
	if meanFP > 0.05 && meanF > -0.1 {
		return true
	}

	return false
}

func (e *SimulationEngine) computeConvergence(changes int, fraudDelta float64) float64 {
	base := 2.0
	if changes > 0 {
		base += float64(changes) * 1.5
	}
	if math.Abs(fraudDelta) > 0.1 {
		base += 4.0
	}
	return base
}

func (e *SimulationEngine) computeConfidenceInterval(pointEstimate float64, convergence float64) [2]float64 {
	margin := 1.96 * convergence / math.Sqrt(float64(e.iterations))
	return [2]float64{
		math.Round((pointEstimate-margin)*100) / 100,
		math.Round((pointEstimate+margin)*100) / 100,
	}
}

func meanStd(values []float64) (float64, float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / n
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= n
	return mean, math.Sqrt(variance)
}
