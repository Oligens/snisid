package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Detention struct {
	DetentionID          uuid.UUID      `json:"detention_id"`
	InmateID             uuid.UUID      `json:"inmate_id"`
	Facility             string         `json:"facility"`
	DetentionBasis       DetentionBasis `json:"detention_basis"`
	LegalStatus          LegalStatus    `json:"legal_status"`
	CaseReference        string         `json:"case_reference"`
	CourtName            string         `json:"court_name"`
	ArrestingAuthority   string         `json:"arresting_authority"`
	WarrantNumber        string         `json:"warrant_number"`
	IntakeDate           time.Time      `json:"intake_date"`
	IntakeOfficer        uuid.UUID      `json:"intake_officer"`
	SentenceDurationDays *int           `json:"sentence_duration_days,omitempty"`
	ReleaseDate          *time.Time     `json:"release_date,omitempty"`
	ReleaseType          ReleaseType    `json:"release_type,omitempty"`
	ReleasingAuthority   string         `json:"releasing_authority"`
	Notes                string         `json:"notes"`
	CreatedAt            time.Time      `json:"created_at"`
}

type DetentionRepository interface {
	Create(ctx context.Context, detention *Detention) error
	FindByID(ctx context.Context, id uuid.UUID) (*Detention, error)
	FindByInmateID(ctx context.Context, inmateID uuid.UUID) ([]*Detention, error)
	GetActiveDetention(ctx context.Context, inmateID uuid.UUID) (*Detention, error)
	Update(ctx context.Context, detention *Detention) error
}

type FacilityOccupancy struct {
	Facility     string `json:"facility"`
	CurrentCount int    `json:"current_count"`
	DeptCode     string `json:"dept_code"`
}
