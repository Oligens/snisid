package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type Aggregator struct {
	consumer   events.ConsumerInterface
	bridge     *ClickHouseBridge
	windows    map[string]*MetricWindow
	mu         sync.RWMutex
	windowSize time.Duration
}

func NewAggregator(brokers []string, topics []string, bridge *ClickHouseBridge) *Aggregator {
	// Consumes from multiple topics for stream fusion
	// In a real system, we'd loop over topics or use a pattern
	consumer := events.NewConsumer(brokers, "analytics-group", topics[0])

	return &Aggregator{
		consumer:   consumer,
		bridge:     bridge,
		windows:    make(map[string]*MetricWindow),
		windowSize: 1 * time.Minute,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	go a.bridge.Start(ctx)
	return a.consumer.Start(ctx, a.processEvent)
}

func (a *Aggregator) processEvent(ctx context.Context, payload []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return err
	}

	// 1. Stream Fusion: Map raw event to FusedEvent
	event := FusedEvent{
		EventID:       fmt.Sprintf("%v", raw["eventId"]),
		CorrelationID: fmt.Sprintf("%v", raw["correlationId"]),
		Type:          fmt.Sprintf("%v", raw["action"]),
		Source:        fmt.Sprintf("%v", raw["source"]),
		Timestamp:     time.Now(), // In prod: parse from event
		Data:          raw,
	}

	// 2. Batch to ClickHouse
	a.bridge.BufferEvent(event)

	// 3. Stateful Windowed Aggregation
	a.aggregate(event.Type)

	return nil
}

func (a *Aggregator) aggregate(metricName string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	win, ok := a.windows[metricName]
	if !ok {
		win = &MetricWindow{
			MetricName: metricName,
			Timestamp:  time.Now().Truncate(a.windowSize),
		}
		a.windows[metricName] = win
	}

	win.Count++

	// Spike Detection Logic
	if win.Count > 1000 { // Simple threshold for demonstration
		logger.Warn(context.Background(), "ANALYTICAL SPIKE DETECTED",
			zap.String("metric", metricName),
			zap.Int64("count", win.Count),
		)
	}
}

func (a *Aggregator) GetWindows() map[string]MetricWindow {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make(map[string]MetricWindow)
	for k, v := range a.windows {
		results[k] = *v
	}
	return results
}
