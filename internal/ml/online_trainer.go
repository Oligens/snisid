package ml

import (
	"fmt"
	"math"
	"sync"
)

type OnlineModel struct {
	mu       sync.RWMutex
	weights  []float64
	lr       float64
	lambda   float64
	nUpdates int64
}

func NewOnlineModel(lr, lambda float64) *OnlineModel {
	return &OnlineModel{
		weights: []float64{0.5, 0.3, 0.2, -0.5},
		lr:      lr,
		lambda:  lambda,
	}
}

func (m *OnlineModel) Predict(fv FeatureVector) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.unsafePredictRaw(fv)
}

func (m *OnlineModel) unsafePredictRaw(fv FeatureVector) float64 {
	z := m.weights[0]*fv.Velocity +
		m.weights[1]*fv.Amount +
		m.weights[2]*fv.GraphRisk +
		m.weights[3]
	return sigmoid(z)
}

func (m *OnlineModel) Update(fv FeatureVector, label float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pred := m.unsafePredictRaw(fv)
	err := pred - label
	features := []float64{fv.Velocity, fv.Amount, fv.GraphRisk, 1.0}

	for i, f := range features {
		grad := err*f + m.lambda*m.weights[i]
		m.weights[i] -= m.lr * grad
	}
	m.nUpdates++
}

func sigmoid(z float64) float64 {
	return 1.0 / (1.0 + math.Exp(-z))
}

func (m *OnlineModel) GetWeights() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]float64, len(m.weights))
	copy(cp, m.weights)
	return cp
}

func (m *OnlineModel) SetWeights(w []float64) error {
	if len(w) != len(m.weights) {
		return fmt.Errorf("weight dimension mismatch: got %d, want %d", len(w), len(m.weights))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	copy(m.weights, w)
	return nil
}

func (m *OnlineModel) GetUpdateCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nUpdates
}
