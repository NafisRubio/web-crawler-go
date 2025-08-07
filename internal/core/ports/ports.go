package ports

import (
	"context"
	"io"
	"web-crawler-go/internal/core/domain"
)

// --- Primary/Driving Ports ---

// ProductService is the interface for the application's business logic.
// It's called by primary adapters (e.g., HTTP handlers).
type ProductService interface {
	CrawlAndSaveProductsFromURL(ctx context.Context, domainUrl string) (int, error)
	GetProviderFromURL(ctx context.Context, domainUrl string) (ProductProvider, error)
	GetProductsByDomainName(ctx context.Context, domainName string, page, pageSize int) ([]*domain.Product, int, error)
}

// --- Secondary/Driven Ports ---

// HTMLFetcher is an interface for fetching HTML content from a URL.
type HTMLFetcher interface {
	Fetch(ctx context.Context, domainUrl string) (io.ReadCloser, error)
}

// ProductProvider is an interface for parsing product data from HTML.
type ProductProvider interface {
	Parse(ctx context.Context, html io.Reader) (*domain.Product, error)
	ProcessProducts(ctx context.Context, domainUrl string) ([]*domain.Product, error)
}

// ProductRepository is an interface for persisting products.
type ProductRepository interface {
	UpsertProduct(ctx context.Context, product *domain.Product) error
	Close(ctx context.Context) error
	GetProducts(ctx context.Context, domainName string, page, pageSize int) ([]*domain.Product, error)
	GetTotalProducts(ctx context.Context, domainName string) (int, error)
}

// SSEService is an interface for Server-Sent Events functionality.
// It allows broadcasting real-time messages to connected clients.
type SSEService interface {
	// Subscribe adds a new client connection for receiving SSE messages
	Subscribe(ctx context.Context, clientID string, messageChan chan<- SSEMessage) error
	// Unsubscribe removes a client connection
	Unsubscribe(clientID string)
	// Broadcast sends a message to all connected clients
	Broadcast(ctx context.Context, message SSEMessage) error
	// BroadcastToClient sends a message to a specific client
	BroadcastToClient(ctx context.Context, clientID string, message SSEMessage) error
	// GetConnectedClients returns the number of connected clients
	GetConnectedClients() int
}

// SSEMessage represents a Server-Sent Event message
type SSEMessage struct {
	ID    string                 `json:"id,omitempty"`
	Event string                 `json:"event,omitempty"`
	Data  map[string]interface{} `json:"data"`
}
