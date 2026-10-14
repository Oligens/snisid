package domain

import (
	"time"

	"github.com/google/uuid"
)

type FIPSLevel string

const (
	FIPSLevel1 FIPSLevel = "LEVEL_1"
	FIPSLevel2 FIPSLevel = "LEVEL_2"
	FIPSLevel3 FIPSLevel = "LEVEL_3"
	FIPSLevel4 FIPSLevel = "LEVEL_4"
)

type ValidationStatus string

const (
	StatusPending   ValidationStatus = "PENDING"
	StatusInReview  ValidationStatus = "IN_REVIEW"
	StatusValidated ValidationStatus = "VALIDATED"
	StatusRejected  ValidationStatus = "REJECTED"
	StatusExpired   ValidationStatus = "EXPIRED"
)

type CertAlgo string

const (
	AlgoAES   CertAlgo = "AES"
	AlgoRSA   CertAlgo = "RSA"
	AlgoECDSA CertAlgo = "ECDSA"
	AlgoSHA   CertAlgo = "SHA"
	AlgoHMAC  CertAlgo = "HMAC"
	AlgoDRBG  CertAlgo = "DRBG"
)

type CryptoModule struct {
	ModuleID       uuid.UUID        `json:"module_id"`
	Name           string           `json:"name"`
	Version        string           `json:"version"`
	Vendor         string           `json:"vendor"`
	FIPSLevel      FIPSLevel        `json:"fips_level"`
	Algorithms     []CertAlgo       `json:"algorithms"`
	CertNumber     *string          `json:"cert_number,omitempty"`
	ValidationDate *time.Time       `json:"validation_date,omitempty"`
	ExpiryDate     *time.Time       `json:"expiry_date,omitempty"`
	Status         ValidationStatus `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type SecurityPolicy struct {
	PolicyID     uuid.UUID `json:"policy_id"`
	ModuleID     uuid.UUID `json:"module_id"`
	DocumentRef  string    `json:"document_ref"`
	Version      string    `json:"version"`
	ApprovedDate time.Time `json:"approved_date"`
	Hash         string    `json:"hash"`
}

type CVEScanResult struct {
	ScanID     uuid.UUID `json:"scan_id"`
	ModuleID   uuid.UUID `json:"module_id"`
	CVEID      string    `json:"cve_id"`
	Severity   string    `json:"severity"`
	Discovered time.Time `json:"discovered"`
	Patched    *bool     `json:"patched,omitempty"`
	Notes      *string   `json:"notes,omitempty"`
}

type ComplianceReport struct {
	ServiceName    string    `json:"service_name"`
	OverallStatus  string    `json:"overall_status"`
	ModuleCount    int       `json:"module_count"`
	ValidatedCount int       `json:"validated_count"`
	PendingCount   int       `json:"pending_count"`
	ExpiredCount   int       `json:"expired_count"`
	OpenCVEs       int       `json:"open_cves"`
	LastChecked    time.Time `json:"last_checked"`
}
