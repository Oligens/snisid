package domain

import (
	"time"

	"github.com/google/uuid"
)

type KidnappingIncident struct {
	IncidentID         uuid.UUID        `json:"incident_id" db:"incident_id"`
	AlertID            uuid.UUID        `json:"alert_id" db:"alert_id"`
	VictimCount        int16            `json:"victim_count" db:"victim_count"`
	VictimSnisidIDs    []uuid.UUID      `json:"victim_snisid_ids" db:"victim_snisid_ids"`
	VictimsNationality []string         `json:"victims_nationality" db:"victims_nationality"`
	VictimsDescription *string          `json:"victims_description,omitempty" db:"victims_description"`
	AbductionDate      time.Time        `json:"abduction_date" db:"abduction_date"`
	AbductionLocation  *string          `json:"abduction_location,omitempty" db:"abduction_location"`
	AbductionDeptCode  *string          `json:"abduction_dept_code,omitempty" db:"abduction_dept_code"`
	AbductionCommune   *string          `json:"abduction_commune,omitempty" db:"abduction_commune"`
	AbductionContext   *string          `json:"abduction_context,omitempty" db:"abduction_context"`
	RansomDemanded     bool             `json:"ransom_demanded" db:"ransom_demanded"`
	RansomAmount       *float64         `json:"ransom_amount,omitempty" db:"ransom_amount"`
	RansomCurrency     string           `json:"ransom_currency" db:"ransom_currency"`
	RansomChannel      *string          `json:"ransom_channel,omitempty" db:"ransom_channel"`
	IncidentStatus     KidnappingStatus `json:"incident_status" db:"incident_status"`
	ResolutionDate     *time.Time       `json:"resolution_date,omitempty" db:"resolution_date"`
	ResolutionLocation *string          `json:"resolution_location,omitempty" db:"resolution_location"`
	ResolutionNotes    *string          `json:"resolution_notes,omitempty" db:"resolution_notes"`
	CaeCaseNumber      *string          `json:"cae_case_number,omitempty" db:"cae_case_number"`
	DcpjCaseNumber     *string          `json:"dcpj_case_number,omitempty" db:"dcpj_case_number"`
	CreatedBy          uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}
