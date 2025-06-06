package cache

import (
	"context"
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

type CacheService struct {
	redis *redis.Client
}

func NewCacheService() abstract.ICacheService {
	return &CacheService{
		redis: RedisInstance(),
	}
}

// Clear implements abstract.ICacheService
func (c *CacheService) Clear(ctx context.Context) error {
	if err := c.redis.FlushAll(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// Delete implements abstract.ICacheService
func (c *CacheService) Delete(ctx context.Context, key string) error {
	if err := c.redis.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

// Exists implements abstract.ICacheService
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// Get implements abstract.ICacheService
func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

// GetBit implements abstract.ICacheService
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

// Set implements abstract.ICacheService
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

// SetBit implements abstract.ICacheService
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

// GetWithDefault implements abstract.ICacheService
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

// SetWithDefaultExpiration implements abstract.ICacheService
func (c *CacheService) SetWithDefaultExpiration(ctx context.Context, key string, value any) error {
	return c.Set(ctx, key, value, DefaultExpiration)
}

// SetNX implements abstract.ICacheService
func (c *CacheService) SetNX(ctx context.Context, key string, value any, expiration int) (bool, error) {
	duration := time.Duration(expiration) * time.Second
	if expiration == NoExpiration {
		duration = 0
	}

	return c.redis.SetNX(ctx, key, value, duration).Result()
}

// GetTTL implements abstract.ICacheService
func (c *CacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.redis.TTL(ctx, key).Result()
}

// Increment implements abstract.ICacheService
func (c *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	return c.redis.Incr(ctx, key).Result()
}

// IncrementBy implements abstract.ICacheService
func (c *CacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.redis.IncrBy(ctx, key, value).Result()
}

// Keys implements abstract.ICacheService
func (c *CacheService) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.redis.Keys(ctx, pattern).Result()
}
