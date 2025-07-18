package services

import (
	"context"
	"errors"
	"golang.org/x/net/html"
	"web-crawler-go/internal/core/domain"
	"web-crawler-go/internal/core/ports"
)

var ErrProviderNotFound = errors.New("suitable provider not found for the given URL")

// productService implements the ProductService port.
type productService struct {
	fetcher          ports.HTMLFetcher
	providerRegistry map[string]ports.ProductProvider // Maps hostname -> provider
	logger           ports.Logger
}

// NewProductService creates a new instance of the product service.
func NewProductService(fetcher ports.HTMLFetcher, registry map[string]ports.ProductProvider, logger ports.Logger) ports.ProductService {
	return &productService{
		fetcher:          fetcher,
		providerRegistry: registry,
		logger:           logger,
	}
}

func (p *productService) GetProviderFromURL(ctx context.Context, rawUrl string) (ports.ProductProvider, error) {
	p.logger.Info("fetching HTML", "url", rawUrl)
	// 1. Fetch the HTML
	htmlBody, err := p.fetcher.Fetch(ctx, rawUrl)
	if err != nil {
		p.logger.Error("failed to fetch HTML", "error", err)
		return nil, err
	}
	defer htmlBody.Close()

	// 2. Parse the HTML document
	doc, err := html.Parse(htmlBody)
	if err != nil {
		return nil, err
	}

	if isDNSPrefetchShopLine(doc) {
		provider := "shopline.tw"
		p.logger.Info("provider identified", "provider", provider)
		return p.providerRegistry[provider], nil
	}

	return nil, ErrProviderNotFound
}

func isDNSPrefetchShopLine(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "link" {
		var rel, href string
		for _, attr := range n.Attr {
			if attr.Key == "rel" {
				rel = attr.Val
			}
			if attr.Key == "href" {
				href = attr.Val
			}
		}
		if rel == "dns-prefetch" && href == "https://cdn.shoplineapp.com" {
			return true
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := isDNSPrefetchShopLine(c); found {
			return true
		}
	}
	return false
}

func (p *productService) GetProductsFromURL(ctx context.Context, rawURL string) ([]*domain.Product, error) {
	p.logger.Info("getting products from url", "url", rawURL)
	// 1. Identify the provider from the URL
	provider, err := p.GetProviderFromURL(ctx, rawURL)
	if err != nil || provider == nil {
		p.logger.Error("failed to get provider from url", "error", err)
		return nil, ErrProviderNotFound
	}
	p.logger.Info("provider found", "provider", provider)
	// 2. Fetch the HTML content using the fetcher port
	products, err := provider.ProcessProducts(ctx, rawURL)
	if err != nil {
		p.logger.Error("failed to process products", "error", err)
		return nil, err
	}
	p.logger.Info("successfully fetched products", "count", len(products))

	return products, nil
}
