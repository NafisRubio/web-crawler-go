package services

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"time"
	"web-crawler-go/internal/core/domain"
	"web-crawler-go/internal/core/ports"
)

const (
	DNSPrefetchRel = "dns-prefetch"
	ShopLineURL    = "https://cdn.shoplineapp.com"
)

var ErrProviderNotFound = errors.New("suitable provider not found for the given URL")

// productService implements the ProductService port.
type productService struct {
	fetcher          ports.HTMLFetcher
	providerRegistry map[string]ports.ProductProvider // Maps hostname -> provider
	repository       ports.ProductRepository
	sseService       ports.SSEService
	logger           ports.Logger
}

// NewProductService creates a new instance of the product service.
func NewProductService(fetcher ports.HTMLFetcher, registry map[string]ports.ProductProvider, repository ports.ProductRepository, sseService ports.SSEService, logger ports.Logger) ports.ProductService {
	return &productService{
		fetcher:          fetcher,
		providerRegistry: registry,
		repository:       repository,
		sseService:       sseService,
		logger:           logger,
	}
}

func (p *productService) GetProviderFromURL(ctx context.Context, domainUrl string) (ports.ProductProvider, error) {
	p.logger.Info("fetching HTML", "domainUrl", domainUrl)
	// 1. Fetch the HTML
	htmlBody, err := p.fetcher.Fetch(ctx, domainUrl)
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

	if hasDNSPrefetchLink(doc, ShopLineURL) {
		provider := "shopline.tw"
		p.logger.Info("provider identified", "provider", provider)
		return p.providerRegistry[provider], nil
	}

	return nil, ErrProviderNotFound
}

func hasDNSPrefetchLink(n *html.Node, targetURL string) bool {
	if n.Type == html.ElementNode && n.Data == "link" {
		var rel, href string
		for _, attr := range n.Attr {
			switch attr.Key {
			case "rel":
				rel = attr.Val
			case "href":
				href = attr.Val
			}
		}
		if rel == DNSPrefetchRel && href == targetURL {
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasDNSPrefetchLink(c, targetURL) {
			return true
		}
	}
	return false
}

func (p *productService) CrawlAndSaveProductsFromURL(ctx context.Context, domainUrl string) (int, error) {
	p.logger.Info("getting products from domainUrl", "domainUrl", domainUrl)

	// Send crawling started notification
	p.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("crawl-start-%d", time.Now().Unix()),
		Event: "crawl_started",
		Data: map[string]interface{}{
			"domain_url": domainUrl,
			"status":     "started",
			"message":    "Starting to crawl domain",
		},
	})

	// 1. Identify the provider from the URL
	provider, err := p.GetProviderFromURL(ctx, domainUrl)
	if err != nil || provider == nil {
		p.logger.Error("failed to get provider from domainUrl", "error", err)
		// Send error notification
		p.sseService.Broadcast(ctx, ports.SSEMessage{
			ID:    fmt.Sprintf("crawl-error-%d", time.Now().Unix()),
			Event: "crawl_error",
			Data: map[string]interface{}{
				"domain_url": domainUrl,
				"status":     "error",
				"message":    "Failed to identify provider for domain",
				"error":      err.Error(),
			},
		})
		return 0, ErrProviderNotFound
	}
	p.logger.Info("provider found", "provider", provider)

	// Send provider identified notification
	p.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("provider-found-%d", time.Now().Unix()),
		Event: "provider_identified",
		Data: map[string]interface{}{
			"domain_url": domainUrl,
			"status":     "provider_found",
			"message":    "Provider identified, starting product extraction",
		},
	})

	// 2. Fetch the HTML content using the fetcher port
	products, err := provider.ProcessProducts(ctx, domainUrl)
	if err != nil {
		p.logger.Error("failed to process products", "error", err)
		// Send error notification
		p.sseService.Broadcast(ctx, ports.SSEMessage{
			ID:    fmt.Sprintf("crawl-error-%d", time.Now().Unix()),
			Event: "crawl_error",
			Data: map[string]interface{}{
				"domain_url": domainUrl,
				"status":     "error",
				"message":    "Failed to process products from domain",
				"error":      err.Error(),
			},
		})
		return 0, err
	}
	p.logger.Info("successfully fetched products", "count", len(products))

	// Send products fetched notification
	p.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("products-fetched-%d", time.Now().Unix()),
		Event: "products_fetched",
		Data: map[string]interface{}{
			"domain_url":     domainUrl,
			"status":         "products_fetched",
			"message":        "Products extracted, starting database save",
			"products_count": len(products),
		},
	})

	// 3. Save each product to DB
	savedCount := 0
	for i, product := range products {
		if err := p.repository.UpsertProduct(ctx, product); err != nil {
			p.logger.Error("failed to save product to DB", "error", err, "product", product.Name)
			// Continue processing other products even if one fails
			continue
		}
		savedCount++

		// Send progress update every 10 products or on the last product
		if (i+1)%10 == 0 || i == len(products)-1 {
			p.sseService.Broadcast(ctx, ports.SSEMessage{
				ID:    fmt.Sprintf("save-progress-%d", time.Now().Unix()),
				Event: "save_progress",
				Data: map[string]interface{}{
					"domain_url":       domainUrl,
					"status":           "saving",
					"message":          "Saving products to database",
					"saved_count":      savedCount,
					"total_count":      len(products),
					"progress_percent": float64(savedCount) / float64(len(products)) * 100,
				},
			})
		}
	}

	productsCount := savedCount

	// Send crawling completed notification
	p.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("crawl-completed-%d", time.Now().Unix()),
		Event: "crawl_completed",
		Data: map[string]interface{}{
			"domain_url":     domainUrl,
			"status":         "completed",
			"message":        "Domain crawling completed successfully",
			"products_count": productsCount,
		},
	})

	return productsCount, nil
}

// GetProductsByDomainName return saved products with pagination
func (p *productService) GetProductsByDomainName(ctx context.Context, domainName string, page, pageSize int) ([]*domain.Product, int, error) {
	products, err := p.repository.GetProducts(ctx, domainName, page, pageSize)
	if err != nil {
		p.logger.Error("failed to get products from DB", "error", err)
		// Continue processing other products even if one fails
		return nil, 0, err
	}

	total, err := p.repository.GetTotalProducts(ctx, domainName)

	return products, total, nil
}
