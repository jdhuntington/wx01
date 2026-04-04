package notify

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// Hub listens for PostgreSQL NOTIFY events on the wx01_data channel
// and broadcasts them to connected SSE clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan string]struct{}),
	}
}

// Subscribe returns a channel that receives notification payloads.
func (h *Hub) Subscribe() chan string {
	ch := make(chan string, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe removes a client channel and closes it.
func (h *Hub) Unsubscribe(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

func (h *Hub) broadcast(payload string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- payload:
		default:
		}
	}
}

// Listen connects to PostgreSQL and listens for notifications on wx01_data.
// It reconnects automatically on failure. Blocks until ctx is cancelled.
func (h *Hub) Listen(ctx context.Context, connString string) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := h.listenOnce(ctx, connString); err != nil {
			log.Printf("notify listener error: %v (reconnecting in 2s)", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (h *Hub) listenOnce(ctx context.Context, connString string) error {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "LISTEN wx01_data"); err != nil {
		return err
	}

	log.Println("notify listener connected")

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		h.broadcast(notification.Payload)
	}
}
