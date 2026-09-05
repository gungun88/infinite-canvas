package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tigerowo/infinite-canvas/handler"
)

// MaxBodyBytes limits request bodies before handlers parse multipart or JSON data.
func MaxBodyBytes(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

type rateLimitBucket struct {
	started time.Time
	count   int
}

type rateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	limit   int
	buckets map[string]rateLimitBucket
}

func RateLimit(window time.Duration, limit int) gin.HandlerFunc {
	limiter := &rateLimiter{window: window, limit: limit, buckets: map[string]rateLimitBucket{}}
	return func(c *gin.Context) {
		key := requestIP(c)
		if !limiter.allow(key) {
			c.Header("Retry-After", strconvItoa(int(window.Seconds())))
			handler.FailWithStatus(c.Writer, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (limiter *rateLimiter) allow(key string) bool {
	now := time.Now()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	bucket := limiter.buckets[key]
	if bucket.started.IsZero() || now.Sub(bucket.started) >= limiter.window {
		limiter.buckets[key] = rateLimitBucket{started: now, count: 1}
		return true
	}
	if bucket.count >= limiter.limit {
		return false
	}
	bucket.count++
	limiter.buckets[key] = bucket
	return true
}

func requestIP(c *gin.Context) string {
	return c.ClientIP()
}

func strconvItoa(value int) string {
	if value <= 0 {
		return "1"
	}
	return fmt.Sprintf("%d", value)
}
