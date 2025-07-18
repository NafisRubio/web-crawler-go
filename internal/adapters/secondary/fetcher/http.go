package fetcher

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"web-crawler-go/internal/core/ports"
)

// Default cache expiration time
const defaultCacheExpiration = 730 * time.Hour

type HTTPFetcher struct {
	cache  ports.CacheService
	logger ports.Logger
}

func NewHTTPFetcher(cache ports.CacheService, logger ports.Logger) *HTTPFetcher {
	return &HTTPFetcher{
		cache:  cache,
		logger: logger,
	}
}

// generateCacheKey creates a unique key for caching based on the URL
func generateCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return "url:" + hex.EncodeToString(hash[:])
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	// Generate cache key
	cacheKey := generateCacheKey(url)

	// Try to get from cache first
	if f.cache != nil {
		cachedData, found, err := f.cache.Get(ctx, cacheKey)
		if err != nil {
			f.logger.Error("cache get error", "error", err)
		} else if found {
			f.logger.Info("cache hit", "key", cacheKey)
			return cachedData, nil
		}
	}

	f.logger.Info("cache miss, making HTTP request", "url", url)
	// If not in cache or cache error, make HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If we have a cache, store the response
	if f.cache != nil {
		// We need to read the body to store it in cache
		// and then provide it to the caller
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			// If we can't read the body, just return the original response
			return resp.Body, nil
		}

		// Close the original body
		resp.Body.Close()

		// Create two readers: one for cache, one to return
		cacheReader := io.NopCloser(bytes.NewReader(bodyBytes))
		returnReader := io.NopCloser(bytes.NewReader(bodyBytes))

		// Store in cache asynchronously to not block the response
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			f.logger.Info("setting cache", "key", cacheKey)
			err := f.cache.Set(cacheCtx, cacheKey, cacheReader, defaultCacheExpiration)
			if err != nil {
				f.logger.Error("cache set error", "error", err)
			}
		}()

		return returnReader, nil
	}

	// Caller is responsible for closing the body
	return resp.Body, nil
}
