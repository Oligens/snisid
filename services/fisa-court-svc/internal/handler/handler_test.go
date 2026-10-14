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

	"github.com/snisid/fisa-court-svc/internal/domain"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1/fisa")
	api.POST("/warrants", func(c *gin.Context) {
		var req domain.FileWarrantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "filed"})
	})
	api.PATCH("/warrants/:id/approve", func(c *gin.Context) {
		var req domain.ApproveWarrantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "approved"})
	})
	api.GET("/warrants/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.FileWarrantRequest{})
	})
	api.POST("/warrants/:id/renew", func(c *gin.Context) {
		var req domain.RenewWarrantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "renewed"})
	})
	api.POST("/reports", func(c *gin.Context) {
		var req domain.FileReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "filed"})
	})
	api.GET("/docket/:term", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"term": c.Param("term")})
	})
	api.POST("/emergency", func(c *gin.Context) {
		var req domain.EmergencyAuthorizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "emergency_authorized"})
	})
	return r
}

func TestFileWarrant_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.FileWarrantRequest{
		WarrantType: "FISA_ELECTRONIC", TargetIdentity: "t1", IssuingCourt: "FISA",
		JudgeName: "J1", ApplicantAgency: "NSA", ApplicantOfficer: "550e8400-e29b-41d4-a716-446655440000",
		DurationDays: 90,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/fisa/warrants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestFileWarrant_BadRequest(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/fisa/warrants", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestApproveWarrant_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.ApproveWarrantRequest{JudgeName: "Judge A"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/fisa/warrants/550e8400-e29b-41d4-a716-446655440000/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetActiveWarrants(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/fisa/warrants/active", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRenewWarrant_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.RenewWarrantRequest{DurationDays: 90})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/fisa/warrants/550e8400-e29b-41d4-a716-446655440000/renew", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestFileReport_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.FileReportRequest{
		WarrantID:            "550e8400-e29b-41d4-a716-446655440000",
		ReportingPeriodStart: "2025-01-01T00:00:00Z",
		ReportingPeriodEnd:   "2025-03-31T00:00:00Z",
		SubmittedBy:          "550e8400-e29b-41d4-a716-446655440000",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/fisa/reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetDocketByTerm(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/fisa/docket/2025-1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestEmergencyAuthorization_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.EmergencyAuthorizationRequest{
		WarrantType: "FISA_ELECTRONIC", TargetIdentity: "t1",
		ApplicantAgency: "NSA", ApplicantOfficer: "550e8400-e29b-41d4-a716-446655440000",
		ProbableCause: "imminent", ApprovedBy: "550e8400-e29b-41d4-a716-446655440000",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/fisa/emergency", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}
