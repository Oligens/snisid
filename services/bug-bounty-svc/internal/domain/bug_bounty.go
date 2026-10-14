package domain

import (
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

type ProgramScope struct {
	ScopeID   uuid.UUID `json:"scope_id"`
	ProgramID uuid.UUID `json:"program_id"`
	Target    string    `json:"target"`
	ScopeType string    `json:"scope_type"`
	InScope   bool      `json:"in_scope"`
	RewardMin *float64  `json:"reward_min,omitempty"`
	RewardMax *float64  `json:"reward_max,omitempty"`
}

type VulnerabilityReport struct {
	ReportID    uuid.UUID  `json:"report_id"`
	ProgramID   uuid.UUID  `json:"program_id"`
	Submitter   string     `json:"submitter"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Severity    Severity   `json:"severity"`
	ScopeID     *uuid.UUID `json:"scope_id,omitempty"`
	Status      string     `json:"status"`
	SubmittedAt time.Time  `json:"submitted_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TriageResult struct {
	TriageID     uuid.UUID  `json:"triage_id"`
	ReportID     uuid.UUID  `json:"report_id"`
	Triager      string     `json:"triager"`
	Severity     Severity   `json:"severity"`
	Reproducible bool       `json:"reproducible"`
	DuplicateOf  *uuid.UUID `json:"duplicate_of,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	TriagedAt    time.Time  `json:"triaged_at"`
}

type Reward struct {
	RewardID   uuid.UUID `json:"reward_id"`
	ReportID   uuid.UUID `json:"report_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	PaidTo     string    `json:"paid_to"`
	ApprovedBy string    `json:"approved_by"`
	PaidAt     time.Time `json:"paid_at"`
}

type PentestEngagement struct {
	EngagementID uuid.UUID  `json:"engagement_id"`
	ProgramID    uuid.UUID  `json:"program_id"`
	Title        string     `json:"title"`
	Scope        string     `json:"scope"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	TeamLead     string     `json:"team_lead"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

type RetestSchedule struct {
	ScheduleID   uuid.UUID  `json:"schedule_id"`
	ReportID     uuid.UUID  `json:"report_id"`
	ScheduledFor time.Time  `json:"scheduled_for"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	AssignedTo   string     `json:"assigned_to"`
	Status       string     `json:"status"`
}
