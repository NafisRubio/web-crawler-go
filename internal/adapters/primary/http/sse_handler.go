package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"web-crawler-go/internal/core/ports"
)

// SSEHandler handles Server-Sent Events HTTP connections
type SSEHandler struct {
	sseService ports.SSEService
	logger     ports.Logger
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(sseService ports.SSEService, logger ports.Logger) *SSEHandler {
	return &SSEHandler{
		sseService: sseService,
		logger:     logger,
	}
}

// HandleSSE handles SSE connection requests
func (h *SSEHandler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Get client ID from query parameter or generate one
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}

	h.logger.Info("SSE connection request", "clientID", clientID, "remoteAddr", r.RemoteAddr)

	// Create a channel for this client
	messageChan := make(chan ports.SSEMessage, 10) // Buffer of 10 messages
	defer close(messageChan)

	// Subscribe the client to the SSE service
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := h.sseService.Subscribe(ctx, clientID, messageChan); err != nil {
		h.logger.Error("failed to subscribe client to SSE", "error", err, "clientID", clientID)
		http.Error(w, "Failed to establish SSE connection", http.StatusInternalServerError)
		return
	}

	// Ensure client is unsubscribed when connection closes
	defer func() {
		h.sseService.Unsubscribe(clientID)
		h.logger.Info("SSE connection closed", "clientID", clientID)
	}()

	// Get the flusher to send data immediately
	h.logger.Info("Checking for http.Flusher support")
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.Error("streaming unsupported", "clientID", clientID)
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	h.logger.Info("http.Flusher support is available")

	// Send initial connection confirmation
	h.writeSSEMessage(w, flusher, ports.SSEMessage{
		ID:    fmt.Sprintf("init-%d", time.Now().Unix()),
		Event: "connection",
		Data: map[string]interface{}{
			"status":    "connected",
			"clientID":  clientID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})

	// Keep connection alive and send messages
	ticker := time.NewTicker(30 * time.Second) // Heartbeat every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				h.logger.Info("message channel closed", "clientID", clientID)
				return
			}
			h.writeSSEMessage(w, flusher, message)

		case <-ticker.C:
			// Send heartbeat to keep connection alive
			heartbeat := ports.SSEMessage{
				ID:    fmt.Sprintf("heartbeat-%d", time.Now().Unix()),
				Event: "heartbeat",
				Data: map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"clients":   h.sseService.GetConnectedClients(),
				},
			}
			h.writeSSEMessage(w, flusher, heartbeat)

		case <-ctx.Done():
			h.logger.Info("SSE context cancelled", "clientID", clientID)
			return
		}
	}
}

// writeSSEMessage writes an SSE message to the HTTP response writer
func (h *SSEHandler) writeSSEMessage(w http.ResponseWriter, flusher http.Flusher, message ports.SSEMessage) {
	// Convert message data to JSON
	dataJSON, err := json.Marshal(message.Data)
	if err != nil {
		h.logger.Error("failed to marshal SSE message data", "error", err)
		return
	}

	// Write SSE format
	if message.ID != "" {
		fmt.Fprintf(w, "id: %s\n", message.ID)
	}
	if message.Event != "" {
		fmt.Fprintf(w, "event: %s\n", message.Event)
	}
	fmt.Fprintf(w, "data: %s\n\n", string(dataJSON))

	// Flush the data to the client immediately
	flusher.Flush()
}

// GetSSEStatus returns the current SSE service status
func (h *SSEHandler) GetSSEStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("SSE status request", "remoteAddr", r.RemoteAddr)

	status := map[string]interface{}{
		"connected_clients": h.sseService.GetConnectedClients(),
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
		"status":            "active",
	}

	RespondSuccess(w, h.logger, http.StatusOK, "SSE service status", status, nil)
}
