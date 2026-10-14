package events

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/internal/platform/tracing"
	"go.uber.org/zap"
)

type ProducerConfig struct {
	Brokers      []string
	Topic        string
	BatchSize    int
	BatchTimeout time.Duration
	MaxRetries   int
	Codec        Codec
}

type ProducerInterface interface {
	Publish(ctx context.Context, key string, event any) error
	Close() error
}

type Producer struct {
	config ProducerConfig
	writer *kafka.Writer
	dlq    *kafka.Writer
}

func DefaultConfig(brokers []string, topic string) ProducerConfig {
	return ProducerConfig{
		Brokers:      brokers,
		Topic:        topic,
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		MaxRetries:   3,
		Codec:        &JSONCodec{}, // Default to JSON
	}
}

// NewProducer remains compatible with existing code but allows config upgrades
func NewProducer(brokers []string, topic string) *Producer {
	cfg := DefaultConfig(brokers, topic)
	return NewProducerWithConfig(cfg)
}

func NewProducerWithConfig(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,     // Compression
		BatchSize:    cfg.BatchSize,    // Batching
		BatchTimeout: cfg.BatchTimeout, // Async flush threshold
		Async:        true,             // High throughput async
	}

	dlq := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic + "-dlq", // Dead Letter Queue
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}

	return &Producer{
		config: cfg,
		writer: writer,
		dlq:    dlq,
	}
}

// Publish serializes and pushes an event to Kafka.
// It leverages asynchronous batching, so the function returns quickly.
// Actual delivery errors trigger the DLQ retry loop.
func (p *Producer) Publish(ctx context.Context, key string, event any) error {
	payload, err := p.config.Codec.Encode(event)
	if err != nil {
		return fmt.Errorf("failed to encode event: %w", err)
	}

	// Trace Injection
	corrID := tracing.ExtractCorrelationID(ctx)
	headers := []kafka.Header{
		{Key: tracing.CorrelationHeader, Value: []byte(corrID)},
	}

	msg := kafka.Message{
		Key:     []byte(key),
		Value:   payload,
		Headers: headers,
	}

	// Async publish via segmentio/kafka-go handles its own internal batching logic
	// If you require synchronous publish, disable Async: true in Config.
	// For resilience, we wrap WriteMessages in a retry loop if synchronous mode is ever enabled.

	go p.publishWithRetry(context.Background(), msg, corrID)

	return nil
}

func (p *Producer) publishWithRetry(ctx context.Context, msg kafka.Message, corrID string) {
	var err error
	for attempt := 1; attempt <= p.config.MaxRetries; attempt++ {
		// Attempt publish (sync context)
		publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = p.writer.WriteMessages(publishCtx, msg)
		cancel()

		if err == nil {
			return // Success
		}

		logger.Error(ctx, "kafka publish failed, retrying", err,
			zap.Int("attempt", attempt), zap.String("topic", p.config.Topic))
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond) // Exponential-ish backoff
	}

	// Exhausted retries -> DLQ
	logger.Error(ctx, "kafka max retries exhausted, routing to DLQ", err, zap.String("topic", p.config.Topic))

	dlqCtx, dlqCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dlqCancel()

	if dlqErr := p.dlq.WriteMessages(dlqCtx, msg); dlqErr != nil {
		logger.Fatal(ctx, "CRITICAL: failed to write to DLQ, event lost", dlqErr, zap.String("topic", p.config.Topic))
	}
}

func (p *Producer) Close() error {
	if err := p.dlq.Close(); err != nil {
		fmt.Printf("Error closing dlq: %v\n", err)
	}
	return p.writer.Close()
}
