package biometrics

import (
	"context"
	"errors"
	"testing"
)

type mockMilvus struct {
	insertFn func(ctx context.Context, collection, identityID string, vector []float32) error
	searchFn func(ctx context.Context, collection string, vector []float32) (string, float32, error)
}

func (m *mockMilvus) InsertBiometric(ctx context.Context, collection, identityID string, vector []float32) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, collection, identityID, vector)
	}
	return nil
}

func (m *mockMilvus) Search(ctx context.Context, collection string, vector []float32) (string, float32, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, collection, vector)
	}
	return "", 0, nil
}

type mockInference struct {
	generateFn func(ctx context.Context, rawData []byte, bType BiometricType) ([]float32, error)
}

func (m *mockInference) GenerateEmbedding(ctx context.Context, rawData []byte, bType BiometricType) ([]float32, error) {
	if m.generateFn != nil {
		return m.generateFn(ctx, rawData, bType)
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func newTestService(milvus *mockMilvus, inference *mockInference) *BiometricService {
	return NewBiometricService(milvus, inference)
}

func TestEnroll_Success(t *testing.T) {
	var capturedCollection, capturedID string

	milvus := &mockMilvus{
		insertFn: func(ctx context.Context, collection, identityID string, vector []float32) error {
			capturedCollection = collection
			capturedID = identityID
			return nil
		},
	}
	inference := &mockInference{
		generateFn: func(ctx context.Context, rawData []byte, bType BiometricType) ([]float32, error) {
			return []float32{0.5, 0.3, 0.8}, nil
		},
	}

	svc := newTestService(milvus, inference)
	err := svc.Enroll(context.Background(), "ID-123", []byte("face-data"), "face")
	if err != nil {
		t.Fatalf("Enroll failed: %v", err)
	}

	if capturedCollection != "snisid_biometrics_face" {
		t.Errorf("collection = %s, want snisid_biometrics_face", capturedCollection)
	}
	if capturedID != "ID-123" {
		t.Errorf("identityID = %s, want ID-123", capturedID)
	}
}

func TestEnroll_MilvusError(t *testing.T) {
	milvus := &mockMilvus{
		insertFn: func(ctx context.Context, collection, identityID string, vector []float32) error {
			return errors.New("milvus connection failed")
		},
	}
	inference := &mockInference{}
	svc := newTestService(milvus, inference)

	err := svc.Enroll(context.Background(), "ID-123", []byte("data"), "fingerprint")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestEnroll_InferenceError(t *testing.T) {
	milvus := &mockMilvus{}
	inference := &mockInference{
		generateFn: func(ctx context.Context, rawData []byte, bType BiometricType) ([]float32, error) {
			return nil, errors.New("inference failed")
		},
	}
	svc := newTestService(milvus, inference)

	err := svc.Enroll(context.Background(), "ID-123", []byte("data"), "face")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestVerify_Success(t *testing.T) {
	milvus := &mockMilvus{
		searchFn: func(ctx context.Context, collection string, vector []float32) (string, float32, error) {
			return "ID-123", 0.15, nil
		},
	}
	inference := &mockInference{
		generateFn: func(ctx context.Context, rawData []byte, bType BiometricType) ([]float32, error) {
			return []float32{0.4, 0.6}, nil
		},
	}

	svc := newTestService(milvus, inference)
	matchID, confidence, err := svc.Verify(context.Background(), []byte("face-data"), "face")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if matchID != "ID-123" {
		t.Errorf("matchID = %s, want ID-123", matchID)
	}
	if confidence <= 0 || confidence > 100 {
		t.Errorf("confidence = %f, want in (0, 100]", confidence)
	}
}

func TestVerify_NoMatch(t *testing.T) {
	milvus := &mockMilvus{
		searchFn: func(ctx context.Context, collection string, vector []float32) (string, float32, error) {
			return "", 5.0, nil
		},
	}
	inference := &mockInference{}
	svc := newTestService(milvus, inference)

	matchID, confidence, err := svc.Verify(context.Background(), []byte("unknown"), "face")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if matchID != "" {
		t.Errorf("matchID = %s, want empty for no match", matchID)
	}
	// distance=5.0 → confidence = 100 - (5.0 * 10.0) = 50
	if confidence != 50 {
		t.Errorf("confidence = %f, want 50", confidence)
	}
}

func TestVerify_MilvusError(t *testing.T) {
	milvus := &mockMilvus{
		searchFn: func(ctx context.Context, collection string, vector []float32) (string, float32, error) {
			return "", 0, errors.New("search failed")
		},
	}
	inference := &mockInference{}
	svc := newTestService(milvus, inference)

	_, _, err := svc.Verify(context.Background(), []byte("data"), "face")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestVerify_ConfidenceNormalization(t *testing.T) {
	tests := []struct {
		distance float32
		wantMin  float32
		wantMax  float32
	}{
		{0.0, 90.0, 100.0},
		{5.0, 50.0, 50.0},
		{2.5, 75.0, 75.0},
	}

	for _, tt := range tests {
		milvus := &mockMilvus{
			searchFn: func(ctx context.Context, collection string, vector []float32) (string, float32, error) {
				return "match", tt.distance, nil
			},
		}
		inference := &mockInference{}
		svc := newTestService(milvus, inference)

		_, confidence, err := svc.Verify(context.Background(), []byte("data"), "face")
		if err != nil {
			t.Fatalf("Verify failed: %v", err)
		}
		if confidence < tt.wantMin || confidence > tt.wantMax {
			t.Errorf("distance=%f: confidence=%f, want in [%f, %f]",
				tt.distance, confidence, tt.wantMin, tt.wantMax)
		}
	}
}

func TestNewBiometricService(t *testing.T) {
	svc := NewBiometricService(nil, nil)
	if svc == nil {
		t.Fatal("NewBiometricService returned nil")
	}
}
