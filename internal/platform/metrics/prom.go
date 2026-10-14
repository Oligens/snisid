package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RiskEvaluationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "snisid_risk_evaluations_total",
		Help: "The total number of risk evaluation requests processed.",
	})

	RiskEvaluationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "snisid_risk_evaluation_duration_seconds",
		Help:    "Latency of risk evaluation processing.",
		Buckets: prometheus.DefBuckets,
	})

	PolicyDenialsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snisid_policy_denials_total",
		Help: "The total number of requests denied by OPA policy.",
	}, []string{"agency", "reason"})
)
