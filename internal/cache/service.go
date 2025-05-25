package cache

import (
	"context"
	"sync"

	"github.com/SekiroKenjii/go-blog-engine/internal/abstract"
	"github.com/redis/go-redis/v9"
)

var (
	cacheSvcInstance abstract.ICacheService
	cacheSvcOnce     sync.Once
)

type CacheService struct {
	Redis *redis.Client
}

func CacheServiceInstance() abstract.ICacheService {
	cacheSvcOnce.Do(func() {
		cacheSvcInstance = newCacheService()
	})

	return cacheSvcInstance
}

func newCacheService() abstract.ICacheService {
	return &CacheService{
		Redis: RedisInstance(),
	}
}

// Clear implements abstract.ICacheService.
func (c *CacheService) Clear(context.Context) error {
	// TODO: Implement the Clear method to flush all keys in Redis.
	panic("unimplemented")
}

// Delete implements abstract.ICacheService.
func (c *CacheService) Delete(context.Context, string) error {
	// TODO: Implement the Delete method to remove a key from Redis.
	panic("unimplemented")
}

// Exists implements abstract.ICacheService.
func (c *CacheService) Exists(context.Context, string) (bool, error) {
	// TODO: Implement the Exists method to check if a key exists in Redis.
	panic("unimplemented")
}

// Get implements abstract.ICacheService.
func (c *CacheService) Get(context.Context, string) (string, error) {
	// TODO: Implement the Get method to retrieve data from Redis.
	panic("unimplemented")
}

// Set implements abstract.ICacheService.
func (c *CacheService) Set(context.Context, string, any, int) error {
	// TODO: Implement the Set method to store data in Redis with an expiration time.
	panic("unimplemented")
}
