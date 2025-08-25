package cache

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/redis/go-redis/v9"
	"web-crawler-go/internal/core/ports"
)

// RedisCache implements the CacheService interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache service
func NewRedisCache(addr string, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
	}
}

// Get retrieves data from the cache for the given key
func (c *RedisCache) Get(ctx context.Context, key string) (io.ReadCloser, bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Key does not exist
			return nil, false, nil
		}
		return nil, false, err
	}

	return io.NopCloser(bytes.NewReader(data)), true, nil
}

// Set stores data in the cache with the given key and expiration time
func (c *RedisCache) Set(ctx context.Context, key string, value io.ReadCloser, expiration time.Duration) error {
	// Read the data from the ReadCloser
	data, err := io.ReadAll(value)
	if err != nil {
		return err
	}

	// Reset the ReadCloser for further use
	value = io.NopCloser(bytes.NewReader(data))

	// Store the data in Redis
	return c.client.Set(ctx, key, data, expiration).Err()
}

// Delete removes data from the cache for the given key
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Ensure RedisCache implements CacheService
var _ ports.CacheService = (*RedisCache)(nil)
