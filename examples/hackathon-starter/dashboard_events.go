package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ============================================================================
// DASHBOARD EVENTS - Server-Sent Events for Real-Time Updates
// ============================================================================

// DashboardEvent represents an event to broadcast to clients
type DashboardEvent struct {
	Type      string `json:"type"`   // "budget", "savings_goal", "subscription", "transaction"
	Action    string `json:"action"` // "created", "updated", "deleted"
	Timestamp int64  `json:"timestamp"`
}

// SSEBroadcaster manages SSE client connections
type SSEBroadcaster struct {
	clients    map[chan DashboardEvent]bool
	register   chan chan DashboardEvent
	unregister chan chan DashboardEvent
	broadcast  chan DashboardEvent
	mu         sync.RWMutex
}

// Global broadcaster instance
var dashboardBroadcaster = NewSSEBroadcaster()

// NewSSEBroadcaster creates a new broadcaster
func NewSSEBroadcaster() *SSEBroadcaster {
	b := &SSEBroadcaster{
		clients:    make(map[chan DashboardEvent]bool),
		register:   make(chan chan DashboardEvent),
		unregister: make(chan chan DashboardEvent),
		broadcast:  make(chan DashboardEvent, 100),
	}
	go b.run()
	return b
}

// run handles client registration and broadcasting
func (b *SSEBroadcaster) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()
			log.Printf("[SSE] Client connected. Total: %d", len(b.clients))

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.mu.Unlock()
			log.Printf("[SSE] Client disconnected. Total: %d", len(b.clients))

		case event := <-b.broadcast:
			b.mu.RLock()
			for client := range b.clients {
				select {
				case client <- event:
				default:
					// Client buffer full, skip
				}
			}
			b.mu.RUnlock()
		}
	}
}

// NotifyDashboardUpdate broadcasts an update event to all connected clients
func NotifyDashboardUpdate(eventType, action string) {
	event := DashboardEvent{
		Type:      eventType,
		Action:    action,
		Timestamp: time.Now().Unix(),
	}

	select {
	case dashboardBroadcaster.broadcast <- event:
		log.Printf("[SSE] Broadcasting: %s %s", action, eventType)
	default:
		log.Printf("[SSE] Broadcast channel full, dropping event")
	}
}

// handleSSE handles SSE connections for dashboard updates
func handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Create client channel
	clientChan := make(chan DashboardEvent, 10)
	dashboardBroadcaster.register <- clientChan

	// Cleanup on disconnect
	defer func() {
		dashboardBroadcaster.unregister <- clientChan
	}()

	// Get the request context for cancellation
	ctx := r.Context()

	// Flusher for immediate writes
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Keep-alive ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send keep-alive
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case event := <-clientChan:
			// Send event
			fmt.Fprintf(w, "event: update\ndata: {\"type\":\"%s\",\"action\":\"%s\",\"timestamp\":%d}\n\n",
				event.Type, event.Action, event.Timestamp)
			flusher.Flush()
		}
	}
}

// RegisterSSERoute registers the SSE endpoint
func RegisterSSERoute(mux *http.ServeMux) {
	mux.HandleFunc("/api/dashboard/events", handleSSE)
}
