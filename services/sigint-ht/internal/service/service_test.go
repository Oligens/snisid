package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/sigint-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSigintRepo struct {
	mock.Mock
}

func (m *mockSigintRepo) CreateTarget(t domain.InterceptionTarget) (domain.InterceptionTarget, error) {
	args := m.Called(t)
	return args.Get(0).(domain.InterceptionTarget), args.Error(1)
}

func (m *mockSigintRepo) GetActiveTargets() ([]domain.InterceptionTarget, error) {
	args := m.Called()
	return args.Get(0).([]domain.InterceptionTarget), args.Error(1)
}

func (m *mockSigintRepo) GetTargetByID(id string) (domain.InterceptionTarget, error) {
	args := m.Called(id)
	return args.Get(0).(domain.InterceptionTarget), args.Error(1)
}

func (m *mockSigintRepo) RecordCommunication(c domain.InterceptedCommunication) (domain.InterceptedCommunication, error) {
	args := m.Called(c)
	return args.Get(0).(domain.InterceptedCommunication), args.Error(1)
}

func (m *mockSigintRepo) GetCommunicationsByTarget(targetID string) ([]domain.InterceptedCommunication, error) {
	args := m.Called(targetID)
	return args.Get(0).([]domain.InterceptedCommunication), args.Error(1)
}

func (m *mockSigintRepo) AnalyzeCDR(phone string) ([]domain.CDRAnalysis, error) {
	args := m.Called(phone)
	return args.Get(0).([]domain.CDRAnalysis), args.Error(1)
}

func (m *mockSigintRepo) CreateEmergencyTarget(t domain.InterceptionTarget) (domain.InterceptionTarget, error) {
	args := m.Called(t)
	return args.Get(0).(domain.InterceptionTarget), args.Error(1)
}

func (m *mockSigintRepo) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

type noopProducer struct{}

func (p *noopProducer) Publish(ctx context.Context, eventType string, payload interface{}) error {
	return nil
}

func (p *noopProducer) Close() error {
	return nil
}

func newTestService(repo *mockSigintRepo) *SigintService {
	return &SigintService{repo: repo, producer: &noopProducer{}}
}

func TestCreateTarget(t *testing.T) {
	mockRepo := new(mockSigintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		req     domain.CreateTargetRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "valid target creation",
			req: domain.CreateTargetRequest{
				TargetType:       "PHONE_NUMBER",
				AuthorizationRef: uuid.New().String(),
				JudgeName:        "Judge Smith",
				IssuingCourt:     "FISA Court",
				StartDate:        time.Now().UTC().Format(time.RFC3339),
				EndDate:          time.Now().UTC().Add(90 * 24 * time.Hour).Format(time.RFC3339),
				TargetIdentifier: "+1-555-0123",
			},
			setup: func() {
				mockRepo.On("CreateTarget", mock.AnythingOfType("domain.InterceptionTarget")).
					Return(domain.InterceptionTarget{
						ID:         uuid.New().String(),
						TargetType: "PHONE_NUMBER",
						Status:     "ACTIVE",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid start date format",
			req: domain.CreateTargetRequest{
				TargetType:       "EMAIL",
				AuthorizationRef: uuid.New().String(),
				StartDate:        "invalid-date",
				EndDate:          time.Now().UTC().Format(time.RFC3339),
			},
			setup:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.CreateTarget(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetActiveTargets(t *testing.T) {
	mockRepo := new(mockSigintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		setup   func()
		wantLen int
		wantErr bool
	}{
		{
			name: "returns active targets",
			setup: func() {
				mockRepo.On("GetActiveTargets").Return([]domain.InterceptionTarget{
					{ID: uuid.New().String(), TargetType: "EMAIL", Status: "ACTIVE"},
					{ID: uuid.New().String(), TargetType: "PHONE_NUMBER", Status: "ACTIVE"},
				}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "empty list",
			setup: func() {
				mockRepo.On("GetActiveTargets").Return([]domain.InterceptionTarget{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.On("GetActiveTargets").Return([]domain.InterceptionTarget{}, errors.New("db error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			targets, err := svc.GetActiveTargets()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, targets, tt.wantLen)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAnalyzeCDR(t *testing.T) {
	mockRepo := new(mockSigintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		phone   string
		setup   func()
		wantLen int
		wantErr bool
	}{
		{
			name:  "phone with CDR records",
			phone: "+1-555-0123",
			setup: func() {
				mockRepo.On("AnalyzeCDR", "+1-555-0123").Return([]domain.CDRAnalysis{
					{ID: uuid.New().String(), Caller: "+1-555-0123", Duration: 120},
				}, nil)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:  "no records found",
			phone: "+1-555-9999",
			setup: func() {
				mockRepo.On("AnalyzeCDR", "+1-555-9999").Return([]domain.CDRAnalysis{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			records, err := svc.AnalyzeCDR(tt.phone)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, records, tt.wantLen)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRecordInterception(t *testing.T) {
	mockRepo := new(mockSigintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name     string
		targetID string
		req      domain.InterceptRequest
		setup    func()
		wantErr  bool
	}{
		{
			name:     "valid interception",
			targetID: uuid.New().String(),
			req: domain.InterceptRequest{
				CommType:      "CALL",
				ContentRef:    "s3://bucket/rec1.mp4",
				InterceptedAt: time.Now().UTC().Format(time.RFC3339),
				CollectorNode: "node-01",
				CaseNumber:    "CASE-001",
			},
			setup: func() {
				mockRepo.On("RecordCommunication", mock.AnythingOfType("domain.InterceptedCommunication")).
					Return(domain.InterceptedCommunication{ID: uuid.New().String(), CommType: "CALL"}, nil)
			},
			wantErr: false,
		},
		{
			name:     "invalid date",
			targetID: uuid.New().String(),
			req: domain.InterceptRequest{
				CommType:      "SMS",
				InterceptedAt: "bad-date",
			},
			setup:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.RecordInterception(tt.targetID, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEmergencyAuthorization(t *testing.T) {
	mockRepo := new(mockSigintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		req     domain.EmergencyRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "valid emergency",
			req: domain.EmergencyRequest{
				TargetIdentifier:   "+1-555-0123",
				TargetType:         "PHONE_NUMBER",
				Reason:             "Imminent threat",
				AuthorizingOfficer: "DIR-001",
			},
			setup: func() {
				mockRepo.On("CreateEmergencyTarget", mock.AnythingOfType("domain.InterceptionTarget")).
					Return(domain.InterceptionTarget{
						ID:         uuid.New().String(),
						TargetType: "PHONE_NUMBER",
						Status:     "ACTIVE",
					}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := svc.EmergencyAuthorization(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Approved)
				assert.NotEmpty(t, resp.AuthRef)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
