package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs each HTTP request using slog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
		}
		if rid, ok := c.Get("request_id"); ok {
			if rs, ok := rid.(string); ok && rs != "" {
				attrs = append(attrs, slog.String("request_id", rs))
			}
		}
		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		args := make([]any, len(attrs))
		for i, a := range attrs {
			args[i] = a
		}

		if status >= 500 {
			slog.Error("request", args...)
		} else if status >= 400 {
			slog.Warn("request", args...)
		} else {
			slog.Info("request", args...)
		}
	}
}
