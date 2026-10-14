package domain

import (
	"time"

	"github.com/google/uuid"
)

type AlertCreatedEvent struct {
	AlertID       uuid.UUID     `json:"alert_id"`
	PlateNumber   string        `json:"plate_number"`
	CrimeCategory CrimeCategory `json:"crime_category"`
	AlertLevel    AlertLevel    `json:"alert_level"`
	ReportingUnit string        `json:"reporting_unit"`
	Timestamp     time.Time     `json:"timestamp"`
}

type AlertUpdatedEvent struct {
	AlertID   uuid.UUID   `json:"alert_id"`
	OldStatus AlertStatus `json:"old_status"`
	NewStatus AlertStatus `json:"new_status"`
	UpdatedBy uuid.UUID   `json:"updated_by"`
	Timestamp time.Time   `json:"timestamp"`
}

type SightingCreatedEvent struct {
	SightingID     uuid.UUID  `json:"sighting_id"`
	PlateNumber    string     `json:"plate_number"`
	MatchedAlertID *uuid.UUID `json:"matched_alert_id,omitempty"`
	AlertTriggered bool       `json:"alert_triggered"`
	Timestamp      time.Time  `json:"timestamp"`
}

type PlateCheckEvent struct {
	PlateNumber      string     `json:"plate_number"`
	HasCriminalAlert bool       `json:"has_criminal_alert"`
	HasStolenPlate   bool       `json:"has_stolen_plate"`
	AlertLevel       AlertLevel `json:"alert_level,omitempty"`
	Source           string     `json:"source"`
	LatencyMs        float64    `json:"latency_ms"`
	Timestamp        time.Time  `json:"timestamp"`
}

type AuditEvent struct {
	AgentID    interface{} `json:"agent_id"`
	Unit       interface{} `json:"unit"`
	Action     string      `json:"action"`
	ResourceID string      `json:"resource_id,omitempty"`
	ClientIP   string      `json:"client_ip"`
	Timestamp  time.Time   `json:"timestamp"`
}
