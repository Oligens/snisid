package domain

import (
	"time"

	"github.com/google/uuid"
)

type KeyAlgorithm string

const (
	AlgorithmRSA  KeyAlgorithm = "RSA"
	AlgorithmEC   KeyAlgorithm = "EC"
	AlgorithmPQC  KeyAlgorithm = "PQC"
	AlgorithmAES  KeyAlgorithm = "AES"
	AlgorithmHMAC KeyAlgorithm = "HMAC"
)

type KeyState string

const (
	KeyStateActive        KeyState = "ACTIVE"
	KeyStateDeactivated   KeyState = "DEACTIVATED"
	KeyStateCompromised   KeyState = "COMPROMISED"
	KeyStateDestroyed     KeyState = "DESTROYED"
	KeyStatePendingRotate KeyState = "PENDING_ROTATE"
)

type KeyUsage string

const (
	UsageSign      KeyUsage = "SIGN"
	UsageVerify    KeyUsage = "VERIFY"
	UsageEncrypt   KeyUsage = "ENCRYPT"
	UsageDecrypt   KeyUsage = "DECRYPT"
	UsageWrapKey   KeyUsage = "WRAP_KEY"
	UsageUnwrapKey KeyUsage = "UNWRAP_KEY"
)

type HSMKey struct {
	KeyID         uuid.UUID    `json:"key_id"`
	KeyLabel      string       `json:"key_label"`
	Algorithm     KeyAlgorithm `json:"algorithm"`
	KeySize       int          `json:"key_size"`
	State         KeyState     `json:"state"`
	Usages        []KeyUsage   `json:"usages"`
	SlotID        int          `json:"slot_id"`
	IsExtractable bool         `json:"is_extractable"`
	PublicKeyPEM  string       `json:"public_key_pem,omitempty"`
	KeyHash       string       `json:"key_hash"`
	RotatedAt     *time.Time   `json:"rotated_at,omitempty"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	CreatedBy     string       `json:"created_by"`
}

type HSMSlot struct {
	SlotID        int    `json:"slot_id"`
	Label         string `json:"label"`
	Manufacturer  string `json:"manufacturer"`
	Model         string `json:"model"`
	SerialNumber  string `json:"serial_number"`
	FirmwareVer   string `json:"firmware_version"`
	IsLoggedIn    bool   `json:"is_logged_in"`
	TokenPresent  bool   `json:"token_present"`
	HardwareModel string `json:"hardware_model"`
}

type KeyGenerationRequest struct {
	Label       string       `json:"label"`
	Algorithm   KeyAlgorithm `json:"algorithm"`
	KeySize     int          `json:"key_size"`
	Usages      []KeyUsage   `json:"usages"`
	SlotID      int          `json:"slot_id"`
	Extractable bool         `json:"extractable"`
	ExpiresIn   string       `json:"expires_in,omitempty"`
	CreatedBy   string       `json:"created_by"`
}

type KeyGenerationResponse struct {
	KeyID        uuid.UUID    `json:"key_id"`
	KeyLabel     string       `json:"key_label"`
	Algorithm    KeyAlgorithm `json:"algorithm"`
	KeySize      int          `json:"key_size"`
	SlotID       int          `json:"slot_id"`
	PublicKeyPEM string       `json:"public_key_pem"`
	CreatedAt    time.Time    `json:"created_at"`
}

type KeyWrapRequest struct {
	TargetKeyID uuid.UUID `json:"target_key_id"`
	WrapKeyID   uuid.UUID `json:"wrap_key_id"`
	Plaintext   string    `json:"plaintext"`
	Algorithm   string    `json:"algorithm,omitempty"`
}

type KeySignRequest struct {
	KeyID     uuid.UUID `json:"key_id"`
	Data      string    `json:"data"`
	Algorithm string    `json:"algorithm,omitempty"`
	Digest    string    `json:"digest,omitempty"`
}
