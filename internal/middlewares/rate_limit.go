package middlewares

import (
	"context"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"github.com/SekiroKenjii/go-blog-engine/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

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

func RateLimit() gin.HandlerFunc {
	logger := logger.Instance()
	ctx := context.Background()

	// Clean up visitors every minute
	// and remove those not seen for 3 minutes
	go startCleanup(ctx, 3*time.Minute, 1*time.Minute)

	return func(c *gin.Context) {
		v := getVisitor(c.ClientIP(), c.Request.Method)

		if !v.limiter.Allow() {
			logger.Warn("Rate limit exceeded", zap.String("ip", c.ClientIP()), zap.String("path", c.Request.URL.Path))
			response.TooManyRequest(c)

			return
		}

		c.Next()
	}
}
