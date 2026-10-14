package orchestrator

import (
	"context"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type Brain struct {
	engine *WorkflowEngine
}

func NewBrain(engine *WorkflowEngine) *Brain {
	return &Brain{engine: engine}
}

func (b *Brain) ProcessEvent(ctx context.Context, eventType string, payload []byte) error {
	logger.Info(ctx, "Brain received system event", zap.String("type", eventType))

	switch eventType {
	case "identity.created":
		// Auto-launch onboarding for new identities
		return b.engine.OnboardIdentity(ctx, "CITIZEN-TEMP-ID", payload)

	case "soc.resolution.approved":
		// Process manual resolution from SOC
		logger.Info(ctx, "Processing SOC manual resolution")
		return nil

	default:
		logger.Debug(ctx, "Ignoring non-orchestrated event", zap.String("type", eventType))
		return nil
	}
}
