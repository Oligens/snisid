package orchestrator

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type WorkflowEngine struct {
	router   *AIRouter
	producer events.ProducerInterface
}

func NewWorkflowEngine(router *AIRouter, producer events.ProducerInterface) *WorkflowEngine {
	return &WorkflowEngine{
		router:   router,
		producer: producer,
	}
}

func (e *WorkflowEngine) OnboardIdentity(ctx context.Context, id string, media []byte) error {
	logger.Info(ctx, "Starting Identity Onboarding Workflow", zap.String("id", id))

	// Step 1: AI Forensic Analysis
	verdict, err := e.router.DispatchAnalysis(ctx, media)
	if err != nil {
		return err
	}

	// Step 2: Policy Evaluation (Logic based on AI results)
	status := "VERIFIED"
	if verdict.DeepfakeProb > 0.8 || verdict.FraudScore > 50 {
		status = "FLAGGED"
	}

	// Step 3: Trigger Lifecycle State Change
	trigger := map[string]interface{}{
		"identityId":  id,
		"targetState": status,
		"reason":      "Automated onboarding forensic analysis complete",
		"evidence":    verdict,
	}

	if err := e.producer.Publish(ctx, id, trigger); err != nil {
		return fmt.Errorf("failed to trigger lifecycle update: %w", err)
	}

	// Step 4: If FLAGGED, trigger SOC Incident
	if status == "FLAGGED" {
		incident := map[string]interface{}{
			"identityId": id,
			"severity":   "HIGH",
			"type":       "AI_ANOMALY_DETECTED",
		}
		_ = e.producer.Publish(ctx, id, incident)
	}

	logger.Info(ctx, "Identity Onboarding Workflow complete", zap.String("id", id), zap.String("final_status", status))
	return nil
}
