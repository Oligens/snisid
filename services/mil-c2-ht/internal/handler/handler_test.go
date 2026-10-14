package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/mil-c2-ht/internal/domain"
)

type mockService struct {
	createUnitFn  func(domain.MilitaryUnit) error
	getDeployedFn func() ([]domain.MilitaryUnit, error)
	createOpFn    func(domain.Operation) error
	getActiveFn   func() ([]domain.Operation, error)
	submitFn      func(uuid.UUID, domain.TacticalReport) error
	timelineFn    func(uuid.UUID) ([]domain.TacticalReport, error)
	copFn         func() (*domain.CommonOperatingPicture, error)
}

func (m *mockService) CreateUnit(u domain.MilitaryUnit) error           { return m.createUnitFn(u) }
func (m *mockService) GetDeployedUnits() ([]domain.MilitaryUnit, error) { return m.getDeployedFn() }
func (m *mockService) CreateOperation(o domain.Operation) error         { return m.createOpFn(o) }
func (m *mockService) GetActiveOperations() ([]domain.Operation, error) { return m.getActiveFn() }
func (m *mockService) SubmitReport(id uuid.UUID, r domain.TacticalReport) error {
	return m.submitFn(id, r)
}
func (m *mockService) GetOperationTimeline(id uuid.UUID) ([]domain.TacticalReport, error) {
	return m.timelineFn(id)
}
func (m *mockService) GetCommonOperatingPicture() (*domain.CommonOperatingPicture, error) {
	return m.copFn()
}

func setupRouter(h *MilC2Handler) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1/milc2")
	{
		v1.POST("/units", h.CreateUnit)
		v1.GET("/units/deployed", h.GetDeployedUnits)
		v1.POST("/operations", h.CreateOperation)
		v1.GET("/operations/active", h.GetActiveOperations)
		v1.POST("/operations/:id/reports", h.SubmitReport)
		v1.GET("/operations/:id/timeline", h.GetOperationTimeline)
		v1.GET("/common-operating-picture", h.GetCommonOperatingPicture)
	}
	return r
}

func TestCreateUnitHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{createUnitFn: func(u domain.MilitaryUnit) error { return nil }},
	}
	router := setupRouter(h)
	body, _ := json.Marshal(map[string]interface{}{
		"unit_name": "1st Battalion", "branch": "ARMY",
	})
	req := httptest.NewRequest("POST", "/api/v1/milc2/units", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestCreateUnitHandlerBadRequest(t *testing.T) {
	h := &MilC2Handler{svc: &mockService{}}
	router := setupRouter(h)
	req := httptest.NewRequest("POST", "/api/v1/milc2/units", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestGetDeployedUnitsHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{
			getDeployedFn: func() ([]domain.MilitaryUnit, error) {
				return []domain.MilitaryUnit{{UnitName: "Recon"}}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/milc2/units/deployed", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestCreateOperationHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{createOpFn: func(o domain.Operation) error { return nil }},
	}
	router := setupRouter(h)
	body, _ := json.Marshal(map[string]interface{}{
		"operation_name": "Op Guardian", "operation_type": "SECURITY",
		"commander_id": uuid.New().String(), "start_date": "2026-06-20T00:00:00Z",
	})
	req := httptest.NewRequest("POST", "/api/v1/milc2/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestCreateOperationHandlerBadRequest(t *testing.T) {
	h := &MilC2Handler{svc: &mockService{}}
	router := setupRouter(h)
	req := httptest.NewRequest("POST", "/api/v1/milc2/operations", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestGetActiveOperationsHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{
			getActiveFn: func() ([]domain.Operation, error) {
				return []domain.Operation{{OperationName: "Op Guardian"}}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/milc2/operations/active", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestSubmitReportHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{
			submitFn: func(id uuid.UUID, r domain.TacticalReport) error { return nil },
		},
	}
	router := setupRouter(h)
	body, _ := json.Marshal(map[string]interface{}{
		"reporting_unit_id": uuid.New().String(), "report_type": "SITREP",
	})
	req := httptest.NewRequest("POST", "/api/v1/milc2/operations/"+uuid.New().String()+"/reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestGetOperationTimelineHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{
			timelineFn: func(id uuid.UUID) ([]domain.TacticalReport, error) {
				return []domain.TacticalReport{{ReportType: domain.ReportSITREP}}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/milc2/operations/"+uuid.New().String()+"/timeline", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestGetCommonOperatingPictureHandler(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{
			copFn: func() (*domain.CommonOperatingPicture, error) {
				return &domain.CommonOperatingPicture{
					Units:      []domain.MilitaryUnit{{UnitName: "Unit A"}},
					Operations: []domain.Operation{{OperationName: "Op A"}},
					Reports:    []domain.TacticalReport{{ReportType: domain.ReportSITREP}},
				}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/milc2/common-operating-picture", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestHandlerInternalError(t *testing.T) {
	h := &MilC2Handler{
		svc: &mockService{
			getDeployedFn: func() ([]domain.MilitaryUnit, error) {
				return nil, errors.New("internal error")
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/milc2/units/deployed", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}
