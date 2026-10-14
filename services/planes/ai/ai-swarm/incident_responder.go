package aiswarm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/snisid/platform/backend/shared/events"
	"github.com/snisid/platform/backend/shared/logger"
)

type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "LOW"
	SeverityMedium   IncidentSeverity = "MEDIUM"
	SeverityHigh     IncidentSeverity = "HIGH"
	SeverityCritical IncidentSeverity = "CRITICAL"
)

type AgentTask struct {
	ID        string          `json:"id"`
	AgentType string          `json:"agentType"`
	Command   string          `json:"command"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

type IncidentPayload struct {
	IncidentID   string           `json:"incident_id"`
	Severity     IncidentSeverity `json:"severity"`
	TargetType   string           `json:"target_type"`
	TargetID     string           `json:"target_id"`
	Reason       string           `json:"reason"`
	SourceSystem string           `json:"source_system"`
}

type PlaybookAction struct {
	Action       string `json:"action"`
	Target       string `json:"target"`
	Duration     string `json:"duration,omitempty"`
	RequireHuman bool   `json:"require_human"`
}

func RunAgent() {
	broker := getEnv("KAFKA_BROKER", "localhost:9092")

	consumer := events.NewConsumer([]string{broker}, "swarm-ir-group", "swarm.tasks")
	producer := events.NewProducer([]string{broker}, "swarm.responses")

	logger.Info("Incident Responder Agent started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Incident Responder shutting down")
		cancel()
	}()

	err := consumer.Read(ctx, func(msg kafka.Message) error {
		var task AgentTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			logger.Error("IR: invalid task message", "error", err)
			return nil
		}

		if task.AgentType != "INCIDENT_RESPONDER" {
			return nil
		}

		var payload IncidentPayload
		if task.Payload != nil {
			json.Unmarshal(task.Payload, &payload)
		}

		playbook := selectPlaybook(payload.Severity, task.Command)
		results := executePlaybook(playbook, payload)

		response := map[string]interface{}{
			"task_id":      task.ID,
			"status":       "completed",
			"actions":      results,
			"responded_at": time.Now().Unix(),
		}

		respBytes, _ := json.Marshal(response)
		if err := producer.Publish(ctx, kafka.Message{
			Key:   []byte(task.ID),
			Value: respBytes,
		}); err != nil {
			logger.Error("IR: failed to publish response", "error", err)
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		logger.Fatal("IR: consumer error", err)
	}
}

func selectPlaybook(severity IncidentSeverity, command string) []PlaybookAction {
	switch severity {
	case SeverityCritical:
		return []PlaybookAction{
			{Action: "ISOLATE", Target: "identity", Duration: "24h", RequireHuman: false},
			{Action: "REVOKE_TOKENS", Target: "all", RequireHuman: false},
			{Action: "ALERT_SOC", Target: "human_operator", RequireHuman: false},
		}
	case SeverityHigh:
		return []PlaybookAction{
			{Action: "RESTRICT", Target: "identity", Duration: "4h", RequireHuman: false},
			{Action: "FLAG_REVIEW", Target: "identity", RequireHuman: true},
		}
	case SeverityMedium:
		return []PlaybookAction{
			{Action: "MONITOR", Target: "identity", Duration: "24h", RequireHuman: false},
		}
	default:
		return []PlaybookAction{
			{Action: "LOG_ONLY", Target: "identity", RequireHuman: false},
		}
	}
}

func executePlaybook(playbook []PlaybookAction, payload IncidentPayload) []map[string]interface{} {
	var results []map[string]interface{}

	for _, action := range playbook {
		result := map[string]interface{}{
			"action":   action.Action,
			"target":   action.Target,
			"status":   "executed",
			"duration": action.Duration,
		}

		logger.Info("IR: executing playbook action",
			"action", action.Action,
			"target", action.Target,
			"incident", payload.IncidentID,
		)

		switch action.Action {
		case "ISOLATE":
			logger.Warn("IR: isolating identity", "id", payload.TargetID)
		case "REVOKE_TOKENS":
			logger.Warn("IR: revoking all tokens for", "id", payload.TargetID)
		case "RESTRICT":
			logger.Warn("IR: restricting access for", "id", payload.TargetID)
		case "ALERT_SOC":
			logger.Warn("IR: SOC alert triggered",
				"incident", payload.IncidentID,
				"severity", payload.Severity,
			)
		}

		results = append(results, result)
	}

	return results
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func init() {
	fmt.Println("SNISID Incident Responder Agent v1.0")
}
