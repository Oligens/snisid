package domain

import "time"

// Target types
const (
	TargetTypePhoneNumber = "PHONE_NUMBER"
	TargetTypeEmail       = "EMAIL"
	TargetTypeSocialMedia = "SOCIAL_MEDIA"
	TargetTypeRadioFreq   = "RADIO_FREQUENCY"
	TargetTypeIPAddress   = "IP_ADDRESS"
)

// Target statuses
const (
	TargetStatusActive    = "ACTIVE"
	TargetStatusSuspended = "SUSPENDED"
	TargetStatusExpired   = "EXPIRED"
	TargetStatusRevoked   = "REVOKED"
)

// Comm types
const (
	CommTypeCall        = "CALL"
	CommTypeSMS         = "SMS"
	CommTypeEmail       = "EMAIL"
	CommTypeRadio       = "RADIO"
	CommTypeSocialMedia = "SOCIAL_MEDIA"
)

type InterceptionTarget struct {
	ID               string    `json:"id"`
	TargetType       string    `json:"target_type"`
	Status           string    `json:"status"`
	AuthorizationRef string    `json:"authorization_ref"`
	JudgeName        string    `json:"judge_name"`
	IssuingCourt     string    `json:"issuing_court"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	TargetIdentifier string    `json:"target_identifier"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type InterceptedCommunication struct {
	ID             string    `json:"id"`
	SourceTargetID string    `json:"source_target_id"`
	CommType       string    `json:"comm_type"`
	Metadata       string    `json:"metadata"`
	ContentRef     string    `json:"content_ref"`
	InterceptedAt  time.Time `json:"intercepted_at"`
	CollectorNode  string    `json:"collector_node"`
	CaseNumber     string    `json:"case_number"`
	CreatedAt      time.Time `json:"created_at"`
}

type CDRAnalysis struct {
	ID            string    `json:"id"`
	Caller        string    `json:"caller"`
	Callee        string    `json:"callee"`
	Duration      int       `json:"duration"`
	TowerLocation string    `json:"tower_location"`
	IMSI          string    `json:"imsi"`
	IMEI          string    `json:"imei"`
	Timestamp     time.Time `json:"timestamp"`
	CreatedAt     time.Time `json:"created_at"`
}

// Request structs
type CreateTargetRequest struct {
	TargetType       string `json:"target_type" binding:"required"`
	AuthorizationRef string `json:"authorization_ref" binding:"required"`
	JudgeName        string `json:"judge_name" binding:"required"`
	IssuingCourt     string `json:"issuing_court" binding:"required"`
	StartDate        string `json:"start_date" binding:"required"`
	EndDate          string `json:"end_date" binding:"required"`
	TargetIdentifier string `json:"target_identifier" binding:"required"`
}

type InterceptRequest struct {
	CommType      string `json:"comm_type" binding:"required"`
	Metadata      string `json:"metadata"`
	ContentRef    string `json:"content_ref" binding:"required"`
	InterceptedAt string `json:"intercepted_at" binding:"required"`
	CollectorNode string `json:"collector_node" binding:"required"`
	CaseNumber    string `json:"case_number"`
}

type EmergencyRequest struct {
	TargetIdentifier   string `json:"target_identifier" binding:"required"`
	TargetType         string `json:"target_type" binding:"required"`
	Reason             string `json:"reason" binding:"required"`
	AuthorizingOfficer string `json:"authorizing_officer" binding:"required"`
}

// Response structs
type TargetResponse struct {
	Target InterceptionTarget `json:"target"`
}

type TargetsResponse struct {
	Targets []InterceptionTarget `json:"targets"`
	Total   int                  `json:"total"`
}

type CommunicationResponse struct {
	Communication InterceptedCommunication `json:"communication"`
}

type CommunicationsResponse struct {
	Communications []InterceptedCommunication `json:"communications"`
	Total          int                        `json:"total"`
}

type CDRAnalysisResponse struct {
	Records []CDRAnalysis `json:"records"`
	Total   int           `json:"total"`
}

type EmergencyResponse struct {
	Target   InterceptionTarget `json:"target"`
	Approved bool               `json:"approved"`
	AuthRef  string             `json:"authorization_ref"`
}
