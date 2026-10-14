package domain

import (
	"time"

	"github.com/google/uuid"
)

type CyberCrimeType string

const (
	MoncashFraud            CyberCrimeType = "MONCASH_FRAUD"
	SIMSwapping             CyberCrimeType = "SIM_SWAPPING"
	Phishing                CyberCrimeType = "PHISHING"
	IdentityTheftDigital    CyberCrimeType = "IDENTITY_THEFT_DIGITAL"
	SystemIntrusion         CyberCrimeType = "SYSTEM_INTRUSION"
	Ransomware              CyberCrimeType = "RANSOMWARE"
	SocialMediaManipulation CyberCrimeType = "SOCIAL_MEDIA_MANIPULATION"
	OnlineScam              CyberCrimeType = "ONLINE_SCAM"
	DigitalExtortion        CyberCrimeType = "DIGITAL_EXTORTION"
	CryptoFraud             CyberCrimeType = "CRYPTO_FRAUD"
	ChildExploitationOnline CyberCrimeType = "CHILD_EXPLOITATION_ONLINE"
	StateSystemAttack       CyberCrimeType = "STATE_SYSTEM_ATTACK"
	OtherCyber              CyberCrimeType = "OTHER"
)

type CyberSeverity string

const (
	CyberLow      CyberSeverity = "LOW"
	CyberMedium   CyberSeverity = "MEDIUM"
	CyberHigh     CyberSeverity = "HIGH"
	CyberCritical CyberSeverity = "CRITICAL"
)

type CyberIncident struct {
	ID                    uuid.UUID      `json:"incident_id" db:"incident_id"`
	NationalCybreID       string         `json:"national_cybre_id" db:"national_cybre_id"`
	CrimeType             CyberCrimeType `json:"crime_type" db:"crime_type"`
	Severity              CyberSeverity  `json:"severity" db:"severity"`
	Status                *string        `json:"status,omitempty" db:"status"`
	VictimCount           *int           `json:"victim_count,omitempty" db:"victim_count"`
	TotalFinancialLossUSD *float64       `json:"total_financial_loss_usd,omitempty" db:"total_financial_loss_usd"`
	IncidentDate          time.Time      `json:"incident_date" db:"incident_date"`
	ReportedDate          time.Time      `json:"reported_date" db:"reported_date"`
	AttackVector          *string        `json:"attack_vector,omitempty" db:"attack_vector"`
	TargetedPlatform      *string        `json:"targeted_platform,omitempty" db:"targeted_platform"`
	SuspectPhone          []string       `json:"suspect_phone" db:"suspect_phone"`
	SuspectEmail          []string       `json:"suspect_email" db:"suspect_email"`
	CaseReference         *string        `json:"case_reference,omitempty" db:"case_reference"`
	CreatedBy             uuid.UUID      `json:"created_by" db:"created_by"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at" db:"updated_at"`
}

type IntrusionAttempt struct {
	ID              uuid.UUID  `json:"attempt_id" db:"attempt_id"`
	IncidentID      *uuid.UUID `json:"incident_id,omitempty" db:"incident_id"`
	TargetSystem    string     `json:"target_system" db:"target_system"`
	AttackTimestamp time.Time  `json:"attack_timestamp" db:"attack_timestamp"`
	AttackType      *string    `json:"attack_type,omitempty" db:"attack_type"`
	SourceIPHash    *string    `json:"source_ip_hash,omitempty" db:"source_ip_hash"`
	SourceCountry   *string    `json:"source_country,omitempty" db:"source_country"`
	WasSuccessful   *bool      `json:"was_successful,omitempty" db:"was_successful"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type ThreatIndicator struct {
	ID              uuid.UUID       `json:"threat_id" db:"threat_id"`
	IndicatorType   string          `json:"indicator_type" db:"indicator_type"`
	IndicatorValue  string          `json:"indicator_value" db:"indicator_value"`
	ThreatCategory  *CyberCrimeType `json:"threat_category,omitempty" db:"threat_category"`
	ConfidenceScore *int            `json:"confidence_score,omitempty" db:"confidence_score"`
	Source          *string         `json:"source,omitempty" db:"source"`
	IsActive        *bool           `json:"is_active,omitempty" db:"is_active"`
	FirstSeen       *time.Time      `json:"first_seen,omitempty" db:"first_seen"`
	LastSeen        *time.Time      `json:"last_seen,omitempty" db:"last_seen"`
	MispRef         *string         `json:"misp_ref,omitempty" db:"misp_ref"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type CyberStats struct {
	CrimeType CyberCrimeType `json:"crime_type" db:"crime_type"`
	Count     int            `json:"count" db:"count"`
}

type DeclareIncidentRequest struct {
	CrimeType             string   `json:"crime_type" binding:"required"`
	Severity              string   `json:"severity"`
	VictimCount           *int     `json:"victim_count"`
	TotalFinancialLossUSD *float64 `json:"total_financial_loss_usd"`
	IncidentDate          string   `json:"incident_date" binding:"required"`
	AttackVector          string   `json:"attack_vector"`
	TargetedPlatform      string   `json:"targeted_platform"`
	CaseReference         string   `json:"case_reference"`
}

type AddThreatIntelRequest struct {
	IndicatorType   string `json:"indicator_type" binding:"required"`
	IndicatorValue  string `json:"indicator_value" binding:"required"`
	ThreatCategory  string `json:"threat_category"`
	ConfidenceScore *int   `json:"confidence_score"`
	Source          string `json:"source"`
}

type CybreRepository interface {
	CreateIncident(incident *CyberIncident) (*CyberIncident, error)
	FindByID(id uuid.UUID) (*CyberIncident, error)
	FindRecentIntrusions() ([]IntrusionAttempt, error)
	CreateThreatIndicator(ti *ThreatIndicator) (*ThreatIndicator, error)
	FindActiveIndicator(indicatorType string, value string) (*ThreatIndicator, error)
	GetStatsByType() ([]CyberStats, error)
}
