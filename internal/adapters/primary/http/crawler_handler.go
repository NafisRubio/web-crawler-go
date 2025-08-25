package http

import (
	http2 "net/http"
	"regexp"
	"unicode/utf8"
	"web-crawler-go/internal/core/ports"
)

var validDomainPattern = regexp.MustCompile(`^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,61}[a-zA-Z0-9]{1}|[a-zA-Z0-9]{1,2})(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,61}[a-zA-Z0-9]{1}|\.[a-zA-Z0-9]{1,2})*$`)

// CrawlerHandler handles Server-Sent Events HTTP connections
type CrawlerHandler struct {
	productService ports.ProductService
	logger         ports.Logger
}

// NewCrawlerHandler creates a new Crawler handler
func NewCrawlerHandler(productService ports.ProductService, logger ports.Logger) *CrawlerHandler {
	return &CrawlerHandler{
		productService: productService,
		logger:         logger,
	}
}

func (h *CrawlerHandler) CrawlDomain(w http2.ResponseWriter, r *http2.Request) {
	h.logger.Info("received request", "method", r.Method, "domainName", r.URL.String())

	// 1. Get URL parameter
	domainName := r.URL.Query().Get("domain_name")
	if domainName == "" {
		h.logger.Error("missing Domain parameter")
		RespondError(w, h.logger, http2.StatusBadRequest, "URL parameter is required", nil)
		return
	}
	// Check the overall length of the domain.
	// A domain name can be a maximum of 253 characters.
	if utf8.RuneCountInString(domainName) > 253 {
		h.logger.Error("invalid domain name length", "domainName", domainName)
		RespondError(w, h.logger, http2.StatusBadRequest, "Invalid domain name length", nil)
		return
	}
	// Validate domain name format
	matched := validDomainPattern.MatchString(domainName)
	if !matched {
		h.logger.Error("invalid domain name format", "domainName", domainName)
		RespondError(w, h.logger, http2.StatusBadRequest, "Invalid domain name format", nil)
		return
	}

	// 2. Get products from the service
	domainUrl := "https://" + domainName
	productsCount, err := h.productService.CrawlAndSaveProductsFromURL(r.Context(), domainUrl)
	if err != nil {
		h.logger.Error("failed to get productsCount", "error", err)
		RespondError(w, h.logger, http2.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	h.logger.Info("successfully crawled domainName")

	RespondSuccess(w, h.logger, http2.StatusOK, "Domain crawled successfully", map[string]int{"productsCount": productsCount}, nil)
}
