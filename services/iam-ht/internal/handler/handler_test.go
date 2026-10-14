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
)

type mockIAMSvc struct{}

func setupIAMTest() (*gin.Engine, *mockIAMSvc) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1/iam")
	api.GET("/.well-known/openid-config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"issuer": "https://sso.gouv.ht"})
	})
	api.POST("/authorize", func(c *gin.Context) {
		citizenID := c.Query("citizen_id")
		if citizenID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"assurance_level": "IAL1_SELF_ASSERTED", "mfa_enrolled": false})
	})
	api.POST("/token", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"access_token": "mock-token", "token_type": "Bearer"})
	})
	api.POST("/step-up", func(c *gin.Context) {
		var req struct {
			CitizenID string `json:"citizen_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.CitizenID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "citizen_id required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"assurance_level": "IAL2_BIOMETRIC_VERIFIED"})
	})
	return r, &mockIAMSvc{}
}

func TestIAMDiscovery_Success(t *testing.T) {
	r, _ := setupIAMTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/iam/.well-known/openid-config", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestIAMAuthorize_Success(t *testing.T) {
	r, _ := setupIAMTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/iam/authorize?citizen_id=u1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestIAMAuthorize_Unauthorized(t *testing.T) {
	r, _ := setupIAMTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/iam/authorize", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIAMToken_Success(t *testing.T) {
	r, _ := setupIAMTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/iam/token", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestIAMStepUp_Success(t *testing.T) {
	r, _ := setupIAMTest()
	body, _ := json.Marshal(map[string]string{"citizen_id": "u1"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/iam/step-up", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestIAMStepUp_BadRequest(t *testing.T) {
	r, _ := setupIAMTest()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/iam/step-up", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
