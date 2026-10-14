package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/internal/domain/soc/entity"
	"github.com/snisid/platform/internal/platform/logger"
)

type SOAROrchestrator interface {
	HandleAlert(ctx context.Context, alertType string, payload map[string]interface{}) error
}

type orchestrator struct {
	// Add repositories and external clients (e.g., Auth service to block users)
}

func NewSOAROrchestrator() SOAROrchestrator {
	return &orchestrator{}
}

func (o *orchestrator) HandleAlert(ctx context.Context, alertType string, payload map[string]interface{}) error {
	logger.Info(ctx, fmt.Sprintf("SOC: processing alert %s", alertType))

	// 1. Create Incident
	incident := &entity.Incident{
		ID:            uuid.NewString(),
		Title:         fmt.Sprintf("Autonomous Alert: %s", alertType),
		Status:        entity.StatusNew,
		CreatedAt:     time.Now(),
		CorrelationID: payload["correlation_id"].(string),
	}

	// 2. Identify Severity
	if alertType == "soc.alert.critical" {
		incident.Severity = entity.SeverityCritical
		incident.PlaybookID = "PB-CRITICAL-CONTAINMENT"
		return o.executeCriticalPlaybook(ctx, incident, payload)
	}

	incident.Severity = entity.SeverityMedium
	return nil
}

func (o *orchestrator) executeCriticalPlaybook(ctx context.Context, incident *entity.Incident, payload map[string]interface{}) error {
	logger.Warn(ctx, fmt.Sprintf("SOC: executing Critical Playbook %s", incident.PlaybookID))

	// Step 1: Revoke Sessions (Simulated)
	userID, _ := payload["user_id"].(string)
	action := entity.Action{
		Timestamp: time.Now(),
		Action:    "REVOKE_USER_SESSIONS",
		Result:    fmt.Sprintf("Sessions revoked for user %s", userID),
		Success:   true,
	}
	incident.ActionsTaken = append(incident.ActionsTaken, action)

	// Step 2: Flag for Manual Review
	incident.Status = entity.StatusContained

	logger.Info(ctx, fmt.Sprintf("SOC: Incident %s contained autonomously", incident.ID))
	return nil
}
