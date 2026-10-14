package domain

import (
	"time"

	"github.com/google/uuid"
)

type IALLevel string

const (
	IALNone IALLevel = "IAL_NONE"
	IAL1    IALLevel = "IAL1"
	IAL2    IALLevel = "IAL2"
	IAL3    IALLevel = "IAL3"
)

type AALLevel string

const (
	AALNone AALLevel = "AAL_NONE"
	AAL1    AALLevel = "AAL1"
	AAL2    AALLevel = "AAL2"
	AAL3    AALLevel = "AAL3"
)

type FALLevel string

const (
	FALNone FALLevel = "FAL_NONE"
	FAL1    FALLevel = "FAL1"
	FAL2    FALLevel = "FAL2"
	FAL3    FALLevel = "FAL3"
)

type AssuranceProfile struct {
	ProfileID    uuid.UUID  `json:"profile_id"`
	IdentityID   uuid.UUID  `json:"identity_id"`
	IAL          IALLevel   `json:"ial"`
	AAL          AALLevel   `json:"aal"`
	FAL          FALLevel   `json:"fal"`
	IsActive     bool       `json:"is_active"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
	LastAssessed time.Time  `json:"last_assessed"`
	AssessorID   string     `json:"assessor_id"`
	AssessorOrg  string     `json:"assessor_org"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type TrustFrameworkClaim struct {
	ClaimID         uuid.UUID  `json:"claim_id"`
	IdentityID      uuid.UUID  `json:"identity_id"`
	FrameworkName   string     `json:"framework_name"`
	ClaimType       string     `json:"claim_type"`
	ClaimValue      string     `json:"claim_value"`
	Issuer          string     `json:"issuer"`
	IssuedAt        time.Time  `json:"issued_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IsVerified      bool       `json:"is_verified"`
	VerificationRef string     `json:"verification_ref,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CertificationAudit struct {
	AuditID     uuid.UUID `json:"audit_id"`
	IdentityID  uuid.UUID `json:"identity_id"`
	Action      string    `json:"action"`
	Field       string    `json:"field,omitempty"`
	OldValue    string    `json:"old_value,omitempty"`
	NewValue    string    `json:"new_value,omitempty"`
	PerformedBy string    `json:"performed_by"`
	PerformedAt time.Time `json:"performed_at"`
	Notes       string    `json:"notes,omitempty"`
}

type ComplianceCheck struct {
	CheckID     uuid.UUID `json:"check_id"`
	IdentityID  uuid.UUID `json:"identity_id"`
	CheckType   string    `json:"check_type"`
	Requirement string    `json:"requirement"`
	IsCompliant bool      `json:"is_compliant"`
	Details     string    `json:"details,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
	CheckedBy   string    `json:"checked_by"`
}
