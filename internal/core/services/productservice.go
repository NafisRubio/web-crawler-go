package services

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"web-crawler-go/internal/core/domain"
	"web-crawler-go/internal/core/ports"
)

var ErrProviderNotFound = errors.New("suitable provider not found for the given URL")

// productService implements the ProductService port.
type productService struct {
	fetcher          ports.HTMLFetcher
	providerRegistry map[string]ports.ProductProvider // Maps hostname -> provider
}

// NewProductService creates a new instance of the product service.
func NewProductService(fetcher ports.HTMLFetcher, registry map[string]ports.ProductProvider) ports.ProductService {
	return &productService{
		fetcher:          fetcher,
		providerRegistry: registry,
	}
}

func (s *productService) GetProductFromURL(ctx context.Context, rawURL string) (*domain.Product, error) {
	// 1. Identify the provider from the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	// e.g., "www.shopify.com" -> "shopify.com"
	host := strings.TrimPrefix(parsedURL.Hostname(), "www.")
	provider, ok := s.providerRegistry[host]
	if !ok {
		return nil, ErrProviderNotFound
	}

	// 2. Fetch the HTML content using the fetcher port
	htmlBody, err := s.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	defer htmlBody.Close()

	// 3. Parse the data using the selected provider port
	product, err := provider.Parse(ctx, htmlBody)
	if err != nil {
		return nil, err
	}

	return product, nil
}
