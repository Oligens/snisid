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

	"github.com/snisid/classification-mgmt-ht/internal/domain"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1/classification")
	api.POST("/rules", func(c *gin.Context) {
		var req domain.CreateRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})
	api.GET("/rules/:data_type", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.CreateRuleRequest{})
	})
	api.POST("/tags", func(c *gin.Context) {
		var req domain.TagResourceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "tagged"})
	})
	api.GET("/tags/check", func(c *gin.Context) {
		uri := c.Query("uri")
		if uri == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "uri required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"uri": uri})
	})
	api.POST("/audit", func(c *gin.Context) {
		var req domain.LogAuditRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "logged"})
	})
	api.GET("/audit/recent", func(c *gin.Context) {
		c.JSON(http.StatusOK, []domain.LogAuditRequest{})
	})
	api.GET("/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, domain.DashboardStats{TotalRules: 10})
	})
	return r
}

func TestCreateRule_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.CreateRuleRequest{
		DataType: "SSN", SensitivityLevel: "TOP_SECRET",
		CreatedBy: "550e8400-e29b-41d4-a716-446655440000",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/classification/rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateRule_BadRequest(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/classification/rules", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRulesByDataType(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/classification/rules/SSN", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestTagResource_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.TagResourceRequest{
		ResourceURI: "snisid://doc/1", ClassificationTop: "SECRET",
		OwnerAgency: "DHS", TaggedBy: "550e8400-e29b-41d4-a716-446655440000",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/classification/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetClassificationByURI(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/classification/tags/check?uri=snisid://doc/1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetClassificationByURI_MissingParam(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/classification/tags/check", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogAudit_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(domain.LogAuditRequest{
		ResourceURI: "snisid://doc/1", Action: "CLASSIFY",
		AuthorizedBy:            "550e8400-e29b-41d4-a716-446655440000",
		ClassificationAuthority: "EO 13526", IPAddress: "10.0.0.1",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/classification/audit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetRecentAuditLogs(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/classification/audit/recent", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetDashboard(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/classification/dashboard", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
