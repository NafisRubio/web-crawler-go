package shopline

import (
	"context"
	"encoding/json"
	"encoding/xml"
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

	merchantID, productID, err := p.parseMerchantIDAndProductID(body)
	if err != nil {
		fmt.Printf("Error processing %s: %v\n", productURL, err)
	}
	fmt.Printf("merchantID: %s\n", *merchantID)
	fmt.Printf("productID: %s\n", *productID)

	// Inside your ProcessProducts function:
	parsedURL, err := url.Parse(productURL)
	if err != nil {
		// Handle the error appropriately
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	productData, err := p.fetchProductData(ctx, parsedURL.Host, merchantID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product data: %w", err)
	}
	fmt.Printf("productData: %+v\n", productData)
	// Use your existing Parse method to extract product data
	return p.Parse(ctx, body)
}

func (p *Parser) parseMerchantIDAndProductID(htmlBody io.ReadCloser) (*string, *string, error) {
	bodyBytes, err := io.ReadAll(htmlBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read html body: %w", err)
	}

	// product', JSON.parse
	//re := regexp.MustCompile(`\\"merchantId\\":\\"([a-zA-Z0-9]+)\\"`)
	re := regexp.MustCompile(`app\.value\('product', JSON\.parse\('({\\"_id\\".+\})`)

	// The FindSubmatch methods also have a version that accepts a byte slice.
	jsonMatches := re.FindSubmatch(bodyBytes)
	if len(jsonMatches) < 1 {
		return nil, nil, fmt.Errorf("merchantId not found in html body")
	}

	rawJson := jsonMatches[1]
	// Use strings.ReplaceAll on the string representation
	validJsonString := strings.ReplaceAll(string(rawJson), `\"`, `"`)

	var config struct {
		ProductID  string `json:"_id"`
		MerchantID string `json:"owner_id"`
	}

	// Unmarshal expects a byte slice, so we convert the cleaned string back.
	err = json.Unmarshal([]byte(validJsonString), &config)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, nil, err
	}

	return &config.MerchantID, &config.ProductID, nil
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

func (p *Parser) fetchProductData(ctx context.Context, hostname string, merchantID *string, productID *string) (*Product, error) {
	//       const productModelUrl = `https://${this.domainName}/api/merchants/${productApplicationLdJson.owner_id}/products/${productApplicationLdJson._id}`
	productDataURL := fmt.Sprintf("https://%s/api/merchants/%s/products/%s", hostname, *merchantID, *productID)
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
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	// Map the data from the nested structure to your flat Product struct.
	productShopLine := &Product{
		Name:            apiResponse.Data.TitleTranslations["zh-hant"],
		Tags:            apiResponse.Data.CategoryIDs,
		Price:           apiResponse.Data.Price.Cents,
		PriceDiscounted: apiResponse.Data.PriceSale.Cents,
		Description:     apiResponse.Data.DescriptionTranslations["zh-hant"],
	}

	for _, media := range apiResponse.Data.Media {
		productShopLine.ImagesURL = append(productShopLine.ImagesURL, media.Images.Original.URL)
	}

	return productShopLine, nil
}
