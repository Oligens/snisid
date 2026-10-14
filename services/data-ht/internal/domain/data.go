package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Destination string

const (
	DestinationClickHouse   Destination = "CLICKHOUSE"
	DestinationS3Parquet    Destination = "S3_PARQUET"
	DestinationFeatureStore Destination = "FEATURE_STORE"
)

type Pipeline struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	SourceTopics pq.StringArray `json:"source_topics"`
	Destination  Destination    `json:"destination"`
	Config       []byte         `json:"config"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
}

type MLModel struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	ModelType    string     `json:"model_type"`
	Version      string     `json:"version"`
	MlflowRunID  string     `json:"mlflow_run_id"`
	BiasMetric   *string    `json:"bias_metric,omitempty"`
	BiasScore    *float64   `json:"bias_score,omitempty"`
	TrainingDate *time.Time `json:"training_date,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
}

type GovernanceAudit struct {
	ID          uuid.UUID `json:"id"`
	ModelID     uuid.UUID `json:"model_id"`
	AuditType   string    `json:"audit_type"`
	Findings    []byte    `json:"findings"`
	ConductedBy uuid.UUID `json:"conducted_by"`
	ConductedAt time.Time `json:"conducted_at"`
}

type RegisterModelRequest struct {
	Name         string  `json:"name"`
	ModelType    string  `json:"model_type"`
	Version      string  `json:"version"`
	MlflowRunID  string  `json:"mlflow_run_id"`
	BiasMetric   string  `json:"bias_metric,omitempty"`
	BiasScore    float64 `json:"bias_score,omitempty"`
	TrainingDate string  `json:"training_date,omitempty"`
}

type BiasAuditResult struct {
	ModelID     uuid.UUID  `json:"model_id"`
	ModelName   string     `json:"model_name"`
	BiasMetric  *string    `json:"bias_metric"`
	BiasScore   *float64   `json:"bias_score"`
	AuditCount  int        `json:"audit_count"`
	LastAudited *time.Time `json:"last_audited"`
}

type NationalDashboard struct {
	TotalPipelines     int            `json:"total_pipelines"`
	ActiveModels       int            `json:"active_models"`
	ModelTypeBreakdown map[string]int `json:"model_type_breakdown"`
	RecentAudits       int            `json:"recent_audits"`
}
