package cache

import (
	"context"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/redis/go-redis/v9"
)

// Default values for cache operations
const (
	DefaultExpiration = 3600 // 1 hour in seconds
	NoExpiration      = -1   // Value for no expiration
)

// ErrKeyNotFound is returned when a key is not found in the cache
var ErrKeyNotFound = redis.Nil

var (
	cacheSvcInstance abstract.ICacheService
	cacheSvcOnce     sync.Once
)

type CacheService struct {
	redis *redis.Client
}

func CacheServiceInstance() abstract.ICacheService {
	cacheSvcOnce.Do(func() {
		cacheSvcInstance = newCacheService()
	})

	return cacheSvcInstance
}

func newCacheService() abstract.ICacheService {
	return &CacheService{
		redis: RedisInstance(),
	}
}

// Clear removes all keys from the cache
func (c *CacheService) Clear(ctx context.Context) error {
	if err := c.redis.FlushAll(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// Delete removes a key from the cache
func (c *CacheService) Delete(ctx context.Context, key string) error {
	if err := c.redis.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

// Exists checks if a key exists in the cache
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// Get retrieves a value from the cache by key
func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

// GetBit retrieves the bit value at the specified offset in a string stored at key
func (c *CacheService) GetBit(ctx context.Context, key string, offset int) (int64, error) {
	val, err := c.redis.GetBit(ctx, key, int64(offset)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrKeyNotFound
		}

		return 0, err
	}

	return val, nil
}

// Set stores a value in the cache with the given key and expiration time
// If expiration is set to NoExpiration, the key will not expire
func (c *CacheService) Set(ctx context.Context, key string, value any, expiration int) error {
	duration := time.Duration(expiration) * time.Second
	if expiration == NoExpiration {
		duration = 0
	}

	if err := c.redis.Set(ctx, key, value, duration).Err(); err != nil {
		return err
	}

	return nil
}

// SetBit sets the bit value at the specified offset in a string stored at key
func (c *CacheService) SetBit(ctx context.Context, key string, offset int, value int) (int64, error) {
	val, err := c.redis.SetBit(ctx, key, int64(offset), value).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrKeyNotFound
		}

		return 0, err
	}

	return val, nil
}

// GetWithDefault retrieves a value from the cache, returning the default value if not found
func (c *CacheService) GetWithDefault(ctx context.Context, key string, defaultValue string) (string, error) {
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return defaultValue, nil
	}

	if err != nil {
		return defaultValue, err
	}

	return val, nil
}

// SetWithDefaultExpiration stores a value with the default expiration time
func (c *CacheService) SetWithDefaultExpiration(ctx context.Context, key string, value any) error {
	return c.Set(ctx, key, value, DefaultExpiration)
}

// SetNX sets a value in the cache only if the key does not exist
func (c *CacheService) SetNX(ctx context.Context, key string, value any, expiration int) (bool, error) {
	duration := time.Duration(expiration) * time.Second
	if expiration == NoExpiration {
		duration = 0
	}

	return c.redis.SetNX(ctx, key, value, duration).Result()
}

// GetTTL returns the remaining time-to-live of a key
func (c *CacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.redis.TTL(ctx, key).Result()
}

// Increment increments the integer value of a key by one
func (c *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	return c.redis.Incr(ctx, key).Result()
}

// IncrementBy increments the integer value of a key by the given amount
func (c *CacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.redis.IncrBy(ctx, key, value).Result()
}

// Keys returns all keys matching the pattern
func (c *CacheService) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.redis.Keys(ctx, pattern).Result()
}
