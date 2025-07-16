package main

import (
	"log"
	"net/http"

	// Adapters
	http_adapter "web-crawler-go/internal/adapters/primary/http"
	"web-crawler-go/internal/adapters/secondary/fetcher"
	"web-crawler-go/internal/adapters/secondary/providers/shopify"
	"web-crawler-go/internal/adapters/secondary/providers/shopline"

	// Core
	"web-crawler-go/internal/core/ports"
	"web-crawler-go/internal/core/services"
)

func main() {
	// 1. Initialize Secondary/Driven Adapters
	htmlFetcher := fetcher.NewHTTPFetcher()
	shopifyProvider := shopify.NewParser()
	shoplineProvider := shopline.NewParser()
	// When you add Wix: wixProvider := wix.NewParser()

	// 2. Create the Provider Registry
	// The key should match the hostname you want to associate with the provider.
	providerRegistry := make(map[string]ports.ProductProvider)
	providerRegistry["shopify.com"] = shopifyProvider
	providerRegistry["shopline.tw"] = shoplineProvider
	// When you add Wix: providerRegistry["wix.com"] = wixProvider

	// 3. Initialize the Core Service (injecting dependencies)
	productService := services.NewProductService(htmlFetcher, providerRegistry)

	// 4. Initialize Primary/Driving Adapters (injecting service)
	productHandler := http_adapter.NewProductHandler(productService)

	// 5. Setup Router and Start Server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /product", productHandler.GetProduct)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
