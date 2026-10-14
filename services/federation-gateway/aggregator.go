package federationgateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type ModelUpdate struct {
	NodeID      string    `json:"nodeId"`
	Weights     []float64 `json:"weights"`
	Gradients   []float64 `json:"gradients"`
	SampleCount int       `json:"sampleCount"`
	Country     string    `json:"country"`
	Signature   string    `json:"signature"`
	Timestamp   time.Time `json:"timestamp"`
	Loss        float64   `json:"loss"`
	Accuracy    float64   `json:"accuracy"`
}

type AggregatedModel struct {
	GlobalWeights    []float64          `json:"globalWeights"`
	Round            int                `json:"round"`
	NodeCount        int                `json:"nodeCount"`
	TotalSamples     int                `json:"totalSamples"`
	AvgLoss          float64            `json:"avgLoss"`
	WeightedAccuracy float64            `json:"weightedAccuracy"`
	NodeWeights      map[string]float64 `json:"nodeWeights"`
	CreatedAt        time.Time          `json:"createdAt"`
	ConvergenceDelta float64            `json:"convergenceDelta"`
}

type NodeReputation struct {
	NodeID       string    `json:"nodeId"`
	TotalUpdates int       `json:"totalUpdates"`
	AvgAccuracy  float64   `json:"avgAccuracy"`
	TrustScore   float64   `json:"trustScore"`
	LastSeen     time.Time `json:"lastSeen"`
}

type Aggregator struct {
	ID                  string
	privateKey          *ecdsa.PrivateKey
	publicKey           *ecdsa.PublicKey
	mu                  sync.RWMutex
	nodeReputations     map[string]*NodeReputation
	previousWeights     []float64
	round               int
	minNodes            int
	minSamples          int
	aggregationFn       string
	differentialPrivacy bool
	noiseScale          float64
}

func NewAggregator(id string) (*Aggregator, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate aggregator key: %w", err)
	}

	return &Aggregator{
		ID:                  id,
		privateKey:          privateKey,
		publicKey:           &privateKey.PublicKey,
		nodeReputations:     make(map[string]*NodeReputation),
		round:               0,
		minNodes:            2,
		minSamples:          100,
		aggregationFn:       "fedavg",
		differentialPrivacy: true,
		noiseScale:          0.01,
	}, nil
}

func (a *Aggregator) AggregateWeights(updates []ModelUpdate) (*AggregatedModel, error) {
	if len(updates) < a.minNodes {
		return nil, fmt.Errorf("insufficient nodes: got %d, need %d", len(updates), a.minNodes)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	validUpdates := make([]ModelUpdate, 0, len(updates))
	totalSamples := 0

	for _, update := range updates {
		if err := a.verifyUpdate(update); err != nil {
			logger.Warn(context.Background(), "federation: invalid update rejected",
				zap.String("node", update.NodeID),
				zap.Error(err))
			continue
		}
		validUpdates = append(validUpdates, update)
		totalSamples += update.SampleCount
		a.updateReputation(update.NodeID, update.Accuracy, true)
	}

	if len(validUpdates) < a.minNodes {
		return nil, fmt.Errorf("insufficient valid updates: got %d, need %d", len(validUpdates), a.minNodes)
	}

	if len(validUpdates) == 0 {
		return nil, fmt.Errorf("no valid updates to aggregate")
	}

	size := len(validUpdates[0].Weights)
	globalWeights := make([]float64, size)
	totalWeight := 0.0
	avgLoss := 0.0
	weightedAccuracy := 0.0

	sampleWeights := make([]float64, len(validUpdates))
	for i, u := range validUpdates {
		trustScore := a.getTrustScore(u.NodeID)
		sampleWeight := float64(u.SampleCount) * trustScore
		sampleWeights[i] = sampleWeight
		totalWeight += sampleWeight
		avgLoss += u.Loss * sampleWeight
		weightedAccuracy += u.Accuracy * sampleWeight
	}

	if totalWeight > 0 {
		for i := range globalWeights {
			var weightedSum float64
			for j, u := range validUpdates {
				weightedSum += u.Weights[i] * sampleWeights[j]
			}
			globalWeights[i] = weightedSum / totalWeight
		}

		avgLoss /= totalWeight
		weightedAccuracy /= totalWeight
	}

	if a.differentialPrivacy {
		for i := range globalWeights {
			noise := a.generateLaplaceNoise(a.noiseScale)
			globalWeights[i] += noise
		}
	}

	a.round++

	convergenceDelta := 0.0
	if a.previousWeights != nil {
		for i := range globalWeights {
			diff := globalWeights[i] - a.previousWeights[i]
			convergenceDelta += diff * diff
		}
		convergenceDelta = math.Sqrt(convergenceDelta) / float64(len(globalWeights))
	}
	a.previousWeights = globalWeights

	result := &AggregatedModel{
		GlobalWeights:    globalWeights,
		Round:            a.round,
		NodeCount:        len(validUpdates),
		TotalSamples:     totalSamples,
		AvgLoss:          math.Round(avgLoss*10000) / 10000,
		WeightedAccuracy: math.Round(weightedAccuracy*10000) / 10000,
		NodeWeights:      make(map[string]float64),
		CreatedAt:        time.Now().UTC(),
		ConvergenceDelta: math.Round(convergenceDelta*100000) / 100000,
	}
	for _, u := range validUpdates {
		result.NodeWeights[u.NodeID] = a.getTrustScore(u.NodeID)
	}

	logger.Info(context.Background(), "federation: model aggregated",
		zap.Int("round", a.round),
		zap.Int("nodes", len(validUpdates)),
		zap.Int("samples", totalSamples),
		zap.Float64("accuracy", result.WeightedAccuracy),
		zap.Float64("delta", result.ConvergenceDelta),
	)

	return result, nil
}

func (a *Aggregator) verifyUpdate(update ModelUpdate) error {
	if len(update.Weights) == 0 {
		return fmt.Errorf("empty weights")
	}
	if update.SampleCount < 1 {
		return fmt.Errorf("invalid sample count: %d", update.SampleCount)
	}
	if update.Timestamp.IsZero() || time.Since(update.Timestamp) > 24*time.Hour {
		return fmt.Errorf("stale update timestamp")
	}
	if update.Accuracy < 0 || update.Accuracy > 1 {
		return fmt.Errorf("invalid accuracy: %.2f", update.Accuracy)
	}
	if update.Loss < 0 {
		return fmt.Errorf("negative loss: %.2f", update.Loss)
	}
	for _, w := range update.Weights {
		if math.IsNaN(w) || math.IsInf(w, 0) {
			return fmt.Errorf("invalid weight value")
		}
	}

	return nil
}

func (a *Aggregator) updateReputation(nodeID string, accuracy float64, success bool) {
	rep, exists := a.nodeReputations[nodeID]
	if !exists {
		rep = &NodeReputation{
			NodeID: nodeID,
		}
		a.nodeReputations[nodeID] = rep
	}

	if success {
		rep.TotalUpdates++
		rep.AvgAccuracy = (rep.AvgAccuracy*float64(rep.TotalUpdates-1) + accuracy) / float64(rep.TotalUpdates)
		rep.TrustScore = math.Min(1.0, rep.TrustScore+0.05)
	} else {
		rep.TrustScore = math.Max(0.1, rep.TrustScore-0.2)
	}
	rep.LastSeen = time.Now().UTC()
}

func (a *Aggregator) getTrustScore(nodeID string) float64 {
	rep, ok := a.nodeReputations[nodeID]
	if !ok {
		return 0.5
	}
	return rep.TrustScore
}

func (a *Aggregator) GetNodeReputations() []NodeReputation {
	a.mu.RLock()
	defer a.mu.RUnlock()

	reputations := make([]NodeReputation, 0, len(a.nodeReputations))
	for _, rep := range a.nodeReputations {
		reputations = append(reputations, *rep)
	}
	return reputations
}

func (a *Aggregator) SignModel(model *AggregatedModel) (string, error) {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v%f", model.GlobalWeights, model.CreatedAt)))
	r, s, err := ecdsa.Sign(rand.Reader, a.privateKey, hash[:])
	if err != nil {
		return "", err
	}
	sig := append(r.Bytes(), s.Bytes()...)
	return fmt.Sprintf("%x", sig), nil
}

func (a *Aggregator) generateLaplaceNoise(scale float64) float64 {
	u, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	uniform := float64(u.Int64()) / float64(1<<62)
	noise := -scale * math.Log(1.0-uniform)
	if uniform < 0.5 {
		noise = -noise
	}
	return noise
}

func (a *Aggregator) ExportPublicKey() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(a.publicKey)
	if err != nil {
		return "", err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}
	return string(pem.EncodeToMemory(block)), nil
}
