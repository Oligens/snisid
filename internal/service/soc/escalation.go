package soc

import (
	"context"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type EscalationManager struct {
	slackWebhook string
	pdServiceKey string
}

func NewEscalationManager(slack, pd string) *EscalationManager {
	return &EscalationManager{
		slackWebhook: slack,
		pdServiceKey: pd,
	}
}

func (m *EscalationManager) Escalate(ctx context.Context, incidentID string, sev Severity, description string) error {
	logger.Info(ctx, "Escalating security incident",
		zap.String("incident_id", incidentID),
		zap.String("severity", string(sev)),
	)

	if sev == SeverityCritical {
		if err := m.notifyPagerDuty(ctx, incidentID, description); err != nil {
			logger.Error(ctx, "Failed to notify PagerDuty", err)
		}
	}

	if sev >= SeverityHigh {
		if err := m.notifySlack(ctx, incidentID, sev, description); err != nil {
			logger.Error(ctx, "Failed to notify Slack", err)
		}
	}

	return nil
}

func (m *EscalationManager) notifyPagerDuty(ctx context.Context, id, desc string) error {
	// Placeholder for PagerDuty API call
	logger.Warn(ctx, "PAGERDUTY ALERT TRIGGERED", zap.String("incident_id", id), zap.String("desc", desc))
	return nil
}

func (m *EscalationManager) notifySlack(ctx context.Context, id string, sev Severity, desc string) error {
	// Placeholder for Slack Webhook call
	logger.Info(ctx, "SLACK NOTIFICATION SENT",
		zap.String("incident_id", id),
		zap.String("severity", string(sev)),
		zap.String("desc", desc),
	)
	return nil
}
