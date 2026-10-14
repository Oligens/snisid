package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/counterintel-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	createInvestigationFn      func(ctx context.Context, inv *domain.BackgroundInvestigation) error
	getInvestigationFn         func(ctx context.Context, id uuid.UUID) (*domain.BackgroundInvestigation, error)
	getPendingInvestigationsFn func(ctx context.Context) ([]domain.BackgroundInvestigation, error)
	updateInvestigationFn      func(ctx context.Context, inv *domain.BackgroundInvestigation) error
	createThreatAlertFn        func(ctx context.Context, alert *domain.InsiderThreatAlert) error
	getActiveThreatsFn         func(ctx context.Context) ([]domain.InsiderThreatAlert, error)
	createForeignContactFn     func(ctx context.Context, fc *domain.ForeignContact) error
	getContactsBySubjectFn     func(ctx context.Context, subjectID string) ([]domain.ForeignContact, error)
}

func (m *mockRepo) CreateInvestigation(ctx context.Context, inv *domain.BackgroundInvestigation) error {
	return m.createInvestigationFn(ctx, inv)
}
func (m *mockRepo) GetInvestigation(ctx context.Context, id uuid.UUID) (*domain.BackgroundInvestigation, error) {
	return m.getInvestigationFn(ctx, id)
}
func (m *mockRepo) GetPendingInvestigations(ctx context.Context) ([]domain.BackgroundInvestigation, error) {
	return m.getPendingInvestigationsFn(ctx)
}
func (m *mockRepo) UpdateInvestigation(ctx context.Context, inv *domain.BackgroundInvestigation) error {
	return m.updateInvestigationFn(ctx, inv)
}
func (m *mockRepo) CreateThreatAlert(ctx context.Context, alert *domain.InsiderThreatAlert) error {
	return m.createThreatAlertFn(ctx, alert)
}
func (m *mockRepo) GetActiveThreats(ctx context.Context) ([]domain.InsiderThreatAlert, error) {
	return m.getActiveThreatsFn(ctx)
}
func (m *mockRepo) CreateForeignContact(ctx context.Context, fc *domain.ForeignContact) error {
	return m.createForeignContactFn(ctx, fc)
}
func (m *mockRepo) GetContactsBySubject(ctx context.Context, subjectID string) ([]domain.ForeignContact, error) {
	return m.getContactsBySubjectFn(ctx, subjectID)
}

func TestCreateInvestigation(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateInvestigationRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			req: domain.CreateInvestigationRequest{
				SubjectIdentityRef: "sub-001",
				InvestigationType:  "STANDARD",
			},
		},
		{
			name: "repo error",
			req: domain.CreateInvestigationRequest{
				SubjectIdentityRef: "sub-002",
				InvestigationType:  "ENHANCED",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				createInvestigationFn: func(ctx context.Context, inv *domain.BackgroundInvestigation) error {
					return tt.repoErr
				},
			}
			svc := NewCounterintelService(repo, nil)
			inv, err := svc.CreateInvestigation(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, inv.ID)
			assert.Equal(t, tt.req.SubjectIdentityRef, inv.SubjectIdentityRef)
			assert.Equal(t, domain.InvPending, inv.Status)
		})
	}
}

func TestGetPendingInvestigations(t *testing.T) {
	repo := &mockRepo{
		getPendingInvestigationsFn: func(ctx context.Context) ([]domain.BackgroundInvestigation, error) {
			return []domain.BackgroundInvestigation{{ID: uuid.New(), Status: domain.InvPending}}, nil
		},
	}
	svc := NewCounterintelService(repo, nil)
	invs, err := svc.GetPendingInvestigations(context.Background())
	require.NoError(t, err)
	assert.Len(t, invs, 1)
}

func TestReportThreat(t *testing.T) {
	req := domain.ReportThreatRequest{
		SubjectID:   "sub-001",
		AlertType:   "DATA_EXFIL",
		Severity:    "HIGH",
		Description: "test threat",
		DetectedBy:  "soc-01",
	}
	repo := &mockRepo{
		createThreatAlertFn: func(ctx context.Context, alert *domain.InsiderThreatAlert) error {
			return nil
		},
	}
	svc := NewCounterintelService(repo, nil)
	alert, err := svc.ReportThreat(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, alert.ID)
	assert.Equal(t, domain.ThreatOpen, alert.Status)
}

func TestReportContact(t *testing.T) {
	req := domain.ReportContactRequest{
		SubjectID:         "sub-001",
		ContactName:       "John Doe",
		ForeignGovernment: "Atlantis",
		RelationshipType:  "DIPLOMATIC",
	}
	repo := &mockRepo{
		createForeignContactFn: func(ctx context.Context, fc *domain.ForeignContact) error {
			return nil
		},
	}
	svc := NewCounterintelService(repo, nil)
	fc, err := svc.ReportContact(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, fc.ID)
	assert.Equal(t, req.ContactName, fc.ContactName)
}
