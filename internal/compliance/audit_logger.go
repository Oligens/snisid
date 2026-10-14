package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
)

type AuditEntry struct {
	Actor         string    `json:"actor"`
	Action        string    `json:"action"`
	Resource      string    `json:"resource"`
	Timestamp     time.Time `json:"timestamp"`
	Justification string    `json:"justification"`
}

func LogAccess(actor, action, resource, justification string) {
	entry := AuditEntry{
		Actor:         actor,
		Action:        action,
		Resource:      resource,
		Timestamp:     time.Now(),
		Justification: justification,
	}

	data, _ := json.Marshal(entry)

	// In production, this goes to an immutable WORM S3 bucket or Audit Kafka topic
	logger.Info(context.Background(), fmt.Sprintf("NEXUS-AUDIT: %s", string(data)))
}

func (e *AuditEntry) IsBiometricAccess() bool {
	return e.Action == "VIEW_BIOMETRICS"
}
