package services

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html"
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

func (p *productService) GetProviderFromURL(ctx context.Context, rawUrl string) (ports.ProductProvider, error) {
	// 1. Fetch the HTML
	htmlBody, err := p.fetcher.Fetch(ctx, rawUrl)
	if err != nil {
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

func (p *productService) GetProductFromURL(ctx context.Context, rawURL string) (*domain.Product, error) {
	// 1. Identify the provider from the URL
	provider, err := p.GetProviderFromURL(ctx, rawURL)
	if err != nil || provider == nil {
		return nil, ErrProviderNotFound
	}
	// 2. Fetch the HTML content using the fetcher port
	products, err := provider.ProcessProducts(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	fmt.Printf("products: %+v\n", products)

	// 2. Fetch the HTML content using the fetcher port
	htmlBody, err := p.fetcher.Fetch(ctx, rawURL)
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
