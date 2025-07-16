package fetcher

import (
	"context"
	"io"
	"net/http"
)

type HTTPFetcher struct{}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Caller is responsible for closing the body
	return resp.Body, nil
}
