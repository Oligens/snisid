package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CriminalRecord struct {
	RecordID          uuid.UUID  `json:"record_id"`
	NationalFIRID     string     `json:"national_fir_id"`
	SNISIDPersonID    uuid.UUID  `json:"snisid_person_id"`
	AfisSubjectID     *uuid.UUID `json:"afis_subject_id,omitempty"`
	IsHaitianNational bool       `json:"is_haitian_national"`
	Aliases           []string   `json:"aliases"`
	IsActive          bool       `json:"is_active"`
	IsExpunged        bool       `json:"is_expunged"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (r *CriminalRecord) HasActiveConvictions() bool {
	return false
}

type CriminalRecordRepository interface {
	Create(ctx context.Context, record *CriminalRecord) error
	FindByID(ctx context.Context, id uuid.UUID) (*CriminalRecord, error)
	FindByPersonID(ctx context.Context, personID uuid.UUID) (*CriminalRecord, error)
	FindByFIRID(ctx context.Context, firID string) (*CriminalRecord, error)
	Update(ctx context.Context, record *CriminalRecord) error
	NextSequence(ctx context.Context) (int64, error)
	SaveCertificate(ctx context.Context, cert *Certificate) error
	FindCertificateByNumber(ctx context.Context, num string) (*Certificate, error)
}

type ArrestRepository interface {
	Create(ctx context.Context, arrest *Arrest) error
	FindByID(ctx context.Context, id uuid.UUID) (*Arrest, error)
	FindByRecordID(ctx context.Context, recordID uuid.UUID) ([]*Arrest, error)
	Update(ctx context.Context, arrest *Arrest) error
}

type ConvictionRepository interface {
	Create(ctx context.Context, conviction *Conviction) error
	FindByID(ctx context.Context, id uuid.UUID) (*Conviction, error)
	FindByRecordID(ctx context.Context, recordID uuid.UUID) ([]*Conviction, error)
	Update(ctx context.Context, conviction *Conviction) error
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
	DateOfBirth string    `json:"date_of_birth"`
}
