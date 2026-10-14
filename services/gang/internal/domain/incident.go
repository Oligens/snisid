package domain

import (
	"time"

	"github.com/google/uuid"
)

type Incident struct {
	IncidentID         uuid.UUID   `json:"incident_id"`
	GangID             uuid.UUID   `json:"gang_id"`
	IncidentType       string      `json:"incident_type"`
	IncidentDate       time.Time   `json:"incident_date"`
	LocationDesc       *string     `json:"location_desc,omitempty"`
	DeptCode           *string     `json:"dept_code,omitempty"`
	Commune            *string     `json:"commune,omitempty"`
	Lat                *float64    `json:"lat,omitempty"`
	Lng                *float64    `json:"lng,omitempty"`
	Casualties         int16       `json:"casualties"`
	VictimIDs          []uuid.UUID `json:"victim_ids"`
	SIVCAlertID        *uuid.UUID  `json:"sivc_alert_id,omitempty"`
	Description        *string     `json:"description,omitempty"`
	IntelligenceSource *string     `json:"intelligence_source,omitempty"`
	CreatedBy          uuid.UUID   `json:"created_by"`
	CreatedAt          time.Time   `json:"created_at"`
}

type CreateIncidentRequest struct {
	GangID             uuid.UUID `json:"gang_id" validate:"required"`
	IncidentType       string    `json:"incident_type" validate:"required"`
	IncidentDate       time.Time `json:"incident_date" validate:"required"`
	LocationDesc       *string   `json:"location_desc"`
	DeptCode           *string   `json:"dept_code"`
	Commune            *string   `json:"commune"`
	Lat                *float64  `json:"lat"`
	Lng                *float64  `json:"lng"`
	Casualties         *int16    `json:"casualties"`
	Description        *string   `json:"description"`
	IntelligenceSource *string   `json:"intelligence_source"`
}
