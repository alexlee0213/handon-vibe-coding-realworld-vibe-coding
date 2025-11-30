package service

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
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

func newTestAuthService(t *testing.T) (*AuthService, *sql.DB) {
	t.Helper()
	db := setupTestDB(t)
	logger := newTestLogger()
	userRepo := repository.NewSQLiteUserRepository(db, logger)

	authService := NewAuthService(userRepo, "test-jwt-secret", 24*time.Hour, logger)
	return authService, db
}

// =============================================================================
// TDD: Register Tests
// =============================================================================

func TestRegister(t *testing.T) {
	t.Run("successfully registers a new user", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()
		input := &domain.CreateUserInput{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "password123",
		}

		user, token, err := authService.Register(ctx, input)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if user == nil {
			t.Fatal("expected user to be returned")
		}
		if user.Email != input.Email {
			t.Errorf("expected email %s, got %s", input.Email, user.Email)
		}
		if user.Username != input.Username {
			t.Errorf("expected username %s, got %s", input.Username, user.Username)
		}
		if token == "" {
			t.Error("expected token to be returned")
		}
		// Password should be hashed, not stored plain
		if user.PasswordHash == input.Password {
			t.Error("password should be hashed, not stored plain")
		}
	})

	t.Run("returns error for duplicate email", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()
		input := &domain.CreateUserInput{
			Email:    "duplicate@example.com",
			Username: "user1",
			Password: "password123",
		}

		_, _, err := authService.Register(ctx, input)
		if err != nil {
			t.Fatalf("first registration should succeed: %v", err)
		}

		input2 := &domain.CreateUserInput{
			Email:    "duplicate@example.com",
			Username: "user2",
			Password: "password123",
		}

		_, _, err = authService.Register(ctx, input2)
		if err == nil {
			t.Error("expected error for duplicate email")
		}
		if err != domain.ErrEmailAlreadyTaken {
			t.Errorf("expected ErrEmailAlreadyTaken, got %v", err)
		}
	})

	t.Run("returns error for duplicate username", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()
		input := &domain.CreateUserInput{
			Email:    "user1@example.com",
			Username: "duplicateuser",
			Password: "password123",
		}

		_, _, err := authService.Register(ctx, input)
		if err != nil {
			t.Fatalf("first registration should succeed: %v", err)
		}

		input2 := &domain.CreateUserInput{
			Email:    "user2@example.com",
			Username: "duplicateuser",
			Password: "password123",
		}

		_, _, err = authService.Register(ctx, input2)
		if err == nil {
			t.Error("expected error for duplicate username")
		}
		if err != domain.ErrUsernameAlreadyTaken {
			t.Errorf("expected ErrUsernameAlreadyTaken, got %v", err)
		}
	})

	t.Run("returns validation error for empty email", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()
		input := &domain.CreateUserInput{
			Email:    "",
			Username: "testuser",
			Password: "password123",
		}

		_, _, err := authService.Register(ctx, input)
		if err == nil {
			t.Error("expected validation error for empty email")
		}
	})

	t.Run("returns validation error for empty username", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()
		input := &domain.CreateUserInput{
			Email:    "test@example.com",
			Username: "",
			Password: "password123",
		}

		_, _, err := authService.Register(ctx, input)
		if err == nil {
			t.Error("expected validation error for empty username")
		}
	})

	t.Run("returns validation error for empty password", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()
		input := &domain.CreateUserInput{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "",
		}

		_, _, err := authService.Register(ctx, input)
		if err == nil {
			t.Error("expected validation error for empty password")
		}
	})
}

// =============================================================================
// TDD: Login Tests
// =============================================================================

func TestLogin(t *testing.T) {
	t.Run("successfully logs in with correct credentials", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		// First register a user
		registerInput := &domain.CreateUserInput{
			Email:    "login@example.com",
			Username: "loginuser",
			Password: "password123",
		}
		_, _, err := authService.Register(ctx, registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Then try to login
		user, token, err := authService.Login(ctx, "login@example.com", "password123")

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if user == nil {
			t.Fatal("expected user to be returned")
		}
		if user.Email != registerInput.Email {
			t.Errorf("expected email %s, got %s", registerInput.Email, user.Email)
		}
		if token == "" {
			t.Error("expected token to be returned")
		}
	})

	t.Run("returns error for wrong password", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		// First register a user
		registerInput := &domain.CreateUserInput{
			Email:    "wrongpass@example.com",
			Username: "wrongpassuser",
			Password: "correctpassword",
		}
		_, _, err := authService.Register(ctx, registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Try to login with wrong password
		_, _, err = authService.Login(ctx, "wrongpass@example.com", "wrongpassword")

		if err == nil {
			t.Error("expected error for wrong password")
		}
		if err != domain.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		_, _, err := authService.Login(ctx, "nonexistent@example.com", "password123")

		if err == nil {
			t.Error("expected error for non-existent email")
		}
		if err != domain.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

// =============================================================================
// TDD: JWT Token Tests
// =============================================================================

func TestGenerateToken(t *testing.T) {
	t.Run("generates a valid JWT token", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		token, err := authService.GenerateToken(123)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected token to be returned")
		}
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("validates a valid token", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		// Generate a token
		token, err := authService.GenerateToken(123)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Validate the token
		userID, err := authService.ValidateToken(token)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if userID != 123 {
			t.Errorf("expected userID 123, got %d", userID)
		}
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		_, err := authService.ValidateToken("invalid.token.here")

		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("returns error for expired token", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		logger := newTestLogger()
		userRepo := repository.NewSQLiteUserRepository(db, logger)

		// Create service with very short expiry
		authService := NewAuthService(userRepo, "test-jwt-secret", -1*time.Hour, logger)

		// Generate a token (already expired)
		token, err := authService.GenerateToken(123)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Validate the token - should fail
		_, err = authService.ValidateToken(token)

		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("returns error for token with wrong secret", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		logger := newTestLogger()
		userRepo := repository.NewSQLiteUserRepository(db, logger)

		// Create two services with different secrets
		authService1 := NewAuthService(userRepo, "secret1", 24*time.Hour, logger)
		authService2 := NewAuthService(userRepo, "secret2", 24*time.Hour, logger)

		// Generate a token with service1
		token, err := authService1.GenerateToken(123)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Try to validate with service2 (different secret)
		_, err = authService2.ValidateToken(token)

		if err == nil {
			t.Error("expected error for token with wrong secret")
		}
	})
}

// =============================================================================
// TDD: GetCurrentUser Tests
// =============================================================================

func TestGetCurrentUser(t *testing.T) {
	t.Run("gets current user by ID", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		// Register a user
		registerInput := &domain.CreateUserInput{
			Email:    "current@example.com",
			Username: "currentuser",
			Password: "password123",
		}
		user, _, err := authService.Register(ctx, registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Get current user
		currentUser, err := authService.GetCurrentUser(ctx, user.ID)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if currentUser == nil {
			t.Fatal("expected user to be returned")
		}
		if currentUser.Email != registerInput.Email {
			t.Errorf("expected email %s, got %s", registerInput.Email, currentUser.Email)
		}
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		_, err := authService.GetCurrentUser(ctx, 999999)

		if err == nil {
			t.Error("expected error for non-existent user")
		}
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =============================================================================
// TDD: UpdateUser Tests
// =============================================================================

func TestUpdateUser(t *testing.T) {
	t.Run("updates user email", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		// Register a user
		registerInput := &domain.CreateUserInput{
			Email:    "update@example.com",
			Username: "updateuser",
			Password: "password123",
		}
		user, _, err := authService.Register(ctx, registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Update email
		newEmail := "newemail@example.com"
		updateInput := &domain.UpdateUserInput{
			Email: &newEmail,
		}

		updatedUser, err := authService.UpdateUser(ctx, user.ID, updateInput)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if updatedUser.Email != newEmail {
			t.Errorf("expected email %s, got %s", newEmail, updatedUser.Email)
		}
	})

	t.Run("updates user password", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		// Register a user
		registerInput := &domain.CreateUserInput{
			Email:    "passupdate@example.com",
			Username: "passupdateuser",
			Password: "oldpassword",
		}
		user, _, err := authService.Register(ctx, registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Update password
		newPassword := "newpassword"
		updateInput := &domain.UpdateUserInput{
			Password: &newPassword,
		}

		_, err = authService.UpdateUser(ctx, user.ID, updateInput)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify login with new password works
		_, _, err = authService.Login(ctx, "passupdate@example.com", "newpassword")
		if err != nil {
			t.Errorf("login with new password should work: %v", err)
		}

		// Verify login with old password fails
		_, _, err = authService.Login(ctx, "passupdate@example.com", "oldpassword")
		if err == nil {
			t.Error("login with old password should fail")
		}
	})

	t.Run("updates user bio and image", func(t *testing.T) {
		authService, db := newTestAuthService(t)
		defer db.Close()

		ctx := context.Background()

		// Register a user
		registerInput := &domain.CreateUserInput{
			Email:    "bioupdate@example.com",
			Username: "bioupdateuser",
			Password: "password123",
		}
		user, _, err := authService.Register(ctx, registerInput)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Update bio and image
		newBio := "This is my bio"
		newImage := "https://example.com/avatar.png"
		updateInput := &domain.UpdateUserInput{
			Bio:   &newBio,
			Image: &newImage,
		}

		updatedUser, err := authService.UpdateUser(ctx, user.ID, updateInput)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if updatedUser.Bio != newBio {
			t.Errorf("expected bio %s, got %s", newBio, updatedUser.Bio)
		}
		if updatedUser.Image != newImage {
			t.Errorf("expected image %s, got %s", newImage, updatedUser.Image)
		}
	})
}
