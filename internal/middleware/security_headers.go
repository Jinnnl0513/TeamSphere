package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds standard HTTP security response headers to every response.
//
// Header coverage:
//   - X-Frame-Options: DENY          — prevents clickjacking via <iframe>
//   - X-Content-Type-Options: nosniff — prevents MIME-type sniffing attacks
//   - X-XSS-Protection: 1; mode=block — legacy IE/Chrome XSS filter (belt-and-suspenders)
//   - Referrer-Policy                 — avoids leaking URL info to third-party origins
//   - Content-Security-Policy         — restricts inline scripts/styles; tighten per-feature
//   - Permissions-Policy              — opt-out of browser features the app doesn't use
//
// NOTE: HSTS (Strict-Transport-Security) is intentionally omitted here because
// it must only be sent over HTTPS connections. Enable it in your reverse proxy
// (Nginx/Caddy) or add an HTTPS-only middleware layer in production.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clickjacking protection
		c.Header("X-Frame-Options", "DENY")

		// MIME-sniffing prevention
		c.Header("X-Content-Type-Options", "nosniff")

		// Legacy XSS filter (belt-and-suspenders; modern browsers rely on CSP instead)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer: send origin only for same-origin requests, nothing for cross-origin
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		// 'self' + 'unsafe-inline' for styles is needed by inline Tailwind/emotion at runtime.
		// Adjust script-src if you add external analytics/widgets.
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src 'self' https://fonts.gstatic.com data:; "+
				"img-src 'self' data: blob: https:; "+
				"media-src 'self' blob:; "+
				"connect-src 'self' ws: wss:; "+
				"frame-ancestors 'none'",
		)

		// Restrict browser feature access (opt-out of sensitive APIs)
		c.Header("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=(), usb=()")

		c.Next()
	}
}
