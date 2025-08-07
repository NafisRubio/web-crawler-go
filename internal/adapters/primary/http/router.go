package http

import (
	"net/http"
	"web-crawler-go/internal/core/ports"
)

// Router handles HTTP routing configuration
type Router struct {
	productHandler *ProductHandler
	logger         ports.Logger
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewRouter creates a new router with the given dependencies
func NewRouter(productService ports.ProductService, logger ports.Logger) *Router {
	productHandler := NewProductHandler(productService, logger)

	return &Router{
		productHandler: productHandler,
		logger:         logger,
	}
}

// SetupRoutes configures all HTTP routes and returns the configured handler with middleware
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Crawler
	mux.HandleFunc("GET /api/v1/crawl", r.productHandler.CrawlDomain)

	// Product endpoints
	mux.HandleFunc("GET /api/v1/products", r.productHandler.GetProduct)

	// Apply middleware pipeline
	return r.pipeline(mux, r.loggingMiddleware, r.corsMiddleware)
}
