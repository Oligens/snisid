package domain

import (
	"time"

	"github.com/google/uuid"
)

type AgeBracket string

const (
	AgeBracketOver18 AgeBracket = "OVER_18"
	AgeBracketOver21 AgeBracket = "OVER_21"
	AgeBracketOver65 AgeBracket = "OVER_65"
)

type AgeAttestationRequest struct {
	RequestID   uuid.UUID `json:"request_id"`
	IdentityID  uuid.UUID `json:"identity_id"`
	DateOfBirth time.Time `json:"date_of_birth"`
	RequestedAt time.Time `json:"requested_at"`
}

type AgeAttestation struct {
	AttestationID uuid.UUID  `json:"attestation_id"`
	IdentityID    uuid.UUID  `json:"identity_id"`
	DateOfBirth   time.Time  `json:"date_of_birth"`
	IssuedAt      time.Time  `json:"issued_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	IsRevoked     bool       `json:"is_revoked"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

type AgeClaim struct {
	ClaimID       uuid.UUID  `json:"claim_id"`
	AttestationID uuid.UUID  `json:"attestation_id"`
	VerifierID    string     `json:"verifier_id"`
	Bracket       AgeBracket `json:"bracket"`
	IsSatisfied   bool       `json:"is_satisfied"`
	ClaimedAt     time.Time  `json:"claimed_at"`
}

type ZKProof struct {
	ProofID       uuid.UUID `json:"proof_id"`
	AttestationID uuid.UUID `json:"attestation_id"`
	ProofData     string    `json:"proof_data"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type VerifierRequest struct {
	RequestID   uuid.UUID  `json:"request_id"`
	VerifierID  string     `json:"verifier_id"`
	Bracket     AgeBracket `json:"bracket"`
	RequestedAt time.Time  `json:"requested_at"`
	IsApproved  bool       `json:"is_approved"`
}
