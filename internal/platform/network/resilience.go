package network

import (
	"context"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/sony/gobreaker"
)

type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

func NewCircuitBreaker(name string) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Info(context.Background(), "Circuit breaker state changed")
		},
	}
	return &CircuitBreaker{cb: gobreaker.NewCircuitBreaker(settings)}
}

func (b *CircuitBreaker) Execute(operation func() (interface{}, error)) (interface{}, error) {
	return b.cb.Execute(operation)
}

func WithRetry(operation func() error) error {
	return retry.Do(
		operation,
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.LastErrorOnly(true),
	)
}
