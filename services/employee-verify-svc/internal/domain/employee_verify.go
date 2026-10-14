package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerificationStatus string

const (
	StatusPending          VerificationStatus = "PENDING"
	StatusInProgress       VerificationStatus = "IN_PROGRESS"
	StatusVerified         VerificationStatus = "VERIFIED"
	StatusNotVerified      VerificationStatus = "NOT_VERIFIED"
	StatusAdditionalAction VerificationStatus = "ADDITIONAL_ACTION"
	StatusClosed           VerificationStatus = "CLOSED"
)

type VerificationRequest struct {
	TCN            string             `json:"tcn"`
	EmployerID     uuid.UUID          `json:"employer_id"`
	EmployeeName   string             `json:"employee_name"`
	DocumentNumber string             `json:"document_number"`
	DocumentType   string             `json:"document_type"`
	Status         VerificationStatus `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type EmploymentRecord struct {
	RecordID     uuid.UUID `json:"record_id"`
	EmployerID   uuid.UUID `json:"employer_id"`
	EmployeeName string    `json:"employee_name"`
	HireDate     time.Time `json:"hire_date"`
	Position     string    `json:"position"`
	IsActive     bool      `json:"is_active"`
}

type EmployerRegistration struct {
	EmployerID   uuid.UUID `json:"employer_id"`
	CompanyName  string    `json:"company_name"`
	EIN          string    `json:"ein"`
	Address      string    `json:"address"`
	ContactEmail string    `json:"contact_email"`
	ContactPhone string    `json:"contact_phone"`
	RegisteredAt time.Time `json:"registered_at"`
	IsActive     bool      `json:"is_active"`
}

type VerificationResult struct {
	ResultID    uuid.UUID          `json:"result_id"`
	TCN         string             `json:"tcn"`
	SSAMatch    bool               `json:"ssa_match"`
	DHSMatch    bool               `json:"dhs_match"`
	IsEligible  bool               `json:"is_eligible"`
	Reason      string             `json:"reason,omitempty"`
	CompletedAt time.Time          `json:"completed_at"`
	Status      VerificationStatus `json:"status"`
}

type CaseStatus struct {
	TCN          string             `json:"tcn"`
	Status       VerificationStatus `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	LastActionBy string             `json:"last_action_by"`
}

type CaseHistory struct {
	HistoryID  uuid.UUID `json:"history_id"`
	TCN        string    `json:"tcn"`
	Action     string    `json:"action"`
	ActionedBy string    `json:"actioned_by"`
	ActionedAt time.Time `json:"actioned_at"`
	Details    string    `json:"details,omitempty"`
}
