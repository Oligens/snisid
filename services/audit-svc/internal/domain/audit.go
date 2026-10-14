package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventSource string

const (
	EventSourceServiceDesk EventSource = "SERVICE_DESK"
	EventSourceCivilHT     EventSource = "CIVIL_HT"
	EventSourceBioADN      EventSource = "BIO_ADN"
	EventSourceGovernance  EventSource = "GOVERNANCE"
	EventSourceSLA         EventSource = "SLA"
	EventSourceIdentity    EventSource = "IDENTITY"
	EventSourceSystem      EventSource = "SYSTEM"
)

type EventType string

const (
	EventTypeCreated  EventType = "CREATED"
	EventTypeUpdated  EventType = "UPDATED"
	EventTypeDeleted  EventType = "DELETED"
	EventTypeAccessed EventType = "ACCESSED"
	EventTypeVerified EventType = "VERIFIED"
	EventTypeBreach   EventType = "BREACH"
	EventTypeEscalate EventType = "ESCALATE"
)

type AuditCategory string

const (
	AuditCategoryIdentity    AuditCategory = "IDENTITY"
	AuditCategorySecurity    AuditCategory = "SECURITY"
	AuditCategoryCompliance  AuditCategory = "COMPLIANCE"
	AuditCategoryOperational AuditCategory = "OPERATIONAL"
	AuditCategoryGovernance  AuditCategory = "GOVERNANCE"
)

type AuditEvent struct {
	EventID    uuid.UUID      `json:"event_id"`
	Source     EventSource    `json:"source"`
	EventType  EventType      `json:"event_type"`
	Category   AuditCategory  `json:"category"`
	ActorID    *uuid.UUID     `json:"actor_id,omitempty"`
	ResourceID string         `json:"resource_id"`
	Action     string         `json:"action"`
	Payload    map[string]any `json:"payload,omitempty"`
	Hash       string         `json:"hash"`
	PrevHash   string         `json:"prev_hash"`
	Timestamp  time.Time      `json:"timestamp"`
}

type ImmutableEntry struct {
	EntryID   uuid.UUID      `json:"entry_id"`
	EventID   uuid.UUID      `json:"event_id"`
	Hash      string         `json:"hash"`
	PrevHash  string         `json:"prev_hash"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

type AuditQuery struct {
	Source    *EventSource   `json:"source,omitempty"`
	EventType *EventType     `json:"event_type,omitempty"`
	Category  *AuditCategory `json:"category,omitempty"`
	ActorID   *uuid.UUID     `json:"actor_id,omitempty"`
	From      *time.Time     `json:"from,omitempty"`
	To        *time.Time     `json:"to,omitempty"`
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
}

type RetentionPolicy struct {
	PolicyID      uuid.UUID     `json:"policy_id"`
	Category      AuditCategory `json:"category"`
	RetentionDays int           `json:"retention_days"`
	CreatedAt     time.Time     `json:"created_at"`
}

type AuditReport struct {
	ReportID       uuid.UUID     `json:"report_id"`
	Category       AuditCategory `json:"category"`
	EventCount     int           `json:"event_count"`
	From           time.Time     `json:"from"`
	To             time.Time     `json:"to"`
	IntegrityVerif bool          `json:"integrity_verified"`
	GeneratedAt    time.Time     `json:"generated_at"`
}
