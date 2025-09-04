package middlewares

import (
	"context"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const rateLimitExceededMsg = "Rate limit exceeded"

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors    = make(map[string]*visitor)
	visitorsMux sync.Mutex
)

func newGetVisitor() *visitor {
	return &visitor{
		limiter:  rate.NewLimiter(rate.Every(2*time.Second), 10), // 10 req/2s
		lastSeen: time.Now(),
	}
}

func newPostVisitor() *visitor {
	return &visitor{
		limiter:  rate.NewLimiter(rate.Every(time.Minute), 30), // 30 req/1m
		lastSeen: time.Now(),
	}
}

func getVisitor(ip string, method string) *visitor {
	visitorsMux.Lock()
	defer visitorsMux.Unlock()

	key := ip + method
	v, exists := visitors[key]

	if !exists {
		if method == "POST" {
			v = newPostVisitor()
		} else {
			v = newGetVisitor()
		}

		visitors[key] = v
	}

	v.lastSeen = time.Now()

	return v
}

// startCleanup periodically cleans up old visitors
// based on the TTL (time-to-live) and interval.
// It removes visitors that have not been seen for the TTL duration.
func startCleanup(ctx context.Context, ttl, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			visitorsMux.Lock()
			now := time.Now()
			for key, v := range visitors {
				if now.Sub(v.lastSeen) > ttl {
					delete(visitors, key)
				}
			}
			visitorsMux.Unlock()
		}
	}
}

// RateLimitWithConfig creates a rate limiter with custom configuration
func RateLimitWithConfig(requestsPerSecond float64, burstSize int) gin.HandlerFunc {
	visitors := make(map[string]*visitor)
	visitorsMux := sync.Mutex{}
	ctx := context.Background()

	// Custom cleanup function for this instance
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				visitorsMux.Lock()
				now := time.Now()
				for key, v := range visitors {
					if now.Sub(v.lastSeen) > 3*time.Minute {
						delete(visitors, key)
					}
				}
				visitorsMux.Unlock()
			}
		}
	}()

	return func(c *gin.Context) {
		clientIP := utils.ExtractIPAddress(c.Request)

		visitorsMux.Lock()
		v, exists := visitors[clientIP]
		if !exists {
			v = &visitor{
				limiter:  rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize),
				lastSeen: time.Now(),
			}
			visitors[clientIP] = v
		}
		v.lastSeen = time.Now()
		visitorsMux.Unlock()

		if !v.limiter.Allow() {
			logger.Warn(rateLimitExceededMsg,
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
				zap.Float64("rate", requestsPerSecond),
				zap.Int("burst", burstSize))
			response.TooManyRequest(c)
			return
		}

		c.Next()
	}
}

// AuthRateLimit provides stricter rate limiting for authentication endpoints
func AuthRateLimit() gin.HandlerFunc {
	// 5 requests per minute for auth endpoints
	return RateLimitWithConfig(5.0/60.0, 5)
}

// RateLimitExcludingPaths creates a rate limiter that excludes specific path prefixes
func RateLimitExcludingPaths(excludePaths ...string) gin.HandlerFunc {
	ctx := context.Background()

	// Clean up visitors every minute
	// and remove those not seen for 3 minutes
	go startCleanup(ctx, 3*time.Minute, 1*time.Minute)

	return func(c *gin.Context) {
		// Check if the current path should be excluded
		currentPath := c.Request.URL.Path
		for _, excludePath := range excludePaths {
			if len(currentPath) >= len(excludePath) && currentPath[:len(excludePath)] == excludePath {
				// Skip rate limiting for this path
				c.Next()
				return
			}
		}

		clientIP := utils.ExtractIPAddress(c.Request)
		v := getVisitor(clientIP, c.Request.Method)

		if !v.limiter.Allow() {
			logger.Warn(rateLimitExceededMsg, zap.String("ip", clientIP), zap.String("path", c.Request.URL.Path))
			response.TooManyRequest(c)
			return
		}

		c.Next()
	}
}
