package domain

import "time"

// Source statuses
const (
	SourceStatusActive      = "ACTIVE"
	SourceStatusCompromised = "COMPROMISED"
	SourceStatusTerminated  = "TERMINATED"
	SourceStatusDead        = "DEAD"
)

// Payment frequencies
const (
	PaymentFreqOneTime   = "ONE_TIME"
	PaymentFreqMonthly   = "MONTHLY"
	PaymentFreqPerReport = "PER_REPORT"
)

// Risk levels
const (
	RiskLevelLow      = "LOW"
	RiskLevelMedium   = "MEDIUM"
	RiskLevelHigh     = "HIGH"
	RiskLevelCritical = "CRITICAL"
)

// Report classifications
const (
	ReportClassUnclassified = "UNCLASSIFIED"
	ReportClassConfidential = "CONFIDENTIAL"
	ReportClassSecret       = "SECRET"
	ReportClassTopSecret    = "TOP_SECRET"
)

// Debriefing methods
const (
	DebriefMethodInPerson     = "IN_PERSON"
	DebriefMethodPhone        = "PHONE"
	DebriefMethodEncryptedApp = "ENCRYPTED_APP"
	DebriefMethodDeadDrop     = "DEAD_DROP"
)

type Source struct {
	CodeName          string    `json:"code_name"`
	CredibilityRating int       `json:"credibility_rating"`
	ReliabilityRating string    `json:"reliability_rating"`
	Status            string    `json:"status"`
	HandlingOfficerID string    `json:"handling_officer_id"`
	PaymentAmount     float64   `json:"payment_amount"`
	PaymentFrequency  string    `json:"payment_frequency"`
	RiskLevel         string    `json:"risk_level"`
	Compartment       string    `json:"compartment"`
	ReportsCount      int       `json:"reports_count"`
	FirstRecruitedAt  time.Time `json:"first_recruited_at"`
	LastContactAt     time.Time `json:"last_contact_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type IntelligenceReport struct {
	ID              string    `json:"id"`
	SourceCode      string    `json:"source_code"`
	Classification  string    `json:"classification"`
	ContentHash     string    `json:"content_hash"`
	ThreatActors    []string  `json:"threat_actors"`
	SectorsTargeted []string  `json:"sectors_targeted"`
	VeracityScore   float64   `json:"veracity_score"`
	VerifiedBy      []string  `json:"verified_by"`
	CreatedAt       time.Time `json:"created_at"`
}

type DebriefingSession struct {
	ID                   string    `json:"id"`
	SourceCode           string    `json:"source_code"`
	OfficerID            string    `json:"officer_id"`
	SessionDate          time.Time `json:"session_date"`
	LocationMethod       string    `json:"location_method"`
	TopicsCovered        []string  `json:"topics_covered"`
	NextMeetingPlannedAt time.Time `json:"next_meeting_planned_at"`
	RiskAssessment       string    `json:"risk_assessment"`
	CreatedAt            time.Time `json:"created_at"`
}

// Request structs
type CreateSourceRequest struct {
	CodeName          string  `json:"code_name" binding:"required"`
	CredibilityRating int     `json:"credibility_rating" binding:"required"`
	ReliabilityRating string  `json:"reliability_rating" binding:"required"`
	HandlingOfficerID string  `json:"handling_officer_id" binding:"required"`
	PaymentAmount     float64 `json:"payment_amount"`
	PaymentFrequency  string  `json:"payment_frequency"`
	RiskLevel         string  `json:"risk_level" binding:"required"`
	Compartment       string  `json:"compartment"`
}

type UpdateCredibilityRequest struct {
	CredibilityRating int    `json:"credibility_rating" binding:"required"`
	ReliabilityRating string `json:"reliability_rating"`
}

type SubmitReportRequest struct {
	SourceCode      string   `json:"source_code" binding:"required"`
	Classification  string   `json:"classification" binding:"required"`
	ContentHash     string   `json:"content_hash" binding:"required"`
	ThreatActors    []string `json:"threat_actors"`
	SectorsTargeted []string `json:"sectors_targeted"`
	VeracityScore   float64  `json:"veracity_score"`
	VerifiedBy      []string `json:"verified_by"`
}

type LogDebriefingRequest struct {
	SourceCode           string   `json:"source_code" binding:"required"`
	OfficerID            string   `json:"officer_id" binding:"required"`
	SessionDate          string   `json:"session_date" binding:"required"`
	LocationMethod       string   `json:"location_method" binding:"required"`
	TopicsCovered        []string `json:"topics_covered"`
	NextMeetingPlannedAt string   `json:"next_meeting_planned_at"`
	RiskAssessment       string   `json:"risk_assessment"`
}

// Response structs
type SourceResponse struct {
	Source Source `json:"source"`
}

type SourcesResponse struct {
	Sources []Source `json:"sources"`
	Total   int      `json:"total"`
}

type ReportResponse struct {
	Report IntelligenceReport `json:"report"`
}

type ReportsResponse struct {
	Reports []IntelligenceReport `json:"reports"`
	Total   int                  `json:"total"`
}

type DebriefingResponse struct {
	Debriefing DebriefingSession `json:"debriefing"`
}

type HighRiskResponse struct {
	Sources []Source `json:"sources"`
	Total   int      `json:"total"`
}

type SourceNetworkResponse struct {
	Nodes []SourceNetworkNode `json:"nodes"`
	Edges []SourceNetworkEdge `json:"edges"`
}

type SourceNetworkNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type SourceNetworkEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}
