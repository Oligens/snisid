package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	IncidentID   uuid.UUID   `json:"incident_id"`
	GangID       uuid.UUID   `json:"gang_id"`
	IncidentType string      `json:"incident_type"`
	IncidentDate time.Time   `json:"incident_date"`
	LocationDesc string      `json:"location_desc"`
	DeptCode     string      `json:"dept_code"`
	Commune      string      `json:"commune"`
	Lat          float64     `json:"lat"`
	Lng          float64     `json:"lng"`
	Casualties   int         `json:"casualties"`
	VictimIDs    []uuid.UUID `json:"victim_ids"`
	SIVCAlertID  *uuid.UUID  `json:"sivc_alert_id,omitempty"`
	Description  string      `json:"description"`
	IntelSource  string      `json:"intel_source"`
	CreatedBy    uuid.UUID   `json:"created_by"`
	CreatedAt    time.Time   `json:"created_at"`
}

type IncidentRepository interface {
	Create(ctx context.Context, incident *Incident) error
	FindByID(ctx context.Context, id uuid.UUID) (*Incident, error)
	FindByGangID(ctx context.Context, gangID uuid.UUID) ([]*Incident, error)
	FindByDeptCode(ctx context.Context, deptCode string) ([]*Incident, error)
}

type Alliance struct {
	AllianceID      uuid.UUID  `json:"alliance_id"`
	GangAID         uuid.UUID  `json:"gang_a_id"`
	GangBID         uuid.UUID  `json:"gang_b_id"`
	AllianceType    string     `json:"alliance_type"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	ConfidenceLevel int        `json:"confidence_level"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AllianceRepository interface {
	Create(ctx context.Context, alliance *Alliance) error
	FindByID(ctx context.Context, id uuid.UUID) (*Alliance, error)
	FindByGangID(ctx context.Context, gangID uuid.UUID) ([]*Alliance, error)
	GetAllianceMap(ctx context.Context) ([]*Alliance, error)
}
