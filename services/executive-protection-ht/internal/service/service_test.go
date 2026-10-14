package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/executive-protection-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExecRepo struct {
	createProtecteeFn             func(ctx context.Context, p *domain.Protectee) error
	getActiveProtecteesFn         func(ctx context.Context) ([]domain.Protectee, error)
	createMovementPlanFn          func(ctx context.Context, m *domain.MovementPlan) error
	getUpcomingMovementsFn        func(ctx context.Context) ([]domain.MovementPlan, error)
	createThreatAssessmentFn      func(ctx context.Context, t *domain.ThreatAssessment) error
	getActiveThreatsByProtecteeFn func(ctx context.Context, id uuid.UUID) ([]domain.ThreatAssessment, error)
	getDashboardFn                func(ctx context.Context) (*domain.DashboardProtection, error)
}

func (m *mockExecRepo) CreateProtectee(ctx context.Context, p *domain.Protectee) error {
	return m.createProtecteeFn(ctx, p)
}
func (m *mockExecRepo) GetActiveProtectees(ctx context.Context) ([]domain.Protectee, error) {
	return m.getActiveProtecteesFn(ctx)
}
func (m *mockExecRepo) CreateMovementPlan(ctx context.Context, pl *domain.MovementPlan) error {
	return m.createMovementPlanFn(ctx, pl)
}
func (m *mockExecRepo) GetUpcomingMovements(ctx context.Context) ([]domain.MovementPlan, error) {
	return m.getUpcomingMovementsFn(ctx)
}
func (m *mockExecRepo) CreateThreatAssessment(ctx context.Context, t *domain.ThreatAssessment) error {
	return m.createThreatAssessmentFn(ctx, t)
}
func (m *mockExecRepo) GetActiveThreatsByProtectee(ctx context.Context, id uuid.UUID) ([]domain.ThreatAssessment, error) {
	return m.getActiveThreatsByProtecteeFn(ctx, id)
}
func (m *mockExecRepo) GetDashboard(ctx context.Context) (*domain.DashboardProtection, error) {
	return m.getDashboardFn(ctx)
}

func TestCreateProtectee(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateProtecteeRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.CreateProtecteeRequest{
				FullName:           "Jean Dupont",
				OfficialTitle:      "Ministre de la Santé",
				ProtectionLevel:    "CABINET_MINISTER",
				RiskAssessment:     "HIGH",
				PrimaryAgentID:     uuid.New().String(),
				SecureVehiclePlate: "AA-001-BB",
				ResidenceLocation:  "Pétion-Ville",
			},
		},
		{
			name: "invalid primary agent id",
			req: domain.CreateProtecteeRequest{
				FullName:           "Bad Agent",
				OfficialTitle:      "Test",
				ProtectionLevel:    "WITNESS",
				RiskAssessment:     "LOW",
				PrimaryAgentID:     "not-a-uuid",
				SecureVehiclePlate: "XX-000-XX",
			},
			wantErr: true,
		},
		{
			name: "repo error",
			req: domain.CreateProtecteeRequest{
				FullName:           "Fail",
				OfficialTitle:      "Test",
				ProtectionLevel:    "JUDGE",
				RiskAssessment:     "MEDIUM",
				PrimaryAgentID:     uuid.New().String(),
				SecureVehiclePlate: "FF-999-FF",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExecRepo{
				createProtecteeFn: func(ctx context.Context, p *domain.Protectee) error {
					return tt.repoErr
				},
			}
			svc := NewExecutiveProtectionService(repo, nil)
			p, err := svc.CreateProtectee(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.req.FullName, p.FullName)
			assert.Equal(t, domain.ProtectionLevel(tt.req.ProtectionLevel), p.ProtectionLevel)
		})
	}
}

func TestGetActiveProtectees(t *testing.T) {
	tests := []struct {
		name    string
		repoRes []domain.Protectee
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: []domain.Protectee{
				{ID: uuid.New(), FullName: "VIP 1", RiskAssessment: domain.RiskHigh},
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
			repo := &mockExecRepo{
				getActiveProtecteesFn: func(ctx context.Context) ([]domain.Protectee, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewExecutiveProtectionService(repo, nil)
			got, err := svc.GetActiveProtectees(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.repoRes))
		})
	}
}

func TestCreateMovementPlan(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateMovementPlanRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.CreateMovementPlanRequest{
				ProtecteeID:       uuid.New().String(),
				EventName:         "Sommet des Amériques",
				Date:              "2026-07-15T09:00:00Z",
				DepartureLocation: "Palais National",
				ArrivalLocation:   "Aéroport Toussaint Louverture",
				TransportMode:     "MOTORCADE",
			},
		},
		{
			name: "invalid protectee id",
			req: domain.CreateMovementPlanRequest{
				ProtecteeID: "bad-uuid",
				EventName:   "Test",
				Date:        "2026-07-15T09:00:00Z",
			},
			wantErr: true,
		},
		{
			name: "invalid date",
			req: domain.CreateMovementPlanRequest{
				ProtecteeID: uuid.New().String(),
				EventName:   "Test",
				Date:        "not-a-date",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExecRepo{
				createMovementPlanFn: func(ctx context.Context, m *domain.MovementPlan) error {
					return tt.repoErr
				},
			}
			svc := NewExecutiveProtectionService(repo, nil)
			pl, err := svc.CreateMovementPlan(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.MovementDraft, pl.Status)
			assert.Equal(t, tt.req.EventName, pl.EventName)
		})
	}
}

func TestCreateThreatAssessment(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateThreatAssessmentRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.CreateThreatAssessmentRequest{
				ProtecteeID:  uuid.New().String(),
				ThreatType:   "DIRECT_THREAT",
				ThreatLevel:  "CRITICAL",
				AssessedBy:   uuid.New().String(),
				ThreatDetail: "Menace directe reçue par canal diplomatique",
			},
		},
		{
			name: "invalid protectee id",
			req: domain.CreateThreatAssessmentRequest{
				ProtecteeID: "bad-uuid",
				ThreatType:  "STALKER",
				ThreatLevel: "LOW",
				AssessedBy:  uuid.New().String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockExecRepo{
				createThreatAssessmentFn: func(ctx context.Context, t *domain.ThreatAssessment) error {
					return tt.repoErr
				},
			}
			svc := NewExecutiveProtectionService(repo, nil)
			th, err := svc.CreateThreatAssessment(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domain.ThreatPending, th.Status)
			assert.Equal(t, domain.ThreatType(tt.req.ThreatType), th.ThreatType)
		})
	}
}

func TestGetDashboard(t *testing.T) {
	tests := []struct {
		name    string
		repoRes *domain.DashboardProtection
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			repoRes: &domain.DashboardProtection{
				TotalProtectees: 10, ActiveProtectees: 4, UpcomingMovements: 3, ActiveThreats: 7,
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
			repo := &mockExecRepo{
				getDashboardFn: func(ctx context.Context) (*domain.DashboardProtection, error) {
					return tt.repoRes, tt.repoErr
				},
			}
			svc := NewExecutiveProtectionService(repo, nil)
			got, err := svc.GetDashboard(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.repoRes.TotalProtectees, got.TotalProtectees)
		})
	}
}
