package shopline

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"web-crawler-go/internal/core/domain"
	"web-crawler-go/internal/core/ports"
)

// Sitemap XML structure
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

type Parser struct {
	fetcher ports.HTMLFetcher
}

func (p *Parser) ProcessProducts(ctx context.Context, url string) (*domain.Product, error) {
	sitemapUrl := fmt.Sprintf("%s/sitemap.xml", url)
	body, err := p.fetcher.Fetch(ctx, sitemapUrl)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// Parse the sitemap XML
	productURLs, err := p.parseProductURLsFromSitemap(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sitemap: %w", err)
	}

	// For now, just log the URLs. You can process them as needed
	var products []*domain.Product
	for _, productURL := range productURLs {
		// Fetch and parse each product page
		product, err := p.fetchAndParseProduct(ctx, productURL)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", productURL, err)
			continue
		}
		products = append(products, product)
	}

	// TODO: Process each product URL to extract product information
	// For now, returning mock data
	return &domain.Product{
		Name:        "Shopline Product",
		Price:       123.45,
		Description: "Parsed from a Shopline page.",
		ImageURL:    "http://example.com/shopline-image.png",
	}, nil
}

func (p *Parser) parseProductURLsFromSitemap(body io.Reader) ([]string, error) {
	var sitemap Sitemap
	decoder := xml.NewDecoder(body)

	if err := decoder.Decode(&sitemap); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	var productURLs []string
	for _, url := range sitemap.URLs {
		if strings.Contains(url.Loc, "/products/") {
			productURLs = append(productURLs, url.Loc)
		}
	}

	return productURLs, nil
}

func (p *Parser) fetchAndParseProduct(ctx context.Context, productURL string) (*domain.Product, error) {
	body, err := p.fetcher.Fetch(ctx, productURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// Use your existing Parse method to extract product data
	return p.Parse(ctx, body)
}

func NewParser(fetcher ports.HTMLFetcher) *Parser {
	return &Parser{
		fetcher: fetcher,
	}
}

func (p *Parser) Parse(ctx context.Context, html io.Reader) (*domain.Product, error) {
	return &domain.Product{
		Name:        "Shopline Product",
		Price:       123.45,
		Description: "Parsed from a Shopline page.",
		ImageURL:    "http://example.com/shopline-image.png",
	}, nil
}
