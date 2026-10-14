package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/snisid/humint-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHumintRepo struct {
	mock.Mock
}

func (m *mockHumintRepo) CreateSource(s domain.Source) (domain.Source, error) {
	args := m.Called(s)
	return args.Get(0).(domain.Source), args.Error(1)
}

func (m *mockHumintRepo) UpdateCredibility(code string, rating int, reliability string) (domain.Source, error) {
	args := m.Called(code, rating, reliability)
	return args.Get(0).(domain.Source), args.Error(1)
}

func (m *mockHumintRepo) GetSourceByCode(code string) (domain.Source, error) {
	args := m.Called(code)
	return args.Get(0).(domain.Source), args.Error(1)
}

func (m *mockHumintRepo) GetReportsBySource(code string) ([]domain.IntelligenceReport, error) {
	args := m.Called(code)
	return args.Get(0).([]domain.IntelligenceReport), args.Error(1)
}

func (m *mockHumintRepo) SubmitReport(r domain.IntelligenceReport) (domain.IntelligenceReport, error) {
	args := m.Called(r)
	return args.Get(0).(domain.IntelligenceReport), args.Error(1)
}

func (m *mockHumintRepo) LogDebriefing(d domain.DebriefingSession) (domain.DebriefingSession, error) {
	args := m.Called(d)
	return args.Get(0).(domain.DebriefingSession), args.Error(1)
}

func (m *mockHumintRepo) GetHighRiskSources() ([]domain.Source, error) {
	args := m.Called()
	return args.Get(0).([]domain.Source), args.Error(1)
}

func (m *mockHumintRepo) GetSourceNetwork() ([]domain.Source, []domain.IntelligenceReport, error) {
	args := m.Called()
	return args.Get(0).([]domain.Source), args.Get(1).([]domain.IntelligenceReport), args.Error(2)
}

func (m *mockHumintRepo) HealthCheck() error {
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

func newTestService(repo *mockHumintRepo) *HumintService {
	return &HumintService{repo: repo, producer: &noopProducer{}}
}

func TestCreateSource(t *testing.T) {
	mockRepo := new(mockHumintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		req     domain.CreateSourceRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "valid source creation",
			req: domain.CreateSourceRequest{
				CodeName:          "RAVEN-1",
				CredibilityRating: 4,
				ReliabilityRating: "B",
				HandlingOfficerID: "550e8400-e29b-41d4-a716-446655440000",
				PaymentAmount:     5000.00,
				PaymentFrequency:  "MONTHLY",
				RiskLevel:         "HIGH",
				Compartment:       "ALPHA",
			},
			setup: func() {
				mockRepo.On("CreateSource", mock.AnythingOfType("domain.Source")).
					Return(domain.Source{
						CodeName:          "RAVEN-1",
						CredibilityRating: 4,
						ReliabilityRating: "B",
						Status:            "ACTIVE",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate code name",
			req: domain.CreateSourceRequest{
				CodeName:          "EXISTING-1",
				CredibilityRating: 3,
				ReliabilityRating: "C",
				HandlingOfficerID: "550e8400-e29b-41d4-a716-446655440000",
				RiskLevel:         "MEDIUM",
			},
			setup: func() {
				mockRepo.On("CreateSource", mock.AnythingOfType("domain.Source")).
					Return(domain.Source{}, errors.New("duplicate key"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.CreateSource(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateCredibility(t *testing.T) {
	mockRepo := new(mockHumintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		code    string
		req     domain.UpdateCredibilityRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "valid update",
			code: "RAVEN-1",
			req: domain.UpdateCredibilityRequest{
				CredibilityRating: 5,
				ReliabilityRating: "A",
			},
			setup: func() {
				mockRepo.On("UpdateCredibility", "RAVEN-1", 5, "A").
					Return(domain.Source{CodeName: "RAVEN-1", CredibilityRating: 5}, nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			code: "NONEXIST",
			req: domain.UpdateCredibilityRequest{
				CredibilityRating: 6,
				ReliabilityRating: "A",
			},
			setup: func() {
				mockRepo.On("UpdateCredibility", "NONEXIST", 6, "A").
					Return(domain.Source{}, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.UpdateCredibility(tt.code, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetHighRiskSources(t *testing.T) {
	mockRepo := new(mockHumintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		setup   func()
		wantLen int
		wantErr bool
	}{
		{
			name: "returns high risk sources",
			setup: func() {
				mockRepo.On("GetHighRiskSources").Return([]domain.Source{
					{CodeName: "VIPER-1", RiskLevel: "CRITICAL", Status: "ACTIVE"},
					{CodeName: "COBRA-2", RiskLevel: "HIGH", Status: "ACTIVE"},
				}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "no high risk sources",
			setup: func() {
				mockRepo.On("GetHighRiskSources").Return([]domain.Source{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			sources, err := svc.GetHighRiskSources()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sources, tt.wantLen)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubmitReport(t *testing.T) {
	mockRepo := new(mockHumintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		req     domain.SubmitReportRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "valid report submission",
			req: domain.SubmitReportRequest{
				SourceCode:     "RAVEN-1",
				Classification: "SECRET",
				ContentHash:    "a1b2c3d4e5f6",
				ThreatActors:   []string{"APT-29", "FancyBear"},
				VeracityScore:  0.85,
			},
			setup: func() {
				mockRepo.On("SubmitReport", mock.AnythingOfType("domain.IntelligenceReport")).
					Return(domain.IntelligenceReport{
						SourceCode:     "RAVEN-1",
						Classification: "SECRET",
						ContentHash:    "a1b2c3d4e5f6",
						ThreatActors:   []string{"APT-29", "FancyBear"},
						VeracityScore:  0.85,
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: domain.SubmitReportRequest{
				SourceCode:     "NONEXIST",
				Classification: "TOP_SECRET",
				ContentHash:    "hash",
			},
			setup: func() {
				mockRepo.On("SubmitReport", mock.AnythingOfType("domain.IntelligenceReport")).
					Return(domain.IntelligenceReport{}, errors.New("source not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.SubmitReport(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogDebriefing(t *testing.T) {
	mockRepo := new(mockHumintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name    string
		req     domain.LogDebriefingRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "valid debriefing log",
			req: domain.LogDebriefingRequest{
				SourceCode:     "RAVEN-1",
				OfficerID:      "550e8400-e29b-41d4-a716-446655440000",
				SessionDate:    time.Now().UTC().Format(time.RFC3339),
				LocationMethod: "IN_PERSON",
				TopicsCovered:  []string{"Asset movements", "Financial status"},
				RiskAssessment: "Source shows signs of compromise",
			},
			setup: func() {
				mockRepo.On("LogDebriefing", mock.AnythingOfType("domain.DebriefingSession")).
					Return(domain.DebriefingSession{
						SourceCode:     "RAVEN-1",
						LocationMethod: "IN_PERSON",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid date format",
			req: domain.LogDebriefingRequest{
				SourceCode:     "RAVEN-1",
				OfficerID:      "550e8400-e29b-41d4-a716-446655440000",
				SessionDate:    "not-a-date",
				LocationMethod: "PHONE",
			},
			setup:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := svc.LogDebriefing(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetSourceNetwork(t *testing.T) {
	mockRepo := new(mockHumintRepo)
	svc := newTestService(mockRepo)

	tests := []struct {
		name      string
		setup     func()
		wantErr   bool
		wantNodes int
		wantEdges int
	}{
		{
			name: "returns network data",
			setup: func() {
				mockRepo.On("GetSourceNetwork").Return(
					[]domain.Source{
						{CodeName: "RAVEN-1"},
						{CodeName: "VIPER-1"},
					},
					[]domain.IntelligenceReport{
						{
							SourceCode:   "RAVEN-1",
							ThreatActors: []string{"APT-29", "FancyBear"},
						},
					},
					nil,
				)
			},
			wantErr:   false,
			wantNodes: 4,
			wantEdges: 2,
		},
		{
			name: "empty network",
			setup: func() {
				mockRepo.On("GetSourceNetwork").Return(
					[]domain.Source{},
					[]domain.IntelligenceReport{},
					nil,
				)
			},
			wantErr:   false,
			wantNodes: 0,
			wantEdges: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := svc.GetSourceNetwork()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Nodes, tt.wantNodes)
				assert.Len(t, resp.Edges, tt.wantEdges)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
