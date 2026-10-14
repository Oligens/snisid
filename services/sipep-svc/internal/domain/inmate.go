package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Inmate struct {
	InmateID            uuid.UUID  `json:"inmate_id"`
	NationalInmateID    string     `json:"national_inmate_id"`
	SNISIDPersonID      uuid.UUID  `json:"snisid_person_id"`
	FIRRecordID         *uuid.UUID `json:"fir_record_id,omitempty"`
	AFISSubjectID       *uuid.UUID `json:"afis_subject_id,omitempty"`
	CurrentFacility     string     `json:"current_facility"`
	CurrentDeptCode     string     `json:"current_dept_code"`
	CellBlock           string     `json:"cell_block"`
	IsCurrentlyDetained bool       `json:"is_currently_detained"`
	IsMinor             bool       `json:"is_minor"`
	IsFemale            bool       `json:"is_female"`
	HasSpecialNeeds     bool       `json:"has_special_needs"`
	SpecialNeedsNotes   string     `json:"special_needs_notes"`
	IntakeDate          time.Time  `json:"intake_date"`
	ExpectedReleaseDate *time.Time `json:"expected_release_date,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type InmateRepository interface {
	Create(ctx context.Context, inmate *Inmate) error
	FindByID(ctx context.Context, id uuid.UUID) (*Inmate, error)
	FindByPersonID(ctx context.Context, personID uuid.UUID) (*Inmate, error)
	Update(ctx context.Context, inmate *Inmate) error
	Search(ctx context.Context, query string) ([]*Inmate, error)
}

type EventPublisher interface {
	Publish(topic string, event interface{}) error
}

type SNISIDClient interface {
	GetPerson(personID uuid.UUID) (*PersonInfo, error)
}

type PersonInfo struct {
	PersonID    uuid.UUID `json:"person_id"`
	FullName    string    `json:"full_name"`
	Nationality string    `json:"nationality"`
}
