package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/data-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDataRepo struct {
	listPipelinesFn              func(ctx context.Context) ([]domain.Pipeline, error)
	createModelFn                func(ctx context.Context, m *domain.MLModel) error
	getModelFn                   func(ctx context.Context, id uuid.UUID) (*domain.MLModel, error)
	getGovernanceAuditsByModelFn func(ctx context.Context, modelID uuid.UUID) ([]domain.GovernanceAudit, error)
	getNationalDashboardFn       func(ctx context.Context) (*domain.NationalDashboard, error)
}

func (m *mockDataRepo) ListPipelines(ctx context.Context) ([]domain.Pipeline, error) {
	return m.listPipelinesFn(ctx)
}
func (m *mockDataRepo) CreateModel(ctx context.Context, model *domain.MLModel) error {
	return m.createModelFn(ctx, model)
}
func (m *mockDataRepo) GetModel(ctx context.Context, id uuid.UUID) (*domain.MLModel, error) {
	return m.getModelFn(ctx, id)
}
func (m *mockDataRepo) GetGovernanceAuditsByModel(ctx context.Context, modelID uuid.UUID) ([]domain.GovernanceAudit, error) {
	return m.getGovernanceAuditsByModelFn(ctx, modelID)
}
func (m *mockDataRepo) GetNationalDashboard(ctx context.Context) (*domain.NationalDashboard, error) {
	return m.getNationalDashboardFn(ctx)
}

func TestListPipelines(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.Pipeline
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: []domain.Pipeline{
				{ID: uuid.New(), Name: "Main Pipeline", Destination: domain.DestinationClickHouse},
			},
		},
		{
			name:    "empty",
			repoRes: []domain.Pipeline{},
		},
		{
			name:    "repo error",
			repoErr: errors.New("query error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDataRepo{
				listPipelinesFn: func(ctx context.Context) ([]domain.Pipeline, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewDataService(repo, nil)
			got, err := svc.ListPipelines(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestRegisterModel(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.RegisterModelRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success basic",
			req: domain.RegisterModelRequest{
				Name:        "FraudDetector",
				ModelType:   "CLASSIFICATION",
				Version:     "1.0.0",
				MlflowRunID: "run-123",
			},
		},
		{
			name: "success with bias",
			req: domain.RegisterModelRequest{
				Name:         "BiasTest",
				ModelType:    "REGRESSION",
				Version:      "2.0",
				MlflowRunID:  "run-456",
				BiasMetric:   "demographic_parity",
				BiasScore:    0.05,
				TrainingDate: "2026-01-15T00:00:00Z",
			},
		},
		{
			name: "repo error",
			req: domain.RegisterModelRequest{
				Name:        "FailModel",
				ModelType:   "CLUSTERING",
				Version:     "1.0",
				MlflowRunID: "run-999",
			},
			repoErr: errors.New("insert error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDataRepo{
				createModelFn: func(ctx context.Context, m *domain.MLModel) error {
					return tt.repoErr
				},
			}
			svc := NewDataService(repo, nil)
			m, err := svc.RegisterModel(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.req.Name, m.Name)
			assert.True(t, m.IsActive)
			if tt.req.BiasMetric != "" {
				require.NotNil(t, m.BiasMetric)
				assert.Equal(t, tt.req.BiasMetric, *m.BiasMetric)
			}
		})
	}
}

func TestGetBiasAudit(t *testing.T) {
	mid := uuid.New()
	model := &domain.MLModel{
		ID:        mid,
		Name:      "FraudDetector",
		ModelType: "CLASSIFICATION",
	}
	audit := domain.GovernanceAudit{
		ID:          uuid.New(),
		ModelID:     mid,
		AuditType:   "BIAS",
		ConductedBy: uuid.New(),
		ConductedAt: time.Now(),
	}
	tests := []struct {
		name     string
		modelID  string
		modelRes *domain.MLModel
		modelErr error
		auditRes []domain.GovernanceAudit
		auditErr error
		wantErr  bool
	}{
		{
			name:     "success with audits",
			modelID:  mid.String(),
			modelRes: model,
			auditRes: []domain.GovernanceAudit{audit},
		},
		{
			name:     "success no audits",
			modelID:  mid.String(),
			modelRes: model,
			auditRes: []domain.GovernanceAudit{},
		},
		{
			name:    "invalid uuid",
			modelID: "bad-uuid",
			wantErr: true,
		},
		{
			name:     "model not found",
			modelID:  mid.String(),
			modelErr: errors.New("model not found"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDataRepo{
				getModelFn: func(ctx context.Context, id uuid.UUID) (*domain.MLModel, error) {
					return tt.modelRes, tt.modelErr
				},
				getGovernanceAuditsByModelFn: func(ctx context.Context, modelID uuid.UUID) ([]domain.GovernanceAudit, error) {
					return tt.auditRes, tt.auditErr
				},
			}
			svc := NewDataService(repo, nil)
			result, err := svc.GetBiasAudit(context.Background(), tt.modelID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, model.Name, result.ModelName)
			assert.Equal(t, len(tt.auditRes), result.AuditCount)
		})
	}
}

func TestGetNationalDashboard(t *testing.T) {
	tests := []struct {
		name    string
		repoRes *domain.NationalDashboard
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: &domain.NationalDashboard{
				TotalPipelines: 12,
				ActiveModels:   8,
				ModelTypeBreakdown: map[string]int{
					"CLASSIFICATION": 5,
					"REGRESSION":     3,
				},
				RecentAudits: 4,
			},
		},
		{
			name:    "repo error",
			repoErr: errors.New("query error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDataRepo{
				getNationalDashboardFn: func(ctx context.Context) (*domain.NationalDashboard, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewDataService(repo, nil)
			got, err := svc.GetNationalDashboard(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.TotalPipelines, got.TotalPipelines)
		})
	}
}
