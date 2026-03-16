package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LimitBody returns a middleware that rejects requests whose body exceeds maxBytes.
// This prevents memory exhaustion from oversized JSON/form payloads.
// Note: multipart file uploads should be controlled via gin.Engine.MaxMultipartMemory instead.
func LimitBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":    41300,
				"message": "请求体过大",
			})
			return
		}
		// Also wrap the body with LimitReader to protect against lying Content-Length
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
