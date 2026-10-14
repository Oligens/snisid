package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/snisid/platform/internal/platform/logger"
	"math"
	"sync"
)

var (
	modelAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "snisid_model_accuracy",
			Help: "Current predictive accuracy of the ML model",
		},
		[]string{"model_version"},
	)

	modelDrift = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "snisid_model_drift_score",
			Help: "Calculated drift score (PSI/KL) for the model",
		},
		[]string{"model_version"},
	)
)

type Metrics struct {
	Total       int
	Correct     int
	Predictions []float64
	mu          sync.Mutex
}

func (m *Metrics) Update(pred float64, label int, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Total++
	predicted := 0
	if pred > 0.5 {
		predicted = 1
	}

	if predicted == label {
		m.Correct++
	}

	m.Predictions = append(m.Predictions, pred)

	// Update Prometheus
	acc := float64(m.Correct) / float64(m.Total)
	modelAccuracy.WithLabelValues(version).Set(acc)
}

func (m *Metrics) CalculateDrift(baseline float64, version string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Predictions) == 0 {
		return 0
	}

	var sum float64
	for _, v := range m.Predictions {
		sum += v
	}
	currentMean := sum / float64(len(m.Predictions))
	drift := math.Abs(currentMean - baseline)

	modelDrift.WithLabelValues(version).Set(drift)
	return drift
}
