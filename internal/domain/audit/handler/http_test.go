package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/domain/audit/entity"
	"github.com/snisid/platform/internal/domain/audit/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockForensicsService struct {
	verifyIntegrityFn      func(ctx context.Context, startSeq, endSeq int64) (bool, error)
	queryByCorrelationIDFn func(ctx context.Context, correlationID string) ([]entity.AuditEvent, error)
}

func (m *mockForensicsService) VerifyIntegrity(ctx context.Context, startSeq, endSeq int64) (bool, error) {
	if m.verifyIntegrityFn != nil {
		return m.verifyIntegrityFn(ctx, startSeq, endSeq)
	}
	return true, nil
}

func (m *mockForensicsService) QueryByCorrelationID(ctx context.Context, correlationID string) ([]entity.AuditEvent, error) {
	if m.queryByCorrelationIDFn != nil {
		return m.queryByCorrelationIDFn(ctx, correlationID)
	}
	return []entity.AuditEvent{}, nil
}

func setupAuditRouter(svc usecase.ForensicsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHttpHandler(svc)
	h.RegisterRoutes(r.Group("/v1/audit"))
	return r
}

func TestVerifyIntegrity_Success(t *testing.T) {
	svc := &mockForensicsService{}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/verify?start=1&end=100", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "verified", resp["status"])
	assert.Equal(t, true, resp["valid"])
}

func TestVerifyIntegrity_TamperDetected(t *testing.T) {
	svc := &mockForensicsService{
		verifyIntegrityFn: func(ctx context.Context, startSeq, endSeq int64) (bool, error) {
			return false, assert.AnError
		},
	}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/verify?start=1&end=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "tamper_detected", resp["status"])
}

func TestVerifyIntegrity_DefaultEnd(t *testing.T) {
	svc := &mockForensicsService{}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/verify?start=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVerifyIntegrity_NoParams(t *testing.T) {
	svc := &mockForensicsService{}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/verify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler defaults end=9999999999 when missing, so it succeeds
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQueryByCorrelationID_Success(t *testing.T) {
	svc := &mockForensicsService{}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/query?correlationId=corr-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQueryByCorrelationID_MissingParam(t *testing.T) {
	svc := &mockForensicsService{}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/query", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQueryByCorrelationID_ServiceError(t *testing.T) {
	svc := &mockForensicsService{
		queryByCorrelationIDFn: func(ctx context.Context, correlationID string) ([]entity.AuditEvent, error) {
			return nil, assert.AnError
		},
	}
	router := setupAuditRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/query?correlationId=corr-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
