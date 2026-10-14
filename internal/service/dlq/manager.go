package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type Manager struct {
	consumer  events.ConsumerInterface
	producers map[string]events.ProducerInterface
	brokers   []string
	mu        sync.RWMutex
}

func NewManager(brokers []string, dlqTopic string) *Manager {
	consumer := events.NewConsumer(brokers, "dlq-manager-group", dlqTopic)

	return &Manager{
		consumer:  consumer,
		producers: make(map[string]events.ProducerInterface),
		brokers:   brokers,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	return m.consumer.Start(ctx, m.handleDLQEvent)
}

func (m *Manager) handleDLQEvent(ctx context.Context, payload []byte) error {
	var dlqEvt map[string]interface{}
	if err := json.Unmarshal(payload, &dlqEvt); err != nil {
		return fmt.Errorf("failed to unmarshal dlq event: %w", err)
	}

	retryCount, _ := dlqEvt["retryCount"].(float64)
	originalTopic, _ := dlqEvt["originalTopic"].(string)

	if retryCount < 3 {
		logger.Info(ctx, "Auto-retrying event",
			zap.String("topic", originalTopic),
			zap.Float64("attempt", retryCount+1),
		)

		// Increment retry and re-publish to original topic
		dlqEvt["retryCount"] = retryCount + 1
		return m.getOrCreateProducer(originalTopic).Publish(ctx, fmt.Sprintf("%v", dlqEvt["originalKey"]), dlqEvt)
	}

	// Max retries reached -> Quarantine
	logger.Warn(ctx, "EVENT QUARANTINED: max retries reached",
		zap.String("topic", originalTopic),
		zap.Any("correlation_id", dlqEvt["header"].(map[string]interface{})["correlationId"]),
	)

	// In a real system, we'd save to Postgres here
	return m.saveToQuarantine(ctx, dlqEvt)
}

func (m *Manager) getOrCreateProducer(topic string) events.ProducerInterface {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.producers[topic]; ok {
		return p
	}

	p := events.NewProducer(m.brokers, topic)
	m.producers[topic] = p
	return p
}

func (m *Manager) saveToQuarantine(ctx context.Context, event map[string]interface{}) error {
	// Mock: log for now, in prod this goes to Postgres failed_events table
	logger.Info(ctx, "Event saved to forensic quarantine ledger", zap.Any("event", event))
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.producers {
		_ = p.Close()
	}
	return m.consumer.Close()
}
