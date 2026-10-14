package domain

import (
	"github.com/google/uuid"
	"time"
)

type Agency struct {
	AgencyID          uuid.UUID `json:"agency_id"`
	AgencyCode        string    `json:"agency_code"`
	AgencyName        string    `json:"agency_name"`
	SecurityServerURL string    `json:"security_server_url"`
	PublicKeyCertRef  *string   `json:"public_key_cert_ref,omitempty"`
	IsActive          bool      `json:"is_active"`
	OnboardedAt       time.Time `json:"onboarded_at"`
}

type DataExchangeAgreement struct {
	AgreementID      uuid.UUID `json:"agreement_id"`
	ProviderAgencyID uuid.UUID `json:"provider_agency_id"`
	ConsumerAgencyID uuid.UUID `json:"consumer_agency_id"`
	ServiceName      string    `json:"service_name"`
	AllowedFields    []string  `json:"allowed_fields"`
	LegalBasis       *string   `json:"legal_basis,omitempty"`
	RateLimitPerMin  int       `json:"rate_limit_per_min"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

type ExchangeLog struct {
	LogID             uuid.UUID `json:"log_id"`
	AgreementID       uuid.UUID `json:"agreement_id"`
	RequestHash       string    `json:"request_hash"`
	ResponseSizeBytes int       `json:"response_size_bytes,omitempty"`
	StatusCode        int       `json:"status_code"`
	DurationMs        int       `json:"duration_ms"`
	ExchangedAt       time.Time `json:"exchanged_at"`
}
