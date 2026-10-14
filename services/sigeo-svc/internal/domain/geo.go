package domain

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	ID             uuid.UUID  `json:"event_id" db:"event_id"`
	SourceModule   string     `json:"source_module" db:"source_module"`
	SourceRecordID uuid.UUID  `json:"source_record_id" db:"source_record_id"`
	EventType      string     `json:"event_type" db:"event_type"`
	EventDate      time.Time  `json:"event_date" db:"event_date"`
	Lat            float64    `json:"lat" db:"lat"`
	Lng            float64    `json:"lng" db:"lng"`
	DeptCode       *string    `json:"dept_code,omitempty" db:"dept_code"`
	Commune        *string    `json:"commune,omitempty" db:"commune"`
	H3Index8       *string    `json:"h3_index_8,omitempty" db:"h3_index_8"`
	H3Index10      *string    `json:"h3_index_10,omitempty" db:"h3_index_10"`
	Severity       *int       `json:"severity,omitempty" db:"severity"`
	GangID         *uuid.UUID `json:"gang_id,omitempty" db:"gang_id"`
	Description    *string    `json:"description,omitempty" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type Checkpoint struct {
	ID                uuid.UUID  `json:"cp_id" db:"cp_id"`
	CPType            string     `json:"cp_type" db:"cp_type"`
	DeptCode          *string    `json:"dept_code,omitempty" db:"dept_code"`
	RoadNumber        *string    `json:"road_number,omitempty" db:"road_number"`
	Description       *string    `json:"description,omitempty" db:"description"`
	ControllingGangID *uuid.UUID `json:"controlling_gang_id,omitempty" db:"controlling_gang_id"`
	IsActive          *bool      `json:"is_active,omitempty" db:"is_active"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type ZoneSecurityReport struct {
	DeptCode         string        `json:"dept_code"`
	Commune          string        `json:"commune"`
	Period           time.Duration `json:"period"`
	GeneratedAt      time.Time     `json:"generated_at"`
	IncidentCount    int           `json:"incident_count"`
	OverallRiskScore float64       `json:"overall_risk_score"`
	RiskLevel        string        `json:"risk_level"`
}

type IngestIncidentRequest struct {
	SourceModule   string  `json:"source_module" binding:"required"`
	SourceRecordID string  `json:"source_record_id" binding:"required"`
	EventType      string  `json:"event_type" binding:"required"`
	EventDate      string  `json:"event_date" binding:"required"`
	Lat            float64 `json:"lat" binding:"required"`
	Lng            float64 `json:"lng" binding:"required"`
	DeptCode       string  `json:"dept_code"`
	Commune        string  `json:"commune"`
	Severity       *int    `json:"severity"`
	GangID         string  `json:"gang_id"`
	Description    string  `json:"description"`
}

type GeoRepository interface {
	CreateIncident(incident *Incident) (*Incident, error)
	FindIncidents(deptCode string, since time.Time) ([]Incident, error)
	FindCheckpoints() ([]Checkpoint, error)
	CountIncidentsByZone(deptCode string, since time.Time) (int, error)
}
