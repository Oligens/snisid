package domain

import (
	"time"

	"github.com/google/uuid"
)

type Extradition struct {
	ExtraditionID  uuid.UUID  `json:"extradition_id"`
	NationalRDEPID string     `json:"national_rdep_id"`
	SNISIDPersonID uuid.UUID  `json:"snisid_person_id"`
	FIRRecordID    *uuid.UUID `json:"fir_record_id,omitempty"`

	RequestingCountry DeportationCountry `json:"requesting_country"`
	ExtraditionStatus ExtraditionStatus  `json:"extradition_status"`
	RequestDate       time.Time          `json:"request_date"`
	ApprovalDate      *time.Time         `json:"approval_date,omitempty"`
	ExecutionDate     *time.Time         `json:"execution_date,omitempty"`

	ChargesSummary string  `json:"charges_summary"`
	LegalReference *string `json:"legal_reference,omitempty"`
	TreatyArticle  *string `json:"treaty_article,omitempty"`

	DeparturePort      *string    `json:"departure_port,omitempty"`
	DepartureDeptCode  *string    `json:"departure_dept_code,omitempty"`
	EscortingAgency    *string    `json:"escorting_agency,omitempty"`
	ExtraditionOfficer *uuid.UUID `json:"extradition_officer,omitempty"`

	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateExtraditionRequest struct {
	SNISIDPersonID    uuid.UUID          `json:"snisid_person_id" binding:"required"`
	FIRRecordID       *uuid.UUID         `json:"fir_record_id,omitempty"`
	RequestingCountry DeportationCountry `json:"requesting_country" binding:"required"`
	RequestDate       time.Time          `json:"request_date" binding:"required"`
	ChargesSummary    string             `json:"charges_summary" binding:"required"`
	LegalReference    *string            `json:"legal_reference,omitempty"`
	TreatyArticle     *string            `json:"treaty_article,omitempty"`
	DeparturePort     *string            `json:"departure_port,omitempty"`
	DepartureDeptCode *string            `json:"departure_dept_code,omitempty"`
	EscortingAgency   *string            `json:"escorting_agency,omitempty"`
	Notes             *string            `json:"notes,omitempty"`
}

type UpdateExtraditionStatusRequest struct {
	Status             ExtraditionStatus `json:"status" binding:"required"`
	ExecutionDate      *time.Time        `json:"execution_date,omitempty"`
	ExtraditionOfficer *uuid.UUID        `json:"extradition_officer,omitempty"`
	Notes              *string           `json:"notes,omitempty"`
}
