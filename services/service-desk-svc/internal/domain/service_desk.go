package domain

import (
	"time"

	"github.com/google/uuid"
)

type CaseStatus string

const (
	CaseStatusOpen       CaseStatus = "OPEN"
	CaseStatusInProgress CaseStatus = "IN_PROGRESS"
	CaseStatusResolved   CaseStatus = "RESOLVED"
	CaseStatusClosed     CaseStatus = "CLOSED"
)

type RecoveryMethod string

const (
	RecoveryMethodEmail     RecoveryMethod = "EMAIL"
	RecoveryMethodSMS       RecoveryMethod = "SMS"
	RecoveryMethodDocument  RecoveryMethod = "DOCUMENT"
	RecoveryMethodBiometric RecoveryMethod = "BIOMETRIC"
	RecoveryMethodInPerson  RecoveryMethod = "IN_PERSON"
)

type SupportCase struct {
	CaseID      uuid.UUID  `json:"case_id"`
	CitizenID   uuid.UUID  `json:"citizen_id"`
	Subject     string     `json:"subject"`
	Description string     `json:"description"`
	Status      CaseStatus `json:"status"`
	AssignedTo  *string    `json:"assigned_to,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

type IdentityRecoveryRequest struct {
	RequestID       uuid.UUID        `json:"request_id"`
	CaseID          uuid.UUID        `json:"case_id"`
	CitizenID       uuid.UUID        `json:"citizen_id"`
	PreferredMethod RecoveryMethod   `json:"preferred_method"`
	VerifiedMethods []RecoveryMethod `json:"verified_methods"`
	IsVerified      bool             `json:"is_verified"`
	CreatedAt       time.Time        `json:"created_at"`
	ResolvedAt      *time.Time       `json:"resolved_at,omitempty"`
}

type VerificationChallenge struct {
	ChallengeID uuid.UUID      `json:"challenge_id"`
	CaseID      uuid.UUID      `json:"case_id"`
	Method      RecoveryMethod `json:"method"`
	Challenge   string         `json:"challenge"`
	ExpiresAt   time.Time      `json:"expires_at"`
	IsResolved  bool           `json:"is_resolved"`
	CreatedAt   time.Time      `json:"created_at"`
}

type CaseNote struct {
	NoteID    uuid.UUID `json:"note_id"`
	CaseID    uuid.UUID `json:"case_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Resolution struct {
	ResolutionID uuid.UUID `json:"resolution_id"`
	CaseID       uuid.UUID `json:"case_id"`
	Action       string    `json:"action"`
	Details      string    `json:"details"`
	ResolvedBy   string    `json:"resolved_by"`
	CreatedAt    time.Time `json:"created_at"`
}
