package domain

import (
	"time"

	"github.com/google/uuid"
)

type SensitivityLevel string

const (
	SensPublic       SensitivityLevel = "PUBLIC"
	SensInternal     SensitivityLevel = "INTERNAL"
	SensConfidential SensitivityLevel = "CONFIDENTIAL"
	SensSecret       SensitivityLevel = "SECRET"
	SensTopSecret    SensitivityLevel = "TOP_SECRET"
)

type AuditAction string

const (
	ActionClassify   AuditAction = "CLASSIFY"
	ActionDowngrade  AuditAction = "DOWNGRADE"
	ActionUpgrade    AuditAction = "UPGRADE"
	ActionDeclassify AuditAction = "DECLASSIFY"
	ActionDestroy    AuditAction = "DESTROY"
)

type ClassificationRule struct {
	ID                  uuid.UUID        `json:"id"`
	DataType            string           `json:"data_type"`
	SensitivityLevel    SensitivityLevel `json:"sensitivity_level"`
	HandlingCaveats     []string         `json:"handling_caveats"`
	DisseminationLimit  *string          `json:"dissemination_limit,omitempty"`
	EncryptionRequired  bool             `json:"encryption_required"`
	AccessControlMFA    bool             `json:"access_control_mfa"`
	AuditLogging        bool             `json:"audit_logging"`
	RetentionDays       int              `json:"retention_days"`
	DestructionRequired bool             `json:"destruction_required"`
	CreatedBy           uuid.UUID        `json:"created_by"`
	Active              bool             `json:"active"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type DataTag struct {
	ID                   uuid.UUID        `json:"id"`
	ResourceURI          string           `json:"resource_uri"`
	ClassificationTop    SensitivityLevel `json:"classification_top_level"`
	ClassificationAtomic string           `json:"classification_atomic"`
	HandlingCaveats      []string         `json:"handling_caveats"`
	OwnerAgency          string           `json:"owner_agency"`
	TaggedBy             uuid.UUID        `json:"tagged_by"`
	TaggedAt             time.Time        `json:"tagged_at"`
	ExpiresAt            *time.Time       `json:"expires_at,omitempty"`
}

type ClassificationAudit struct {
	ID                      uuid.UUID   `json:"id"`
	ResourceURI             string      `json:"resource_uri"`
	Action                  AuditAction `json:"action"`
	FromLevel               *string     `json:"from_level,omitempty"`
	ToLevel                 *string     `json:"to_level,omitempty"`
	Rationale               *string     `json:"rationale,omitempty"`
	AuthorizedBy            uuid.UUID   `json:"authorized_by"`
	ClassificationAuthority string      `json:"classification_authority"`
	Timestamp               time.Time   `json:"timestamp"`
	IPAddress               string      `json:"ip_address"`
}

type CreateRuleRequest struct {
	DataType            string   `json:"data_type" binding:"required"`
	SensitivityLevel    string   `json:"sensitivity_level" binding:"required"`
	HandlingCaveats     []string `json:"handling_caveats"`
	DisseminationLimit  string   `json:"dissemination_limit"`
	EncryptionRequired  bool     `json:"encryption_required"`
	AccessControlMFA    bool     `json:"access_control_mfa"`
	AuditLogging        bool     `json:"audit_logging"`
	RetentionDays       int      `json:"retention_days"`
	DestructionRequired bool     `json:"destruction_required"`
	CreatedBy           string   `json:"created_by" binding:"required"`
}

type TagResourceRequest struct {
	ResourceURI          string   `json:"resource_uri" binding:"required"`
	ClassificationTop    string   `json:"classification_top_level" binding:"required"`
	ClassificationAtomic string   `json:"classification_atomic"`
	HandlingCaveats      []string `json:"handling_caveats"`
	OwnerAgency          string   `json:"owner_agency" binding:"required"`
	TaggedBy             string   `json:"tagged_by" binding:"required"`
}

type LogAuditRequest struct {
	ResourceURI             string `json:"resource_uri" binding:"required"`
	Action                  string `json:"action" binding:"required"`
	FromLevel               string `json:"from_level"`
	ToLevel                 string `json:"to_level"`
	Rationale               string `json:"rationale"`
	AuthorizedBy            string `json:"authorized_by" binding:"required"`
	ClassificationAuthority string `json:"classification_authority" binding:"required"`
	IPAddress               string `json:"ip_address" binding:"required"`
}

type DashboardStats struct {
	TotalRules      int `json:"total_rules"`
	ActiveRules     int `json:"active_rules"`
	TotalTags       int `json:"total_tags"`
	TotalAuditLogs  int `json:"total_audit_logs"`
	ClassifiedCount int `json:"classified_count"`
	SecretCount     int `json:"secret_count"`
	TopSecretCount  int `json:"top_secret_count"`
}
