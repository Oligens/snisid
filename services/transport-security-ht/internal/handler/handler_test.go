package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/transport-security-ht/internal/domain"
	"github.com/snisid/transport-security-ht/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTransportSvc struct {
	service.TransportService
	logScreeningFn        func(ctx context.Context, s *domain.PassengerScreening) error
	getRecentScreeningsFn func(ctx context.Context) ([]domain.PassengerScreening, error)
	addNoFlyEntryFn       func(ctx context.Context, p *domain.NoFlyPassenger) error
	checkNoFlyFn          func(ctx context.Context, identityRef string) (*domain.NoFlyPassenger, error)
	getZonesByAirportFn   func(ctx context.Context, airportCode string) ([]domain.AirportSecurityZone, error)
	reportBreachFn        func(ctx context.Context, zoneID uuid.UUID) error
}

func (m *mockTransportSvc) LogScreening(ctx context.Context, s *domain.PassengerScreening) error {
	return m.logScreeningFn(ctx, s)
}
func (m *mockTransportSvc) GetRecentScreenings(ctx context.Context) ([]domain.PassengerScreening, error) {
	return m.getRecentScreeningsFn(ctx)
}
func (m *mockTransportSvc) AddNoFlyEntry(ctx context.Context, p *domain.NoFlyPassenger) error {
	return m.addNoFlyEntryFn(ctx, p)
}
func (m *mockTransportSvc) CheckNoFly(ctx context.Context, identityRef string) (*domain.NoFlyPassenger, error) {
	return m.checkNoFlyFn(ctx, identityRef)
}
func (m *mockTransportSvc) GetZonesByAirport(ctx context.Context, airportCode string) ([]domain.AirportSecurityZone, error) {
	return m.getZonesByAirportFn(ctx, airportCode)
}
func (m *mockTransportSvc) ReportBreach(ctx context.Context, zoneID uuid.UUID) error {
	return m.reportBreachFn(ctx, zoneID)
}

func setupTransportHandler(svc service.TransportService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewTransportHandler(svc)
	r := gin.New()
	h.RegisterRoutes(r)
	return r
}

func TestLogScreeningHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
	}{
		{name: "success", body: `{"traveler_identity_ref":"T123","document_type":"PASSPORT","document_number":"AB123","nationality":"US","travel_mode":"AIR","screening_point_type":"AIRPORT","screening_point_name":"JFK","departure_at":"2025-01-01T00:00:00Z","arrival_at":"2025-01-01T02:00:00Z","watchlist_match":false,"screening_result":"CLEAR","screening_officer":"OFF1"}`, wantStatus: http.StatusCreated},
		{name: "bad json", body: `invalid`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"traveler_identity_ref":"T123","document_type":"PASSPORT","document_number":"AB123","nationality":"US","travel_mode":"AIR","screening_point_type":"AIRPORT","screening_point_name":"JFK","departure_at":"2025-01-01T00:00:00Z","arrival_at":"2025-01-01T02:00:00Z","watchlist_match":false,"screening_result":"CLEAR","screening_officer":"OFF1"}`, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTransportSvc{
				logScreeningFn: func(_ context.Context, _ *domain.PassengerScreening) error {
					return tt.svcErr
				},
			}
			r := setupTransportHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transport/screenings", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetRecentScreeningsHandler(t *testing.T) {
	tests := []struct {
		name       string
		result     []domain.PassengerScreening
		svcErr     error
		wantStatus int
	}{
		{name: "success", result: []domain.PassengerScreening{}, wantStatus: http.StatusOK},
		{name: "service error", svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTransportSvc{
				getRecentScreeningsFn: func(_ context.Context) ([]domain.PassengerScreening, error) {
					return tt.result, tt.svcErr
				},
			}
			r := setupTransportHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/transport/screenings/recent", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAddNoFlyHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
	}{
		{name: "success", body: `{"identity_ref":"T123","list_type":"NO_FLY","added_by":"00000000-0000-0000-0000-000000000001","reason":"threat"}`, wantStatus: http.StatusCreated},
		{name: "bad json", body: `invalid`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"identity_ref":"T123","list_type":"NO_FLY","added_by":"00000000-0000-0000-0000-000000000001","reason":"threat"}`, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTransportSvc{
				addNoFlyEntryFn: func(_ context.Context, _ *domain.NoFlyPassenger) error {
					return tt.svcErr
				},
			}
			r := setupTransportHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transport/no-fly", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCheckNoFlyHandler(t *testing.T) {
	tests := []struct {
		name       string
		identity   string
		result     *domain.NoFlyPassenger
		svcErr     error
		wantStatus int
		wantMatch  bool
	}{
		{name: "match found", identity: "T123", result: &domain.NoFlyPassenger{IdentityRef: "T123"}, wantStatus: http.StatusOK, wantMatch: true},
		{name: "no match", identity: "T456", result: nil, wantStatus: http.StatusOK, wantMatch: false},
		{name: "missing param", identity: "", wantStatus: http.StatusBadRequest},
		{name: "service error", identity: "T123", svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTransportSvc{
				checkNoFlyFn: func(_ context.Context, _ string) (*domain.NoFlyPassenger, error) {
					return tt.result, tt.svcErr
				},
			}
			r := setupTransportHandler(svc)
			url := "/api/v1/transport/no-fly/check"
			if tt.identity != "" {
				url += "?identity=" + tt.identity
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var body map[string]interface{}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Equal(t, tt.wantMatch, body["match"])
			}
		})
	}
}

func TestGetZonesByAirportHandler(t *testing.T) {
	tests := []struct {
		name       string
		airport    string
		result     []domain.AirportSecurityZone
		svcErr     error
		wantStatus int
	}{
		{name: "success", airport: "JFK", result: []domain.AirportSecurityZone{}, wantStatus: http.StatusOK},
		{name: "missing airport", airport: "", wantStatus: http.StatusBadRequest},
		{name: "service error", airport: "JFK", svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTransportSvc{
				getZonesByAirportFn: func(_ context.Context, _ string) ([]domain.AirportSecurityZone, error) {
					return tt.result, tt.svcErr
				},
			}
			r := setupTransportHandler(svc)
			url := "/api/v1/transport/zones/" + tt.airport
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestReportZoneBreachHandler(t *testing.T) {
	validID := uuid.New().String()
	tests := []struct {
		name       string
		zoneID     string
		svcErr     error
		wantStatus int
	}{
		{name: "success", zoneID: validID, wantStatus: http.StatusOK},
		{name: "invalid id", zoneID: "bad-id", wantStatus: http.StatusBadRequest},
		{name: "service error", zoneID: validID, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTransportSvc{
				reportBreachFn: func(_ context.Context, _ uuid.UUID) error {
					return tt.svcErr
				},
			}
			r := setupTransportHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/transport/zones/"+tt.zoneID+"/breach", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestNewTransportHandler(t *testing.T) {
	h := NewTransportHandler(&mockTransportSvc{})
	require.NotNil(t, h)
}
