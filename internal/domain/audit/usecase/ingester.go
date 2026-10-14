package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/internal/domain/audit/entity"
	"github.com/snisid/platform/internal/domain/audit/repository"
	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"github.com/snisid/platform/internal/platform/security"
)

type AuditIngester interface {
	Start(ctx context.Context)
}

type kafkaIngester struct {
	repo     repository.AuditRepository
	consumer events.ConsumerInterface
}

func NewKafkaIngester(repo repository.AuditRepository, consumer events.ConsumerInterface) AuditIngester {
	return &kafkaIngester{
		repo:     repo,
		consumer: consumer,
	}
}

func (i *kafkaIngester) Start(ctx context.Context) {
	handler := func(handlerCtx context.Context, msg []byte) error {
		// Assuming payload structure is generic for auditing
		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			logger.Error(handlerCtx, "failed to unmarshal audit message", err)
			return err
		}

		// Serialize payload strictly for hashing stability
		stablePayload, _ := json.Marshal(payload)

		// Fetch previous hash (critical section, assumes singleton consumer)
		lastEvent, err := i.repo.GetLastEvent(handlerCtx)
		if err != nil {
			return err
		}

		previousHash := "genesis-hash-snisid"
		if lastEvent != nil {
			previousHash = lastEvent.Hash
		}

		hash := security.GenerateHashChain(previousHash, string(stablePayload))

		// Extract meta if present
		corrID, _ := payload["correlationId"].(string)
		actor, _ := payload["userId"].(string)
		action, _ := payload["action"].(string)
		resource, _ := payload["resource"].(string)
		eventType, _ := payload["eventType"].(string)
		status, _ := payload["status"].(string)
		if status == "" {
			if allowed, ok := payload["allowed"].(bool); ok {
				if allowed {
					status = "success"
				} else {
					status = "denied"
				}
			}
		}

		event := &entity.AuditEvent{
			EventID:       uuid.NewString(),
			CorrelationID: corrID,
			EventType:     eventType,
			Actor:         actor,
			Action:        action,
			Resource:      resource,
			Status:        status,
			Payload:       string(stablePayload),
			PreviousHash:  previousHash,
			Hash:          hash,
			Timestamp:     time.Now().UTC(),
		}

		if err := i.repo.Append(handlerCtx, event); err != nil {
			logger.Error(handlerCtx, "failed to append audit event to ledger", err)
			return err
		}

		return nil
	}

	i.consumer.Start(ctx, handler)
}
