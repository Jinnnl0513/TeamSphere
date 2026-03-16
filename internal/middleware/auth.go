package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth returns a JWT authentication middleware.
// It extracts the Bearer token from the Authorization header,
// validates it, checks the blacklist, and sets "claims" and "user_id" in the context.
func Auth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithError(c, http.StatusUnauthorized, 40101, "missing authorization header")
			return
		}

		tokenStr, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || tokenStr == "" {
			abortWithError(c, http.StatusUnauthorized, 40101, "invalid authorization format")
			return
		}

		claims, err := authService.ValidateToken(c.Request.Context(), tokenStr)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				abortWithError(c, http.StatusUnauthorized, 40102, "token expired")
				return
			}
			abortWithError(c, http.StatusUnauthorized, 40101, "invalid token")
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func abortWithError(c *gin.Context, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
