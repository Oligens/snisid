package biometrics

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type InferenceEngine interface {
	GenerateEmbedding(ctx context.Context, imageData []byte, bType BiometricType) ([]float32, error)
}

type ONNXInferenceEngine struct {
	modelPath string
	embedDim  int
	endpoint  string
	client    *http.Client
}

func NewONNXInferenceEngine(modelPath string) (*ONNXInferenceEngine, error) {
	endpoint := os.Getenv("ONNX_SERVICE_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://biometrics-service.snisid.svc.cluster.local:8080"
	}
	return &ONNXInferenceEngine{
		modelPath: modelPath,
		embedDim:  512,
		endpoint:  endpoint,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (e *ONNXInferenceEngine) GenerateEmbedding(ctx context.Context, imageData []byte, bType BiometricType) ([]float32, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	embedding, err := e.callONNXService(ctx, imageData)
	if err != nil {
		return nil, fmt.Errorf("onnx inference: %w", err)
	}

	return l2Normalize(embedding), nil
}

func (e *ONNXInferenceEngine) callONNXService(ctx context.Context, imageData []byte) ([]float32, error) {
	reqBody := map[string]interface{}{
		"image": imageData,
		"model": e.modelPath,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint+"/v1/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("onnx service call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("onnx service error: %s", string(respBody))
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result.Embedding, nil
}

func l2Normalize(v []float32) []float32 {
	var norm float32
	for _, x := range v {
		norm += x * x
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm == 0 {
		return v
	}
	result := make([]float32, len(v))
	for i, x := range v {
		result[i] = x / norm
	}
	return result
}

type TFServingInferenceEngine struct {
	endpoint string
	client   *http.Client
	timeout  time.Duration
}

func NewTFServingEngine(endpoint string, timeout time.Duration) *TFServingInferenceEngine {
	if endpoint == "" {
		endpoint = os.Getenv("TF_SERVING_ENDPOINT")
		if endpoint == "" {
			endpoint = "http://tf-serving.snisid.svc.cluster.local:8501"
		}
	}
	return &TFServingInferenceEngine{
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
		timeout:  timeout,
	}
}

func (e *TFServingInferenceEngine) GenerateEmbedding(ctx context.Context, imageData []byte, bType BiometricType) ([]float32, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	reqBody := map[string]interface{}{
		"instances": []map[string]interface{}{
			{"image_bytes": imageData},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint+"/v1/models/arcface:predict", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tf serving call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tf serving error: %s", string(respBody))
	}

	var result struct {
		Predictions [][]float32 `json:"predictions"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Predictions) == 0 {
		return nil, fmt.Errorf("no predictions returned")
	}

	return l2Normalize(result.Predictions[0]), nil
}

type GRPCInferenceEngine struct {
	conn     *grpc.ClientConn
	endpoint string
	timeout  time.Duration
}

func NewGRPCInferenceEngine(endpoint string, timeout time.Duration) (*GRPCInferenceEngine, error) {
	if endpoint == "" {
		endpoint = os.Getenv("GRPC_INFERENCE_ENDPOINT")
		if endpoint == "" {
			endpoint = "biometrics-service.snisid.svc.cluster.local:8443"
		}
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ServerName: "biometrics-service.snisid.svc.cluster.local",
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.Dial(endpoint,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}
	return &GRPCInferenceEngine{
		conn:     conn,
		endpoint: endpoint,
		timeout:  timeout,
	}, nil
}

func preprocessImage(data []byte, height, width int) ([]byte, error) {
	expected := height * width * 3
	result := make([]byte, expected)
	copy(result, data)
	return result, nil
}

func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
