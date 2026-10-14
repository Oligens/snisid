package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/snisid/fisa-court-svc/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	createWarrantFn     func(ctx context.Context, w *domain.SurveillanceWarrant) error
	updateWarrantFn     func(ctx context.Context, w *domain.SurveillanceWarrant) error
	getWarrantFn        func(ctx context.Context, id uuid.UUID) (*domain.SurveillanceWarrant, error)
	getActiveWarrantsFn func(ctx context.Context) ([]domain.SurveillanceWarrant, error)
	createReportFn      func(ctx context.Context, r *domain.SurveillanceReport) error
	getDocketByTermFn   func(ctx context.Context, term string) (*domain.FISADocket, error)
	upsertDocketFn      func(ctx context.Context, d *domain.FISADocket) error
}

func (m *mockRepo) CreateWarrant(ctx context.Context, w *domain.SurveillanceWarrant) error {
	return m.createWarrantFn(ctx, w)
}
func (m *mockRepo) UpdateWarrant(ctx context.Context, w *domain.SurveillanceWarrant) error {
	return m.updateWarrantFn(ctx, w)
}
func (m *mockRepo) GetWarrant(ctx context.Context, id uuid.UUID) (*domain.SurveillanceWarrant, error) {
	return m.getWarrantFn(ctx, id)
}
func (m *mockRepo) GetActiveWarrants(ctx context.Context) ([]domain.SurveillanceWarrant, error) {
	return m.getActiveWarrantsFn(ctx)
}
func (m *mockRepo) CreateReport(ctx context.Context, r *domain.SurveillanceReport) error {
	return m.createReportFn(ctx, r)
}
func (m *mockRepo) GetDocketByTerm(ctx context.Context, term string) (*domain.FISADocket, error) {
	return m.getDocketByTermFn(ctx, term)
}
func (m *mockRepo) UpsertDocket(ctx context.Context, d *domain.FISADocket) error {
	return m.upsertDocketFn(ctx, d)
}

func TestFileWarrant(t *testing.T) {
	repo := &mockRepo{
		createWarrantFn: func(ctx context.Context, w *domain.SurveillanceWarrant) error { return nil },
	}
	svc := NewFISAService(repo, nil)
	req := domain.FileWarrantRequest{
		WarrantType:      "FISA_ELECTRONIC",
		TargetIdentity:   "target-001",
		IssuingCourt:     "FISA Court",
		JudgeName:        "Judge A",
		ApplicantAgency:  "NSA",
		ApplicantOfficer: uuid.New().String(),
		DurationDays:     90,
	}
	w, err := svc.FileWarrant(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, w.ID)
	assert.Equal(t, domain.WarrantPending, w.Status)
}

func TestApproveWarrant(t *testing.T) {
	wID := uuid.New()
	repo := &mockRepo{
		getWarrantFn: func(ctx context.Context, id uuid.UUID) (*domain.SurveillanceWarrant, error) {
			return &domain.SurveillanceWarrant{ID: wID, DurationDays: 90, Status: domain.WarrantPending}, nil
		},
		updateWarrantFn: func(ctx context.Context, w *domain.SurveillanceWarrant) error { return nil },
	}
	svc := NewFISAService(repo, nil)
	w, err := svc.ApproveWarrant(context.Background(), wID, domain.ApproveWarrantRequest{JudgeName: "Judge A"})
	require.NoError(t, err)
	assert.Equal(t, domain.WarrantActive, w.Status)
}

func TestEmergencyAuthorization(t *testing.T) {
	repo := &mockRepo{
		createWarrantFn: func(ctx context.Context, w *domain.SurveillanceWarrant) error { return nil },
	}
	svc := NewFISAService(repo, nil)
	req := domain.EmergencyAuthorizationRequest{
		WarrantType:      "FISA_ELECTRONIC",
		TargetIdentity:   "target-001",
		ApplicantAgency:  "NSA",
		ApplicantOfficer: uuid.New().String(),
		ProbableCause:    "imminent threat",
		ApprovedBy:       uuid.New().String(),
	}
	w, err := svc.EmergencyAuthorization(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, w.EmergencyAuthorized)
	assert.Equal(t, domain.WarrantActive, w.Status)
}
