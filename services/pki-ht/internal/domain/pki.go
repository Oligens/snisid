package domain

import (
	"time"

	"github.com/google/uuid"
)

type CAType string

const (
	CARoot                 CAType = "ROOT"
	CAIntermediateCitizens CAType = "INTERMEDIATE_CITIZENS"
	CAIntermediateServices CAType = "INTERMEDIATE_SERVICES"
	CAIntermediateAgencies CAType = "INTERMEDIATE_AGENCIES"
)

type CertStatus string

const (
	CertValid     CertStatus = "VALID"
	CertRevoked   CertStatus = "REVOKED"
	CertExpired   CertStatus = "EXPIRED"
	CertSuspended CertStatus = "SUSPENDED"
)

type SubjectType string

const (
	SubjectCitizen SubjectType = "CITIZEN"
	SubjectService SubjectType = "SERVICE"
	SubjectAgency  SubjectType = "AGENCY"
)

type CertificateAuthority struct {
	CAID         uuid.UUID  `json:"ca_id"`
	CAType       CAType     `json:"ca_type"`
	CommonName   string     `json:"common_name"`
	SerialNumber string     `json:"serial_number"`
	PublicKeyPEM string     `json:"public_key_pem"`
	HSMKeyRef    *string    `json:"hsm_key_ref,omitempty"`
	ParentCAID   *uuid.UUID `json:"parent_ca_id,omitempty"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidUntil   time.Time  `json:"valid_until"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
}

type IssuedCertificate struct {
	CertID           uuid.UUID   `json:"cert_id"`
	SerialNumber     string      `json:"serial_number"`
	IssuingCAID      uuid.UUID   `json:"issuing_ca_id"`
	SubjectType      SubjectType `json:"subject_type"`
	SubjectRef       *uuid.UUID  `json:"subject_ref,omitempty"`
	CommonName       *string     `json:"common_name,omitempty"`
	Status           CertStatus  `json:"status"`
	ValidFrom        time.Time   `json:"valid_from"`
	ValidUntil       time.Time   `json:"valid_until"`
	RevokedAt        *time.Time  `json:"revoked_at,omitempty"`
	RevocationReason *string     `json:"revocation_reason,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

type CRL struct {
	CRLID          uuid.UUID `json:"crl_id"`
	CAID           uuid.UUID `json:"ca_id"`
	CRLNumber      int64     `json:"crl_number"`
	RevokedSerials []string  `json:"revoked_serials"`
	PublishedAt    time.Time `json:"published_at"`
	NextUpdate     time.Time `json:"next_update"`
}

type IssueRequest struct {
	SubjectType string `json:"subject_type"`
	SubjectRef  string `json:"subject_ref,omitempty"`
	CommonName  string `json:"common_name"`
}

type RevokeRequest struct {
	SerialNumber string `json:"serial_number"`
	Reason       string `json:"reason"`
}
