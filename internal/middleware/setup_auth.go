package middleware

import "github.com/gin-gonic/gin"

// RequireSetupAccess intentionally leaves setup routes open while the system
// is uninitialized. The entire setup API is removed once initialization
// completes and the server restarts in normal mode.
func RequireSetupAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
