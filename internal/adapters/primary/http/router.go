package http

import (
	"net/http"
	"web-crawler-go/internal/core/ports"
)

// Router handles HTTP routing configuration
type Router struct {
	productHandler *ProductHandler
}

// NewRouter creates a new router with the given dependencies
func NewRouter(productService ports.ProductService, logger ports.Logger) *Router {
	productHandler := NewProductHandler(productService, logger)

	return &Router{
		productHandler: productHandler,
	}
}

// SetupRoutes configures all HTTP routes and returns the configured mux
func (r *Router) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Crawler
	mux.HandleFunc("GET /crawl", r.productHandler.CrawlDomain)

	// Product endpoints
	mux.HandleFunc("GET /products", r.productHandler.GetProduct)

	return mux
}
