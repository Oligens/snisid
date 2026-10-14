package domain

import (
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SevLow      Severity = "LOW"
	SevMedium   Severity = "MEDIUM"
	SevHigh     Severity = "HIGH"
	SevCritical Severity = "CRITICAL"
)

type IncidentStatus string

const (
	IncDetected   IncidentStatus = "DETECTED"
	IncTriaging   IncidentStatus = "TRIAGING"
	IncContained  IncidentStatus = "CONTAINED"
	IncEradicated IncidentStatus = "ERADICATED"
	IncRecovered  IncidentStatus = "RECOVERED"
	IncClosed     IncidentStatus = "CLOSED"
)

type Incident struct {
	ID          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Description *string        `json:"description,omitempty"`
	Severity    Severity       `json:"severity"`
	Status      IncidentStatus `json:"status"`
	SourceIP    *string        `json:"source_ip,omitempty"`
	TargetAsset *string        `json:"target_asset,omitempty"`
	DetectedBy  string         `json:"detected_by"`
	AssignedTo  *string        `json:"assigned_to,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ClosedAt    *time.Time     `json:"closed_at,omitempty"`
}

type ZeroTrustPolicy struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PolicyType  string    `json:"policy_type"`
	Rules       []string  `json:"rules"`
	Enabled     bool      `json:"enabled"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ThreatIndicator struct {
	ID          uuid.UUID  `json:"id"`
	Indicator   string     `json:"indicator"`
	Type        string     `json:"type"`
	ThreatLevel string     `json:"threat_level"`
	Source      string     `json:"source"`
	Description *string    `json:"description,omitempty"`
	Tags        []string   `json:"tags"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateIncidentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description,omitempty"`
	Severity    string `json:"severity" binding:"required"`
	SourceIP    string `json:"source_ip,omitempty"`
	TargetAsset string `json:"target_asset,omitempty"`
	DetectedBy  string `json:"detected_by" binding:"required"`
	AssignedTo  string `json:"assigned_to,omitempty"`
}

type CreatePolicyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description" binding:"required"`
	PolicyType  string   `json:"policy_type" binding:"required"`
	Rules       []string `json:"rules" binding:"required"`
	Enabled     bool     `json:"enabled"`
	CreatedBy   string   `json:"created_by" binding:"required"`
}

type ThreatCheckRequest struct {
	Indicator string `json:"indicator" form:"indicator" binding:"required"`
}
