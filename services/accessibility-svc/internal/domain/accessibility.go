package domain

import (
	"time"

	"github.com/google/uuid"
)

type WCAGLevel string

const (
	WCAGA   WCAGLevel = "A"
	WCAGAA  WCAGLevel = "AA"
	WCAGAAA WCAGLevel = "AAA"
)

type ElementSelector struct {
	Tag      string `json:"tag"`
	ID       string `json:"id,omitempty"`
	Class    string `json:"class,omitempty"`
	Selector string `json:"selector"`
}

type Violation struct {
	ViolationID  uuid.UUID       `json:"violation_id"`
	AuditRunID   uuid.UUID       `json:"audit_run_id"`
	WCAGLevel    WCAGLevel       `json:"wcag_level"`
	Guideline    string          `json:"guideline"`
	Description  string          `json:"description"`
	Element      ElementSelector `json:"element"`
	Severity     string          `json:"severity"`
	Remediated   bool            `json:"remediated"`
	RemediatedAt *time.Time      `json:"remediated_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AuditRun struct {
	AuditRunID      uuid.UUID  `json:"audit_run_id"`
	TargetURL       string     `json:"target_url"`
	WCAGLevel       WCAGLevel  `json:"wcag_level"`
	Status          string     `json:"status"`
	TotalViolations int        `json:"total_violations"`
	Passed          int        `json:"passed"`
	Failed          int        `json:"failed"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type RemediationTrack struct {
	TrackID      uuid.UUID  `json:"track_id"`
	ViolationID  uuid.UUID  `json:"violation_id"`
	AssignedTo   string     `json:"assigned_to"`
	Notes        *string    `json:"notes,omitempty"`
	RemediatedAt *time.Time `json:"remediated_at,omitempty"`
	VerifiedBy   *string    `json:"verified_by,omitempty"`
	Status       string     `json:"status"`
}

type AuditSchedule struct {
	ScheduleID uuid.UUID  `json:"schedule_id"`
	TargetURL  string     `json:"target_url"`
	WCAGLevel  WCAGLevel  `json:"wcag_level"`
	CronExpr   string     `json:"cron_expr"`
	Enabled    bool       `json:"enabled"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
	NextRunAt  *time.Time `json:"next_run_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type AccessibilityReport struct {
	ReportID    uuid.UUID   `json:"report_id"`
	AuditRunID  uuid.UUID   `json:"audit_run_id"`
	TargetURL   string      `json:"target_url"`
	WCAGLevel   WCAGLevel   `json:"wcag_level"`
	Summary     string      `json:"summary"`
	PassRate    float64     `json:"pass_rate"`
	Violations  []Violation `json:"violations"`
	GeneratedAt time.Time   `json:"generated_at"`
}
