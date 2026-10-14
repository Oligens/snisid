package domain

import (
	"time"

	"github.com/google/uuid"
)

type LicenseType string

const (
	LicenseTypeOSIApproved     LicenseType = "OSI_APPROVED"
	LicenseTypeProprietary     LicenseType = "PROPRIETARY"
	LicenseTypeCreativeCommons LicenseType = "CREATIVE_COMMONS"
	LicenseTypePublicDomain    LicenseType = "PUBLIC_DOMAIN"
	LicenseTypeOther           LicenseType = "OTHER"
)

type ComplianceStatus string

const (
	ComplianceStatusCompliant     ComplianceStatus = "COMPLIANT"
	ComplianceStatusNonCompliant  ComplianceStatus = "NON_COMPLIANT"
	ComplianceStatusPendingReview ComplianceStatus = "PENDING_REVIEW"
	ComplianceStatusExempted      ComplianceStatus = "EXEMPTED"
)

type SoftwareLicense struct {
	LicenseID     uuid.UUID   `json:"license_id"`
	Name          string      `json:"name"`
	SPDXID        string      `json:"spdx_id"`
	LicenseType   LicenseType `json:"license_type"`
	Version       string      `json:"version"`
	Publisher     string      `json:"publisher"`
	IsOsiApproved bool        `json:"is_osi_approved"`
	Text          string      `json:"text,omitempty"`
	RegisteredAt  time.Time   `json:"registered_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type GovernancePolicy struct {
	PolicyID    uuid.UUID `json:"policy_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PolicyRule struct {
	RuleID    uuid.UUID `json:"rule_id"`
	PolicyID  uuid.UUID `json:"policy_id"`
	RuleType  string    `json:"rule_type"`
	Condition string    `json:"condition"`
	Action    string    `json:"action"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

type LicenseAudit struct {
	AuditID    uuid.UUID        `json:"audit_id"`
	LicenseID  uuid.UUID        `json:"license_id"`
	PolicyID   uuid.UUID        `json:"policy_id"`
	Status     ComplianceStatus `json:"status"`
	Findings   string           `json:"findings,omitempty"`
	AuditedAt  time.Time        `json:"audited_at"`
	ReviewedBy string           `json:"reviewed_by,omitempty"`
}

type AttributionReport struct {
	ReportID      uuid.UUID         `json:"report_id"`
	Components    []SoftwareLicense `json:"components"`
	GeneratedAt   time.Time         `json:"generated_at"`
	TotalLicenses int               `json:"total_licenses"`
}
