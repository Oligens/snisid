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
	"github.com/snisid/air-defense-ht/internal/domain"
)

type mockService struct {
	ingestFn    func(domain.RadarContact) error
	getActiveFn func() ([]domain.RadarContact, error)
	getByIDFn   func(uuid.UUID) (*domain.RadarContact, error)
	openIncFn   func(domain.AirDefenseIncident) error
	resolveFn   func(uuid.UUID) error
	addNoFlyFn  func(domain.NoFlyListEntry) error
	checkFn     func(string) (*domain.NoFlyListEntry, error)
}

func (m *mockService) IngestRadarContact(c domain.RadarContact) error  { return m.ingestFn(c) }
func (m *mockService) GetActiveTracks() ([]domain.RadarContact, error) { return m.getActiveFn() }
func (m *mockService) GetTrackByID(id uuid.UUID) (*domain.RadarContact, error) {
	return m.getByIDFn(id)
}
func (m *mockService) OpenIncident(i domain.AirDefenseIncident) error { return m.openIncFn(i) }
func (m *mockService) ResolveIncident(id uuid.UUID) error             { return m.resolveFn(id) }
func (m *mockService) AddNoFlyEntry(e domain.NoFlyListEntry) error    { return m.addNoFlyFn(e) }
func (m *mockService) CheckNoFly(identity string) (*domain.NoFlyListEntry, error) {
	return m.checkFn(identity)
}

func setupRouter(h *AirDefenseHandler) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1/airdef")
	{
		v1.POST("/tracks", h.IngestTrack)
		v1.GET("/tracks/active", h.GetActiveTracks)
		v1.GET("/tracks/:track_id", h.GetTrackByID)
		v1.POST("/incidents", h.OpenIncident)
		v1.PATCH("/incidents/:id/resolve", h.ResolveIncident)
		v1.POST("/no-fly", h.AddNoFly)
		v1.GET("/no-fly/check", h.CheckNoFly)
	}
	return r
}

func TestIngestTrackHandler(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			ingestFn: func(c domain.RadarContact) error { return nil },
		},
	}
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]interface{}{
		"track_number": "TRK001", "latitude": 10.0, "longitude": 20.0, "source_radar": "RDR-A",
	})
	req := httptest.NewRequest("POST", "/api/v1/airdef/tracks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestIngestTrackHandlerBadRequest(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{ingestFn: func(c domain.RadarContact) error { return nil }},
	}
	router := setupRouter(h)

	req := httptest.NewRequest("POST", "/api/v1/airdef/tracks", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestGetActiveTracksHandler(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			getActiveFn: func() ([]domain.RadarContact, error) {
				return []domain.RadarContact{{TrackNumber: "TRK001"}}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/tracks/active", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestGetTrackByIDHandler(t *testing.T) {
	id := uuid.New()
	h := &AirDefenseHandler{
		svc: &mockService{
			getByIDFn: func(uid uuid.UUID) (*domain.RadarContact, error) {
				return &domain.RadarContact{ContactID: uid}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/tracks/"+id.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestGetTrackByIDHandlerNotFound(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			getByIDFn: func(uid uuid.UUID) (*domain.RadarContact, error) { return nil, nil },
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/tracks/"+uuid.New().String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestOpenIncidentHandler(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			openIncFn: func(i domain.AirDefenseIncident) error { return nil },
		},
	}
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]interface{}{
		"aircraft_id": uuid.New().String(),
	})
	req := httptest.NewRequest("POST", "/api/v1/airdef/incidents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestResolveIncidentHandler(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			resolveFn: func(id uuid.UUID) error { return nil },
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("PATCH", "/api/v1/airdef/incidents/"+uuid.New().String()+"/resolve", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestAddNoFlyHandler(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			addNoFlyFn: func(e domain.NoFlyListEntry) error { return nil },
		},
	}
	router := setupRouter(h)

	body, _ := json.Marshal(map[string]interface{}{
		"identity_ref": "ID123", "full_name": "John Doe",
		"reason": "threat", "added_by": "admin", "expires_at": "2027-01-01T00:00:00Z",
	})
	req := httptest.NewRequest("POST", "/api/v1/airdef/no-fly", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestCheckNoFlyHandlerRestricted(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			checkFn: func(identity string) (*domain.NoFlyListEntry, error) {
				return &domain.NoFlyListEntry{IdentityRef: identity}, nil
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/no-fly/check?identity=ID123", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["restricted"] != true {
		t.Fatal("expected restricted=true")
	}
}

func TestCheckNoFlyHandlerNotRestricted(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			checkFn: func(identity string) (*domain.NoFlyListEntry, error) { return nil, nil },
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/no-fly/check?identity=UNKNOWN", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["restricted"] != false {
		t.Fatal("expected restricted=false")
	}
}

func TestCheckNoFlyHandlerMissingParam(t *testing.T) {
	h := &AirDefenseHandler{svc: &mockService{}}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/no-fly/check", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestHandlerInternalError(t *testing.T) {
	h := &AirDefenseHandler{
		svc: &mockService{
			getActiveFn: func() ([]domain.RadarContact, error) {
				return nil, errors.New("internal error")
			},
		},
	}
	router := setupRouter(h)
	req := httptest.NewRequest("GET", "/api/v1/airdef/tracks/active", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
}
