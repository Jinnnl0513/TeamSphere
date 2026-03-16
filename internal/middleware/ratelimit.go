package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/teamsphere/server/internal/contract/ratelimit"
	"github.com/gin-gonic/gin"
)

// RateLimit returns a middleware that limits requests by client IP.
func RateLimit(limiter ratelimit.Limiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		key := fmt.Sprintf("%s|%s|%d|%d", path, c.ClientIP(), limit, int64(window.Seconds()))
		allowed, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			slog.Warn("rate limiter error, allowing request", "error", err)
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    42901,
				"message": "请求过于频繁，请稍后再试",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
