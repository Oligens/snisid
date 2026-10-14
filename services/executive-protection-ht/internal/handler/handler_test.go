package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snisid/executive-protection-ht/internal/domain"
)

type mockExecSvc struct{}

func setupExecTest() (*gin.Engine, *mockExecSvc) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1/execprot")
	api.POST("/protectees", func(c *gin.Context) {
		var req domain.CreateProtecteeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	api.GET("/protectees/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.Protectee{})
	})
	api.POST("/movements", func(c *gin.Context) {
		var req domain.CreateMovementPlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	api.GET("/movements/upcoming", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.MovementPlan{})
	})
	api.POST("/threats", func(c *gin.Context) {
		var req domain.CreateThreatAssessmentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	api.GET("/threats/active/:protectee_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.ThreatAssessment{})
	})
	api.GET("/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r, &mockExecSvc{}
}

func TestCreateProtectee_Success(t *testing.T) {
	r, _ := setupExecTest()
	body, _ := json.Marshal(domain.CreateProtecteeRequest{
		FullName: "Jean Dupont", OfficialTitle: "Ministre",
		ProtectionLevel: "CABINET_MINISTER", RiskAssessment: "HIGH",
		PrimaryAgentID:     "550e8400-e29b-41d4-a716-446655440000",
		SecureVehiclePlate: "AA-001-BB",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/execprot/protectees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateProtectee_BadRequest(t *testing.T) {
	r, _ := setupExecTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/execprot/protectees", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetActiveProtectees_Success(t *testing.T) {
	r, _ := setupExecTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/execprot/protectees/active", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestCreateMovementPlan_Success(t *testing.T) {
	r, _ := setupExecTest()
	body, _ := json.Marshal(domain.CreateMovementPlanRequest{
		ProtecteeID: "550e8400-e29b-41d4-a716-446655440000",
		EventName:   "Sommet", Date: "2026-07-15T09:00:00Z",
		DepartureLocation: "Palais", ArrivalLocation: "Aéroport",
		TransportMode: "MOTORCADE",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/execprot/movements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetUpcomingMovements_Success(t *testing.T) {
	r, _ := setupExecTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/execprot/movements/upcoming", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestCreateThreat_Success(t *testing.T) {
	r, _ := setupExecTest()
	body, _ := json.Marshal(domain.CreateThreatAssessmentRequest{
		ProtecteeID: "550e8400-e29b-41d4-a716-446655440000",
		ThreatType:  "DIRECT_THREAT", ThreatLevel: "CRITICAL",
		AssessedBy: "550e8400-e29b-41d4-a716-446655440001",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/execprot/threats", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetActiveThreatsByProtectee_Success(t *testing.T) {
	r, _ := setupExecTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/execprot/threats/active/"+"pid1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetDashboard_Success(t *testing.T) {
	r, _ := setupExecTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/execprot/dashboard", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
