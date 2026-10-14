package replay

import (
	"context"
	"fmt"
	"time"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type ReplayJob struct {
	ID          string
	SourceTopic string
	StartTime   time.Time
	EndTime     time.Time
	Filters     map[string]string
	Status      string
}

type Engine struct {
	producer events.ProducerInterface
	brokers  []string
}

func NewEngine(brokers []string) *Engine {
	// Replay engine re-publishes to a specialized replay topic or original topic
	return &Engine{
		producer: events.NewProducer(brokers, "snisid.replay.ingress"),
		brokers:  brokers,
	}
}

func (e *Engine) RunJob(ctx context.Context, job ReplayJob) error {
	logger.Info(ctx, "Starting replay job",
		zap.String("job_id", job.ID),
		zap.Time("start", job.StartTime),
	)

	// 1. Operational Replay (from Kafka)
	// In a real system, we'd use a Kafka Reader with StartOffset set to timestamp
	if time.Since(job.StartTime) < 7*24*time.Hour {
		return e.replayFromKafka(ctx, job)
	}

	// 2. Forensic Replay (from Audit Service)
	return e.replayFromAudit(ctx, job)
}

func (e *Engine) replayFromKafka(ctx context.Context, job ReplayJob) error {
	// Placeholder: Using the resilient framework to read and re-publish
	logger.Info(ctx, "Replaying events from Kafka offsets", zap.String("topic", job.SourceTopic))

	// Simulation of processing loop
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Inject Replay Headers
			event := map[string]interface{}{
				"eventId":   fmt.Sprintf("replay-%s-%d", job.ID, i),
				"isReplay":  true,
				"original":  "...",
				"timestamp": time.Now().UnixMilli(),
			}

			if err := e.producer.Publish(ctx, job.ID, event); err != nil {
				logger.Error(ctx, "Failed to re-publish event", err)
			}
			time.Sleep(10 * time.Millisecond) // Throttling
		}
	}

	return nil
}

func (e *Engine) replayFromAudit(ctx context.Context, job ReplayJob) error {
	logger.Info(ctx, "Replaying events from Audit Service (Cold Storage)", zap.String("job_id", job.ID))
	// In prod: SELECT * FROM audit_events WHERE timestamp BETWEEN ? AND ? AND ...
	return nil
}
