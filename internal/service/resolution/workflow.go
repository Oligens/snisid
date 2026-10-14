package resolution

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/events"
	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type Workflow struct {
	arbitrator *Arbitrator
	producer   events.ProducerInterface
}

func NewWorkflow(arb *Arbitrator, prod events.ProducerInterface) *Workflow {
	return &Workflow{
		arbitrator: arb,
		producer:   prod,
	}
}

func (w *Workflow) Merge(ctx context.Context, primaryID string, secondaryIDs []string) error {
	logger.Info(ctx, "Starting identity merge workflow",
		zap.String("primary_id", primaryID),
		zap.Strings("secondary_ids", secondaryIDs),
	)

	// Mock: Fetching data and performing arbitration
	evidence := map[string]string{
		"full_name": "Resolved via national_registry trust level 100",
		"dob":       "Resolved via passport_office trust level 90",
	}

	event := map[string]interface{}{
		"operationType":        "MERGE",
		"primaryIdentityId":    primaryID,
		"secondaryIdentityIds": secondaryIDs,
		"arbitrationEvidence":  evidence,
		"reason":               "Identity consolidation requested via SOC forensic investigation",
	}

	return w.producer.Publish(ctx, primaryID, event)
}

func (w *Workflow) Split(ctx context.Context, originalID string) (string, string, error) {
	logger.Info(ctx, "Starting identity split workflow", zap.String("original_id", originalID))

	newID1 := fmt.Sprintf("%s-A", originalID)
	newID2 := fmt.Sprintf("%s-B", originalID)

	event := map[string]interface{}{
		"operationType":        "SPLIT",
		"primaryIdentityId":    originalID,
		"secondaryIdentityIds": []string{newID1, newID2},
		"arbitrationEvidence":  map[string]string{"split_logic": "Evidence of multiple individuals using the same primary record"},
		"reason":               "Correction of identity overlap detected via biometric anomaly",
	}

	if err := w.producer.Publish(ctx, originalID, event); err != nil {
		return "", "", err
	}

	return newID1, newID2, nil
}
