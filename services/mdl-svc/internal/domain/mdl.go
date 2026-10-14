package domain

import (
	"time"

	"github.com/google/uuid"
)

type MDLIssuance struct {
	IssuanceID uuid.UUID  `json:"issuance_id"`
	IdentityID uuid.UUID  `json:"identity_id"`
	DeviceID   string     `json:"device_id"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	IsRevoked  bool       `json:"is_revoked"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type MDLPresentation struct {
	PresentationID     uuid.UUID `json:"presentation_id"`
	IssuanceID         uuid.UUID `json:"issuance_id"`
	ReaderID           string    `json:"reader_id"`
	PresentedAt        time.Time `json:"presented_at"`
	IsVerified         bool      `json:"is_verified"`
	VerificationResult string    `json:"verification_result,omitempty"`
}

type MDLDataElement struct {
	ElementID    uuid.UUID `json:"element_id"`
	IssuanceID   uuid.UUID `json:"issuance_id"`
	ElementName  string    `json:"element_name"`
	ElementValue string    `json:"element_value"`
	IsMandatory  bool      `json:"is_mandatory"`
}

type DeviceEngagement struct {
	EngagementID   uuid.UUID `json:"engagement_id"`
	IssuanceID     uuid.UUID `json:"issuance_id"`
	QRPayload      string    `json:"qr_payload"`
	EngagementCode string    `json:"engagement_code"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type QRBarcode struct {
	BarcodeID    uuid.UUID `json:"barcode_id"`
	EngagementID uuid.UUID `json:"engagement_id"`
	EncodedData  string    `json:"encoded_data"`
	Format       string    `json:"format"`
	GeneratedAt  time.Time `json:"generated_at"`
}

type MDLVerification struct {
	VerificationID uuid.UUID `json:"verification_id"`
	PresentationID uuid.UUID `json:"presentation_id"`
	VerifiedBy     string    `json:"verified_by"`
	VerifiedAt     time.Time `json:"verified_at"`
	IsAuthentic    bool      `json:"is_authentic"`
	Reason         string    `json:"reason,omitempty"`
}

type MDLTrustRegistry struct {
	EntryID      uuid.UUID `json:"entry_id"`
	ReaderID     string    `json:"reader_id"`
	ReaderName   string    `json:"reader_name"`
	PublicKey    string    `json:"public_key"`
	IsTrusted    bool      `json:"is_trusted"`
	RegisteredAt time.Time `json:"registered_at"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}
