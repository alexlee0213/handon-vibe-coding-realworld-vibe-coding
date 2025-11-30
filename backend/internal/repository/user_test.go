package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create users table
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
		Level: slog.LevelError, // Only show errors in tests
	}))
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db, newTestLogger())
	ctx := context.Background()

	t.Run("successfully creates a user", func(t *testing.T) {
		user := &domain.User{
			Email:        "test@example.com",
			Username:     "testuser",
			PasswordHash: "hashedpassword",
			Bio:          "Test bio",
			Image:        "https://example.com/image.png",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if user.ID == 0 {
			t.Error("expected user ID to be set")
		}

		if user.CreatedAt.IsZero() {
			t.Error("expected created_at to be set")
		}

		if user.UpdatedAt.IsZero() {
			t.Error("expected updated_at to be set")
		}
	})

	t.Run("returns error for duplicate email", func(t *testing.T) {
		user := &domain.User{
			Email:        "duplicate@example.com",
			Username:     "unique1",
			PasswordHash: "hashedpassword",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("expected no error on first create, got %v", err)
		}

		user2 := &domain.User{
			Email:        "duplicate@example.com",
			Username:     "unique2",
			PasswordHash: "hashedpassword",
		}

		err = repo.CreateUser(ctx, user2)
		if err == nil {
			t.Error("expected error for duplicate email")
		}
		if err != domain.ErrEmailAlreadyTaken {
			t.Errorf("expected ErrEmailAlreadyTaken, got %v", err)
		}
	})

	t.Run("returns error for duplicate username", func(t *testing.T) {
		user := &domain.User{
			Email:        "unique1@example.com",
			Username:     "duplicateuser",
			PasswordHash: "hashedpassword",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("expected no error on first create, got %v", err)
		}

		user2 := &domain.User{
			Email:        "unique2@example.com",
			Username:     "duplicateuser",
			PasswordHash: "hashedpassword",
		}

		err = repo.CreateUser(ctx, user2)
		if err == nil {
			t.Error("expected error for duplicate username")
		}
		if err != domain.ErrUsernameAlreadyTaken {
			t.Errorf("expected ErrUsernameAlreadyTaken, got %v", err)
		}
	})
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db, newTestLogger())
	ctx := context.Background()

	t.Run("successfully gets user by ID", func(t *testing.T) {
		user := &domain.User{
			Email:        "getbyid@example.com",
			Username:     "getbyiduser",
			PasswordHash: "hashedpassword",
			Bio:          "Test bio",
			Image:        "https://example.com/image.png",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		found, err := repo.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if found.Email != user.Email {
			t.Errorf("expected email %s, got %s", user.Email, found.Email)
		}

		if found.Username != user.Username {
			t.Errorf("expected username %s, got %s", user.Username, found.Username)
		}
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, 999999)
		if err == nil {
			t.Error("expected error for non-existent user")
		}
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestGetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db, newTestLogger())
	ctx := context.Background()

	t.Run("successfully gets user by email", func(t *testing.T) {
		user := &domain.User{
			Email:        "getbyemail@example.com",
			Username:     "getbyemailuser",
			PasswordHash: "hashedpassword",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		found, err := repo.GetUserByEmail(ctx, user.Email)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if found.ID != user.ID {
			t.Errorf("expected ID %d, got %d", user.ID, found.ID)
		}
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		_, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("expected error for non-existent email")
		}
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestGetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db, newTestLogger())
	ctx := context.Background()

	t.Run("successfully gets user by username", func(t *testing.T) {
		user := &domain.User{
			Email:        "getbyusername@example.com",
			Username:     "getbyusernameuser",
			PasswordHash: "hashedpassword",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		found, err := repo.GetUserByUsername(ctx, user.Username)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if found.ID != user.ID {
			t.Errorf("expected ID %d, got %d", user.ID, found.ID)
		}
	})

	t.Run("returns error for non-existent username", func(t *testing.T) {
		_, err := repo.GetUserByUsername(ctx, "nonexistentuser")
		if err == nil {
			t.Error("expected error for non-existent username")
		}
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db, newTestLogger())
	ctx := context.Background()

	t.Run("successfully updates user", func(t *testing.T) {
		user := &domain.User{
			Email:        "update@example.com",
			Username:     "updateuser",
			PasswordHash: "hashedpassword",
			Bio:          "Original bio",
		}

		err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		user.Bio = "Updated bio"
		user.Image = "https://example.com/new-image.png"

		err = repo.UpdateUser(ctx, user)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		found, err := repo.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}

		if found.Bio != "Updated bio" {
			t.Errorf("expected bio 'Updated bio', got %s", found.Bio)
		}

		if found.Image != "https://example.com/new-image.png" {
			t.Errorf("expected updated image, got %s", found.Image)
		}
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		user := &domain.User{
			ID:           999999,
			Email:        "nonexistent@example.com",
			Username:     "nonexistentuser",
			PasswordHash: "hashedpassword",
		}

		err := repo.UpdateUser(ctx, user)
		if err == nil {
			t.Error("expected error for non-existent user")
		}
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("returns error for duplicate email on update", func(t *testing.T) {
		user1 := &domain.User{
			Email:        "user1@example.com",
			Username:     "user1",
			PasswordHash: "hashedpassword",
		}
		err := repo.CreateUser(ctx, user1)
		if err != nil {
			t.Fatalf("failed to create user1: %v", err)
		}

		user2 := &domain.User{
			Email:        "user2@example.com",
			Username:     "user2",
			PasswordHash: "hashedpassword",
		}
		err = repo.CreateUser(ctx, user2)
		if err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		// Try to update user2's email to user1's email
		user2.Email = user1.Email
		err = repo.UpdateUser(ctx, user2)
		if err == nil {
			t.Error("expected error for duplicate email")
		}
		if err != domain.ErrEmailAlreadyTaken {
			t.Errorf("expected ErrEmailAlreadyTaken, got %v", err)
		}
	})
}
