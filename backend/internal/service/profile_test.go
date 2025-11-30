package service

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
)

// setupProfileTestDB creates a test database with all required tables
func setupProfileTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Drop existing tables for clean state
	db.Exec("DROP TABLE IF EXISTS follows")
	db.Exec("DROP TABLE IF EXISTS users")

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
		)
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Create follows table
	_, err = db.Exec(`
		CREATE TABLE follows (
			follower_id INTEGER NOT NULL,
			following_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id),
			FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create follows table: %v", err)
	}

	return db
}

func newProfileTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func newTestProfileService(t *testing.T) (*ProfileService, *sql.DB) {
	t.Helper()
	db := setupProfileTestDB(t)
	logger := newProfileTestLogger()
	userRepo := repository.NewSQLiteUserRepository(db, logger)
	followRepo := repository.NewSQLiteFollowRepository(db, logger)

	profileService := NewProfileService(userRepo, followRepo, logger)
	return profileService, db
}

// createProfileTestUser creates a test user and returns the user ID
func createProfileTestUser(t *testing.T, db *sql.DB, username, email string) int64 {
	t.Helper()
	result, err := db.Exec(`
		INSERT INTO users (email, username, password_hash, bio, image)
		VALUES (?, ?, 'hashedpassword', 'Test bio', 'http://example.com/image.jpg')
	`, email, username)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	userID, _ := result.LastInsertId()
	return userID
}

// =============================================================================
// GetProfileByUsername Tests
// =============================================================================

func TestProfileService_GetProfileByUsername(t *testing.T) {
	t.Run("returns profile for existing user", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		createProfileTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		profile, err := service.GetProfileByUsername(ctx, "testuser", nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if profile == nil {
			t.Fatal("expected profile, got nil")
		}
		if profile.Username != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", profile.Username)
		}
		if profile.Following {
			t.Error("expected following to be false for unauthenticated request")
		}
	})

	t.Run("returns profile with following status when authenticated", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		targetID := createProfileTestUser(t, db, "target", "target@example.com")
		ctx := context.Background()

		// Create follow relationship
		db.Exec("INSERT INTO follows (follower_id, following_id) VALUES (?, ?)", followerID, targetID)

		profile, err := service.GetProfileByUsername(ctx, "target", &followerID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !profile.Following {
			t.Error("expected following to be true")
		}
	})

	t.Run("returns following=false when not following", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		userID := createProfileTestUser(t, db, "user1", "user1@example.com")
		createProfileTestUser(t, db, "user2", "user2@example.com")
		ctx := context.Background()

		profile, err := service.GetProfileByUsername(ctx, "user2", &userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if profile.Following {
			t.Error("expected following to be false when not following")
		}
	})

	t.Run("fails for non-existent user", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		ctx := context.Background()

		_, err := service.GetProfileByUsername(ctx, "nonexistent", nil)
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =============================================================================
// FollowUser Tests
// =============================================================================

func TestProfileService_FollowUser(t *testing.T) {
	t.Run("successfully follows a user", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		createProfileTestUser(t, db, "target", "target@example.com")
		ctx := context.Background()

		profile, err := service.FollowUser(ctx, followerID, "target")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if profile == nil {
			t.Fatal("expected profile, got nil")
		}
		if profile.Username != "target" {
			t.Errorf("expected username 'target', got '%s'", profile.Username)
		}
		if !profile.Following {
			t.Error("expected following to be true after follow")
		}
	})

	t.Run("fails when trying to follow self", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		userID := createProfileTestUser(t, db, "selfuser", "self@example.com")
		ctx := context.Background()

		_, err := service.FollowUser(ctx, userID, "selfuser")
		if err != domain.ErrValidation {
			t.Errorf("expected ErrValidation for self-follow, got %v", err)
		}
	})

	t.Run("fails for non-existent target user", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		ctx := context.Background()

		_, err := service.FollowUser(ctx, followerID, "nonexistent")
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("handles duplicate follow gracefully", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		createProfileTestUser(t, db, "target", "target@example.com")
		ctx := context.Background()

		// Follow first time
		_, err := service.FollowUser(ctx, followerID, "target")
		if err != nil {
			t.Fatalf("first follow failed: %v", err)
		}

		// Follow second time - should handle gracefully (either succeed or return error)
		profile, err := service.FollowUser(ctx, followerID, "target")
		if err != nil {
			// This is expected behavior - duplicate follow may return an error
			t.Logf("duplicate follow returned error (expected): %v", err)
		} else {
			if !profile.Following {
				t.Error("expected following to be true")
			}
		}
	})
}

// =============================================================================
// UnfollowUser Tests
// =============================================================================

func TestProfileService_UnfollowUser(t *testing.T) {
	t.Run("successfully unfollows a user", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		targetID := createProfileTestUser(t, db, "target", "target@example.com")
		ctx := context.Background()

		// Create follow relationship first
		db.Exec("INSERT INTO follows (follower_id, following_id) VALUES (?, ?)", followerID, targetID)

		profile, err := service.UnfollowUser(ctx, followerID, "target")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if profile == nil {
			t.Fatal("expected profile, got nil")
		}
		if profile.Username != "target" {
			t.Errorf("expected username 'target', got '%s'", profile.Username)
		}
		if profile.Following {
			t.Error("expected following to be false after unfollow")
		}
	})

	t.Run("fails for non-existent target user", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		ctx := context.Background()

		_, err := service.UnfollowUser(ctx, followerID, "nonexistent")
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("handles unfollow when not following", func(t *testing.T) {
		service, db := newTestProfileService(t)
		defer db.Close()

		followerID := createProfileTestUser(t, db, "follower", "follower@example.com")
		createProfileTestUser(t, db, "target", "target@example.com")
		ctx := context.Background()

		// Unfollow without following first - should handle gracefully
		profile, err := service.UnfollowUser(ctx, followerID, "target")
		if err != nil {
			// This is expected behavior - unfollow when not following may return an error
			t.Logf("unfollow when not following returned error (expected): %v", err)
		} else {
			if profile.Following {
				t.Error("expected following to be false")
			}
		}
	})
}
