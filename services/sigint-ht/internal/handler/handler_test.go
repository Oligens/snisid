package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/sigint-ht/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSigintSvc struct {
	mock.Mock
}

func (m *mockSigintSvc) CreateTarget(req domain.CreateTargetRequest) (domain.InterceptionTarget, error) {
	args := m.Called(req)
	return args.Get(0).(domain.InterceptionTarget), args.Error(1)
}

func (m *mockSigintSvc) GetActiveTargets() ([]domain.InterceptionTarget, error) {
	args := m.Called()
	return args.Get(0).([]domain.InterceptionTarget), args.Error(1)
}

func (m *mockSigintSvc) RecordInterception(targetID string, req domain.InterceptRequest) (domain.InterceptedCommunication, error) {
	args := m.Called(targetID, req)
	return args.Get(0).(domain.InterceptedCommunication), args.Error(1)
}

func (m *mockSigintSvc) GetCommunications(targetID string) ([]domain.InterceptedCommunication, error) {
	args := m.Called(targetID)
	return args.Get(0).([]domain.InterceptedCommunication), args.Error(1)
}

func (m *mockSigintSvc) AnalyzeCDR(phone string) ([]domain.CDRAnalysis, error) {
	args := m.Called(phone)
	return args.Get(0).([]domain.CDRAnalysis), args.Error(1)
}

func (m *mockSigintSvc) EmergencyAuthorization(req domain.EmergencyRequest) (domain.EmergencyResponse, error) {
	args := m.Called(req)
	return args.Get(0).(domain.EmergencyResponse), args.Error(1)
}

func setupRouter(svc SigintService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &SigintHandler{svc: svc}
	h.RegisterRoutes(r)
	return r
}

func TestCreateTargetHandler(t *testing.T) {
	mockSvc := new(mockSigintSvc)

	tests := []struct {
		name       string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name: "valid creation",
			body: `{"target_type":"PHONE_NUMBER","authorization_ref":"` + uuid.NewString() + `","judge_name":"Judge","issuing_court":"FISA","start_date":"2026-01-01T00:00:00Z","end_date":"2026-04-01T00:00:00Z","target_identifier":"+15550123"}`,
			mockFn: func() {
				mockSvc.On("CreateTarget", mock.AnythingOfType("domain.CreateTargetRequest")).
					Return(domain.InterceptionTarget{ID: uuid.NewString(), TargetType: "PHONE_NUMBER"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid body",
			body:       `{"target_type":""}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"target_type":"PHONE_NUMBER","authorization_ref":"` + uuid.NewString() + `","judge_name":"Judge","issuing_court":"FISA","start_date":"2026-01-01T00:00:00Z","end_date":"2026-04-01T00:00:00Z","target_identifier":"+15550123"}`,
			mockFn: func() {
				mockSvc.On("CreateTarget", mock.AnythingOfType("domain.CreateTargetRequest")).
					Return(domain.InterceptionTarget{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/sigint/targets", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestGetActiveTargetsHandler(t *testing.T) {
	mockSvc := new(mockSigintSvc)

	tests := []struct {
		name       string
		mockFn     func()
		wantStatus int
		wantTotal  int
	}{
		{
			name: "returns targets",
			mockFn: func() {
				mockSvc.On("GetActiveTargets").Return([]domain.InterceptionTarget{
					{ID: uuid.NewString(), TargetType: "PHONE_NUMBER", Status: "ACTIVE"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  1,
		},
		{
			name: "no targets",
			mockFn: func() {
				mockSvc.On("GetActiveTargets").Return([]domain.InterceptionTarget{}, nil)
			},
			wantStatus: http.StatusOK,
			wantTotal:  0,
		},
		{
			name: "service error",
			mockFn: func() {
				mockSvc.On("GetActiveTargets").Return([]domain.InterceptionTarget{}, errors.New("error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/sigint/targets/active", nil)
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp domain.TargetsResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.wantTotal, resp.Total)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestAnalyzeCDRHandler(t *testing.T) {
	mockSvc := new(mockSigintSvc)

	tests := []struct {
		name       string
		phone      string
		mockFn     func()
		wantStatus int
	}{
		{
			name:  "valid phone",
			phone: "+15550123",
			mockFn: func() {
				mockSvc.On("AnalyzeCDR", "+15550123").Return([]domain.CDRAnalysis{
					{ID: uuid.NewString(), Caller: "+15550123", Duration: 60},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing phone",
			phone:      "",
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "no records",
			phone: "+15559999",
			mockFn: func() {
				mockSvc.On("AnalyzeCDR", "+15559999").Return([]domain.CDRAnalysis{}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			url := "/api/v1/sigint/cdr/analysis"
			if tt.phone != "" {
				url += "?phone=" + tt.phone
			}
			req, _ := http.NewRequest("GET", url, nil)
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestRecordInterceptionHandler(t *testing.T) {
	mockSvc := new(mockSigintSvc)
	targetID := uuid.NewString()

	tests := []struct {
		name       string
		targetID   string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name:     "valid interception",
			targetID: targetID,
			body:     `{"comm_type":"CALL","content_ref":"s3://encrypted-bucket/rec1.mp4","intercepted_at":"2026-06-20T12:00:00Z","collector_node":"node-01","case_number":"CASE-001"}`,
			mockFn: func() {
				mockSvc.On("RecordInterception", targetID, mock.AnythingOfType("domain.InterceptRequest")).
					Return(domain.InterceptedCommunication{ID: uuid.NewString(), CommType: "CALL"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing fields",
			targetID:   targetID,
			body:       `{"comm_type":""}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/sigint/targets/"+tt.targetID+"/intercept", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestEmergencyAuthorizationHandler(t *testing.T) {
	mockSvc := new(mockSigintSvc)

	tests := []struct {
		name       string
		body       string
		mockFn     func()
		wantStatus int
	}{
		{
			name: "valid emergency",
			body: `{"target_identifier":"+15550123","target_type":"PHONE_NUMBER","reason":"Imminent threat","authorizing_officer":"DIR-001"}`,
			mockFn: func() {
				mockSvc.On("EmergencyAuthorization", mock.AnythingOfType("domain.EmergencyRequest")).
					Return(domain.EmergencyResponse{Approved: true, AuthRef: uuid.NewString()}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing fields",
			body:       `{}`,
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFn()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/sigint/emergency", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			setupRouter(mockSvc).ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
