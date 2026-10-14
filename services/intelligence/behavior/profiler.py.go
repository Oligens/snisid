package behavior

import (
	"encoding/json"
	"math"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type BehaviorProfile struct {
	UserID          string   `json:"user_id"`
	AnomalyScore    float64  `json:"anomaly_score"`
	NormalizedScore float64  `json:"normalized_score"`
	FeaturesUsed    []string `json:"features_used"`
	IsAnomaly       bool     `json:"is_anomaly"`
}

type BehavioralProfiler struct {
	redis         *redis.Client
	features      []string
	trained       bool
	contamination float64
}

func NewBehavioralProfiler(redisClient *redis.Client, contamination float64) *BehavioralProfiler {
	return &BehavioralProfiler{
		redis:         redisClient,
		features:      []string{"transaction_hour", "transaction_day", "amount_normalized", "velocity_24h", "new_location", "device_count_7d"},
		contamination: contamination,
	}
}

func (p *BehavioralProfiler) Profile(ctx context.Context, userID string, event map[string]interface{}) (*BehaviorProfile, error) {
	features := p.extractFeatures(event)

	rawScore := p.calculateAnomalyScore(features)
	prediction := rawScore < -0.1

	normalized := math.Max(0.0, math.Min(1.0, 0.5-rawScore))

	profile := &BehaviorProfile{
		UserID:          userID,
		AnomalyScore:    rawScore,
		NormalizedScore: normalized,
		FeaturesUsed:    p.features,
		IsAnomaly:       prediction,
	}

	data, _ := json.Marshal(profile)
	p.redis.SetEx(ctx, "snisid:behavior:"+userID+":profile", string(data), 3600)

	return profile, nil
}

func (p *BehavioralProfiler) extractFeatures(event map[string]interface{}) []float64 {
	features := make([]float64, len(p.features))
	defaults := []float64{12, 3, 0.5, 1.0, 0.0, 1.0}

	for i, f := range p.features {
		if val, ok := event[f]; ok {
			switch v := val.(type) {
			case float64:
				features[i] = v
			case int:
				features[i] = float64(v)
			default:
				features[i] = defaults[i]
			}
		} else {
			features[i] = defaults[i]
		}
	}
	return features
}

func (p *BehavioralProfiler) calculateAnomalyScore(features []float64) float64 {
	score := 0.0
	weights := []float64{0.15, 0.1, 0.3, 0.25, 0.1, 0.1}
	means := []float64{12, 3, 0.5, 2, 0.1, 1.5}
	stds := []float64{5, 2, 0.3, 3, 0.3, 0.8}

	for i := range features {
		if stds[i] > 0 {
			z := (features[i] - means[i]) / stds[i]
			score += weights[i] * z
		}
	}

	return score
}

func (p *BehavioralProfiler) GetCachedProfile(ctx context.Context, userID string) (*BehaviorProfile, error) {
	data, err := p.redis.Get(ctx, "snisid:behavior:"+userID+":profile").Result()
	if err != nil {
		return nil, err
	}
	var profile BehaviorProfile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}
