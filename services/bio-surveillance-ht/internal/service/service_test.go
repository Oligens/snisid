package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/bio-surveillance-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBioRepo struct {
	createAlertFn          func(ctx context.Context, a *domain.DiseaseAlert) error
	getActiveAlertsFn      func(ctx context.Context) ([]domain.DiseaseAlert, error)
	getAlertsByRegionFn    func(ctx context.Context, region string) ([]domain.DiseaseAlert, error)
	createCampaignFn       func(ctx context.Context, c *domain.VaccinationCampaign) error
	getCampaignCoverageFn  func(ctx context.Context, id uuid.UUID) (*domain.VaccinationCampaign, error)
	updateFacilityStockFn  func(ctx context.Context, id uuid.UUID, req domain.UpdateFacilityStockRequest) (*domain.HealthFacility, error)
	getDashboardNationalFn func(ctx context.Context) (*domain.DashboardNational, error)
}

func (m *mockBioRepo) CreateAlert(ctx context.Context, a *domain.DiseaseAlert) error {
	return m.createAlertFn(ctx, a)
}
func (m *mockBioRepo) GetActiveAlerts(ctx context.Context) ([]domain.DiseaseAlert, error) {
	return m.getActiveAlertsFn(ctx)
}
func (m *mockBioRepo) GetAlertsByRegion(ctx context.Context, region string) ([]domain.DiseaseAlert, error) {
	return m.getAlertsByRegionFn(ctx, region)
}
func (m *mockBioRepo) CreateCampaign(ctx context.Context, c *domain.VaccinationCampaign) error {
	return m.createCampaignFn(ctx, c)
}
func (m *mockBioRepo) GetCampaignCoverage(ctx context.Context, id uuid.UUID) (*domain.VaccinationCampaign, error) {
	return m.getCampaignCoverageFn(ctx, id)
}
func (m *mockBioRepo) UpdateFacilityStock(ctx context.Context, id uuid.UUID, req domain.UpdateFacilityStockRequest) (*domain.HealthFacility, error) {
	return m.updateFacilityStockFn(ctx, id, req)
}
func (m *mockBioRepo) GetDashboardNational(ctx context.Context) (*domain.DashboardNational, error) {
	return m.getDashboardNationalFn(ctx)
}

func TestCreateAlert(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateDiseaseAlertRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.CreateDiseaseAlertRequest{
				DiseaseName:      "COVID-19",
				PathogenType:     "VIRUS",
				Icd10Code:        "U07.1",
				AlertLevel:       "RED",
				TransmissionMode: "AIRBORNE",
				IncubationDays:   14,
				FatalityRate:     2.5,
				CasesConfirmed:   100,
				CasesSuspected:   50,
				CasesDeaths:      5,
				AffectedRegions:  []string{"Ouest", "Nord"},
			},
		},
		{
			name: "repo error",
			req: domain.CreateDiseaseAlertRequest{
				DiseaseName:      "Ebola",
				PathogenType:     "VIRUS",
				Icd10Code:        "A98.4",
				AlertLevel:       "RED",
				TransmissionMode: "CONTACT",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBioRepo{
				createAlertFn: func(ctx context.Context, a *domain.DiseaseAlert) error {
					return tt.repoErr
				},
			}
			svc := NewBioSurveillanceService(repo, nil)
			a, err := svc.CreateAlert(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.req.DiseaseName, a.DiseaseName)
			assert.Equal(t, domain.AlertLevel(tt.req.AlertLevel), a.AlertLevel)
		})
	}
}

func TestGetActiveAlerts(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.DiseaseAlert
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: []domain.DiseaseAlert{
				{ID: uuid.New(), DiseaseName: "Dengue", AlertLevel: domain.AlertOrange},
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
			repo := &mockBioRepo{
				getActiveAlertsFn: func(ctx context.Context) ([]domain.DiseaseAlert, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewBioSurveillanceService(repo, nil)
			got, err := svc.GetActiveAlerts(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestCreateCampaign(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateVaccinationCampaignRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.CreateVaccinationCampaignRequest{
				CampaignName:      "Vax-Ouest",
				TargetDisease:     "Polio",
				VaccineType:       "OPV",
				TargetPopulation:  50000,
				DosesAdministered: 35000,
				CoveragePct:       70.0,
				RegionsActive:     []string{"Ouest"},
				StartDate:         "2026-01-15T00:00:00Z",
				CoordinatorAgency: "MSPP",
			},
		},
		{
			name: "repo error",
			req: domain.CreateVaccinationCampaignRequest{
				CampaignName:      "FailCamp",
				TargetDisease:     "Measles",
				VaccineType:       "MR",
				TargetPopulation:  1000,
				DosesAdministered: 0,
				CoveragePct:       0,
				StartDate:         "2026-01-01T00:00:00Z",
				CoordinatorAgency: "MSPP",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBioRepo{
				createCampaignFn: func(ctx context.Context, c *domain.VaccinationCampaign) error {
					return tt.repoErr
				},
			}
			svc := NewBioSurveillanceService(repo, nil)
			c, err := svc.CreateCampaign(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.req.CampaignName, c.CampaignName)
		})
	}
}

func TestGetCampaignCoverage(t *testing.T) {
	tests := []struct {
		name     string
		campaign *domain.VaccinationCampaign
		repoErr  error
		wantErr  bool
	}{
		{
			name: "success",
			campaign: &domain.VaccinationCampaign{
				ID: uuid.New(), CampaignName: "Test", CoveragePct: 85.5,
			},
		},
		{
			name:    "invalid id",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBioRepo{
				getCampaignCoverageFn: func(ctx context.Context, id uuid.UUID) (*domain.VaccinationCampaign, error) {
					return tt.campaign, tt.repoErr
				},
			}
			svc := NewBioSurveillanceService(repo, nil)
			id := uuid.New().String()
			if tt.name == "invalid id" {
				id = "bad-uuid"
			}
			got, err := svc.GetCampaignCoverage(context.Background(), id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.campaign.CoveragePct, got.CoveragePct)
		})
	}
}

func TestGetDashboardNational(t *testing.T) {
	tests := []struct {
		name    string
		repoRes *domain.DashboardNational
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: &domain.DashboardNational{
				TotalAlerts: 5, ActiveAlerts: 3, TotalCampaigns: 2, TotalFacilities: 10,
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
			repo := &mockBioRepo{
				getDashboardNationalFn: func(ctx context.Context) (*domain.DashboardNational, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewBioSurveillanceService(repo, nil)
			got, err := svc.GetDashboardNational(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.TotalAlerts, got.TotalAlerts)
		})
	}
}
