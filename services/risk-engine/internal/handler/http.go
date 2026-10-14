package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/snisid/platform/pkg/kafka"
	"nexus-snisid/pkg/eventbus"
)

type RiskRequest struct {
	TransactionID   string  `json:"transaction_id"`
	UserID          string  `json:"user_id"`
	Amount          float64 `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	SourceIP        string  `json:"source_ip"`
	DeviceID        string  `json:"device_id"`
	Location        string  `json:"location"`
	Timestamp       int64   `json:"timestamp"`
}

type RiskResponse struct {
	TransactionID string       `json:"transaction_id"`
	RiskScore     float64      `json:"risk_score"`
	Decision      string       `json:"decision"` // APPROVE, REVIEW, BLOCK
	Factors       []RiskFactor `json:"factors"`
	ComputedAt    int64        `json:"computed_at"`
}

type RiskFactor struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Weight float64 `json:"weight"`
	Reason string  `json:"reason"`
}

type RiskEngine struct {
	mu              sync.RWMutex
	userHistory     map[string][]RiskRequest
	ipBlacklist     map[string]bool
	deviceBlacklist map[string]bool
	thresholds      RiskThresholds
}

type RiskThresholds struct {
	BlockThreshold  float64 `json:"block_threshold"`
	ReviewThreshold float64 `json:"review_threshold"`
	VelocityLimit   int     `json:"velocity_limit"`
	VelocityWindow  int64   `json:"velocity_window_seconds"`
}

var (
	engine   *RiskEngine
	once     sync.Once
	producer = kafka.NewProducer("kafka:9092", "events.risk")
)

func getEngine() *RiskEngine {
	once.Do(func() {
		engine = &RiskEngine{
			userHistory:     make(map[string][]RiskRequest),
			ipBlacklist:     make(map[string]bool),
			deviceBlacklist: make(map[string]bool),
			thresholds: RiskThresholds{
				BlockThreshold:  0.85,
				ReviewThreshold: 0.6,
				VelocityLimit:   10,
				VelocityWindow:  300,
			},
		}
	})
	return engine
}

func RiskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	eng := getEngine()
	response := eng.evaluate(req)

	event := eventbus.Event{
		Type:   "RISK_CALCULATED",
		Source: "risk-engine",
		Payload: map[string]interface{}{
			"transaction_id": req.TransactionID,
			"score":          response.RiskScore,
			"decision":       response.Decision,
			"factors":        response.Factors,
		},
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(event)
	producer.Publish("risk", data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (e *RiskEngine) evaluate(req RiskRequest) RiskResponse {
	e.mu.Lock()
	e.userHistory[req.UserID] = append(e.userHistory[req.UserID], req)
	e.mu.Unlock()

	response := RiskResponse{
		TransactionID: req.TransactionID,
		ComputedAt:    time.Now().Unix(),
	}

	var factors []RiskFactor
	totalScore := 0.0
	totalWeight := 0.0

	amountFactor := e.evaluateAmount(req.Amount)
	factors = append(factors, amountFactor)
	totalScore += amountFactor.Score * amountFactor.Weight
	totalWeight += amountFactor.Weight

	velocityFactor := e.evaluateVelocity(req.UserID)
	factors = append(factors, velocityFactor)
	totalScore += velocityFactor.Score * velocityFactor.Weight
	totalWeight += velocityFactor.Weight

	blacklistFactor := e.evaluateBlacklists(req)
	factors = append(factors, blacklistFactor)
	totalScore += blacklistFactor.Score * blacklistFactor.Weight
	totalWeight += blacklistFactor.Weight

	timeFactor := e.evaluateTime()
	factors = append(factors, timeFactor)
	totalScore += timeFactor.Score * timeFactor.Weight
	totalWeight += timeFactor.Weight

	if totalWeight > 0 {
		response.RiskScore = math.Round((totalScore/totalWeight)*100) / 100
	} else {
		response.RiskScore = 0.0
	}

	switch {
	case response.RiskScore >= e.thresholds.BlockThreshold:
		response.Decision = "BLOCK"
	case response.RiskScore >= e.thresholds.ReviewThreshold:
		response.Decision = "REVIEW"
	default:
		response.Decision = "APPROVE"
	}

	response.Factors = factors
	return response
}

func (e *RiskEngine) evaluateAmount(amount float64) RiskFactor {
	score := 0.0
	var reason string

	switch {
	case amount > 50000:
		score = 1.0
		reason = "amount exceeds 50,000"
	case amount > 10000:
		score = 0.7
		reason = "amount exceeds 10,000"
	case amount > 5000:
		score = 0.4
		reason = "amount exceeds 5,000"
	case amount > 1000:
		score = 0.2
		reason = "amount exceeds 1,000"
	default:
		score = 0.05
		reason = "normal amount"
	}

	return RiskFactor{
		Name:   "amount",
		Score:  score,
		Weight: 0.35,
		Reason: reason,
	}
}

func (e *RiskEngine) evaluateVelocity(userID string) RiskFactor {
	e.mu.RLock()
	history, ok := e.userHistory[userID]
	e.mu.RUnlock()

	if !ok || len(history) < 2 {
		return RiskFactor{
			Name:   "velocity",
			Score:  0.1,
			Weight: 0.25,
			Reason: "insufficient history",
		}
	}

	window := time.Now().Unix() - e.thresholds.VelocityWindow
	recentCount := 0
	for _, h := range history {
		if h.Timestamp >= window {
			recentCount++
		}
	}

	score := float64(recentCount) / float64(e.thresholds.VelocityLimit)
	if score > 1.0 {
		score = 1.0
	}

	return RiskFactor{
		Name:   "velocity",
		Score:  score,
		Weight: 0.25,
		Reason: formatVelocityReason(recentCount, e.thresholds.VelocityLimit),
	}
}

func (e *RiskEngine) evaluateBlacklists(req RiskRequest) RiskFactor {
	e.mu.RLock()
	_, ipBlocked := e.ipBlacklist[req.SourceIP]
	_, deviceBlocked := e.deviceBlacklist[req.DeviceID]
	e.mu.RUnlock()

	if ipBlocked || deviceBlocked {
		return RiskFactor{
			Name:   "blacklist",
			Score:  1.0,
			Weight: 0.30,
			Reason: "source is blacklisted",
		}
	}

	return RiskFactor{
		Name:   "blacklist",
		Score:  0.0,
		Weight: 0.30,
		Reason: "source not blacklisted",
	}
}

func (e *RiskEngine) evaluateTime() RiskFactor {
	hour := time.Now().Hour()
	if hour >= 1 && hour <= 5 {
		return RiskFactor{
			Name:   "time",
			Score:  0.5,
			Weight: 0.10,
			Reason: "off-hours transaction",
		}
	}
	return RiskFactor{
		Name:   "time",
		Score:  0.05,
		Weight: 0.10,
		Reason: "normal hours",
	}
}

func (e *RiskEngine) AddToBlacklist(ip, deviceID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ip != "" {
		e.ipBlacklist[ip] = true
	}
	if deviceID != "" {
		e.deviceBlacklist[deviceID] = true
	}
}

func formatVelocityReason(count, limit int) string {
	if count >= limit {
		return "velocity limit exceeded"
	}
	return formatInt(count) + " transactions in window"
}

func formatInt(v int) string {
	return fmt.Sprintf("%d", v)
}
