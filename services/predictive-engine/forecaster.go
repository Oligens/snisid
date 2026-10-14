package predictive

import (
	"context"
	"fmt"
	"math"

	"github.com/redis/go-redis/v9"
)

type RiskForecast struct {
	UserID      string   `json:"user_id"`
	Probability float64  `json:"probability"`
	RiskLevel   string   `json:"risk_level"`
	Factors     []string `json:"factors"`
}

type RiskWeights struct {
	BehaviorDrift   float64
	GraphCentrality float64
	VelocityTrend   float64
	LocationRisk    float64
}

var DefaultWeights = RiskWeights{
	BehaviorDrift:   0.40,
	GraphCentrality: 0.30,
	VelocityTrend:   0.20,
	LocationRisk:    0.10,
}

type FeatureStore interface {
	GetFloat(ctx context.Context, key string) (float64, error)
}

type RedisFeatureStore struct {
	client *redis.Client
}

func NewRedisFeatureStore(client *redis.Client) *RedisFeatureStore {
	return &RedisFeatureStore{client: client}
}

func (s *RedisFeatureStore) GetFloat(ctx context.Context, key string) (float64, error) {
	val, err := s.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0.0, nil
	}
	if err != nil {
		return 0.0, err
	}
	return val, nil
}

type RiskForecaster struct {
	store   FeatureStore
	weights RiskWeights
}

func NewRiskForecaster(store FeatureStore, weights RiskWeights) *RiskForecaster {
	return &RiskForecaster{store: store, weights: weights}
}

func (f *RiskForecaster) ForecastRisk(ctx context.Context, userID string) (*RiskForecast, error) {
	behaviorDrift, err := f.store.GetFloat(ctx, fmt.Sprintf("snisid:behavior:%s:drift", userID))
	if err != nil {
		return nil, fmt.Errorf("forecast: behavior drift: %w", err)
	}

	graphCentrality, err := f.store.GetFloat(ctx, fmt.Sprintf("snisid:graph:%s:centrality", userID))
	if err != nil {
		return nil, fmt.Errorf("forecast: graph centrality: %w", err)
	}

	velocityTrend, err := f.store.GetFloat(ctx, fmt.Sprintf("snisid:features:%s:velocity", userID))
	if err != nil {
		return nil, fmt.Errorf("forecast: velocity: %w", err)
	}

	locationRisk, err := f.store.GetFloat(ctx, fmt.Sprintf("snisid:features:%s:location_risk", userID))
	if err != nil {
		return nil, fmt.Errorf("forecast: location risk: %w", err)
	}

	score := f.weights.BehaviorDrift*behaviorDrift +
		f.weights.GraphCentrality*graphCentrality +
		f.weights.VelocityTrend*velocityTrend +
		f.weights.LocationRisk*locationRisk

	probability := sigmoid(score)

	return &RiskForecast{
		UserID:      userID,
		Probability: probability,
		RiskLevel:   classifyRisk(probability),
		Factors:     explainFactors(behaviorDrift, graphCentrality, velocityTrend, locationRisk),
	}, nil
}

func sigmoid(x float64) float64 { return 1.0 / (1.0 + math.Exp(-x)) }

func classifyRisk(p float64) string {
	switch {
	case p >= 0.85:
		return "CRITICAL"
	case p >= 0.70:
		return "HIGH"
	case p >= 0.40:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func explainFactors(behavior, graph, velocity, location float64) []string {
	var factors []string
	if behavior > 0.7 {
		factors = append(factors, "high_behavior_drift")
	}
	if graph > 0.7 {
		factors = append(factors, "high_graph_centrality")
	}
	if velocity > 0.7 {
		factors = append(factors, "high_velocity_trend")
	}
	if location > 0.7 {
		factors = append(factors, "high_location_risk")
	}
	return factors
}
