package ports

import (
	"context"
	"io"
	"time"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	// Get retrieves data from the cache for the given key
	Get(ctx context.Context, key string) (io.ReadCloser, bool, error)

	// Set stores data in the cache with the given key and expiration time
	Set(ctx context.Context, key string, value io.ReadCloser, expiration time.Duration) error

	// Delete removes data from the cache for the given key
	Delete(ctx context.Context, key string) error
}
