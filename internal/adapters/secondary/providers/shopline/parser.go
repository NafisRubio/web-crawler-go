package shopline

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
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
	logger  ports.Logger
}

type Product struct {
	Name            string
	Price           int
	PriceDiscounted int
	Description     string
	ImagesURL       []string
	Tags            []string
	Status          string
}

func (p *Parser) ProcessProducts(ctx context.Context, url string) ([]*domain.Product, error) {
	sitemapUrl := fmt.Sprintf("%s/sitemap.xml", url)
	p.logger.Info("processing products from sitemap", "url", sitemapUrl)
	body, err := p.fetcher.Fetch(ctx, sitemapUrl)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	// Parse the sitemap XML
	productURLs, err := p.parseProductURLsFromSitemap(body)
	if err != nil {
		p.logger.Error("failed to parse sitemap", "error", err)
		return nil, fmt.Errorf("failed to parse sitemap: %w", err)
	}

	p.logger.Info("found product urls", "count", len(productURLs))
	if len(productURLs) < 1 {
		p.logger.Error("no product URLs found in sitemap")
		return nil, errors.New("no product URLs found in sitemap")
	}

	products := make([]*domain.Product, len(productURLs))
	for i, productURL := range productURLs {
		p.logger.Info("processing product url", "url", productURL)
		product, err := p.fetchAndParseProduct(ctx, productURL)
		if err != nil {
			p.logger.Error("error processing product", "url", productURL, "error", err)
			return nil, fmt.Errorf("error processing %s: %w", productURL, err)
		}
		products[i] = product
	}

	return products, nil
}

// Parse implements the ProductProvider interface
func (p *Parser) Parse(ctx context.Context, html io.Reader) (*domain.Product, error) {
	// Read the HTML content
	htmlBytes, err := io.ReadAll(html)
	if err != nil {
		p.logger.Error("failed to read HTML", "error", err)
		return nil, fmt.Errorf("failed to read HTML: %w", err)
	}

	// Parse merchant ID and product ID from HTML
	merchantID, productID, err := p.parseMerchantIDAndProductIDFromBytes(htmlBytes)
	if err != nil {
		p.logger.Error("failed to parse merchant and product IDs", "error", err)
		return nil, fmt.Errorf("failed to parse merchant and product IDs: %w", err)
	}

	// This is a bit tricky - we need to extract the hostname from the HTML or use a default
	// For now, let's extract it from the HTML content or use a fallback approach
	hostname, err := p.extractHostnameFromHTML(htmlBytes)
	if err != nil {
		p.logger.Error("failed to extract hostname", "error", err)
		return nil, fmt.Errorf("failed to extract hostname: %w", err)
	}

	// Fetch product data from API
	productData, err := p.fetchProductData(ctx, hostname, merchantID, productID)
	if err != nil {
		p.logger.Error("failed to fetch product data", "error", err)
		return nil, fmt.Errorf("failed to fetch product data: %w", err)
	}

	// Parse the product data
	return p.parseProductResponse(productData)
}

func (p *Parser) parseProductURLsFromSitemap(body io.Reader) ([]string, error) {
	var sitemap Sitemap
	decoder := xml.NewDecoder(body)

	if err := decoder.Decode(&sitemap); err != nil {
		p.logger.Error("failed to decode XML", "error", err)
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
	p.logger.Info("fetching and parsing product", "url", productURL)
	body, err := p.fetcher.Fetch(ctx, productURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	merchantID, productID, err := p.parseMerchantIDAndProductID(body)
	if err != nil {
		p.logger.Error("error parsing merchant/product ID", "url", productURL, "error", err)
		return nil, fmt.Errorf("error parsing merchant/product ID: %w", err)
	}

	// Parse URL to get hostname
	parsedURL, err := url.Parse(productURL)
	if err != nil {
		p.logger.Error("failed to parse URL", "url", productURL, "error", err)
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	productData, err := p.fetchProductData(ctx, parsedURL.Host, merchantID, productID)
	if err != nil {
		p.logger.Error("failed to fetch product data", "host", parsedURL.Host, "merchantID", *merchantID, "productID", *productID, "error", err)
		return nil, fmt.Errorf("failed to fetch product data: %w", err)
	}

	// Use the parseProductResponse method
	return p.parseProductResponse(productData)
}

func (p *Parser) parseMerchantIDAndProductID(htmlBody io.ReadCloser) (*string, *string, error) {
	bodyBytes, err := io.ReadAll(htmlBody)
	if err != nil {
		p.logger.Error("failed to read HTML body", "error", err)
		return nil, nil, fmt.Errorf("failed to read HTML body: %w", err)
	}

	return p.parseMerchantIDAndProductIDFromBytes(bodyBytes)
}

func (p *Parser) parseMerchantIDAndProductIDFromBytes(bodyBytes []byte) (*string, *string, error) {
	re := regexp.MustCompile(`app\.value\('product', JSON\.parse\('({\\"_id\\".+\})`)

	jsonMatches := re.FindSubmatch(bodyBytes)
	if len(jsonMatches) < 2 {
		p.logger.Error("product data not found in HTML body")
		return nil, nil, errors.New("product data not found in HTML body")
	}

	rawJson := jsonMatches[1]
	validJsonString := string(rawJson)
	// Keep unescaping until no more escaped quotes are found
	validJsonString = strings.ReplaceAll(validJsonString, `\"`, `"`)
	validJsonString = strings.ReplaceAll(validJsonString, `\"`, `"`)

	var config struct {
		ProductID  string `json:"_id"`
		MerchantID string `json:"owner_id"`
	}

	err := json.Unmarshal([]byte(validJsonString), &config)
	if err != nil {
		p.logger.Error("error parsing JSON", "error", err)
		return nil, nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &config.MerchantID, &config.ProductID, nil
}

func fixMalformedJSON(jsonData []byte) []byte {
	// This is a simplified example - you'd need more robust regex
	// to handle all cases properly
	re := regexp.MustCompile(`"color":\s*"\{"\s*([^}]+)\s*"\}"`)
	return re.ReplaceAllFunc(jsonData, func(match []byte) []byte {
		// Extract and properly escape the content
		// This is a simplified example
		return bytes.ReplaceAll(match, []byte(`"`), []byte(`\"`))
	})
}

func (p *Parser) extractHostnameFromHTML(htmlBytes []byte) (string, error) {
	// Try to extract hostname from canonical URL or other meta tags
	re := regexp.MustCompile(`<link[^>]+rel="canonical"[^>]+href="https?://([^/]+)`)
	matches := re.FindSubmatch(htmlBytes)
	if len(matches) >= 2 {
		return string(matches[1]), nil
	}

	// Fallback: try to find it in script tags or other places
	re2 := regexp.MustCompile(`https?://([^/\s"']+\.shoplineapp\.com)`)
	matches2 := re2.FindSubmatch(htmlBytes)
	if len(matches2) >= 2 {
		return string(matches2[1]), nil
	}

	p.logger.Error("could not extract hostname from HTML")
	return "", errors.New("could not extract hostname from HTML")
}

func NewParser(fetcher ports.HTMLFetcher, logger ports.Logger) *Parser {
	return &Parser{
		fetcher: fetcher,
		logger:  logger,
	}
}

// parseProductResponse parses the API response into a domain Product
func (p *Parser) parseProductResponse(apiResponse *ProductResponse) (*domain.Product, error) {
	productShopLine := &domain.Product{
		Name:            apiResponse.Data.TitleTranslations["zh-hant"],
		Tags:            apiResponse.Data.CategoryIDs,
		Price:           apiResponse.Data.Price.Cents,
		PriceDiscounted: apiResponse.Data.PriceSale.Cents,
		Description:     apiResponse.Data.DescriptionTranslations["zh-hant"],
		Status:          "active",
	}

	if apiResponse.Data.Quantity < 1 {
		productShopLine.Status = "outOfStock"
	}

	for _, media := range apiResponse.Data.Media {
		productShopLine.ImagesURL = append(productShopLine.ImagesURL, media.Images.Original.URL)
	}

	return productShopLine, nil
}

func (p *Parser) fetchProductData(ctx context.Context, hostname string, merchantID *string, productID *string) (*ProductResponse, error) {
	productDataURL := fmt.Sprintf("https://%s/api/merchants/%s/products/%s", hostname, *merchantID, *productID)
	p.logger.Info("fetching product data", "url", productDataURL)
	fetchResponse, err := p.fetcher.Fetch(ctx, productDataURL)
	if err != nil {
		return nil, err
	}
	defer fetchResponse.Close()

	bodyBytes, err := io.ReadAll(fetchResponse)
	if err != nil {
		return nil, err
	}

	apiResponse := &ProductResponse{}
	err = json.Unmarshal(bodyBytes, apiResponse)
	if err != nil {
		p.logger.Error("error parsing JSON", "error", err)
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return apiResponse, nil
}
