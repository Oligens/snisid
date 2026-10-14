package monitor

import (
	"context"
	"math"
	"sync"

	"github.com/snisid/platform/internal/platform/logger"
)

type ModelMetrics struct {
	mu           sync.Mutex
	Total        int
	Correct      int
	Predictions  []float64
	GroundTruths []int
	DriftScore   float64
}

func (m *ModelMetrics) RecordPrediction(pred float64, actual int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Total++
	m.Predictions = append(m.Predictions, pred)
	m.GroundTruths = append(m.GroundTruths, actual)

	// Simple binary threshold accuracy
	predictedClass := 0
	if pred > 0.5 {
		predictedClass = 1
	}

	if predictedClass == actual {
		m.Correct++
	}

	m.CalculateDrift()
}

func (m *ModelMetrics) Accuracy() float64 {
	if m.Total == 0 {
		return 0
	}
	return float64(m.Correct) / float64(m.Total)
}

func (m *ModelMetrics) CalculateDrift() {
	if len(m.Predictions) < 10 {
		return
	}

	// Mock Drift: Mean absolute deviation from 0.5 baseline
	var sum float64
	for _, p := range m.Predictions {
		sum += p
	}
	avg := sum / float64(len(m.Predictions))
	m.DriftScore = math.Abs(avg - 0.5)

	if m.DriftScore > 0.2 {
		logger.Warn(context.Background(), "MLOPS: Significant model drift detected.")
	}
}
