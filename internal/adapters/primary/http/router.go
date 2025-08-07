package http

import (
	"net/http"
	"web-crawler-go/internal/core/ports"
)

// Router handles HTTP routing configuration
type Router struct {
	productHandler *ProductHandler
	crawlerHandler *CrawlerHandler
	sseHandler     *SSEHandler
	logger         ports.Logger
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewRouter creates a new router with the given dependencies
func NewRouter(productService ports.ProductService, sseService ports.SSEService, logger ports.Logger) *Router {
	productHandler := NewProductHandler(productService, logger)
	sseHandler := NewSSEHandler(sseService, logger)
	crawlerHandler := NewCrawlerHandler(productService, logger) // Create new handler

	return &Router{
		productHandler: productHandler,
		crawlerHandler: crawlerHandler, // Add to router
		sseHandler:     sseHandler,
		logger:         logger,
	}
}

// SetupRoutes configures all HTTP routes and returns the configured handler with middleware
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Crawler
	mux.HandleFunc("GET /api/v1/crawl", r.crawlerHandler.CrawlDomain)

	// Product endpoints
	mux.HandleFunc("GET /api/v1/products", r.productHandler.GetProduct)

	// SSE endpoints
	mux.HandleFunc("GET /api/v1/sse", r.sseHandler.HandleSSE)
	mux.HandleFunc("GET /api/v1/sse/status", r.sseHandler.GetSSEStatus)

	// Apply middleware pipeline
	return r.pipeline(mux, r.loggingMiddleware, r.corsMiddleware)
}
