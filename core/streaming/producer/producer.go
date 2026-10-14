package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type EventBackbone struct {
	Writer *kafka.Writer
	DLQ    *kafka.Writer
}

func NewEventBackbone(brokers []string) *EventBackbone {
	return &EventBackbone{
		Writer: &kafka.Writer{
			Addr:        kafka.TCP(brokers...),
			Topic:       "snisid.fraud.signals",
			Balancer:    &kafka.LeastBytes{},
			Async:       false, // Exactly-once semantics (sync writing)
			MaxAttempts: 5,
		},
		DLQ: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "snisid.fraud.dlq",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (b *EventBackbone) Publish(ctx context.Context, key string, payload []byte) error {
	// Start OTel span
	tr := otel.Tracer("snisid-producer")
	ctx, span := tr.Start(ctx, "kafka-publish")
	defer span.End()

	msg := kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	err := b.Writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("⚠️ KAFKA-BACKBONE: Primary publish failed, routing to DLQ: %v", err)
		_ = b.DLQ.WriteMessages(ctx, msg) // Attempt DLQ routing
		return err
	}

	fmt.Printf("📡 KAFKA-BACKBONE: Event published to snisid.fraud.signals (Key: %s)\n", key)
	return nil
}
