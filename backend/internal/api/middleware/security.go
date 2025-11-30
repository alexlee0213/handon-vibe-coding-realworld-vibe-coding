package middleware

import (
	"net/http"
)

// Security creates a middleware that adds security headers to responses
func Security() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Enable XSS filter in browsers
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// Referrer policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy (API only, strict)
			w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			// Strict Transport Security (only in production with HTTPS)
			// This header will be set by the reverse proxy/load balancer in production

			next.ServeHTTP(w, r)
		})
	}
}
