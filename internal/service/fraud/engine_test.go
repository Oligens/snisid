package fraud

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/snisid/platform/internal/service/router"
	"go.uber.org/zap"
)

type mockMLModel struct{}

func (m *mockMLModel) Predict(ctx context.Context, features FeatureVector) (float64, error) {
	score := features.Velocity*0.4 + features.Amount*0.001*0.3 + features.GraphRisk*0.3
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}
	return score, nil
}

type mockAIClient struct {
	predictFn func(ctx context.Context, features FeatureVector) (float64, error)
}

func (m *mockAIClient) Predict(ctx context.Context, features FeatureVector) (float64, error) {
	if m.predictFn != nil {
		return m.predictFn(ctx, features)
	}
	return 0.05, nil // match DefaultAIClient baseline
}

func newTestEngine(aiClient AIClient) (*ScoringEngine, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	eng, _ := router.NewEngine()
	return &ScoringEngine{
		rules:    eng.Rules(),
		state:    NewRedisStateStore(client),
		aiClient: aiClient,
	}, s
}

func TestScoringEngine_CalculateScore_NoTriggers(t *testing.T) {
	engine, mini := newTestEngine(&mockAIClient{})
	t.Cleanup(func() { mini.Close() })
	event := map[string]interface{}{
		"identityId": "normal-user",
		"action":     "view",
	}

	score, _, riskLevel := engine.CalculateScore(context.Background(), event)
	if score != 5 {
		t.Errorf("score = %d, want 5 (only AI default, 0.05*100=5)", score)
	}
	if riskLevel == "" {
		t.Error("expected non-empty risk level")
	}
}

func TestScoringEngine_AIHighRisk(t *testing.T) {
	aiClient := &mockAIClient{
		predictFn: func(ctx context.Context, features FeatureVector) (float64, error) {
			return 0.60, nil
		},
	}

	engine, mini := newTestEngine(aiClient)
	t.Cleanup(func() { mini.Close() })
	event := map[string]interface{}{
		"identityId": "test-fraud",
		"action":     "enroll",
	}

	score, reasons, riskLevel := engine.CalculateScore(context.Background(), event)
	if score < 60 {
		t.Errorf("score = %d, want >= 60", score)
	}
	if reasons == "" {
		t.Error("expected reasons for high risk")
	}
	if riskLevel == "" {
		t.Error("expected non-empty risk level")
	}
}

func TestScoringEngine_AIErrorFallback(t *testing.T) {
	aiClient := &mockAIClient{
		predictFn: func(ctx context.Context, features FeatureVector) (float64, error) {
			return 0.0, nil
		},
	}

	engine, mini := newTestEngine(aiClient)
	t.Cleanup(func() { mini.Close() })
	event := map[string]interface{}{
		"identityId": "user-001",
	}

	score, _, riskLevel := engine.CalculateScore(context.Background(), event)
	if score < 0 || score > 100 {
		t.Errorf("score = %d, out of expected range [0,100]", score)
	}
	if riskLevel == "" {
		t.Error("expected non-empty risk level")
	}
}

func TestScoringEngine_IntelligenceFusion(t *testing.T) {
	aiClient := &mockAIClient{
		predictFn: func(ctx context.Context, features FeatureVector) (float64, error) {
			return features.GraphRisk, nil
		},
	}
	engine, mini := newTestEngine(aiClient)
	t.Cleanup(func() { mini.Close() })
	event := map[string]interface{}{
		"identityId": "high-graph-risk",
		"graph_risk": 0.95,
	}

	score, reasons, riskLevel := engine.CalculateScore(context.Background(), event)
	if score < 85 {
		t.Errorf("score = %d, want >= 85", score)
	}
	if reasons == "" {
		t.Error("expected reasons from intelligence fusion")
	}
	if riskLevel != "CRITICAL" {
		t.Errorf("riskLevel = %s, want CRITICAL (graph_risk=0.95)", riskLevel)
	}
}

func TestScoringEngine_VelocityScore(t *testing.T) {
	engine, mini := newTestEngine(NewDefaultAIClient(&mockMLModel{}))
	t.Cleanup(func() { mini.Close() })
	ctx := context.Background()

	event := map[string]interface{}{
		"identityId": "vel-user",
		"action":     "enroll",
	}

	for i := 0; i < 6; i++ {
		engine.CalculateScore(ctx, event)
	}

	score, reasons, _ := engine.CalculateScore(ctx, event)
	if score < 50 {
		t.Errorf("score = %d, want >= 50 (velocity penalty)", score)
	}
	if reasons == "" {
		t.Error("expected reasons including velocity")
	}
}

func TestDefaultAIClient_Predict(t *testing.T) {
	client := NewDefaultAIClient(&mockMLModel{})

	tests := []struct {
		name     string
		features FeatureVector
		wantMin  float64
	}{
		{"fraud pattern", FeatureVector{UserID: "test-fraud", Velocity: 100, Amount: 50000, GraphRisk: 0.9}, 60},
		{"normal user", FeatureVector{UserID: "normal-user"}, 0},
		{"missing identityId", FeatureVector{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := client.Predict(context.Background(), tt.features)
			if err != nil {
				t.Fatalf("Predict failed: %v", err)
			}
			if score*100 < tt.wantMin {
				t.Errorf("score = %f, want >= %f", score*100, tt.wantMin)
			}
		})
	}
}

func TestNewScoringEngine(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	engine := NewScoringEngine(&mockAIClient{}, NewStateStore("localhost:6379"), logger)
	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}
}

func TestScoringEngine_ReloadRules(t *testing.T) {
	engine, mini := newTestEngine(&mockAIClient{})
	t.Cleanup(func() { mini.Close() })
	err := engine.ReloadRules([]router.Rule{})
	if err != nil {
		t.Errorf("ReloadRules failed: %v", err)
	}
}
