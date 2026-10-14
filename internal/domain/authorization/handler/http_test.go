package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/snisid/platform/internal/domain/authorization/entity"
)

type mockEngine struct {
	enforceFn func(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error)
	refreshFn func(ctx context.Context) error
}

func (m *mockEngine) Enforce(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error) {
	if m.enforceFn != nil {
		return m.enforceFn(ctx, req)
	}
	return &entity.AuthorizationDecision{Allowed: true}, nil
}

func (m *mockEngine) RefreshPolicies(ctx context.Context) error {
	if m.refreshFn != nil {
		return m.refreshFn(ctx)
	}
	return nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestEnforce_Allowed(t *testing.T) {
	engine := &mockEngine{
		enforceFn: func(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error) {
			return &entity.AuthorizationDecision{Allowed: true, PolicyName: "test"}, nil
		},
	}
	handler := NewHttpHandler(engine)

	r := setupRouter()
	group := r.Group("/v1/authz")
	handler.RegisterRoutes(group)

	body, _ := json.Marshal(entity.AuthorizationRequest{
		Subject:  entity.SubjectData{UserID: "usr-001", Roles: []string{"admin"}},
		Action:   "read",
		Resource: "identity:NNU-123",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/authz/enforce", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp entity.AuthorizationDecision
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Allowed {
		t.Error("Expected allowed = true")
	}
}

func TestEnforce_Denied(t *testing.T) {
	engine := &mockEngine{
		enforceFn: func(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error) {
			return &entity.AuthorizationDecision{Allowed: false, Reason: "Insufficient role"}, nil
		},
	}
	handler := NewHttpHandler(engine)

	r := setupRouter()
	group := r.Group("/v1/authz")
	handler.RegisterRoutes(group)

	body, _ := json.Marshal(entity.AuthorizationRequest{
		Subject:  entity.SubjectData{UserID: "usr-002", Roles: []string{"viewer"}},
		Action:   "delete",
		Resource: "identity:NNU-999",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/authz/enforce", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp entity.AuthorizationDecision
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Allowed {
		t.Error("Expected allowed = false")
	}
}

func TestEnforce_InvalidBody(t *testing.T) {
	handler := NewHttpHandler(&mockEngine{})
	r := setupRouter()
	group := r.Group("/v1/authz")
	handler.RegisterRoutes(group)

	req := httptest.NewRequest(http.MethodPost, "/v1/authz/enforce", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestEnforce_EngineError(t *testing.T) {
	engine := &mockEngine{
		enforceFn: func(ctx context.Context, req *entity.AuthorizationRequest) (*entity.AuthorizationDecision, error) {
			return nil, errors.New("engine failure")
		},
	}
	handler := NewHttpHandler(engine)
	r := setupRouter()
	group := r.Group("/v1/authz")
	handler.RegisterRoutes(group)

	body, _ := json.Marshal(entity.AuthorizationRequest{
		Subject: entity.SubjectData{UserID: "usr-001"},
		Action:  "read",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/authz/enforce", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestRefresh_Success(t *testing.T) {
	refreshed := false
	engine := &mockEngine{
		refreshFn: func(ctx context.Context) error {
			refreshed = true
			return nil
		},
	}
	handler := NewHttpHandler(engine)
	r := setupRouter()
	group := r.Group("/v1/authz")
	handler.RegisterRoutes(group)

	req := httptest.NewRequest(http.MethodPost, "/v1/authz/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
	if !refreshed {
		t.Error("RefreshPolicies was not called")
	}
}

func TestRefresh_Error(t *testing.T) {
	engine := &mockEngine{
		refreshFn: func(ctx context.Context) error {
			return errors.New("refresh failed")
		},
	}
	handler := NewHttpHandler(engine)
	r := setupRouter()
	group := r.Group("/v1/authz")
	handler.RegisterRoutes(group)

	req := httptest.NewRequest(http.MethodPost, "/v1/authz/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
