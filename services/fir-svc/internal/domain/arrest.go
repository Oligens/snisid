package domain

import (
	"time"

	"github.com/google/uuid"
)

type Arrest struct {
	ArrestID         uuid.UUID    `json:"arrest_id"`
	RecordID         uuid.UUID    `json:"record_id"`
	ArrestDate       time.Time    `json:"arrest_date"`
	ArrestingUnit    string       `json:"arresting_unit"`
	ArrestingOfficer *uuid.UUID   `json:"arresting_officer,omitempty"`
	ArrestLocation   string       `json:"arrest_location"`
	DeptCode         string       `json:"dept_code"`
	ChargesText      string       `json:"charges_text"`
	OffenseClass     OffenseClass `json:"offense_class"`
	CaseReference    string       `json:"case_reference"`
	ReleaseDate      *time.Time   `json:"release_date,omitempty"`
	ReleaseReason    string       `json:"release_reason,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
}
