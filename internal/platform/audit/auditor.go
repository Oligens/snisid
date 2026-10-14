package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/snisid/platform/internal/platform/events"
)

type AuditEvent struct {
	ID         string    `json:"id"`
	Action     string    `json:"action"`
	Actor      string    `json:"actor"`
	Target     string    `json:"target"`
	Context    string    `json:"context"`
	Timestamp  time.Time `json:"timestamp"`
	Compliance bool      `json:"compliance"`
}

type Auditor struct {
	producer events.ProducerInterface
}

func NewAuditor(producer events.ProducerInterface) *Auditor {
	return &Auditor{producer: producer}
}

func (a *Auditor) Log(ctx context.Context, action, actor, target, reason string) error {
	event := AuditEvent{
		ID:         generateID(),
		Action:     action,
		Actor:      actor,
		Target:     target,
		Context:    reason,
		Timestamp:  time.Now().UTC(),
		Compliance: true,
	}

	data, _ := json.Marshal(event)
	return a.producer.Publish(ctx, event.ID, data)
}

func generateID() string {
	return time.Now().Format("20060102150405")
}
