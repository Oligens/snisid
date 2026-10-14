package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/critical-infra-protection-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	createAssetFn          func(ctx context.Context, a *domain.CriticalAsset) error
	getAssetsBySectorFn    func(ctx context.Context, sector string) ([]domain.CriticalAsset, error)
	createIncidentFn       func(ctx context.Context, inc *domain.InfrastructureIncident) error
	getActiveIncidentsFn   func(ctx context.Context) ([]domain.InfrastructureIncident, error)
	getIncidentsByAssetFn  func(ctx context.Context, assetID uuid.UUID) ([]domain.InfrastructureIncident, error)
	createAssessmentFn     func(ctx context.Context, a *domain.SectorRiskAssessment) error
	getNationalDashboardFn func(ctx context.Context) ([]domain.SectorRiskAssessment, error)
}

func (m *mockRepo) CreateAsset(ctx context.Context, a *domain.CriticalAsset) error {
	return m.createAssetFn(ctx, a)
}
func (m *mockRepo) GetAssetsBySector(ctx context.Context, sector string) ([]domain.CriticalAsset, error) {
	return m.getAssetsBySectorFn(ctx, sector)
}
func (m *mockRepo) CreateIncident(ctx context.Context, inc *domain.InfrastructureIncident) error {
	return m.createIncidentFn(ctx, inc)
}
func (m *mockRepo) GetActiveIncidents(ctx context.Context) ([]domain.InfrastructureIncident, error) {
	return m.getActiveIncidentsFn(ctx)
}
func (m *mockRepo) GetIncidentsByAsset(ctx context.Context, assetID uuid.UUID) ([]domain.InfrastructureIncident, error) {
	return m.getIncidentsByAssetFn(ctx, assetID)
}
func (m *mockRepo) CreateAssessment(ctx context.Context, a *domain.SectorRiskAssessment) error {
	return m.createAssessmentFn(ctx, a)
}
func (m *mockRepo) GetNationalDashboard(ctx context.Context) ([]domain.SectorRiskAssessment, error) {
	return m.getNationalDashboardFn(ctx)
}

func TestCreateAsset(t *testing.T) {
	repo := &mockRepo{
		createAssetFn: func(ctx context.Context, a *domain.CriticalAsset) error { return nil },
	}
	svc := NewInfraProtService(repo, nil)
	req := domain.CreateAssetRequest{
		AssetName:    "Dam-1",
		Sector:       "ENERGY",
		OwnerEntity:  "GovCo",
		LocationLat:  45.0,
		LocationLng:  -93.0,
		Region:       "Midwest",
		DeptCode:     "EN",
		Criticality:  "HIGH",
		ContactName:  "John",
		ContactPhone: "555-0100",
	}
	a, err := svc.CreateAsset(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, a.ID)
	assert.Equal(t, req.AssetName, a.AssetName)
}

func TestCreateAsset_RepoError(t *testing.T) {
	repo := &mockRepo{
		createAssetFn: func(ctx context.Context, a *domain.CriticalAsset) error {
			return errors.New("db error")
		},
	}
	svc := NewInfraProtService(repo, nil)
	_, err := svc.CreateAsset(context.Background(), domain.CreateAssetRequest{})
	assert.Error(t, err)
}

func TestReportIncident(t *testing.T) {
	repo := &mockRepo{
		createIncidentFn: func(ctx context.Context, inc *domain.InfrastructureIncident) error {
			return nil
		},
	}
	svc := NewInfraProtService(repo, nil)
	req := domain.ReportIncidentRequest{
		AssetID:      uuid.New().String(),
		IncidentType: "CYBER_ATTACK",
		Severity:     "HIGH",
		Description:  "ransomware detected",
	}
	inc, err := svc.ReportIncident(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, inc.ID)
	assert.Equal(t, domain.IncStatusReported, inc.Status)
}

func TestCreateAssessment(t *testing.T) {
	repo := &mockRepo{
		createAssessmentFn: func(ctx context.Context, a *domain.SectorRiskAssessment) error {
			return nil
		},
	}
	svc := NewInfraProtService(repo, nil)
	req := domain.CreateAssessmentRequest{
		Sector:           "ENERGY",
		OverallRiskScore: 7,
		AssessorAgency:   "CISA",
	}
	a, err := svc.CreateAssessment(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, a.ID)
	assert.Equal(t, 7, a.OverallRiskScore)
}

func TestGetActiveIncidents(t *testing.T) {
	repo := &mockRepo{
		getActiveIncidentsFn: func(ctx context.Context) ([]domain.InfrastructureIncident, error) {
			return []domain.InfrastructureIncident{{ID: uuid.New()}}, nil
		},
	}
	svc := NewInfraProtService(repo, nil)
	incidents, err := svc.GetActiveIncidents(context.Background())
	require.NoError(t, err)
	assert.Len(t, incidents, 1)
}
