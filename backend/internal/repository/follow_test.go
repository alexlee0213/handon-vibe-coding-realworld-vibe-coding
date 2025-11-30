package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

func setupFollowTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create users table (needed for foreign key constraints)
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

	// Create follows table
	_, err = db.Exec(`
		CREATE TABLE follows (
			follower_id INTEGER NOT NULL,
			following_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id),
			FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
			CHECK (follower_id != following_id)
		);
		CREATE INDEX idx_follows_follower_id ON follows(follower_id);
		CREATE INDEX idx_follows_following_id ON follows(following_id);
	`)
	if err != nil {
		t.Fatalf("failed to create follows table: %v", err)
	}

	return db
}

func createFollowTestUser(t *testing.T, db *sql.DB, email, username string) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO users (email, username, password_hash, bio, image, created_at, updated_at)
		VALUES (?, ?, 'hashedpassword', '', '', datetime('now'), datetime('now'))
	`, email, username)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	return id
}

func TestFollowUser(t *testing.T) {
	db := setupFollowTestDB(t)
	defer db.Close()

	repo := NewSQLiteFollowRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users
	user1ID := createFollowTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFollowTestUser(t, db, "user2@example.com", "user2")

	t.Run("successfully follows a user", func(t *testing.T) {
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify the follow relationship exists
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error checking follow status, got %v", err)
		}
		if !isFollowing {
			t.Error("expected user1 to be following user2")
		}
	})

	t.Run("following same user again is idempotent", func(t *testing.T) {
		// user1 already follows user2 from previous test
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error for duplicate follow, got %v", err)
		}
	})

	t.Run("returns error for self-follow", func(t *testing.T) {
		err := repo.FollowUser(ctx, user1ID, user1ID)
		if err == nil {
			t.Error("expected error for self-follow")
		}
		if err != domain.ErrValidation {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestUnfollowUser(t *testing.T) {
	db := setupFollowTestDB(t)
	defer db.Close()

	repo := NewSQLiteFollowRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users
	user1ID := createFollowTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFollowTestUser(t, db, "user2@example.com", "user2")

	t.Run("successfully unfollows a user", func(t *testing.T) {
		// First, follow
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("failed to follow: %v", err)
		}

		// Then, unfollow
		err = repo.UnfollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify the follow relationship is removed
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error checking follow status, got %v", err)
		}
		if isFollowing {
			t.Error("expected user1 to not be following user2")
		}
	})

	t.Run("unfollowing when not following is idempotent", func(t *testing.T) {
		// user1 is not following user2 now
		err := repo.UnfollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error for unfollow when not following, got %v", err)
		}
	})
}

func TestIsFollowing(t *testing.T) {
	db := setupFollowTestDB(t)
	defer db.Close()

	repo := NewSQLiteFollowRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users
	user1ID := createFollowTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFollowTestUser(t, db, "user2@example.com", "user2")

	t.Run("returns false when not following", func(t *testing.T) {
		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFollowing {
			t.Error("expected false when not following")
		}
	})

	t.Run("returns true when following", func(t *testing.T) {
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("failed to follow: %v", err)
		}

		isFollowing, err := repo.IsFollowing(ctx, user1ID, user2ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !isFollowing {
			t.Error("expected true when following")
		}
	})

	t.Run("returns false for zero follower ID", func(t *testing.T) {
		isFollowing, err := repo.IsFollowing(ctx, 0, user2ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFollowing {
			t.Error("expected false for zero follower ID")
		}
	})

	t.Run("returns false for zero following ID", func(t *testing.T) {
		isFollowing, err := repo.IsFollowing(ctx, user1ID, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFollowing {
			t.Error("expected false for zero following ID")
		}
	})
}

func TestGetFollowers(t *testing.T) {
	db := setupFollowTestDB(t)
	defer db.Close()

	repo := NewSQLiteFollowRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users
	user1ID := createFollowTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFollowTestUser(t, db, "user2@example.com", "user2")
	user3ID := createFollowTestUser(t, db, "user3@example.com", "user3")

	t.Run("returns empty list when no followers", func(t *testing.T) {
		followers, err := repo.GetFollowers(ctx, user1ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(followers) != 0 {
			t.Errorf("expected 0 followers, got %d", len(followers))
		}
	})

	t.Run("returns list of followers", func(t *testing.T) {
		// user2 and user3 follow user1
		err := repo.FollowUser(ctx, user2ID, user1ID)
		if err != nil {
			t.Fatalf("failed to follow: %v", err)
		}
		err = repo.FollowUser(ctx, user3ID, user1ID)
		if err != nil {
			t.Fatalf("failed to follow: %v", err)
		}

		followers, err := repo.GetFollowers(ctx, user1ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(followers) != 2 {
			t.Errorf("expected 2 followers, got %d", len(followers))
		}

		// Check that both user2 and user3 are in the followers list
		followerMap := make(map[int64]bool)
		for _, id := range followers {
			followerMap[id] = true
		}
		if !followerMap[user2ID] {
			t.Error("expected user2 to be in followers list")
		}
		if !followerMap[user3ID] {
			t.Error("expected user3 to be in followers list")
		}
	})
}

func TestGetFollowing(t *testing.T) {
	db := setupFollowTestDB(t)
	defer db.Close()

	repo := NewSQLiteFollowRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users
	user1ID := createFollowTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFollowTestUser(t, db, "user2@example.com", "user2")
	user3ID := createFollowTestUser(t, db, "user3@example.com", "user3")

	t.Run("returns empty list when not following anyone", func(t *testing.T) {
		following, err := repo.GetFollowing(ctx, user1ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(following) != 0 {
			t.Errorf("expected 0 following, got %d", len(following))
		}
	})

	t.Run("returns list of users being followed", func(t *testing.T) {
		// user1 follows user2 and user3
		err := repo.FollowUser(ctx, user1ID, user2ID)
		if err != nil {
			t.Fatalf("failed to follow: %v", err)
		}
		err = repo.FollowUser(ctx, user1ID, user3ID)
		if err != nil {
			t.Fatalf("failed to follow: %v", err)
		}

		following, err := repo.GetFollowing(ctx, user1ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(following) != 2 {
			t.Errorf("expected 2 following, got %d", len(following))
		}

		// Check that both user2 and user3 are in the following list
		followingMap := make(map[int64]bool)
		for _, id := range following {
			followingMap[id] = true
		}
		if !followingMap[user2ID] {
			t.Error("expected user2 to be in following list")
		}
		if !followingMap[user3ID] {
			t.Error("expected user3 to be in following list")
		}
	})
}

func TestIsFollowingBulk(t *testing.T) {
	db := setupFollowTestDB(t)
	defer db.Close()

	repo := NewSQLiteFollowRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users
	user1ID := createFollowTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFollowTestUser(t, db, "user2@example.com", "user2")
	user3ID := createFollowTestUser(t, db, "user3@example.com", "user3")
	user4ID := createFollowTestUser(t, db, "user4@example.com", "user4")

	// user1 follows user2 and user3 but not user4
	err := repo.FollowUser(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("failed to follow: %v", err)
	}
	err = repo.FollowUser(ctx, user1ID, user3ID)
	if err != nil {
		t.Fatalf("failed to follow: %v", err)
	}

	t.Run("returns correct follow status for multiple users", func(t *testing.T) {
		result, err := repo.IsFollowingBulk(ctx, user1ID, []int64{user2ID, user3ID, user4ID})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if !result[user2ID] {
			t.Error("expected user1 to be following user2")
		}
		if !result[user3ID] {
			t.Error("expected user1 to be following user3")
		}
		if result[user4ID] {
			t.Error("expected user1 to not be following user4")
		}
	})

	t.Run("returns all false for zero follower ID", func(t *testing.T) {
		result, err := repo.IsFollowingBulk(ctx, 0, []int64{user2ID, user3ID})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result[user2ID] {
			t.Error("expected false for user2 with zero follower ID")
		}
		if result[user3ID] {
			t.Error("expected false for user3 with zero follower ID")
		}
	})

	t.Run("returns empty map for empty following IDs", func(t *testing.T) {
		result, err := repo.IsFollowingBulk(ctx, user1ID, []int64{})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})
}
