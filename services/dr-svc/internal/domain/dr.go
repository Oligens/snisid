package domain

import (
	"time"

	"github.com/google/uuid"
)

type HealthStatus string

const (
	HealthHealthy   HealthStatus = "HEALTHY"
	HealthDegraded  HealthStatus = "DEGRADED"
	HealthUnhealthy HealthStatus = "UNHEALTHY"
)

type DRRegion struct {
	RegionID    uuid.UUID    `json:"region_id"`
	Name        string       `json:"name"`
	Endpoint    string       `json:"endpoint"`
	IsActive    bool         `json:"is_active"`
	Health      HealthStatus `json:"health"`
	LastChecked time.Time    `json:"last_checked"`
	CreatedAt   time.Time    `json:"created_at"`
}

type ReplicationStatus struct {
	ReplicationID uuid.UUID `json:"replication_id"`
	SourceRegion  string    `json:"source_region"`
	TargetRegion  string    `json:"target_region"`
	LagSeconds    int       `json:"lag_seconds"`
	IsHealthy     bool      `json:"is_healthy"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}

type FailoverPlan struct {
	PlanID       uuid.UUID `json:"plan_id"`
	Name         string    `json:"name"`
	SourceRegion string    `json:"source_region"`
	TargetRegion string    `json:"target_region"`
	IsAutomated  bool      `json:"is_automated"`
	CreatedAt    time.Time `json:"created_at"`
	IsExecuted   bool      `json:"is_executed"`
}

type FailoverExecution struct {
	ExecutionID  uuid.UUID  `json:"execution_id"`
	PlanID       uuid.UUID  `json:"plan_id"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	IsSuccessful bool       `json:"is_successful"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

type BackupManifest struct {
	ManifestID   uuid.UUID `json:"manifest_id"`
	Region       string    `json:"region"`
	BackupPath   string    `json:"backup_path"`
	BackupSizeMB int64     `json:"backup_size_mb"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
	IsValid      bool      `json:"is_valid"`
}

type RecoveryPoint struct {
	PointID      uuid.UUID  `json:"point_id"`
	ManifestID   uuid.UUID  `json:"manifest_id"`
	RecoveryTime time.Time  `json:"recovery_time"`
	IsRestored   bool       `json:"is_restored"`
	RestoredAt   *time.Time `json:"restored_at,omitempty"`
}

type DRTestResult struct {
	TestID       uuid.UUID `json:"test_id"`
	PlanID       uuid.UUID `json:"plan_id"`
	TestName     string    `json:"test_name"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
	IsSuccessful bool      `json:"is_successful"`
	Details      string    `json:"details,omitempty"`
}
