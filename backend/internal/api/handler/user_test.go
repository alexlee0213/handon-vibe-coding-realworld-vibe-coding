package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
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
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_username ON users(username);
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

type testSetup struct {
	handler     *UserHandler
	authService *service.AuthService
	db          *sql.DB
}

func newTestUserHandler(t *testing.T) *testSetup {
	t.Helper()
	db := setupTestDB(t)
	logger := newTestLogger()
	userRepo := repository.NewSQLiteUserRepository(db, logger)
	authService := service.NewAuthService(userRepo, "test-jwt-secret", 24*time.Hour, logger)
	userHandler := NewUserHandler(authService, logger)

	return &testSetup{
		handler:     userHandler,
		authService: authService,
		db:          db,
	}
}

// =============================================================================
// TDD: POST /api/users (Register) Tests
// =============================================================================

func TestRegisterHandler(t *testing.T) {
	t.Run("successfully registers a new user", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		body := `{"user":{"username":"testuser","email":"test@example.com","password":"password123"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.Register(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		user, ok := response["user"].(map[string]interface{})
		if !ok {
			t.Fatal("expected user object in response")
		}

		if user["email"] != "test@example.com" {
			t.Errorf("expected email test@example.com, got %v", user["email"])
		}
		if user["username"] != "testuser" {
			t.Errorf("expected username testuser, got %v", user["username"])
		}
		if user["token"] == nil || user["token"] == "" {
			t.Error("expected token in response")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		body := `invalid json`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.Register(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("returns error for missing fields", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		body := `{"user":{"email":"test@example.com"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.Register(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("returns error for duplicate email", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		// First registration
		body := `{"user":{"username":"user1","email":"duplicate@example.com","password":"password123"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		setup.handler.Register(w, req)

		// Second registration with same email
		body = `{"user":{"username":"user2","email":"duplicate@example.com","password":"password123"}}`
		req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		setup.handler.Register(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})
}

// =============================================================================
// TDD: POST /api/users/login Tests
// =============================================================================

func TestLoginHandler(t *testing.T) {
	t.Run("successfully logs in with correct credentials", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		// First register a user
		ctx := context.Background()
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "login@example.com",
			Username: "loginuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Then login
		body := `{"user":{"email":"login@example.com","password":"password123"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		user, ok := response["user"].(map[string]interface{})
		if !ok {
			t.Fatal("expected user object in response")
		}

		if user["token"] == nil || user["token"] == "" {
			t.Error("expected token in response")
		}
	})

	t.Run("returns error for wrong password", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		// First register a user
		ctx := context.Background()
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "wrongpass@example.com",
			Username: "wrongpassuser",
			Password: "correctpassword",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Try login with wrong password
		body := `{"user":{"email":"wrongpass@example.com","password":"wrongpassword"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.Login(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		body := `{"user":{"email":"nonexistent@example.com","password":"password123"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.Login(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})
}

// =============================================================================
// TDD: GET /api/user (Current User) Tests
// =============================================================================

func TestGetCurrentUserHandler(t *testing.T) {
	t.Run("successfully gets current user with valid token", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		// Register and get token
		ctx := context.Background()
		user, token, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "current@example.com",
			Username: "currentuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		// Add user ID to context (simulating auth middleware)
		ctx = context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.GetCurrentUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		userResp, ok := response["user"].(map[string]interface{})
		if !ok {
			t.Fatal("expected user object in response")
		}

		if userResp["email"] != "current@example.com" {
			t.Errorf("expected email current@example.com, got %v", userResp["email"])
		}

		_ = token // token is valid but we're simulating middleware
	})

	t.Run("returns error when user ID not in context", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
		w := httptest.NewRecorder()

		setup.handler.GetCurrentUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

// =============================================================================
// TDD: PUT /api/user (Update User) Tests
// =============================================================================

func TestUpdateUserHandler(t *testing.T) {
	t.Run("successfully updates user email", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		// Register a user
		ctx := context.Background()
		user, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "update@example.com",
			Username: "updateuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		body := `{"user":{"email":"newemail@example.com"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/user", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx = context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UpdateUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		userResp, ok := response["user"].(map[string]interface{})
		if !ok {
			t.Fatal("expected user object in response")
		}

		if userResp["email"] != "newemail@example.com" {
			t.Errorf("expected email newemail@example.com, got %v", userResp["email"])
		}
	})

	t.Run("successfully updates user bio and image", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		// Register a user
		ctx := context.Background()
		user, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "bio@example.com",
			Username: "biouser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		body := `{"user":{"bio":"My bio","image":"https://example.com/avatar.png"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/user", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx = context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UpdateUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		userResp, ok := response["user"].(map[string]interface{})
		if !ok {
			t.Fatal("expected user object in response")
		}

		if userResp["bio"] != "My bio" {
			t.Errorf("expected bio 'My bio', got %v", userResp["bio"])
		}
		if userResp["image"] != "https://example.com/avatar.png" {
			t.Errorf("expected image url, got %v", userResp["image"])
		}
	})

	t.Run("returns error when user ID not in context", func(t *testing.T) {
		setup := newTestUserHandler(t)
		defer setup.db.Close()

		body := `{"user":{"email":"newemail@example.com"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/user", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.UpdateUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}
