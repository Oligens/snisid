package domain

import (
	"time"

	"github.com/google/uuid"
)

type Officer struct {
	OfficerID uuid.UUID `json:"officer_id"`
	SnisidID  uuid.UUID `json:"snisid_id"`
	Badge     string    `json:"badge,omitempty"`
	FullName  string    `json:"full_name"`
	Unit      string    `json:"unit,omitempty"`
	Rank      string    `json:"rank,omitempty"`

	UnderInvestigation bool       `json:"under_investigation"`
	ActiveCaseID       *uuid.UUID `json:"active_case_id,omitempty"`
	InvestigationCount int        `json:"investigation_count"`
	TotalRiskScore     int16      `json:"total_risk_score"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
