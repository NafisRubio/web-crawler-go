package http

import (
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
	h.logger.Info("received request", "method", r.Method, "domain_name", r.URL.String())

	// 1. Get URL parameter
	domainName := r.URL.Query().Get("domain_name")
	if domainName == "" {
		h.logger.Error("missing URL parameter")
		RespondError(w, h.logger, http.StatusBadRequest, "URL parameter is required", nil)
		return
	}

	// 2. Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	// 3. Get products from the service
	products, totalItems, err := h.service.GetProductsByDomainName(r.Context(), domainName, page, pageSize)
	if err != nil {
		h.logger.Error("failed to get products", "error", err)
		RespondError(w, h.logger, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	// 4. Construct the response
	var nextPage, prevPage string
	if page < totalPages {
		nextPage = fmt.Sprintf("%s?domain_name=%s&page=%d&page_size=%d", r.URL.Path, domainName, page+1, pageSize)
	}
	if page > 1 {
		prevPage = fmt.Sprintf("%s?domain_name=%s&page=%d&page_size=%d", r.URL.Path, domainName, page-1, pageSize)
	}

	pagination := &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		NextPage:   nextPage,
		PrevPage:   prevPage,
	}

	h.logger.Info("successfully retrieved products", "count", len(products), "page", page, "pageSize", pageSize)

	RespondSuccess(w, h.logger, http.StatusOK, "Products retrieved successfully", products, pagination)
}

func (h *ProductHandler) CrawlDomain(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received request", "method", r.Method, "domainName", r.URL.String())

	// 1. Get URL parameter
	domainName := r.URL.Query().Get("domain_name")
	if domainName == "" {
		h.logger.Error("missing Domain parameter")
		RespondError(w, h.logger, http.StatusBadRequest, "URL parameter is required", nil)
		return
	}

	// 2. Get products from the service
	domainUrl := "https://" + domainName
	productsCount, err := h.service.CrawlAndSaveProductsFromURL(r.Context(), domainUrl)
	if err != nil {
		h.logger.Error("failed to get productsCount", "error", err)
		RespondError(w, h.logger, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	h.logger.Info("successfully crawled domainName")

	RespondSuccess(w, h.logger, http.StatusOK, "Domain crawled successfully", map[string]int{"productsCount": productsCount}, nil)
}
