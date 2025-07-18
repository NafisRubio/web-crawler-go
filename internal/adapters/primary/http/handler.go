package http

import (
	"encoding/json"
	"net/http"
	"web-crawler-go/internal/core/ports"
)

type ProductHandler struct {
	service ports.ProductService
	logger  ports.Logger
}

func NewProductHandler(service ports.ProductService, logger ports.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received request", "method", r.Method, "url", r.URL.String())
	url := r.URL.Query().Get("url")
	if url == "" {
		h.logger.Error("missing URL parameter")
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProductsFromURL(r.Context(), url)
	if err != nil {
		h.logger.Error("failed to get products", "error", err)
		// In a real app, check error type for better status codes
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("successfully retrieved products", "count", len(product))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
