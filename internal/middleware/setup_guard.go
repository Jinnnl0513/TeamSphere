package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupGuard blocks all non-setup API routes when the system is not initialized.
// It allows: /api/v1/health, /api/v1/setup/*, and non-API paths (SPA, static).
func SetupGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Allow non-API paths (SPA static files, uploads)
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		// Allow health check and setup routes
		if path == "/api/v1/health" || strings.HasPrefix(path, "/api/v1/setup") {
			c.Next()
			return
		}

		// Block everything else
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"code":    50001,
			"message": "\u7cfb\u7edf\u5c1a\u672a\u521d\u59cb\u5316\uff0c\u8bf7\u5148\u5b8c\u6210\u521d\u59cb\u5316",
			"data":    nil,
		})
	}
}
