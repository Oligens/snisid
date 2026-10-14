package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/radiation-safety-svc/internal/domain"
	"github.com/snisid/radiation-safety-svc/internal/kafka"
	"github.com/snisid/radiation-safety-svc/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRadiationRepo struct {
	repository.RadiationRepository
	createSourceFn           func(ctx context.Context, s *domain.RadioactiveSource) error
	updateSourceStatusFn     func(ctx context.Context, id uuid.UUID, status domain.SourceStatus) error
	createAlertFn            func(ctx context.Context, a *domain.RadiationAlert) error
	getUnrespondedAlertsFn   func(ctx context.Context) ([]domain.RadiationAlert, error)
	createChemicalFn         func(ctx context.Context, c *domain.ChemicalPrecursor) error
	getSuspiciousChemicalsFn func(ctx context.Context) ([]domain.ChemicalPrecursor, error)
	getDashboardStatsFn      func(ctx context.Context) (*repository.DashboardStats, error)
}

func (m *mockRadiationRepo) CreateSource(ctx context.Context, s *domain.RadioactiveSource) error {
	return m.createSourceFn(ctx, s)
}
func (m *mockRadiationRepo) UpdateSourceStatus(ctx context.Context, id uuid.UUID, status domain.SourceStatus) error {
	return m.updateSourceStatusFn(ctx, id, status)
}
func (m *mockRadiationRepo) CreateAlert(ctx context.Context, a *domain.RadiationAlert) error {
	return m.createAlertFn(ctx, a)
}
func (m *mockRadiationRepo) GetUnrespondedAlerts(ctx context.Context) ([]domain.RadiationAlert, error) {
	return m.getUnrespondedAlertsFn(ctx)
}
func (m *mockRadiationRepo) CreateChemical(ctx context.Context, c *domain.ChemicalPrecursor) error {
	return m.createChemicalFn(ctx, c)
}
func (m *mockRadiationRepo) GetSuspiciousChemicals(ctx context.Context) ([]domain.ChemicalPrecursor, error) {
	return m.getSuspiciousChemicalsFn(ctx)
}
func (m *mockRadiationRepo) GetDashboardStats(ctx context.Context) (*repository.DashboardStats, error) {
	return m.getDashboardStatsFn(ctx)
}

type mockKafka struct {
	kafka.Producer
	publishFn func(ctx context.Context, key string, msg interface{}) error
}

func (m *mockKafka) Publish(ctx context.Context, key string, msg interface{}) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, key, msg)
	}
	return nil
}

func TestRegisterSource(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				createSourceFn: func(_ context.Context, _ *domain.RadioactiveSource) error {
					return tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			src := &domain.RadioactiveSource{Isotope: "Cs-137"}
			err := svc.RegisterSource(context.Background(), src)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, src.SourceID)
		})
	}
}

func TestUpdateSourceStatus(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "repo error", repoErr: errors.New("not found"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				updateSourceStatusFn: func(_ context.Context, _ uuid.UUID, _ domain.SourceStatus) error {
					return tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			err := svc.UpdateSourceStatus(context.Background(), id, domain.SourceStatusLost)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestCreateAlert(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				createAlertFn: func(_ context.Context, _ *domain.RadiationAlert) error {
					return tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			a := &domain.RadiationAlert{DetectorID: "DET-01"}
			err := svc.CreateAlert(context.Background(), a)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetUnrespondedAlerts(t *testing.T) {
	alerts := []domain.RadiationAlert{{DetectorID: "DET-01"}}
	tests := []struct {
		name    string
		result  []domain.RadiationAlert
		repoErr error
		wantErr bool
	}{
		{name: "success", result: alerts, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				getUnrespondedAlertsFn: func(_ context.Context) ([]domain.RadiationAlert, error) {
					return tt.result, tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			result, err := svc.GetUnrespondedAlerts(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestRegisterChemical(t *testing.T) {
	tests := []struct {
		name       string
		suspicious bool
		repoErr    error
		wantErr    bool
	}{
		{name: "success not suspicious", wantErr: false},
		{name: "success suspicious", suspicious: true, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				createChemicalFn: func(_ context.Context, _ *domain.ChemicalPrecursor) error {
					return tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			c := &domain.ChemicalPrecursor{SubstanceName: "NH3", ReportedSuspicious: tt.suspicious}
			err := svc.RegisterChemical(context.Background(), c)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.suspicious {
				require.NotNil(t, c.FlaggedAt)
			}
		})
	}
}

func TestGetSuspiciousChemicals(t *testing.T) {
	chems := []domain.ChemicalPrecursor{{SubstanceName: "NH3"}}
	tests := []struct {
		name    string
		result  []domain.ChemicalPrecursor
		repoErr error
		wantErr bool
	}{
		{name: "success", result: chems, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				getSuspiciousChemicalsFn: func(_ context.Context) ([]domain.ChemicalPrecursor, error) {
					return tt.result, tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			result, err := svc.GetSuspiciousChemicals(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestGetDashboard(t *testing.T) {
	stats := &repository.DashboardStats{TotalSources: 10}
	tests := []struct {
		name    string
		result  *repository.DashboardStats
		repoErr error
		wantErr bool
	}{
		{name: "success", result: stats, wantErr: false},
		{name: "repo error", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRadiationRepo{
				getDashboardStatsFn: func(_ context.Context) (*repository.DashboardStats, error) {
					return tt.result, tt.repoErr
				},
			}
			svc := NewRadiationService(repo, &mockKafka{})
			result, err := svc.GetDashboard(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestNewRadiationService(t *testing.T) {
	svc := NewRadiationService(&mockRadiationRepo{}, &mockKafka{})
	require.NotNil(t, svc)
}
