package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type EventType string

const (
	EventIdentity   EventType = "IDENTITY"
	EventFraud      EventType = "FRAUD"
	EventBiometrics EventType = "BIOMETRICS"
	EventFPR        EventType = "FPR"
	EventVehicle    EventType = "VEHICLE"
	EventAlert      EventType = "ALERT"
)

type FederatedEvent struct {
	ID            string    `json:"id"`
	SourceCountry string    `json:"source_country"`
	TargetCountry string    `json:"target_country"`
	EventType     EventType `json:"event_type"`
	Payload       []byte    `json:"payload"`
	Signature     string    `json:"signature"`
	Timestamp     int64     `json:"timestamp"`
	CorrelationID string    `json:"correlation_id"`
	RetryCount    int       `json:"retry_count"`
}

type OutboundQueue struct {
	Events []QueuedEvent `json:"events"`
	mu     sync.Mutex
}

type QueuedEvent struct {
	Event     FederatedEvent `json:"event"`
	QueuedAt  time.Time      `json:"queued_at"`
	NextRetry time.Time      `json:"next_retry"`
}

type Ack struct {
	EventID     string `json:"event_id"`
	Received    bool   `json:"received"`
	Processed   bool   `json:"processed"`
	Error       string `json:"error,omitempty"`
	ProcessedAt int64  `json:"processed_at"`
}

type FederationGateway struct {
	GatewayID      string
	sharedSecret   []byte
	privateKeyPath string
	certPath       string
	peers          map[string]string // country -> endpoint URL
	outbox         *OutboundQueue
	client         *http.Client
	maxRetries     int
	baseBackoff    time.Duration
}

func NewFederationGateway(gatewayID string, sharedSecret string, peers map[string]string) *FederationGateway {
	return &FederationGateway{
		GatewayID:    gatewayID,
		sharedSecret: []byte(sharedSecret),
		peers:        peers,
		outbox:       &OutboundQueue{},
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       20,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		},
		maxRetries:  5,
		baseBackoff: 2 * time.Second,
	}
}

func (g *FederationGateway) Exchange(event FederatedEvent) (*Ack, error) {
	ctx := context.Background()
	logger.Info(ctx, "FEDERATION: processing event",
		zap.String("id", event.ID),
		zap.String("source", event.SourceCountry),
		zap.String("target", event.TargetCountry),
		zap.String("type", string(event.EventType)),
	)

	normalized, err := g.normalize(event.Payload, event.EventType)
	if err != nil {
		return nil, fmt.Errorf("normalization failed: %w", err)
	}
	event.Payload = normalized

	signature, err := g.sign(normalized)
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}
	event.Signature = signature

	ack, err := g.transfer(event.TargetCountry, event)
	if err != nil {
		g.enqueue(event)
		return nil, fmt.Errorf("transfer failed, queued: %w", err)
	}

	return ack, nil
}

func (g *FederationGateway) normalize(data []byte, eventType EventType) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %w", err)
	}

	normalized := make(map[string]interface{})
	normalized["snisid_version"] = "1.0"
	normalized["event_type"] = string(eventType)
	normalized["normalized_at"] = time.Now().Unix()

	switch eventType {
	case EventIdentity:
		g.normalizeIdentity(raw, normalized)
	case EventBiometrics:
		g.normalizeBiometrics(raw, normalized)
	case EventFraud:
		normalized["data"] = raw
	case EventFPR:
		g.normalizeFPR(raw, normalized)
	default:
		normalized["data"] = raw
	}

	return json.Marshal(normalized)
}

func (g *FederationGateway) normalizeIdentity(raw, out map[string]interface{}) {
	fieldMap := map[string]string{
		"last_name":     "family_name",
		"first_name":    "given_name",
		"date_of_birth": "birth_date",
		"national_id":   "national_id",
	}
	for src, dst := range fieldMap {
		if v, ok := raw[src]; ok {
			out[dst] = v
		}
	}
	if nid, ok := raw["national_id"].(string); ok {
		out["country_code"] = strings.ToUpper(nid[:2])
	}
}

func (g *FederationGateway) normalizeBiometrics(raw, out map[string]interface{}) {
	if f, ok := raw["face_template"]; ok {
		out["face_hash"] = g.hashData(fmt.Sprintf("%v", f))
	}
	if fp, ok := raw["fingerprint"]; ok {
		out["fingerprint_hash"] = g.hashData(fmt.Sprintf("%v", fp))
	}
}

func (g *FederationGateway) normalizeFPR(raw, out map[string]interface{}) {
	for _, k := range []string{"category", "last_name", "first_name", "case_reference", "alert_level"} {
		if v, ok := raw[k]; ok {
			out[k] = v
		}
	}
}

func (g *FederationGateway) sign(data []byte) (string, error) {
	mac := hmac.New(sha256.New, g.sharedSecret)
	mac.Write(data)
	mac.Write([]byte(time.Now().UTC().Format(time.RFC3339)))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("SNISID-HMAC-SHA256:%s", sig), nil
}

func (g *FederationGateway) VerifySignature(data []byte, signature string) bool {
	parts := strings.SplitN(signature, ":", 2)
	if len(parts) != 2 || parts[0] != "SNISID-HMAC-SHA256" {
		return false
	}
	mac := hmac.New(sha256.New, g.sharedSecret)
	mac.Write(data)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}

func (g *FederationGateway) transfer(targetCountry string, event FederatedEvent) (*Ack, error) {
	ctx := context.Background()
	endpoint, ok := g.peers[targetCountry]
	if !ok {
		return nil, fmt.Errorf("unknown peer: %s", targetCountry)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint+"/api/v1/federation/events", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SNISID-Gateway", g.GatewayID)
	req.Header.Set("X-SNISID-Timestamp", fmt.Sprintf("%d", event.Timestamp))

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer returned %d: %s", resp.StatusCode, string(respBody))
	}

	var ack Ack
	if err := json.Unmarshal(respBody, &ack); err != nil {
		return nil, fmt.Errorf("invalid ack response: %w", err)
	}

	ack.ProcessedAt = time.Now().Unix()
	logger.Info(ctx, "FEDERATION: event transferred successfully",
		zap.String("target", targetCountry),
		zap.String("event_id", event.ID),
		zap.Bool("ack_processed", ack.Processed),
	)

	return &ack, nil
}

func (g *FederationGateway) enqueue(event FederatedEvent) {
	ctx := context.Background()
	g.outbox.mu.Lock()
	defer g.outbox.mu.Unlock()

	event.RetryCount = 0
	g.outbox.Events = append(g.outbox.Events, QueuedEvent{
		Event:     event,
		QueuedAt:  time.Now(),
		NextRetry: time.Now().Add(g.baseBackoff),
	})

	logger.Warn(ctx, "FEDERATION: event queued for retry",
		zap.String("event_id", event.ID),
		zap.String("target", event.TargetCountry),
	)
}

func (g *FederationGateway) ProcessOutbox() {
	ctx := context.Background()
	g.outbox.mu.Lock()
	var remaining []QueuedEvent
	now := time.Now()

	for _, qe := range g.outbox.Events {
		if qe.Event.RetryCount >= g.maxRetries {
			logger.Error(ctx, "FEDERATION: event dropped after max retries",
				fmt.Errorf("max retries exceeded"),
				zap.String("event_id", qe.Event.ID),
				zap.String("target", qe.Event.TargetCountry),
			)
			continue
		}

		if now.Before(qe.NextRetry) {
			remaining = append(remaining, qe)
			continue
		}

		ack, err := g.transfer(qe.Event.TargetCountry, qe.Event)
		if err != nil {
			qe.Event.RetryCount++
			backoff := g.baseBackoff * (1 << qe.Event.RetryCount)
			qe.NextRetry = time.Now().Add(backoff)
			remaining = append(remaining, qe)
			logger.Warn(ctx, "FEDERATION: retry scheduled",
				zap.String("event_id", qe.Event.ID),
				zap.Int("retry", qe.Event.RetryCount),
				zap.String("backoff", backoff.String()),
			)
		} else {
			logger.Info(ctx, "FEDERATION: queued event delivered",
				zap.String("event_id", qe.Event.ID),
				zap.String("target", qe.Event.TargetCountry),
				zap.Bool("ack", ack.Processed),
			)
		}
	}

	g.outbox.Events = remaining
	g.outbox.mu.Unlock()
}

func (g *FederationGateway) hashData(data string) string {
	h := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(h[:])
}

func (g *FederationGateway) HandleIncoming(event FederatedEvent) (*Ack, error) {
	ctx := context.Background()
	if !g.VerifySignature(event.Payload, event.Signature) {
		return &Ack{
			EventID:   event.ID,
			Received:  true,
			Processed: false,
			Error:     "invalid signature",
		}, fmt.Errorf("invalid signature for event %s", event.ID)
	}

	logger.Info(ctx, "FEDERATION: incoming event verified and accepted",
		zap.String("source", event.SourceCountry),
		zap.String("type", string(event.EventType)),
		zap.String("id", event.ID),
	)

	return &Ack{
		EventID:   event.ID,
		Received:  true,
		Processed: true,
	}, nil
}
