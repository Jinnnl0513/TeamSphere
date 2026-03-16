package middleware

import (
	"net/http"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

// RequireAdmin2FA enforces 2FA for admin routes based on policy.
func RequireAdmin2FA(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if !authService.Is2FARequiredForRole(c.Request.Context(), role) {
			c.Next()
			return
		}
		userID := c.GetInt64("user_id")
		enabled, err := authService.GetTOTPStatus(c.Request.Context(), userID)
		if err != nil || !enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    40302,
				"message": "2fa required for admin access",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
