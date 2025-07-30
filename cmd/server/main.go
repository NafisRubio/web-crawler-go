package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	// Adapters
	httpadapter "web-crawler-go/internal/adapters/primary/http"
	"web-crawler-go/internal/adapters/secondary/cache"
	"web-crawler-go/internal/adapters/secondary/fetcher"
	"web-crawler-go/internal/adapters/secondary/providers/shopify"
	"web-crawler-go/internal/adapters/secondary/providers/shopline"
	"web-crawler-go/internal/adapters/secondary/repository"

	// Core
	"web-crawler-go/internal/core/ports"
	"web-crawler-go/internal/core/services"
	"web-crawler-go/internal/core/services/loggerservice"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env.development"); err != nil {
		log.Println("No .env file found or error loading .env file:", err)
		// Don't fatal here - environment variables might be set elsewhere
	}

	// 0. Initialize Logger
	logger := loggerservice.NewLoggerService()

	// 1. Initialize Secondary/Driven Adapters

	// Initialize Redis cache
	redisHost := getEnvWithDefault("REDIS_HOST", "localhost:6379")
	redisPassword := getEnvWithDefault("REDIS_PASSWORD", "")
	redisCache := cache.NewRedisCache(redisHost, redisPassword, 0)

	// Initialize HTTP fetcher with Redis cache
	htmlFetcher := fetcher.NewHTTPFetcher(redisCache, logger)

	// Initialize MongoDB repository
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoDBURI := getEnvWithDefault("MONGODB_URI", "mongodb://localhost:27017")

	mongoDBRepo, err := repository.NewMongoDBRepository(
		ctx,
		mongoDBURI,
		"web-crawler", // Database name
		"products",    // Collection name
		logger,
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoDBRepo.Close(context.Background()); err != nil {
			logger.Error("Failed to close MongoDB connection", "error", err)
		}
	}()

	shopifyProvider := shopify.NewParser(htmlFetcher, logger)
	shoplineProvider := shopline.NewParser(htmlFetcher, logger)
	// When you add Wix: wixProvider := wix.NewParser()

	// 2. Create the Provider Registry
	// The key should match the hostname you want to associate with the provider.
	providerRegistry := make(map[string]ports.ProductProvider)
	providerRegistry["shopify.com"] = shopifyProvider
	providerRegistry["shopline.tw"] = shoplineProvider
	// When you add Wix: providerRegistry["wix.com"] = wixProvider

	// 3. Initialize the Core Service (injecting dependencies)
	productService := services.NewProductService(htmlFetcher, providerRegistry, mongoDBRepo, logger)

	// 4. Initialize Primary/Driving Adapters (injecting service)
	router := httpadapter.NewRouter(productService, logger)

	// 5. Setup Router and Start Server
	handler := router.SetupRoutes()

	port := getEnvWithDefault("PORT", "8080")
	log.Printf("Server starting on :%s...", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

// Helper function to get environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
