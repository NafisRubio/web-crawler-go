package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	httpadapter "web-crawler-go/internal/adapters/primary/http"
	"web-crawler-go/internal/core/domain"
	"web-crawler-go/internal/core/ports"
	"web-crawler-go/internal/core/services"
	"web-crawler-go/internal/core/services/loggerservice"
)

// MockProductService for testing SSE without external dependencies
type MockProductService struct {
	sseService ports.SSEService
	logger     ports.Logger
}

func (m *MockProductService) CrawlAndSaveProductsFromURL(ctx context.Context, domainUrl string) (int, error) {
	m.logger.Info("Mock crawling started", "domainUrl", domainUrl)

	// Send crawling started notification
	m.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("crawl-start-%d", time.Now().Unix()),
		Event: "crawl_started",
		Data: map[string]interface{}{
			"domain_url": domainUrl,
			"status":     "started",
			"message":    "Starting to crawl domain (mock)",
		},
	})

	// Simulate some processing time
	time.Sleep(1 * time.Second)

	// Send provider identified notification
	m.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("provider-found-%d", time.Now().Unix()),
		Event: "provider_identified",
		Data: map[string]interface{}{
			"domain_url": domainUrl,
			"status":     "provider_found",
			"message":    "Provider identified (mock)",
		},
	})

	time.Sleep(1 * time.Second)

	// Send products fetched notification
	mockProductCount := 57
	m.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("products-fetched-%d", time.Now().Unix()),
		Event: "products_fetched",
		Data: map[string]interface{}{
			"domain_url":     domainUrl,
			"status":         "products_fetched",
			"message":        "Products extracted (mock)",
			"products_count": mockProductCount,
		},
	})

	// Simulate saving products with progress updates
	for i := 1; i <= mockProductCount; i++ {
		time.Sleep(100 * time.Millisecond) // Simulate processing time

		// Send progress update every 5 products or on the last product
		if i%5 == 0 || i == mockProductCount {
			m.sseService.Broadcast(ctx, ports.SSEMessage{
				ID:    fmt.Sprintf("save-progress-%d", time.Now().Unix()),
				Event: "save_progress",
				Data: map[string]interface{}{
					"domain_url":       domainUrl,
					"status":           "saving",
					"message":          "Saving products to database (mock)",
					"saved_count":      i,
					"total_count":      mockProductCount,
					"progress_percent": float64(i) / float64(mockProductCount) * 100,
				},
			})
		}
	}

	// Send crawling completed notification
	m.sseService.Broadcast(ctx, ports.SSEMessage{
		ID:    fmt.Sprintf("crawl-completed-%d", time.Now().Unix()),
		Event: "crawl_completed",
		Data: map[string]interface{}{
			"domain_url":     domainUrl,
			"status":         "completed",
			"message":        "Domain crawling completed successfully (mock)",
			"products_count": mockProductCount,
		},
	})

	return mockProductCount, nil
}

func (m *MockProductService) GetProviderFromURL(ctx context.Context, domainUrl string) (ports.ProductProvider, error) {
	return nil, fmt.Errorf("mock service - not implemented")
}

func (m *MockProductService) GetProductsByDomainName(ctx context.Context, domainName string, page, pageSize int) ([]*domain.Product, int, error) {
	return nil, 0, fmt.Errorf("mock service - not implemented")
}

func main() {
	fmt.Println("Starting SSE Integration Test Server...")

	// Initialize logger
	logger := loggerservice.NewLoggerService()

	// Initialize SSE service
	sseService := services.NewSSEService(logger)

	// Create mock product service with SSE integration
	mockProductService := &MockProductService{
		sseService: sseService,
		logger:     logger,
	}

	// Create router with mock service
	router := httpadapter.NewRouter(mockProductService, sseService, logger)

	// Setup routes
	handler := router.SetupRoutes()

	// Start server
	port := "8080"
	fmt.Printf("SSE Test Server starting on :%s...\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /api/v1/sse - SSE stream endpoint")
	fmt.Println("  GET /api/v1/sse/status - SSE status endpoint")
	fmt.Println("  GET /api/v1/crawl?domain_name=example.com - Mock crawl endpoint")
	fmt.Println("  GET /api/v1/products - Mock products endpoint")
	fmt.Println("\nOpen test_sse.html in your browser to test the SSE functionality")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
