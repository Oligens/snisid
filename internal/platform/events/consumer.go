package events

import (
	"context"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/internal/platform/tracing"
	"go.uber.org/zap"
)

type ConsumerConfig struct {
	Brokers    []string
	GroupID    string
	Topic      string
	MaxWorkers int
	MaxRetries int
	RetryDelay time.Duration
	Codec      Codec
}

type ConsumerInterface interface {
	Start(ctx context.Context, handler func(ctx context.Context, msg []byte) error) error
	Decode(data []byte, v interface{}) error
	Close() error
}

type Consumer struct {
	config  ConsumerConfig
	reader  *kafka.Reader
	dlq     *Producer // Use the new Producer for DLQ routing
	workers sync.WaitGroup
	quit    chan struct{}
}

func NewConsumer(brokers []string, groupID, topic string) *Consumer {
	cfg := ConsumerConfig{
		Brokers:    brokers,
		GroupID:    groupID,
		Topic:      topic,
		MaxWorkers: 10,
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
		Codec:      &JSONCodec{}, // Default to JSON
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: 0, // Manual commit only
		MinBytes:       10e3,
		MaxBytes:       10e6,
	})

	return &Consumer{
		config: cfg,
		reader: reader,
		dlq:    NewProducer(brokers, topic+"-dlq"),
		quit:   make(chan struct{}),
	}
}

// Start initiates the worker pool and starts reading messages
func (c *Consumer) Start(ctx context.Context, handler func(ctx context.Context, msg []byte) error) error {
	msgChan := make(chan kafka.Message, c.config.MaxWorkers)

	// Launch Workers
	for i := 0; i < c.config.MaxWorkers; i++ {
		c.workers.Add(1)
		go c.worker(ctx, msgChan, handler)
	}

	logger.Info(ctx, "Kafka Consumer started", zap.String("topic", c.config.Topic), zap.Int("workers", c.config.MaxWorkers))

	for {
		select {
		case <-c.quit:
			close(msgChan)
			return nil
		case <-ctx.Done():
			close(msgChan)
			return ctx.Err()
		default:
			m, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				logger.Error(ctx, "Failed to fetch message", err)
				continue
			}
			msgChan <- m
		}
	}
}

// Decode uses the configured codec to decode message bytes into v
func (c *Consumer) Decode(data []byte, v interface{}) error {
	return c.config.Codec.Decode(data, v)
}

func (c *Consumer) worker(ctx context.Context, msgChan <-chan kafka.Message, handler func(ctx context.Context, msg []byte) error) {
	defer c.workers.Done()

	for m := range msgChan {
		// Extract Tracing
		var corrID string
		for _, h := range m.Headers {
			if h.Key == tracing.CorrelationHeader {
				corrID = string(h.Value)
				break
			}
		}
		if corrID == "" {
			corrID = tracing.GenerateCorrelationID()
		}

		workerCtx := tracing.WithCorrelationID(ctx, corrID)

		// Process with Retries
		if err := c.processWithRetry(workerCtx, m, handler); err != nil {
			logger.Error(workerCtx, "Message failed after retries, routing to DLQ", err)
			c.routeToDLQ(workerCtx, m)
		}

		// Manual Commit
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			logger.Error(workerCtx, "Failed to commit message", err)
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, m kafka.Message, handler func(ctx context.Context, msg []byte) error) error {
	var lastErr error
	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(c.config.RetryDelay * time.Duration(i))
			logger.Info(ctx, "Retrying message processing", zap.Int("attempt", i))
		}

		if err := handler(ctx, m.Value); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (c *Consumer) routeToDLQ(ctx context.Context, m kafka.Message) {
	if err := c.dlq.Publish(ctx, string(m.Key), m.Value); err != nil {
		logger.Fatal(ctx, "CRITICAL: Failed to route to DLQ", err)
	}
}

func (c *Consumer) Close() error {
	close(c.quit)
	c.workers.Wait()
	if err := c.dlq.Close(); err != nil {
		logger.Error(context.Background(), "Failed to close DLQ producer", err)
	}
	return c.reader.Close()
}
