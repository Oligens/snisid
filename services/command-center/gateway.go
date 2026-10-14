package commandcenter

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/snisid/platform/internal/platform/logger"
	"sync"
)

type LiveEvent struct {
	Type     string      `json:"type"`
	Severity string      `json:"severity"`
	Source   string      `json:"source"`
	Payload  interface{} `json:"payload"`
}

var (
	subscribers = make(map[chan LiveEvent]bool)
	subMu       sync.Mutex
)

func Broadcast(event LiveEvent) {
	subMu.Lock()
	defer subMu.Unlock()
	for ch := range subscribers {
		select {
		case ch <- event:
		default:
			// Buffer full, skip
		}
	}
}

func Stream(ws *websocket.Conn) {
	logger.Info(context.Background(), "COMMAND-CENTER: New national supervisor connected via WebSocket.")
	ch := make(chan LiveEvent, 100)

	subMu.Lock()
	subscribers[ch] = true
	subMu.Unlock()

	defer func() {
		subMu.Lock()
		delete(subscribers, ch)
		subMu.Unlock()
		ws.Close()
	}()

	for event := range ch {
		if err := ws.WriteJSON(event); err != nil {
			logger.Error(context.Background(), "COMMAND-CENTER: WebSocket write failed", err)
			break
		}
	}
}

func TriggerAlert(alertType string, citizenID string) {
	Broadcast(LiveEvent{
		Type:     "CRITICAL_ALERT",
		Severity: "HIGH",
		Source:   "FRAUD_ENGINE",
		Payload: map[string]string{
			"type":       alertType,
			"citizen_id": citizenID,
		},
	})
}
