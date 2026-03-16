package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAdmin returns a middleware that ensures the user has system admin+ role.
// Must be used after the Auth middleware which sets "role" in the context.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" && role != "owner" && role != "system_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "admin access required",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
