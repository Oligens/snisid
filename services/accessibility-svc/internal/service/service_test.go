package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/accessibility-svc/internal/domain"
	"github.com/snisid/accessibility-svc/internal/repository"
)

type mockAccRepo struct {
	createAuditRunFunc          func(ctx context.Context, a *domain.AuditRun) error
	findAuditRunByIDFunc        func(ctx context.Context, id uuid.UUID) (*domain.AuditRun, error)
	listAuditRunsFunc           func(ctx context.Context) ([]domain.AuditRun, error)
	listViolationsByAuditFunc   func(ctx context.Context, auditID uuid.UUID) ([]domain.Violation, error)
	markViolationRemediatedFunc func(ctx context.Context, id uuid.UUID) error
	getComplianceOverviewFunc   func(ctx context.Context) ([]domain.AccessibilityReport, error)
	createAuditScheduleFunc     func(ctx context.Context, s *domain.AuditSchedule) error
	getDashboardFunc            func(ctx context.Context) ([]domain.AccessibilityReport, error)
}

func (m *mockAccRepo) CreateAuditRun(ctx context.Context, a *domain.AuditRun) error {
	return m.createAuditRunFunc(ctx, a)
}
func (m *mockAccRepo) FindAuditRunByID(ctx context.Context, id uuid.UUID) (*domain.AuditRun, error) {
	return m.findAuditRunByIDFunc(ctx, id)
}
func (m *mockAccRepo) ListAuditRuns(ctx context.Context) ([]domain.AuditRun, error) {
	return m.listAuditRunsFunc(ctx)
}
func (m *mockAccRepo) CreateViolation(ctx context.Context, v *domain.Violation) error { return nil }
func (m *mockAccRepo) ListViolationsByAudit(ctx context.Context, auditID uuid.UUID) ([]domain.Violation, error) {
	return m.listViolationsByAuditFunc(ctx, auditID)
}
func (m *mockAccRepo) MarkViolationRemediated(ctx context.Context, id uuid.UUID) error {
	return m.markViolationRemediatedFunc(ctx, id)
}
func (m *mockAccRepo) GetComplianceOverview(ctx context.Context) ([]domain.AccessibilityReport, error) {
	return m.getComplianceOverviewFunc(ctx)
}
func (m *mockAccRepo) CreateAuditSchedule(ctx context.Context, s *domain.AuditSchedule) error {
	return m.createAuditScheduleFunc(ctx, s)
}
func (m *mockAccRepo) ListAuditSchedules(ctx context.Context) ([]domain.AuditSchedule, error) {
	return nil, nil
}
func (m *mockAccRepo) GetDashboard(ctx context.Context) ([]domain.AccessibilityReport, error) {
	return m.getDashboardFunc(ctx)
}

func newTestAccService(repo repository.Repository) *AccessibilityService {
	return NewAccessibilityService(repo, nil)
}

func TestNewAccessibilityService(t *testing.T) {
	svc := NewAccessibilityService(nil, nil)
	require.NotNil(t, svc)
}

func TestRunAudit(t *testing.T) {
	repo := &mockAccRepo{
		createAuditRunFunc: func(ctx context.Context, a *domain.AuditRun) error { return nil },
	}
	svc := newTestAccService(repo)
	result, err := svc.RunAudit(context.Background(), "https://example.com", domain.WCAGAA)
	require.NoError(t, err)
	assert.Equal(t, "RUNNING", result.Status)
	assert.NotEmpty(t, result.AuditRunID)
}

func TestRunAudit_RepoError(t *testing.T) {
	repo := &mockAccRepo{
		createAuditRunFunc: func(ctx context.Context, a *domain.AuditRun) error {
			return errors.New("db error")
		},
	}
	svc := newTestAccService(repo)
	_, err := svc.RunAudit(context.Background(), "https://example.com", domain.WCAGAA)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run audit")
}

func TestGetAuditResult(t *testing.T) {
	auditID := uuid.New()
	repo := &mockAccRepo{
		findAuditRunByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.AuditRun, error) {
			return &domain.AuditRun{AuditRunID: id, TargetURL: "https://example.com"}, nil
		},
		listViolationsByAuditFunc: func(ctx context.Context, auditID uuid.UUID) ([]domain.Violation, error) {
			return []domain.Violation{{ViolationID: uuid.New(), Remediated: false}}, nil
		},
	}
	svc := newTestAccService(repo)
	result, err := svc.GetAuditResult(context.Background(), auditID)
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalViolations)
}

func TestListAudits(t *testing.T) {
	repo := &mockAccRepo{
		listAuditRunsFunc: func(ctx context.Context) ([]domain.AuditRun, error) {
			return []domain.AuditRun{{TargetURL: "https://example.com"}}, nil
		},
	}
	svc := newTestAccService(repo)
	audits, err := svc.ListAudits(context.Background())
	require.NoError(t, err)
	assert.Len(t, audits, 1)
}

func TestMarkRemediated(t *testing.T) {
	violationID := uuid.New()
	repo := &mockAccRepo{
		markViolationRemediatedFunc: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	svc := newTestAccService(repo)
	err := svc.MarkRemediated(context.Background(), violationID)
	require.NoError(t, err)
}

func TestMarkRemediated_RepoError(t *testing.T) {
	violationID := uuid.New()
	repo := &mockAccRepo{
		markViolationRemediatedFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("db error")
		},
	}
	svc := newTestAccService(repo)
	err := svc.MarkRemediated(context.Background(), violationID)
	require.Error(t, err)
}

func TestGetComplianceOverview(t *testing.T) {
	repo := &mockAccRepo{
		getComplianceOverviewFunc: func(ctx context.Context) ([]domain.AccessibilityReport, error) {
			return []domain.AccessibilityReport{{TargetURL: "https://example.com", PassRate: 85.0}}, nil
		},
	}
	svc := newTestAccService(repo)
	reports, err := svc.GetComplianceOverview(context.Background())
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	assert.Equal(t, 85.0, reports[0].PassRate)
}

func TestCreateSchedule(t *testing.T) {
	repo := &mockAccRepo{
		createAuditScheduleFunc: func(ctx context.Context, s *domain.AuditSchedule) error { return nil },
	}
	svc := newTestAccService(repo)
	schedule, err := svc.CreateSchedule(context.Background(), "https://example.com", domain.WCAGAA, "0 0 * * 0")
	require.NoError(t, err)
	assert.True(t, schedule.Enabled)
	assert.Equal(t, "0 0 * * 0", schedule.CronExpr)
}

func TestGetDashboard(t *testing.T) {
	repo := &mockAccRepo{
		getDashboardFunc: func(ctx context.Context) ([]domain.AccessibilityReport, error) {
			return []domain.AccessibilityReport{{TargetURL: "https://example.com"}}, nil
		},
	}
	svc := newTestAccService(repo)
	reports, err := svc.GetDashboard(context.Background())
	require.NoError(t, err)
	assert.Len(t, reports, 1)
}
