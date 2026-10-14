package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/fips-cert-svc/internal/domain"
	"github.com/snisid/fips-cert-svc/internal/repository"
)

type mockFIPSRepo struct {
	createModuleFunc           func(ctx context.Context, m *domain.CryptoModule) error
	findModuleByIDFunc         func(ctx context.Context, id uuid.UUID) (*domain.CryptoModule, error)
	listModulesFunc            func(ctx context.Context) ([]domain.CryptoModule, error)
	updateValidationFunc       func(ctx context.Context, id uuid.UUID, status domain.ValidationStatus, certNumber string, validationDate time.Time) error
	createCVEResultFunc        func(ctx context.Context, r *domain.CVEScanResult) error
	getComplianceByServiceFunc func(ctx context.Context, service string) (*domain.ComplianceReport, error)
	getDashboardFunc           func(ctx context.Context) ([]domain.ComplianceReport, error)
}

func (m *mockFIPSRepo) CreateModule(ctx context.Context, mod *domain.CryptoModule) error {
	return m.createModuleFunc(ctx, mod)
}
func (m *mockFIPSRepo) FindModuleByID(ctx context.Context, id uuid.UUID) (*domain.CryptoModule, error) {
	return m.findModuleByIDFunc(ctx, id)
}
func (m *mockFIPSRepo) ListModules(ctx context.Context) ([]domain.CryptoModule, error) {
	return m.listModulesFunc(ctx)
}
func (m *mockFIPSRepo) UpdateValidation(ctx context.Context, id uuid.UUID, status domain.ValidationStatus, certNumber string, validationDate time.Time) error {
	return m.updateValidationFunc(ctx, id, status, certNumber, validationDate)
}
func (m *mockFIPSRepo) CreateCVEResult(ctx context.Context, r *domain.CVEScanResult) error {
	return m.createCVEResultFunc(ctx, r)
}
func (m *mockFIPSRepo) ListCVEsByModule(ctx context.Context, moduleID uuid.UUID) ([]domain.CVEScanResult, error) {
	return nil, nil
}
func (m *mockFIPSRepo) GetComplianceByService(ctx context.Context, service string) (*domain.ComplianceReport, error) {
	return m.getComplianceByServiceFunc(ctx, service)
}
func (m *mockFIPSRepo) GetDashboard(ctx context.Context) ([]domain.ComplianceReport, error) {
	return m.getDashboardFunc(ctx)
}

func newTestFIPSService(repo repository.Repository) *FIPSService {
	return NewFIPSService(repo, nil)
}

func TestNewFIPSService(t *testing.T) {
	svc := NewFIPSService(nil, nil)
	require.NotNil(t, svc)
}

func TestRegisterModule(t *testing.T) {
	tests := []struct {
		name        string
		mod         domain.CryptoModule
		repo        *mockFIPSRepo
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			mod: domain.CryptoModule{
				Name:      "AES-256-GCM",
				Version:   "2.0",
				Vendor:    "TestCorp",
				FIPSLevel: domain.FIPSLevel1,
			},
			repo: &mockFIPSRepo{
				createModuleFunc: func(ctx context.Context, m *domain.CryptoModule) error { return nil },
			},
		},
		{
			name: "repo error",
			mod: domain.CryptoModule{
				Name:      "FailMod",
				Version:   "1.0",
				Vendor:    "BadCorp",
				FIPSLevel: domain.FIPSLevel2,
			},
			repo: &mockFIPSRepo{
				createModuleFunc: func(ctx context.Context, m *domain.CryptoModule) error { return errors.New("db error") },
			},
			wantErr:     true,
			errContains: "register module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestFIPSService(tt.repo)
			result, err := svc.RegisterModule(context.Background(), tt.mod)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, domain.StatusPending, result.Status)
			assert.NotEmpty(t, result.ModuleID)
		})
	}
}

func TestListModules(t *testing.T) {
	repo := &mockFIPSRepo{
		listModulesFunc: func(ctx context.Context) ([]domain.CryptoModule, error) {
			return []domain.CryptoModule{{Name: "Mod1"}}, nil
		},
	}
	svc := newTestFIPSService(repo)
	modules, err := svc.ListModules(context.Background())
	require.NoError(t, err)
	assert.Len(t, modules, 1)
}

func TestSubmitValidation(t *testing.T) {
	moduleID := uuid.New()
	tests := []struct {
		name    string
		repo    *mockFIPSRepo
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockFIPSRepo{
				updateValidationFunc: func(ctx context.Context, id uuid.UUID, status domain.ValidationStatus, certNumber string, validationDate time.Time) error {
					return nil
				},
				findModuleByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.CryptoModule, error) {
					return &domain.CryptoModule{ModuleID: id, Status: domain.StatusValidated}, nil
				},
			},
		},
		{
			name: "repo error",
			repo: &mockFIPSRepo{
				updateValidationFunc: func(ctx context.Context, id uuid.UUID, status domain.ValidationStatus, certNumber string, validationDate time.Time) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestFIPSService(tt.repo)
			result, err := svc.SubmitValidation(context.Background(), moduleID, "FIPS-001", time.Now())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.StatusValidated, result.Status)
		})
	}
}

func TestReportCVE(t *testing.T) {
	moduleID := uuid.New()
	repo := &mockFIPSRepo{
		createCVEResultFunc: func(ctx context.Context, r *domain.CVEScanResult) error { return nil },
	}
	svc := newTestFIPSService(repo)
	result, err := svc.ReportCVE(context.Background(), moduleID, "CVE-2026-0001", "HIGH", nil)
	require.NoError(t, err)
	assert.Equal(t, "CVE-2026-0001", result.CVEID)
	assert.Equal(t, moduleID, result.ModuleID)
}

func TestGetComplianceByService(t *testing.T) {
	repo := &mockFIPSRepo{
		getComplianceByServiceFunc: func(ctx context.Context, service string) (*domain.ComplianceReport, error) {
			return &domain.ComplianceReport{ServiceName: service, OverallStatus: "COMPLIANT"}, nil
		},
	}
	svc := newTestFIPSService(repo)
	rep, err := svc.GetComplianceByService(context.Background(), "auth-svc")
	require.NoError(t, err)
	assert.Equal(t, "COMPLIANT", rep.OverallStatus)
}

func TestGetDashboard(t *testing.T) {
	repo := &mockFIPSRepo{
		getDashboardFunc: func(ctx context.Context) ([]domain.ComplianceReport, error) {
			return []domain.ComplianceReport{{ServiceName: "svc1"}}, nil
		},
	}
	svc := newTestFIPSService(repo)
	reports, err := svc.GetDashboard(context.Background())
	require.NoError(t, err)
	assert.Len(t, reports, 1)
}
