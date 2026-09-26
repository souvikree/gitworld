package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit applies a simple global token-bucket limiter.
// Per-IP limiting is the better long-term answer, but a global limiter
// is a reasonable first line of defense and simple to reason about.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	var mu sync.Mutex

	return func(c *gin.Context) {
		mu.Lock()
		allowed := limiter.Allow()
		mu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}