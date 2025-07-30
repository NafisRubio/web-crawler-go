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
	repository       ports.ProductRepository
	logger           ports.Logger
}

// NewProductService creates a new instance of the product service.
func NewProductService(fetcher ports.HTMLFetcher, registry map[string]ports.ProductProvider, repository ports.ProductRepository, logger ports.Logger) ports.ProductService {
	return &productService{
		fetcher:          fetcher,
		providerRegistry: registry,
		repository:       repository,
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

func (p *productService) CrawlAndSaveProductsFromURL(ctx context.Context, domainUrl string) ([]*domain.Product, error) {
	p.logger.Info("getting products from domainUrl", "domainUrl", domainUrl)
	// 1. Identify the provider from the URL
	provider, err := p.GetProviderFromURL(ctx, domainUrl)
	if err != nil || provider == nil {
		p.logger.Error("failed to get provider from domainUrl", "error", err)
		return nil, ErrProviderNotFound
	}
	p.logger.Info("provider found", "provider", provider)
	// 2. Fetch the HTML content using the fetcher port
	products, err := provider.ProcessProducts(ctx, domainUrl)
	if err != nil {
		p.logger.Error("failed to process products", "error", err)
		return nil, err
	}
	p.logger.Info("successfully fetched products", "count", len(products))

	// 3. Save each product to DB
	for _, product := range products {
		if err := p.repository.UpsertProduct(ctx, product); err != nil {
			p.logger.Error("failed to save product to DB", "error", err, "product", product.Name)
			// Continue processing other products even if one fails
			continue
		}
	}

	return products, nil
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
