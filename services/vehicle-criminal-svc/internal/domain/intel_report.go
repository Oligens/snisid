package domain

import (
	"time"

	"github.com/google/uuid"
)

type IntelligenceReport struct {
	ReportID        uuid.UUID   `json:"report_id" db:"report_id"`
	ReportNumber    string      `json:"report_number" db:"report_number"`
	Title           string      `json:"title" db:"title"`
	ReportType      ReportType  `json:"report_type" db:"report_type"`
	Classification  string      `json:"classification" db:"classification"`
	Summary         string      `json:"summary" db:"summary"`
	FullReport      interface{} `json:"full_report,omitempty" db:"full_report"`
	AlertIDs        []uuid.UUID `json:"alert_ids" db:"alert_ids"`
	PlateIDs        []uuid.UUID `json:"plate_ids" db:"plate_ids"`
	PersonIDs       []uuid.UUID `json:"person_ids" db:"person_ids"`
	OriginatingUnit string      `json:"originating_unit" db:"originating_unit"`
	AuthorID        uuid.UUID   `json:"author_id" db:"author_id"`
	RecipientUnits  []string    `json:"recipient_units" db:"recipient_units"`
	PublishedAt     *time.Time  `json:"published_at,omitempty" db:"published_at"`
	ExpiryDate      *time.Time  `json:"expiry_date,omitempty" db:"expiry_date"`
	Attachments     []string    `json:"attachments" db:"attachments"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}
