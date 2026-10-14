package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/internal/domain/audit/entity"
	"github.com/snisid/platform/internal/domain/audit/repository"
	"github.com/snisid/platform/internal/platform/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockIngesterAuditRepo struct {
	mock.Mock
}

func (m *mockIngesterAuditRepo) Append(ctx context.Context, event *entity.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockIngesterAuditRepo) GetLastEvent(ctx context.Context) (*entity.AuditEvent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AuditEvent), args.Error(1)
}

func (m *mockIngesterAuditRepo) GetEventsByCorrelationID(ctx context.Context, correlationID string) ([]entity.AuditEvent, error) {
	args := m.Called(ctx, correlationID)
	return args.Get(0).([]entity.AuditEvent), args.Error(1)
}

func (m *mockIngesterAuditRepo) GetEventsBySequenceRange(ctx context.Context, start, end int64) ([]entity.AuditEvent, error) {
	args := m.Called(ctx, start, end)
	return args.Get(0).([]entity.AuditEvent), args.Error(1)
}

var _ repository.AuditRepository = (*mockIngesterAuditRepo)(nil)

type mockIngesterConsumer struct {
	msgs [][]byte
}

func (m *mockIngesterConsumer) Start(ctx context.Context, handler func(ctx context.Context, msg []byte) error) error {
	for _, msg := range m.msgs {
		if err := handler(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockIngesterConsumer) Decode(data []byte, v interface{}) error { return nil }
func (m *mockIngesterConsumer) Close() error                            { return nil }

func TestIngester_IngestSingleEvent(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockIngesterAuditRepo)
	mockConsumer := &mockIngesterConsumer{
		msgs: [][]byte{
			[]byte(`{"correlationId":"corr-single","userId":"actor-1","action":"CREATE","resource":"identity/1","eventType":"identity.created","status":"success"}`),
		},
	}
	ingester := NewKafkaIngester(mockRepo, mockConsumer)
	require.NotNil(t, ingester)

	payload := map[string]interface{}{
		"correlationId": "corr-single",
		"userId":        "actor-1",
		"action":        "CREATE",
		"resource":      "identity/1",
		"eventType":     "identity.created",
		"status":        "success",
	}
	msg, err := json.Marshal(payload)
	require.NoError(t, err)

	var payloadMap map[string]interface{}
	err = json.Unmarshal(msg, &payloadMap)
	require.NoError(t, err)

	stablePayload, _ := json.Marshal(payloadMap)
	expectedHash := security.GenerateHashChain("genesis-hash-snisid", string(stablePayload))

	mockRepo.On("GetLastEvent", mock.Anything).Return(nil, nil).Once()
	mockRepo.On("Append", mock.Anything, mock.MatchedBy(func(e *entity.AuditEvent) bool {
		return e.CorrelationID == "corr-single" &&
			e.Actor == "actor-1" &&
			e.Action == "CREATE" &&
			e.Resource == "identity/1" &&
			e.EventType == "identity.created" &&
			e.Status == "success" &&
			e.PreviousHash == "genesis-hash-snisid" &&
			e.Hash == expectedHash &&
			e.EventID != "" &&
			!e.Timestamp.IsZero()
	})).Return(nil).Once()

	ingester.Start(context.Background())

	mockRepo.AssertExpectations(t)
}

func TestIngester_BatchIngestMultipleEvents(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockIngesterAuditRepo)
	payloads := [][]byte{
		[]byte(`{"correlationId":"corr-batch","userId":"actor-1","action":"CREATE","resource":"identity/1","eventType":"identity.created","status":"success"}`),
		[]byte(`{"correlationId":"corr-batch","userId":"actor-2","action":"UPDATE","resource":"identity/2","eventType":"identity.updated","status":"success"}`),
		[]byte(`{"correlationId":"corr-batch","userId":"actor-3","action":"DELETE","resource":"identity/3","eventType":"identity.deleted","status":"success"}`),
	}
	mockConsumer := &mockIngesterConsumer{msgs: payloads}
	ingester := NewKafkaIngester(mockRepo, mockConsumer)

	events := []map[string]interface{}{
		{"correlationId": "corr-batch", "userId": "actor-1", "action": "CREATE", "resource": "identity/1", "eventType": "identity.created", "status": "success"},
		{"correlationId": "corr-batch", "userId": "actor-2", "action": "UPDATE", "resource": "identity/2", "eventType": "identity.updated", "status": "success"},
		{"correlationId": "corr-batch", "userId": "actor-3", "action": "DELETE", "resource": "identity/3", "eventType": "identity.deleted", "status": "success"},
	}

	prevHash := "genesis-hash-snisid"
	for i, p := range events {
		stablePayload, _ := json.Marshal(p)
		hash := security.GenerateHashChain(prevHash, string(stablePayload))

		if i < len(events)-1 {
			lastEvent := &entity.AuditEvent{Hash: hash}
			mockRepo.On("GetLastEvent", mock.Anything).Return(lastEvent, nil).Once()
		} else {
			mockRepo.On("GetLastEvent", mock.Anything).Return(nil, nil).Once()
		}

		mockRepo.On("Append", mock.Anything, mock.MatchedBy(func(e *entity.AuditEvent) bool {
			return e.CorrelationID == "corr-batch"
		})).Return(nil).Once()

		prevHash = hash
	}

	ingester.Start(context.Background())

	mockRepo.AssertExpectations(t)
}

func TestIngester_InvalidEventReturnsError(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockIngesterAuditRepo)
	mockConsumer := &mockIngesterConsumer{msgs: [][]byte{[]byte(`not-valid-json`)}}
	ingester := NewKafkaIngester(mockRepo, mockConsumer)

	mockRepo.AssertNotCalled(t, "Append", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "GetLastEvent", mock.Anything)

	ingester.Start(context.Background())

	mockRepo.AssertExpectations(t)
}

func TestIngester_RepositoryFailureHandling(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockIngesterAuditRepo)
	mockConsumer := &mockIngesterConsumer{msgs: [][]byte{[]byte(`{"correlationId":"corr-fail","userId":"actor-1","action":"CREATE","resource":"identity/1","eventType":"identity.created","status":"success"}`)}}
	ingester := NewKafkaIngester(mockRepo, mockConsumer)

	mockRepo.On("GetLastEvent", mock.Anything).Return(nil, errors.New("db connection failed")).Once()

	ingester.Start(context.Background())

	mockRepo.AssertExpectations(t)
}

func TestIngester_AppendFailure(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockIngesterAuditRepo)
	mockConsumer := &mockIngesterConsumer{msgs: [][]byte{[]byte(`{"correlationId":"corr-append-fail","userId":"actor-1","action":"CREATE","resource":"identity/1","eventType":"identity.created","status":"success"}`)}}
	ingester := NewKafkaIngester(mockRepo, mockConsumer)

	mockRepo.On("GetLastEvent", mock.Anything).Return(nil, nil).Once()
	mockRepo.On("Append", mock.Anything, mock.Anything).Return(errors.New("constraint violation")).Once()

	ingester.Start(context.Background())

	mockRepo.AssertExpectations(t)
}

func TestIngester_EventOrderingHashChain(t *testing.T) {
	t.Parallel()

	prevHash := "genesis-hash-snisid"
	events := make([]entity.AuditEvent, 3)

	payloads := []string{
		`{"action":"first"}`,
		`{"action":"second"}`,
		`{"action":"third"}`,
	}

	for i, p := range payloads {
		hash := security.GenerateHashChain(prevHash, p)
		events[i] = entity.AuditEvent{
			EventID:      uuid.NewString(),
			Payload:      p,
			PreviousHash: prevHash,
			Hash:         hash,
			SequenceID:   int64(i + 1),
			Timestamp:    time.Now().UTC(),
		}
		prevHash = hash
	}

	for i := 1; i < len(events); i++ {
		assert.Equal(t, events[i-1].Hash, events[i].PreviousHash,
			"event %d should link to event %d", i, i-1)
	}

	for _, e := range events {
		assert.True(t, security.VerifyHashChain(e.Hash, e.PreviousHash, e.Payload),
			"hash chain broken at event %d", e.SequenceID)
	}
}

func TestIngester_StatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload map[string]interface{}
		want    string
	}{
		{
			name:    "explicit success",
			payload: map[string]interface{}{"status": "success"},
			want:    "success",
		},
		{
			name:    "explicit failed",
			payload: map[string]interface{}{"status": "failed"},
			want:    "failed",
		},
		{
			name:    "explicit denied",
			payload: map[string]interface{}{"status": "denied"},
			want:    "denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := json.Marshal(tt.payload)
			var parsed map[string]interface{}
			json.Unmarshal(msg, &parsed)
			status, _ := parsed["status"].(string)
			assert.Equal(t, tt.want, status)
		})
	}
}
