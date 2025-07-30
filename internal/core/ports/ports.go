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
	CrawlAndSaveProductsFromURL(ctx context.Context, domainUrl string) ([]*domain.Product, error)
	GetProviderFromURL(ctx context.Context, domainUrl string) (ProductProvider, error)
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
}
