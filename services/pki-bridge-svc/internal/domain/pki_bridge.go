package domain

import (
	"time"

	"github.com/google/uuid"
)

type ForeignCA struct {
	CAID         uuid.UUID `json:"ca_id"`
	Name         string    `json:"name"`
	Country      string    `json:"country"`
	PublicKeyPEM string    `json:"public_key_pem"`
	CertPolicy   *string   `json:"cert_policy,omitempty"`
	RegisteredAt time.Time `json:"registered_at"`
	Status       string    `json:"status"`
}

type CrossCertificate struct {
	CrossCertID    uuid.UUID `json:"cross_cert_id"`
	Subject        string    `json:"subject"`
	IssuerCAID     uuid.UUID `json:"issuer_ca_id"`
	SerialNumber   string    `json:"serial_number"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	CertificatePEM string    `json:"certificate_pem"`
	CreatedAt      time.Time `json:"created_at"`
}

type TrustAnchor struct {
	AnchorID     uuid.UUID  `json:"anchor_id"`
	Subject      string     `json:"subject"`
	PublicKeyPEM string     `json:"public_key_pem"`
	AddedAt      time.Time  `json:"added_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

type CertificatePath struct {
	PathID        uuid.UUID `json:"path_id"`
	LeafSubject   string    `json:"leaf_subject"`
	Intermediates []string  `json:"intermediates"`
	RootSubject   string    `json:"root_subject"`
	Valid         bool      `json:"valid"`
	ValidatedAt   time.Time `json:"validated_at"`
}

type BridgePolicy struct {
	PolicyID          uuid.UUID `json:"policy_id"`
	Name              string    `json:"name"`
	MinKeySize        int       `json:"min_key_size"`
	AllowedAlgorithms []string  `json:"allowed_algorithms"`
	MaxValidityDays   int       `json:"max_validity_days"`
	RequireCRL        bool      `json:"require_crl"`
}

type BridgeAgreement struct {
	AgreementID uuid.UUID  `json:"agreement_id"`
	Name        string     `json:"name"`
	PartnerCA   string     `json:"partner_ca"`
	PolicyID    uuid.UUID  `json:"policy_id"`
	SignedAt    time.Time  `json:"signed_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Status      string     `json:"status"`
}

type PathValidation struct {
	ValidationID uuid.UUID `json:"validation_id"`
	PathID       uuid.UUID `json:"path_id"`
	Result       bool      `json:"result"`
	Errors       []string  `json:"errors,omitempty"`
	ValidatedAt  time.Time `json:"validated_at"`
}
