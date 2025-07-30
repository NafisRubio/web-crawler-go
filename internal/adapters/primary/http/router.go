package http

import (
	"net/http"
	"time"
	"web-crawler-go/internal/core/ports"
)

// Router handles HTTP routing configuration
type Router struct {
	productHandler *ProductHandler
	logger         ports.Logger
}

// NewRouter creates a new router with the given dependencies
func NewRouter(productService ports.ProductService, logger ports.Logger) *Router {
	productHandler := NewProductHandler(productService, logger)

	return &Router{
		productHandler: productHandler,
		logger:         logger,
	}
}

// loggingMiddleware wraps an HTTP handler with request logging using slog
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log the incoming request
		r.logger.Info("HTTP request started",
			"method", req.Method,
			"path", req.URL.Path,
			"remote_addr", req.RemoteAddr,
			"user_agent", req.UserAgent(),
		)

		// Call the next handler
		next.ServeHTTP(wrapper, req)

		// Log the completed request
		duration := time.Since(start)
		r.logger.Info("HTTP request completed",
			"method", req.Method,
			"path", req.URL.Path,
			"status_code", wrapper.statusCode,
			"duration_ms", int64(duration/time.Millisecond),
			"remote_addr", req.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// SetupRoutes configures all HTTP routes and returns the configured handler with middleware
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Crawler
	mux.HandleFunc("GET /crawl", r.productHandler.CrawlDomain)

	// Product endpoints
	mux.HandleFunc("GET /products", r.productHandler.GetProduct)

	// Wrap the mux with logging middleware
	return r.loggingMiddleware(mux)
}
