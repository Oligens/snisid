package command

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type LiveEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Gateway struct {
	mu          sync.RWMutex
	subscribers map[chan LiveEvent]bool
}

func NewGateway() *Gateway {
	return &Gateway{
		subscribers: make(map[chan LiveEvent]bool),
	}
}

func (g *Gateway) Broadcast(event LiveEvent) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for ch := range g.subscribers {
		ch <- event
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ch := make(chan LiveEvent)
	g.mu.Lock()
	g.subscribers[ch] = true
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.subscribers, ch)
		g.mu.Unlock()
		close(ch)
	}()

	for event := range ch {
		if err := conn.WriteJSON(event); err != nil {
			break
		}
	}
}
