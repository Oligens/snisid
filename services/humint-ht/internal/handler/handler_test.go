package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/snisid/humint-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHumintSvc struct {
	mock.Mock
}

func (m *mockHumintSvc) CreateSource(req domain.CreateSourceRequest) (domain.Source, error) {
	args := m.Called(req)
	return args.Get(0).(domain.Source), args.Error(1)
}

func (m *mockHumintSvc) UpdateCredibility(code string, req domain.UpdateCredibilityRequest) (domain.Source, error) {
	args := m.Called(code, req)
	return args.Get(0).(domain.Source), args.Error(1)
}

func (m *mockHumintSvc) GetReportsBySource(code string) ([]domain.IntelligenceReport, error) {
	args := m.Called(code)
	return args.Get(0).([]domain.IntelligenceReport), args.Error(1)
}

func (m *mockHumintSvc) SubmitReport(req domain.SubmitReportRequest) (domain.IntelligenceReport, error) {
	args := m.Called(req)
	return args.Get(0).(domain.IntelligenceReport), args.Error(1)
}

func (m *mockHumintSvc) LogDebriefing(req domain.LogDebriefingRequest) (domain.DebriefingSession, error) {
	args := m.Called(req)
	return args.Get(0).(domain.DebriefingSession), args.Error(1)
}

func (m *mockHumintSvc) GetHighRiskSources() ([]domain.Source, error) {
	args := m.Called()
	return args.Get(0).([]domain.Source), args.Error(1)
}

func (m *mockHumintSvc) GetSourceNetwork() (domain.SourceNetworkResponse, error) {
	args := m.Called()
	return args.Get(0).(domain.SourceNetworkResponse), args.Error(1)
}

func setupRouter(svc HumintService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &HumintHandler{svc: svc}
	h.RegisterRoutes(r)
	return r
}

func TestCreateSourceHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	tests := []struct {
		name       string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name: "valid source creation",
			body: `{"code_name":"RAVEN-1","credibility_rating":4,"reliability_rating":"B","handling_officer_id":"550e8400-e29b-41d4-a716-446655440000","risk_level":"HIGH","payment_amount":5000,"payment_frequency":"MONTHLY"}`,
			mockFn: func() {
				mockSvc.On("CreateSource", mock.AnythingOfType("domain.CreateSourceRequest")).
					Return(domain.Source{CodeName: "RAVEN-1", CredibilityRating: 4}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid body",
			body:       `{"code_name":""}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"code_name":"FAIL-1","credibility_rating":1,"reliability_rating":"A","handling_officer_id":"550e8400-e29b-41d4-a716-446655440000","risk_level":"LOW"}`,
			mockFn: func() {
				mockSvc.On("CreateSource", mock.AnythingOfType("domain.CreateSourceRequest")).
					Return(domain.Source{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/humint/sources", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestUpdateCredibilityHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	tests := []struct {
		name       string
		code       string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name: "valid update",
			code: "RAVEN-1",
			body: `{"credibility_rating":5,"reliability_rating":"A"}`,
			mockFn: func() {
				mockSvc.On("UpdateCredibility", "RAVEN-1", mock.AnythingOfType("domain.UpdateCredibilityRequest")).
					Return(domain.Source{CodeName: "RAVEN-1", CredibilityRating: 5, ReliabilityRating: "A"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid body",
			code:       "RAVEN-1",
			body:       `{}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/api/v1/humint/sources/"+tt.code+"/credibility", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetHighRiskHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	tests := []struct {
		name       string
		mockFn     func()
		wantStatus int
		wantTotal  int
	}{
		{
			name: "returns high risk sources",
			mockFn: func() {
				mockSvc.On("GetHighRiskSources").Return([]domain.Source{
					{CodeName: "VIPER-1", RiskLevel: "CRITICAL"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name: "no sources",
			mockFn: func() {
				mockSvc.On("GetHighRiskSources").Return([]domain.Source{}, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name: "service error",
			mockFn: func() {
				mockSvc.On("GetHighRiskSources").Return([]domain.Source{}, errors.New("error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/humint/sources/high-risk", nil)
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp domain.HighRiskResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantTotal, resp.Total)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestSubmitReportHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	tests := []struct {
		name       string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name: "valid report",
			body: `{"source_code":"RAVEN-1","classification":"SECRET","content_hash":"abc123","threat_actors":["APT-29"],"veracity_score":0.9}`,
			mockFn: func() {
				mockSvc.On("SubmitReport", mock.AnythingOfType("domain.SubmitReportRequest")).
					Return(domain.IntelligenceReport{SourceCode: "RAVEN-1", Classification: "SECRET"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid body",
			body:       `{"source_code":""}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"source_code":"FAIL-1","classification":"SECRET","content_hash":"abc"}`,
			mockFn: func() {
				mockSvc.On("SubmitReport", mock.AnythingOfType("domain.SubmitReportRequest")).
					Return(domain.IntelligenceReport{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/humint/reports", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetReportsHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	tests := []struct {
		name       string
		code       string
		mockFn     func()
		wantStatus int
		wantTotal  int
	}{
		{
			name: "has reports",
			code: "RAVEN-1",
			mockFn: func() {
				mockSvc.On("GetReportsBySource", "RAVEN-1").Return([]domain.IntelligenceReport{
					{SourceCode: "RAVEN-1", Classification: "SECRET"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name: "no reports",
			code: "NEW-1",
			mockFn: func() {
				mockSvc.On("GetReportsBySource", "NEW-1").Return([]domain.IntelligenceReport{}, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/humint/sources/"+tt.code+"/reports", nil)
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp domain.ReportsResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantTotal, resp.Total)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestLogDebriefingHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	tests := []struct {
		name       string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name: "valid debriefing",
			body: `{"source_code":"RAVEN-1","officer_id":"550e8400-e29b-41d4-a716-446655440000","session_date":"2026-06-20T10:00:00Z","location_method":"IN_PERSON","topics_covered":["Status update"],"risk_assessment":"Low"}`,
			mockFn: func() {
				mockSvc.On("LogDebriefing", mock.AnythingOfType("domain.LogDebriefingRequest")).
					Return(domain.DebriefingSession{SourceCode: "RAVEN-1", LocationMethod: "IN_PERSON"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing fields",
			body:       `{"source_code":""}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/humint/debriefings", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetSourceNetworkHandler(t *testing.T) {
	mockSvc := new(mockHumintSvc)

	t.Run("returns network data", func(t *testing.T) {
		mockSvc.On("GetSourceNetwork").Return(domain.SourceNetworkResponse{
			Nodes: []domain.SourceNetworkNode{
				{ID: "RAVEN-1", Type: "source"},
				{ID: "APT-29", Type: "threat_actor"},
			},
			Edges: []domain.SourceNetworkEdge{
				{Source: "RAVEN-1", Target: "APT-29", Label: "reported"},
			},
		}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/humint/analytics/source-network", nil)
		setupRouter(mockSvc).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
