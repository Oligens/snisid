package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/bug-bounty-svc/internal/domain"
	"github.com/snisid/bug-bounty-svc/internal/repository"
)

type mockBBRepo struct {
	createProgramFunc           func(ctx context.Context, scope *domain.ProgramScope) error
	listProgramsFunc            func(ctx context.Context) ([]domain.ProgramScope, error)
	createReportFunc            func(ctx context.Context, r *domain.VulnerabilityReport) error
	findReportByIDFunc          func(ctx context.Context, id uuid.UUID) (*domain.VulnerabilityReport, error)
	saveTriageResultFunc        func(ctx context.Context, t *domain.TriageResult) error
	issueRewardFunc             func(ctx context.Context, r *domain.Reward) error
	createPentestEngagementFunc func(ctx context.Context, e *domain.PentestEngagement) error
	findPentestByIDFunc         func(ctx context.Context, id uuid.UUID) (*domain.PentestEngagement, error)
}

func (m *mockBBRepo) CreateProgram(ctx context.Context, scope *domain.ProgramScope) error {
	return m.createProgramFunc(ctx, scope)
}
func (m *mockBBRepo) ListPrograms(ctx context.Context) ([]domain.ProgramScope, error) {
	return m.listProgramsFunc(ctx)
}
func (m *mockBBRepo) CreateReport(ctx context.Context, r *domain.VulnerabilityReport) error {
	return m.createReportFunc(ctx, r)
}
func (m *mockBBRepo) FindReportByID(ctx context.Context, id uuid.UUID) (*domain.VulnerabilityReport, error) {
	return m.findReportByIDFunc(ctx, id)
}
func (m *mockBBRepo) SaveTriageResult(ctx context.Context, t *domain.TriageResult) error {
	return m.saveTriageResultFunc(ctx, t)
}
func (m *mockBBRepo) IssueReward(ctx context.Context, r *domain.Reward) error {
	return m.issueRewardFunc(ctx, r)
}
func (m *mockBBRepo) CreatePentestEngagement(ctx context.Context, e *domain.PentestEngagement) error {
	return m.createPentestEngagementFunc(ctx, e)
}
func (m *mockBBRepo) FindPentestByID(ctx context.Context, id uuid.UUID) (*domain.PentestEngagement, error) {
	return m.findPentestByIDFunc(ctx, id)
}
func (m *mockBBRepo) SaveRetestSchedule(ctx context.Context, s *domain.RetestSchedule) error {
	return nil
}

func newTestBBService(repo repository.Repository) *BugBountyService {
	return NewBugBountyService(repo, nil)
}

func TestNewBugBountyService(t *testing.T) {
	svc := NewBugBountyService(nil, nil)
	require.NotNil(t, svc)
}

func TestCreateProgram(t *testing.T) {
	programID := uuid.New()
	repo := &mockBBRepo{
		createProgramFunc: func(ctx context.Context, scope *domain.ProgramScope) error { return nil },
	}
	svc := newTestBBService(repo)
	result, err := svc.CreateProgram(context.Background(), programID, "https://example.com", "URL", true, nil, nil)
	require.NoError(t, err)
	assert.True(t, result.InScope)
	assert.NotEmpty(t, result.ScopeID)
}

func TestListPrograms(t *testing.T) {
	repo := &mockBBRepo{
		listProgramsFunc: func(ctx context.Context) ([]domain.ProgramScope, error) {
			return []domain.ProgramScope{{Target: "https://example.com"}}, nil
		},
	}
	svc := newTestBBService(repo)
	programs, err := svc.ListPrograms(context.Background())
	require.NoError(t, err)
	assert.Len(t, programs, 1)
}

func TestSubmitReport(t *testing.T) {
	programID := uuid.New()
	repo := &mockBBRepo{
		createReportFunc: func(ctx context.Context, r *domain.VulnerabilityReport) error { return nil },
	}
	svc := newTestBBService(repo)
	report, err := svc.SubmitReport(context.Background(), programID, "researcher1", "XSS", "Description", domain.SeverityHigh, nil)
	require.NoError(t, err)
	assert.Equal(t, "SUBMITTED", report.Status)
}

func TestGetReport(t *testing.T) {
	reportID := uuid.New()
	repo := &mockBBRepo{
		findReportByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.VulnerabilityReport, error) {
			return &domain.VulnerabilityReport{ReportID: id, Title: "XSS"}, nil
		},
	}
	svc := newTestBBService(repo)
	report, err := svc.GetReport(context.Background(), reportID)
	require.NoError(t, err)
	assert.Equal(t, "XSS", report.Title)
}

func TestTriageReport(t *testing.T) {
	reportID := uuid.New()
	repo := &mockBBRepo{
		saveTriageResultFunc: func(ctx context.Context, t *domain.TriageResult) error { return nil },
	}
	svc := newTestBBService(repo)
	result, err := svc.TriageReport(context.Background(), reportID, "analyst1", domain.SeverityCritical, true, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, domain.SeverityCritical, result.Severity)
}

func TestIssueReward(t *testing.T) {
	reportID := uuid.New()
	repo := &mockBBRepo{
		issueRewardFunc: func(ctx context.Context, r *domain.Reward) error { return nil },
	}
	svc := newTestBBService(repo)
	reward, err := svc.IssueReward(context.Background(), reportID, 5000, "USD", "researcher1", "admin1")
	require.NoError(t, err)
	assert.Equal(t, float64(5000), reward.Amount)
}

func TestSchedulePentest(t *testing.T) {
	programID := uuid.New()
	repo := &mockBBRepo{
		createPentestEngagementFunc: func(ctx context.Context, e *domain.PentestEngagement) error { return nil },
	}
	svc := newTestBBService(repo)
	engagement, err := svc.SchedulePentest(context.Background(), programID, "Pentest Q3", "All", time.Now(), nil, "lead1")
	require.NoError(t, err)
	assert.Equal(t, "SCHEDULED", engagement.Status)
}

func TestGetPentestResults(t *testing.T) {
	engagementID := uuid.New()
	repo := &mockBBRepo{
		findPentestByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.PentestEngagement, error) {
			return &domain.PentestEngagement{EngagementID: id, Title: "Pentest Q3"}, nil
		},
	}
	svc := newTestBBService(repo)
	result, err := svc.GetPentestResults(context.Background(), engagementID)
	require.NoError(t, err)
	assert.Equal(t, "Pentest Q3", result.Title)
}

func TestSubmitReport_RepoError(t *testing.T) {
	programID := uuid.New()
	repo := &mockBBRepo{
		createReportFunc: func(ctx context.Context, r *domain.VulnerabilityReport) error {
			return errors.New("db error")
		},
	}
	svc := newTestBBService(repo)
	_, err := svc.SubmitReport(context.Background(), programID, "researcher1", "XSS", "Desc", domain.SeverityHigh, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "submit report")
}
