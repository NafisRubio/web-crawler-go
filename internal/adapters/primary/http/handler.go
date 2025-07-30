package http

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
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

	// 1. Get URL parameter
	url := r.URL.Query().Get("url")
	if url == "" {
		h.logger.Error("missing URL parameter")
		response := Response{
			Status:  "error",
			Message: "URL parameter is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 2. Get products from the service
	products, err := h.service.CrawlAndSaveProductsFromURL(r.Context(), url)
	if err != nil {
		h.logger.Error("failed to get products", "error", err)
		response := Response{
			Status:  "error",
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 3. Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	totalItems := int64(len(products))
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	// Calculate start and end indices for slicing
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	if startIndex > len(products) {
		startIndex = len(products)
	}
	if endIndex > len(products) {
		endIndex = len(products)
	}

	pagedProducts := products[startIndex:endIndex]

	// 4. Construct the response
	var nextPage, prevPage string
	if page < totalPages {
		nextPage = fmt.Sprintf("%s?url=%s&page=%d&page_size=%d", r.URL.Path, url, page+1, pageSize)
	}
	if page > 1 {
		prevPage = fmt.Sprintf("%s?url=%s&page=%d&page_size=%d", r.URL.Path, url, page-1, pageSize)
	}

	pagination := &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		NextPage:   nextPage,
		PrevPage:   prevPage,
	}

	h.logger.Info("successfully retrieved products", "count", len(pagedProducts), "page", page, "pageSize", pageSize)

	response := Response{
		Status:     "success",
		Data:       pagedProducts,
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) CrawlDomain(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received request", "method", r.Method, "domainName", r.URL.String())

	// 1. Get URL parameter
	domainName := r.URL.Query().Get("domain_name")
	if domainName == "" {
		h.logger.Error("missing Domain parameter")
		response := Response{
			Status:  "error",
			Message: "URL parameter is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 2. Get products from the service
	domainUrl := "https://" + domainName
	products, err := h.service.CrawlAndSaveProductsFromURL(r.Context(), domainUrl)
	if err != nil {
		h.logger.Error("failed to get products", "error", err)
		response := Response{
			Status:  "error",
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	h.logger.Info("successfully crawled domainName")

	response := Response{
		Status: "success",
		Data:   products,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
