package lifecycle

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type Engine struct {
	validator *Validator
	producer  events.ProducerInterface
}

func NewEngine(v *Validator, p events.ProducerInterface) *Engine {
	return &Engine{
		validator: v,
		producer:  p,
	}
}

func (e *Engine) Transition(ctx context.Context, id string, from, to State, reason, principal string) error {
	logger.Info(ctx, "Attempting state transition",
		zap.String("identity_id", id),
		zap.String("from", string(from)),
		zap.String("to", string(to)),
	)

	// 1. Validate legal move
	if err := e.validator.ValidateTransition(from, to); err != nil {
		logger.Warn(ctx, "Illegal state transition blocked", zap.Error(err))
		return err
	}

	// 2. Mock: Update Primary DB
	// In production: identityRepo.UpdateState(ctx, id, to)

	// 3. Publish Lifecycle Event
	event := map[string]interface{}{
		"identityId": id,
		"fromState":  string(from),
		"toState":    string(to),
		"reason":     reason,
		"principal":  principal,
	}

	if err := e.producer.Publish(ctx, id, event); err != nil {
		return fmt.Errorf("failed to publish lifecycle event: %w", err)
	}

	logger.Info(ctx, "State transition successful", zap.String("id", id), zap.String("new_state", string(to)))
	return nil
}
