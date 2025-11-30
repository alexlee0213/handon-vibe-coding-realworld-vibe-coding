package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alexlee0213/realworld-conduit/backend/internal/api/handler"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// Auth creates a middleware that requires authentication
// It validates the JWT token and adds the user ID to the request context
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := extractToken(r)
			if !ok {
				writeUnauthorizedError(w)
				return
			}

			userID, err := authService.ValidateToken(token)
			if err != nil {
				writeUnauthorizedError(w)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), handler.UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth creates a middleware that optionally authenticates
// If a valid token is provided, the user ID is added to context
// If no token or invalid token, the request continues without user ID
func OptionalAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := extractToken(r)
			if !ok {
				// No token, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			userID, err := authService.ValidateToken(token)
			if err != nil {
				// Invalid token, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), handler.UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts the JWT token from the Authorization header
// Expected format: "Token <jwt-token>"
func extractToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Token" {
		return "", false
	}

	return parts[1], true
}

// writeUnauthorizedError writes a 401 Unauthorized response
func writeUnauthorizedError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"errors":{"token":["authorization required"]}}`))
}
