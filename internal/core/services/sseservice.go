package services

import (
	"context"
	"fmt"
	"sync"
	"time"
	"web-crawler-go/internal/core/ports"
)

// sseService implements the SSEService port.
type sseService struct {
	clients map[string]chan<- ports.SSEMessage
	mutex   sync.RWMutex
	logger  ports.Logger
}

// NewSSEService creates a new instance of the SSE service.
func NewSSEService(logger ports.Logger) ports.SSEService {
	return &sseService{
		clients: make(map[string]chan<- ports.SSEMessage),
		logger:  logger,
	}
}

// Subscribe adds a new client connection for receiving SSE messages
func (s *sseService) Subscribe(ctx context.Context, clientID string, messageChan chan<- ports.SSEMessage) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if client already exists
	if _, exists := s.clients[clientID]; exists {
		s.logger.Warn("client already subscribed, replacing connection", "clientID", clientID)
	}

	s.clients[clientID] = messageChan
	s.logger.Info("client subscribed to SSE", "clientID", clientID, "totalClients", len(s.clients))

	// Send welcome message
	welcomeMsg := ports.SSEMessage{
		ID:    fmt.Sprintf("welcome-%d", time.Now().Unix()),
		Event: "connected",
		Data: map[string]interface{}{
			"message":   "Successfully connected to SSE stream",
			"clientID":  clientID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	select {
	case messageChan <- welcomeMsg:
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Warn("failed to send welcome message, channel might be full", "clientID", clientID)
	}

	return nil
}

// Unsubscribe removes a client connection
func (s *sseService) Unsubscribe(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.clients[clientID]; exists {
		delete(s.clients, clientID)
		s.logger.Info("client unsubscribed from SSE", "clientID", clientID, "totalClients", len(s.clients))
	} else {
		s.logger.Warn("attempted to unsubscribe non-existent client", "clientID", clientID)
	}
}

// Broadcast sends a message to all connected clients
func (s *sseService) Broadcast(ctx context.Context, message ports.SSEMessage) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.clients) == 0 {
		s.logger.Debug("no clients connected, skipping broadcast")
		return nil
	}

	// Add timestamp if not present
	if message.Data == nil {
		message.Data = make(map[string]interface{})
	}
	if _, exists := message.Data["timestamp"]; !exists {
		message.Data["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	}

	s.logger.Info("broadcasting message to all clients", "event", message.Event, "clientCount", len(s.clients))

	// Send to all clients
	var failedClients []string
	for clientID, clientChan := range s.clients {
		select {
		case clientChan <- message:
			// Message sent successfully
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Channel is full or closed, mark for removal
			s.logger.Warn("failed to send message to client, marking for removal", "clientID", clientID)
			failedClients = append(failedClients, clientID)
		}
	}

	// Remove failed clients (do this after the loop to avoid modifying map during iteration)
	if len(failedClients) > 0 {
		s.mutex.RUnlock()
		s.mutex.Lock()
		for _, clientID := range failedClients {
			delete(s.clients, clientID)
			s.logger.Info("removed failed client", "clientID", clientID)
		}
		s.mutex.Unlock()
		s.mutex.RLock()
	}

	return nil
}

// BroadcastToClient sends a message to a specific client
func (s *sseService) BroadcastToClient(ctx context.Context, clientID string, message ports.SSEMessage) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clientChan, exists := s.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	// Add timestamp if not present
	if message.Data == nil {
		message.Data = make(map[string]interface{})
	}
	if _, exists := message.Data["timestamp"]; !exists {
		message.Data["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	}

	s.logger.Info("sending message to specific client", "clientID", clientID, "event", message.Event)

	select {
	case clientChan <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel is full or closed, remove client
		s.mutex.RUnlock()
		s.mutex.Lock()
		delete(s.clients, clientID)
		s.mutex.Unlock()
		s.mutex.RLock()

		s.logger.Warn("failed to send message to client, client removed", "clientID", clientID)
		return fmt.Errorf("failed to send message to client %s, client removed", clientID)
	}
}

// GetConnectedClients returns the number of connected clients
func (s *sseService) GetConnectedClients() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.clients)
}
