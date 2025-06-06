package cache

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/config"
	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	redisCtx        = context.Background()
	redisRetryCount = 0
	maxRetries      = 3
	redisMux        sync.Mutex
	redisInstance   *redis.Client
	redisOnce       sync.Once
)

// RedisInstance returns a singleton instance of the Redis client.
func RedisInstance() *redis.Client {
	redisOnce.Do(func() {
		handlePanic()

		redisInstance = newRedisClient()
	})

	return redisInstance
}

// newRedisClient initializes a new Redis client using Sentinel configuration.
func newRedisClient() *redis.Client {
	redisConf := config.Instance().Redis

	ports := strings.Split(redisConf.SentinelPorts, ",")
	sentinelAddrs := make([]string, len(ports))
	for i := range ports {
		sentinelAddrs[i] = fmt.Sprintf("%s:%s", redisConf.Host, strings.TrimSpace(ports[i]))
	}

	logger.Info("Initializing RedisSentinel",
		zap.String("master_name", redisConf.MasterName),
		zap.Strings("sentinel_addrs", sentinelAddrs),
	)

	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    redisConf.MasterName,
		SentinelAddrs: sentinelAddrs,
		DB:            redisConf.Database,
		Password:      redisConf.Password,
	})

	// Check the connection
	_, err := rdb.Ping(redisCtx).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis Sentinel", zap.Error(err))
	}

	// Set a test key to verify the connection
	err = rdb.Set(redisCtx, "test_key", "Redis Sentinel Instance!", 0).Err()
	if err != nil {
		logger.Fatal("Error setting key", zap.Error(err))
	}

	logger.Info("Initializing RedisSentinel Successfully")

	return rdb
}

// handlePanic recovers from panics that occur during Redis operations.
func handlePanic() {
	defer func() {
		if r := recover(); r != nil {
			redisMux.Lock()
			defer redisMux.Unlock()

			logger.Error("Recovered from Redis panic",
				zap.Any("error", r),
				zap.Int("retry_count", redisRetryCount),
				zap.Int("maxRetries", maxRetries),
				zap.String("stack", string(debug.Stack())),
			)

			if redisRetryCount < maxRetries {
				redisRetryCount++
				backoff := time.Duration(redisRetryCount*redisRetryCount) * time.Second
				logger.Warn("Retrying Redis connection...",
					zap.Int("attempt", redisRetryCount),
					zap.Duration("backoff", backoff),
				)
				time.Sleep(backoff)

				redisInstance = newRedisClient()

				return
			}

			logger.Fatal("Redis connection failed after max retries")
		}
	}()
}
