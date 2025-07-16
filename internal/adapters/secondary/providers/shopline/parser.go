package shopline

import (
	"context"
	"io"
	"web-crawler-go/internal/core/domain"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, html io.Reader) (*domain.Product, error) {
	return &domain.Product{
		Name:        "Shopline Product",
		Price:       123.45,
		Description: "Parsed from a Shopline page.",
		ImageURL:    "http://example.com/shopline-image.png",
	}, nil
}
