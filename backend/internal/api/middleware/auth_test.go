package middleware

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/api/handler"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// Test helpers
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			bio TEXT DEFAULT '',
			image TEXT DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	return db
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func newTestAuthService(t *testing.T) (*service.AuthService, *sql.DB) {
	t.Helper()
	db := setupTestDB(t)
	logger := newTestLogger()
	userRepo := repository.NewSQLiteUserRepository(db, logger)
	authService := service.NewAuthService(userRepo, "test-jwt-secret", 24*time.Hour, logger)
	return authService, db
}

// =============================================================================
// TDD: Auth Middleware Tests
// =============================================================================

func TestAuthMiddleware(t *testing.T) {
	t.Run("allows request with valid token", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		// Generate a valid token
		token, err := authService.GenerateToken(123)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Create test handler that checks for user ID in context
		var capturedUserID int64
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(handler.UserIDContextKey).(int64)
			if ok {
				capturedUserID = userID
			}
			w.WriteHeader(http.StatusOK)
		})

		// Apply auth middleware
		middleware := Auth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		req.Header.Set("Authorization", "Token "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if capturedUserID != 123 {
			t.Errorf("expected user ID 123, got %d", capturedUserID)
		}
	})

	t.Run("returns 401 for missing Authorization header", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := Auth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns 401 for invalid token format", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := Auth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		req.Header.Set("Authorization", "Bearer token") // Wrong format, should be "Token"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns 401 for invalid token", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := Auth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		req.Header.Set("Authorization", "Token invalid.token.here")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestOptionalAuthMiddleware(t *testing.T) {
	t.Run("allows request without token", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := OptionalAuth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("sets user ID in context when valid token provided", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		// Generate a valid token
		token, err := authService.GenerateToken(456)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		var capturedUserID int64
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(handler.UserIDContextKey).(int64)
			if ok {
				capturedUserID = userID
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware := OptionalAuth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
		req.Header.Set("Authorization", "Token "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if capturedUserID != 456 {
			t.Errorf("expected user ID 456, got %d", capturedUserID)
		}
	})

	t.Run("continues without user ID when token is invalid", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		var hasUserID bool
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, hasUserID = r.Context().Value(handler.UserIDContextKey).(int64)
			w.WriteHeader(http.StatusOK)
		})

		middleware := OptionalAuth(authService)
		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
		req.Header.Set("Authorization", "Token invalid.token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if hasUserID {
			t.Error("expected no user ID in context for invalid token")
		}
	})
}
