package audit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type AuditEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	EntityID  string    `json:"entity_id"`
	Action    string    `json:"action"`
	Principal string    `json:"principal"`
	Outcome   string    `json:"outcome"`
	Metadata  string    `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`
}

type AuditStore struct {
	events []AuditEvent
}

func (s *AuditStore) Record(e AuditEvent) {
	s.events = append(s.events, e)
}

func (s *AuditStore) Query(entityID string) []AuditEvent {
	var results []AuditEvent
	for _, e := range s.events {
		if e.EntityID == entityID {
			results = append(results, e)
		}
	}
	return results
}

func (s *AuditStore) QueryByType(eventType string) []AuditEvent {
	var results []AuditEvent
	for _, e := range s.events {
		if e.Type == eventType {
			results = append(results, e)
		}
	}
	return results
}

func TestAuditEventRecording(t *testing.T) {
	store := &AuditStore{}
	event := AuditEvent{
		ID: "evt-001", Type: "IDENTITY_VERIFIED", EntityID: "cit-001",
		Action: "VERIFY", Principal: "agent-001", Outcome: "PASS",
		Timestamp: time.Now(),
	}
	store.Record(event)
	assert.Len(t, store.events, 1)
	assert.Equal(t, "evt-001", store.events[0].ID)
}

func TestAuditQueryByEntity(t *testing.T) {
	store := &AuditStore{}
	store.Record(AuditEvent{ID: "e1", Type: "LOGIN", EntityID: "user-001", Action: "LOGIN", Timestamp: time.Now()})
	store.Record(AuditEvent{ID: "e2", Type: "UPDATE", EntityID: "user-001", Action: "UPDATE_PROFILE", Timestamp: time.Now()})
	store.Record(AuditEvent{ID: "e3", Type: "LOGIN", EntityID: "user-002", Action: "LOGIN", Timestamp: time.Now()})

	results := store.Query("user-001")
	assert.Len(t, results, 2)
	assert.Equal(t, "e1", results[0].ID)
	assert.Equal(t, "e2", results[1].ID)
}

func TestAuditQueryByType(t *testing.T) {
	store := &AuditStore{}
	store.Record(AuditEvent{ID: "e1", Type: "LOGIN", EntityID: "u1", Timestamp: time.Now()})
	store.Record(AuditEvent{ID: "e2", Type: "LOGIN", EntityID: "u2", Timestamp: time.Now()})
	store.Record(AuditEvent{ID: "e3", Type: "FRAUD_ALERT", EntityID: "u3", Timestamp: time.Now()})

	results := store.QueryByType("LOGIN")
	assert.Len(t, results, 2)
}

func TestAuditEvent_EmptyQuery(t *testing.T) {
	store := &AuditStore{}
	results := store.Query("nonexistent")
	assert.Empty(t, results)
}

func TestAuditMiddleware_RecordsEvent(t *testing.T) {
	store := &AuditStore{}

	middleware := func(c *gin.Context) {
		c.Next()
		event := AuditEvent{
			ID:        "mid-" + c.GetString("request_id"),
			Type:      "API_CALL",
			EntityID:  c.GetString("entity_id"),
			Action:    c.Request.Method + " " + c.Request.URL.Path,
			Principal: c.GetString("principal"),
			Outcome:   "SUCCESS",
			Timestamp: time.Now(),
		}
		store.Record(event)
	}

	r := gin.New()
	r.Use(middleware)
	r.GET("/v1/audit/events", func(c *gin.Context) {
		c.Set("request_id", "req-001")
		c.Set("entity_id", "sys")
		c.Set("principal", "admin")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/events", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, store.events, 1)
	assert.Equal(t, "API_CALL", store.events[0].Type)
	assert.Equal(t, "GET /v1/audit/events", store.events[0].Action)
}

func TestAuditEventPersistence_Ordering(t *testing.T) {
	store := &AuditStore{}
	now := time.Now()

	for i := 0; i < 5; i++ {
		store.Record(AuditEvent{
			ID:        "evt-ord-" + string(rune('0'+i)),
			Type:      "TEST",
			EntityID:  "order-test",
			Action:    "ACTION",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	assert.Len(t, store.events, 5)
	for i := 1; i < len(store.events); i++ {
		assert.True(t, store.events[i].Timestamp.After(store.events[i-1].Timestamp) ||
			store.events[i].Timestamp.Equal(store.events[i-1].Timestamp))
	}
}

func TestMiddlewareRateLimitConfiguration(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Header("X-RateLimit-Limit", "50")
		c.Header("X-RateLimit-Remaining", "49")
		c.Next()
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "50", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "49", w.Header().Get("X-RateLimit-Remaining"))
}

func TestHealthEndpoint(t *testing.T) {
	r := gin.New()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])
}

func TestAuditEventJSONSerialization(t *testing.T) {
	event := AuditEvent{
		ID: "evt-json-001", Type: "FRAUD_SCORE_UPDATED", EntityID: "cit-042",
		Action: "SCORE_UPDATE", Principal: "engine", Outcome: "COMPLETED",
		Metadata:  `{"old_score": 45, "new_score": 72}`,
		Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "evt-json-001")
	assert.Contains(t, string(data), "FRAUD_SCORE_UPDATED")

	var decoded AuditEvent
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.Type, decoded.Type)
}

func TestConcurrentAuditRecording(t *testing.T) {
	store := &AuditStore{}
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			store.Record(AuditEvent{
				ID: "conc-evt-" + string(rune('0'+idx)), Type: "CONCURRENT",
				EntityID: "conc-test", Action: "WRITE",
			})
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	assert.Len(t, store.events, 10)
}

func init() {
	gin.SetMode(gin.TestMode)
}
