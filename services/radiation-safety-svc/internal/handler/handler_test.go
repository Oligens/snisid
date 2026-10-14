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
	"github.com/snisid/radiation-safety-svc/internal/domain"
	"github.com/snisid/radiation-safety-svc/internal/repository"
	"github.com/snisid/radiation-safety-svc/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRadiationSvc struct {
	service.RadiationService
	registerSourceFn         func(ctx context.Context, s *domain.RadioactiveSource) error
	updateSourceStatusFn     func(ctx context.Context, id uuid.UUID, status domain.SourceStatus) error
	createAlertFn            func(ctx context.Context, a *domain.RadiationAlert) error
	getUnrespondedAlertsFn   func(ctx context.Context) ([]domain.RadiationAlert, error)
	registerChemicalFn       func(ctx context.Context, c *domain.ChemicalPrecursor) error
	getSuspiciousChemicalsFn func(ctx context.Context) ([]domain.ChemicalPrecursor, error)
	getDashboardFn           func(ctx context.Context) (*repository.DashboardStats, error)
}

func (m *mockRadiationSvc) RegisterSource(ctx context.Context, s *domain.RadioactiveSource) error {
	return m.registerSourceFn(ctx, s)
}
func (m *mockRadiationSvc) UpdateSourceStatus(ctx context.Context, id uuid.UUID, status domain.SourceStatus) error {
	return m.updateSourceStatusFn(ctx, id, status)
}
func (m *mockRadiationSvc) CreateAlert(ctx context.Context, a *domain.RadiationAlert) error {
	return m.createAlertFn(ctx, a)
}
func (m *mockRadiationSvc) GetUnrespondedAlerts(ctx context.Context) ([]domain.RadiationAlert, error) {
	return m.getUnrespondedAlertsFn(ctx)
}
func (m *mockRadiationSvc) RegisterChemical(ctx context.Context, c *domain.ChemicalPrecursor) error {
	return m.registerChemicalFn(ctx, c)
}
func (m *mockRadiationSvc) GetSuspiciousChemicals(ctx context.Context) ([]domain.ChemicalPrecursor, error) {
	return m.getSuspiciousChemicalsFn(ctx)
}
func (m *mockRadiationSvc) GetDashboard(ctx context.Context) (*repository.DashboardStats, error) {
	return m.getDashboardFn(ctx)
}

func setupRadiationHandler(svc service.RadiationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewRadiationHandler(svc)
	r := gin.New()
	h.RegisterRoutes(r)
	return r
}

func TestRegisterSourceHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
	}{
		{name: "success", body: `{"source_type":"MEDICAL","isotope":"Co-60","activity_curie":50,"location_building":"BldgA","location_lat":40.71,"location_lng":-74.00,"custodian_org":"Hospital","license_ref":"LIC-001","status":"REGISTERED"}`, wantStatus: http.StatusCreated},
		{name: "bad json", body: `invalid`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"source_type":"MEDICAL","isotope":"Co-60","activity_curie":50,"location_building":"BldgA","location_lat":40.71,"location_lng":-74.00,"custodian_org":"Hospital","license_ref":"LIC-001","status":"REGISTERED"}`, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				registerSourceFn: func(_ context.Context, _ *domain.RadioactiveSource) error {
					return tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/radiation/sources", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdateSourceStatusHandler(t *testing.T) {
	validID := uuid.New().String()
	tests := []struct {
		name       string
		id         string
		body       string
		svcErr     error
		wantStatus int
	}{
		{name: "success", id: validID, body: `{"status":"LOST"}`, wantStatus: http.StatusOK},
		{name: "invalid id", id: "bad-id", body: `{"status":"LOST"}`, wantStatus: http.StatusBadRequest},
		{name: "bad json", id: validID, body: `invalid`, wantStatus: http.StatusBadRequest},
		{name: "service error", id: validID, body: `{"status":"LOST"}`, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				updateSourceStatusFn: func(_ context.Context, _ uuid.UUID, _ domain.SourceStatus) error {
					return tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/radiation/sources/"+tt.id+"/status", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCreateAlertHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
	}{
		{name: "success", body: `{"detector_id":"DET-01","detector_location":"GateA","detected_isotope":"Cs-137","dose_rate_usv":5.0,"alert_level":"YELLOW"}`, wantStatus: http.StatusCreated},
		{name: "bad json", body: `invalid`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"detector_id":"DET-01","detector_location":"GateA","detected_isotope":"Cs-137","dose_rate_usv":5.0,"alert_level":"YELLOW"}`, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				createAlertFn: func(_ context.Context, _ *domain.RadiationAlert) error {
					return tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/radiation/alerts", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetUnrespondedAlertsHandler(t *testing.T) {
	tests := []struct {
		name       string
		result     []domain.RadiationAlert
		svcErr     error
		wantStatus int
	}{
		{name: "success", result: []domain.RadiationAlert{}, wantStatus: http.StatusOK},
		{name: "service error", svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				getUnrespondedAlertsFn: func(_ context.Context) ([]domain.RadiationAlert, error) {
					return tt.result, tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/radiation/alerts/unresponded", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestRegisterChemicalHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantStatus int
	}{
		{name: "success", body: `{"substance_name":"NH3","cas_number":"7664-41-7","category":"PRECURSOR","quantity_kg":100,"storage_location":"WH-1","importer_entity":"Corp","end_user":"Lab","end_use":"Research","permit_ref":"P-001","reported_suspicious":false}`, wantStatus: http.StatusCreated},
		{name: "bad json", body: `invalid`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"substance_name":"NH3","cas_number":"7664-41-7","category":"PRECURSOR","quantity_kg":100,"storage_location":"WH-1","importer_entity":"Corp","end_user":"Lab","end_use":"Research","permit_ref":"P-001","reported_suspicious":false}`, svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				registerChemicalFn: func(_ context.Context, _ *domain.ChemicalPrecursor) error {
					return tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/radiation/chemicals", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetSuspiciousChemicalsHandler(t *testing.T) {
	tests := []struct {
		name       string
		result     []domain.ChemicalPrecursor
		svcErr     error
		wantStatus int
	}{
		{name: "success", result: []domain.ChemicalPrecursor{}, wantStatus: http.StatusOK},
		{name: "service error", svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				getSuspiciousChemicalsFn: func(_ context.Context) ([]domain.ChemicalPrecursor, error) {
					return tt.result, tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/radiation/chemicals/suspicious", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetDashboardHandler(t *testing.T) {
	dash := &repository.DashboardStats{TotalSources: 10}
	tests := []struct {
		name       string
		result     *repository.DashboardStats
		svcErr     error
		wantStatus int
	}{
		{name: "success", result: dash, wantStatus: http.StatusOK},
		{name: "service error", svcErr: errors.New("svc error"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockRadiationSvc{
				getDashboardFn: func(_ context.Context) (*repository.DashboardStats, error) {
					return tt.result, tt.svcErr
				},
			}
			r := setupRadiationHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/radiation/dashboard", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestNewRadiationHandler(t *testing.T) {
	h := NewRadiationHandler(&mockRadiationSvc{})
	require.NotNil(t, h)
}
