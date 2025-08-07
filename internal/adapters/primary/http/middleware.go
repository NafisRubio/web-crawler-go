package http

import (
	"net/http"
	"time"

	"github.com/rs/cors"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Flush makes the wrapper implement http.Flusher
func (rw *responseWriter) Flush() {
	// Check if the underlying ResponseWriter supports flushing
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
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

// corsMiddleware wraps a handler with CORS protection
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	return c.Handler(next)
}

// pipeline applies multiple middleware functions in order
func (r *Router) pipeline(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
