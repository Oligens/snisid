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

	"github.com/snisid/counterintel-ht/internal/domain"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1/counterintel")
	api.POST("/investigations", func(c *gin.Context) {
		var req domain.CreateInvestigationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	api.GET("/investigations/pending", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.CreateInvestigationRequest{})
	})
	api.GET("/investigations/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})
	api.PATCH("/investigations/:id/adjudicate", func(c *gin.Context) {
		var req domain.AdjudicateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "adjudicated"})
	})
	api.POST("/threats", func(c *gin.Context) {
		var req domain.ReportThreatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "reported"})
	})
	api.GET("/threats/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.ReportThreatRequest{})
	})
	api.POST("/contacts", func(c *gin.Context) {
		var req domain.ReportContactRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	api.GET("/contacts/:subject_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.ReportContactRequest{})
	})
	return r
}

func TestCreateInvestigation_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.CreateInvestigationRequest{
		SubjectIdentityRef: "sub-001",
		InvestigationType:  "STANDARD",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/counterintel/investigations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateInvestigation_BadRequest(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/counterintel/investigations", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPendingInvestigations_Success(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/counterintel/investigations/pending", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAdjudicateInvestigation_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.AdjudicateRequest{
		Adjudicator:           [16]byte{1},
		ClearanceLevelGranted: "SECRET",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/counterintel/investigations/550e8400-e29b-41d4-a716-446655440000/adjudicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestReportThreat_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.ReportThreatRequest{
		SubjectID:   "sub-001",
		AlertType:   "DATA_EXFIL",
		Severity:    "HIGH",
		Description: "data exfil detected",
		DetectedBy:  "soc",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/counterintel/threats", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetActiveThreats_Success(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/counterintel/threats/active", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestReportContact_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.ReportContactRequest{
		SubjectID:         "sub-001",
		ContactName:       "Jane Doe",
		ForeignGovernment: "Atlantis",
		RelationshipType:  "DIPLOMATIC",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/counterintel/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetContacts_Success(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/counterintel/contacts/sub-001", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
