package domain

import (
	"time"

	"github.com/google/uuid"
)

type SLA struct {
	SLAID       uuid.UUID `json:"sla_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SLO struct {
	SLOID          uuid.UUID `json:"slo_id"`
	SLAID          uuid.UUID `json:"sla_id"`
	Name           string    `json:"name"`
	TargetValue    float64   `json:"target_value"`
	Threshold      float64   `json:"threshold"`
	TimeWindowDays int       `json:"time_window_days"`
	CreatedAt      time.Time `json:"created_at"`
}

type ServiceLevelIndicator struct {
	SLIID      uuid.UUID `json:"sli_id"`
	SLOID      uuid.UUID `json:"slo_id"`
	SLAID      uuid.UUID `json:"sla_id"`
	Name       string    `json:"name"`
	Value      float64   `json:"value"`
	RecordedAt time.Time `json:"recorded_at"`
}

type SLIResult struct {
	SLIID       uuid.UUID `json:"sli_id"`
	SLOID       uuid.UUID `json:"slo_id"`
	Name        string    `json:"name"`
	Value       float64   `json:"value"`
	IsBreaching bool      `json:"is_breaching"`
	RecordedAt  time.Time `json:"recorded_at"`
}

type BreachRecord struct {
	BreachID   uuid.UUID  `json:"breach_id"`
	SLAID      uuid.UUID  `json:"sla_id"`
	SLOID      uuid.UUID  `json:"slo_id"`
	SLIValue   float64    `json:"sli_value"`
	Threshold  float64    `json:"threshold"`
	DetectedAt time.Time  `json:"detected_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	IsActive   bool       `json:"is_active"`
}

type UptimeWindow struct {
	WindowID   uuid.UUID `json:"window_id"`
	SLAID      uuid.UUID `json:"sla_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	IsUp       bool      `json:"is_up"`
	DurationMs int64     `json:"duration_ms"`
}

type SLAReport struct {
	ReportID      uuid.UUID          `json:"report_id"`
	SLAID         uuid.UUID          `json:"sla_id"`
	SLOCompliance map[string]float64 `json:"slo_compliance"`
	OverallScore  float64            `json:"overall_score"`
	BreachCount   int                `json:"breach_count"`
	From          time.Time          `json:"from"`
	To            time.Time          `json:"to"`
	GeneratedAt   time.Time          `json:"generated_at"`
}

type EscalationPolicy struct {
	PolicyID      uuid.UUID `json:"policy_id"`
	SLAID         uuid.UUID `json:"sla_id"`
	EscalateAfter int       `json:"escalate_after"`
	NotifyChannel string    `json:"notify_channel"`
	NotifyTarget  string    `json:"notify_target"`
	IsActive      bool      `json:"is_active"`
}
