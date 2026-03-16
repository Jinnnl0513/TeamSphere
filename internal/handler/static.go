package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

// StaticUploads serves user-uploaded files from the local filesystem.
// Fix #3: Validates the filepath to prevent directory traversal (e.g. ../../etc/passwd).
func StaticUploads(uploadDir string) gin.HandlerFunc {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		slog.Error("failed to create upload directory", "dir", uploadDir, "error", err)
	}
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		slog.Error("failed to resolve upload directory", "dir", uploadDir, "error", err)
		absUploadDir = uploadDir
	}
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir)))

	return func(c *gin.Context) {
		fp := c.Param("filepath")

		// Reject paths with traversal sequences before serving
		cleaned := filepath.Clean(fp)
		if strings.Contains(cleaned, "..") {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Double-check resolved absolute path stays within upload dir
		abs, err := filepath.Abs(filepath.Join(absUploadDir, cleaned))
		if err != nil || !strings.HasPrefix(abs, absUploadDir+string(filepath.Separator)) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// StaticUploadsWithAuth serves uploads with either a signed token or a valid JWT.
func StaticUploadsWithAuth(uploadDir string, secret string, authService *service.AuthService) gin.HandlerFunc {
	base := StaticUploads(uploadDir)
	return func(c *gin.Context) {
		// Allow bearer auth
		if authService != nil {
			token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
			if token != "" {
				if _, err := authService.ValidateToken(c.Request.Context(), token); err == nil {
					base(c)
					return
				}
			}
		}

		// Fallback to signed token
		expStr := c.Query("exp")
		token := c.Query("token")
		if expStr == "" || token == "" || secret == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		exp, err := strconv.ParseInt(expStr, 10, 64)
		if err != nil || exp <= 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if time.Now().Unix() > exp {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		fp := strings.TrimPrefix(c.Param("filepath"), "/")
		if !validateFileToken(fp, expStr, token, secret) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		base(c)
	}
}

func validateFileToken(filePath, exp, token, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(filePath))
	mac.Write([]byte("|"))
	mac.Write([]byte(exp))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(token))
}

// SPA serves the embedded frontend SPA with fallback to index.html for client-side routing.
func SPA(embedFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(embedFS))

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API and WebSocket routes
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws") {
			c.Next()
			return
		}

		// Skip uploads (served separately)
		if strings.HasPrefix(path, "/uploads/") {
			c.Next()
			return
		}

		// Try to serve the exact file
		f, err := embedFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Fallback to index.html for SPA client-side routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
