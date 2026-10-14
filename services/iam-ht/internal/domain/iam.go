package domain

import (
	"time"

	"github.com/google/uuid"
)

type AssuranceLevel string

const (
	IAL1SelfAsserted      AssuranceLevel = "IAL1_SELF_ASSERTED"
	IAL2BiometricVerified AssuranceLevel = "IAL2_BIOMETRIC_VERIFIED"
	IAL3InPerson          AssuranceLevel = "IAL3_IN_PERSON"
)

type IdentityAssurance struct {
	AssuranceID         uuid.UUID      `json:"assurance_id"`
	CitizenID           uuid.UUID      `json:"citizen_id"`
	KeycloakUserID      string         `json:"keycloak_user_id"`
	AssuranceLevel      AssuranceLevel `json:"assurance_level"`
	BiometricVerifiedAt *time.Time     `json:"biometric_verified_at,omitempty"`
	MFAEnrolled         bool           `json:"mfa_enrolled"`
	LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
}

type AgencyClient struct {
	ClientID               uuid.UUID      `json:"client_id"`
	AgencyName             string         `json:"agency_name"`
	OAuthClientID          string         `json:"oauth_client_id"`
	AllowedScopes          []string       `json:"allowed_scopes"`
	RedirectURIs           []string       `json:"redirect_uris"`
	RequiredAssuranceLevel AssuranceLevel `json:"required_assurance_level"`
	IsActive               bool           `json:"is_active"`
	CreatedAt              time.Time      `json:"created_at"`
}

type AccessLog struct {
	LogID      uuid.UUID  `json:"log_id"`
	CitizenID  *uuid.UUID `json:"citizen_id,omitempty"`
	ClientID   *uuid.UUID `json:"client_id,omitempty"`
	Action     string     `json:"action"`
	IPHash     *string    `json:"ip_hash,omitempty"`
	AccessedAt time.Time  `json:"accessed_at"`
}
