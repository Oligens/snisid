package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	modelAccuracy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snisid_model_accuracy_total",
		Help: "Current prediction accuracy of the national identity model.",
	})

	modelDrift = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snisid_model_drift_score",
		Help: "Detected drift score (deviation from baseline distribution).",
	})
)

func UpdatePrometheusMetrics(m *ModelMetrics) {
	modelAccuracy.Set(m.Accuracy())
	modelDrift.Set(m.DriftScore)
}
