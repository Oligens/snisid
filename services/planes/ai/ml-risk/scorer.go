package ml

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type ModelInput struct {
	Amount           float64 `json:"amount"`
	Velocity         float64 `json:"velocity"`
	HourOfDay        int     `json:"hour_of_day"`
	DayOfWeek        int     `json:"day_of_week"`
	IsWeekend        bool    `json:"is_weekend"`
	IsNewDevice      bool    `json:"is_new_device"`
	IsNewLocation    bool    `json:"is_new_location"`
	FailedAttempts   int     `json:"failed_attempts"`
	AccountAgeDays   int     `json:"account_age_days"`
	TransactionCount int     `json:"transaction_count"`
	AvgTransaction   float64 `json:"avg_transaction"`
}

type FeatureWeights struct {
	Amount         float64 `json:"amount"`
	Velocity       float64 `json:"velocity"`
	HourAnomaly    float64 `json:"hour_anomaly"`
	NewDevice      float64 `json:"new_device"`
	NewLocation    float64 `json:"new_location"`
	FailedAttempts float64 `json:"failed_attempts"`
	AccountAge     float64 `json:"account_age"`
	TransactionDev float64 `json:"transaction_deviation"`
}

type MLService struct {
	ModelVersion      string
	weights           FeatureWeights
	mu                sync.RWMutex
	predictionHistory []float64
	threshold         float64
}

func NewMLService(modelVersion string) *MLService {
	return &MLService{
		ModelVersion: modelVersion,
		weights: FeatureWeights{
			Amount:         0.25,
			Velocity:       0.20,
			HourAnomaly:    0.10,
			NewDevice:      0.15,
			NewLocation:    0.10,
			FailedAttempts: 0.10,
			AccountAge:     0.05,
			TransactionDev: 0.05,
		},
		threshold: 0.7,
	}
}

func (s *MLService) Predict(input ModelInput) float64 {
	logger.Info(context.Background(), "ML-RISK: scoring input", zap.String("model", s.ModelVersion))

	score := 0.0

	score += s.evaluateAmount(input.Amount, input.AvgTransaction) * s.weights.Amount
	score += s.evaluateVelocity(input.Velocity) * s.weights.Velocity
	score += s.evaluateHourAnomaly(input.HourOfDay) * s.weights.HourAnomaly
	score += s.evaluateBinary(float64(boolToFloat(input.IsNewDevice))) * s.weights.NewDevice
	score += s.evaluateBinary(float64(boolToFloat(input.IsNewLocation))) * s.weights.NewLocation
	score += s.evaluateFailedAttempts(input.FailedAttempts) * s.weights.FailedAttempts
	score += s.evaluateAccountAge(input.AccountAgeDays) * s.weights.AccountAge
	score += s.evaluateTransactionDeviation(input.TransactionCount, input.AvgTransaction, input.Amount) * s.weights.TransactionDev

	score = math.Min(1.0, math.Max(0.0, score))

	s.mu.Lock()
	s.predictionHistory = append(s.predictionHistory, score)
	if len(s.predictionHistory) > 1000 {
		s.predictionHistory = s.predictionHistory[len(s.predictionHistory)-1000:]
	}
	s.mu.Unlock()

	return math.Round(score*100) / 100
}

func (s *MLService) evaluateAmount(amount, avg float64) float64 {
	if avg <= 0 {
		return 0.1
	}
	ratio := amount / avg
	switch {
	case ratio > 10:
		return 1.0
	case ratio > 5:
		return 0.8
	case ratio > 3:
		return 0.5
	case ratio > 1.5:
		return 0.2
	default:
		return 0.05
	}
}

func (s *MLService) evaluateVelocity(velocity float64) float64 {
	switch {
	case velocity > 100:
		return 1.0
	case velocity > 50:
		return 0.8
	case velocity > 20:
		return 0.5
	case velocity > 10:
		return 0.3
	default:
		return velocity / 10 * 0.2
	}
}

func (s *MLService) evaluateHourAnomaly(hour int) float64 {
	if hour >= 1 && hour <= 5 {
		return 0.7
	}
	if hour >= 23 || hour <= 6 {
		return 0.4
	}
	return 0.1
}

func (s *MLService) evaluateBinary(value float64) float64 {
	return value
}

func (s *MLService) evaluateFailedAttempts(attempts int) float64 {
	switch {
	case attempts > 10:
		return 1.0
	case attempts > 5:
		return 0.7
	case attempts > 3:
		return 0.4
	case attempts > 1:
		return 0.2
	default:
		return 0.0
	}
}

func (s *MLService) evaluateAccountAge(days int) float64 {
	switch {
	case days < 1:
		return 0.9
	case days < 7:
		return 0.5
	case days < 30:
		return 0.2
	default:
		return 0.05
	}
}

func (s *MLService) evaluateTransactionDeviation(count int, avg, current float64) float64 {
	if count < 5 || avg <= 0 {
		return 0.5
	}
	std := avg * 0.3
	if std <= 0 {
		return 0
	}
	zScore := math.Abs(current-avg) / std
	switch {
	case zScore > 3:
		return 1.0
	case zScore > 2:
		return 0.7
	case zScore > 1:
		return 0.3
	default:
		return 0.0
	}
}

func (s *MLService) AugmentRisk(ruleScore, mlScore float64) float64 {
	return math.Round((ruleScore*0.4+mlScore*0.6)*100) / 100
}

func (s *MLService) UpdateWeights(weights FeatureWeights) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.weights = weights
	logger.Info(context.Background(), "ML-RISK: weights updated", zap.Any("weights", weights))
}

func (s *MLService) GetPercentile(score float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.predictionHistory) == 0 {
		return 0.5
	}

	below := 0
	for _, v := range s.predictionHistory {
		if v <= score {
			below++
		}
	}
	return float64(below) / float64(len(s.predictionHistory))
}

func (s *MLService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	avg := 0.0
	for _, v := range s.predictionHistory {
		avg += v
	}
	if len(s.predictionHistory) > 0 {
		avg /= float64(len(s.predictionHistory))
	}

	return map[string]interface{}{
		"model_version": s.ModelVersion,
		"predictions":   len(s.predictionHistory),
		"avg_score":     math.Round(avg*100) / 100,
		"threshold":     s.threshold,
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func (s *MLService) Now() int64 {
	return time.Now().Unix()
}
