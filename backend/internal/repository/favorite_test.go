package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupFavoriteTestDB(t *testing.T) *sql.DB {
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

	// Create articles table
	_, err = db.Exec(`
		CREATE TABLE articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			body TEXT NOT NULL,
			author_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("failed to create articles table: %v", err)
	}

	// Create favorites table
	_, err = db.Exec(`
		CREATE TABLE favorites (
			user_id INTEGER NOT NULL,
			article_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, article_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_favorites_user_id ON favorites(user_id);
		CREATE INDEX idx_favorites_article_id ON favorites(article_id);
	`)
	if err != nil {
		t.Fatalf("failed to create favorites table: %v", err)
	}

	return db
}

func createFavoriteTestUser(t *testing.T, db *sql.DB, email, username string) int64 {
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

func createFavoriteTestArticle(t *testing.T, db *sql.DB, authorID int64, slug, title string) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO articles (slug, title, description, body, author_id, created_at, updated_at)
		VALUES (?, ?, 'Test description', 'Test body', ?, datetime('now'), datetime('now'))
	`, slug, title, authorID)
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	return id
}

func TestFavorite(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and article
	userID := createFavoriteTestUser(t, db, "user@example.com", "user")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("successfully favorites an article", func(t *testing.T) {
		err := repo.Favorite(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify the favorite relationship exists
		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error checking favorite status, got %v", err)
		}
		if !isFavorited {
			t.Error("expected article to be favorited")
		}
	})

	t.Run("favoriting same article again is idempotent", func(t *testing.T) {
		// Article is already favorited from previous test
		err := repo.Favorite(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error for duplicate favorite, got %v", err)
		}
	})
}

func TestUnfavorite(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and article
	userID := createFavoriteTestUser(t, db, "user@example.com", "user")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("successfully unfavorites an article", func(t *testing.T) {
		// First, favorite
		err := repo.Favorite(ctx, userID, articleID)
		if err != nil {
			t.Fatalf("failed to favorite: %v", err)
		}

		// Then, unfavorite
		err = repo.Unfavorite(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify the favorite relationship is removed
		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error checking favorite status, got %v", err)
		}
		if isFavorited {
			t.Error("expected article to not be favorited")
		}
	})

	t.Run("unfavoriting when not favorited is idempotent", func(t *testing.T) {
		// Article is not favorited now
		err := repo.Unfavorite(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error for unfavorite when not favorited, got %v", err)
		}
	})
}

func TestIsFavorited(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and article
	userID := createFavoriteTestUser(t, db, "user@example.com", "user")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("returns false when not favorited", func(t *testing.T) {
		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFavorited {
			t.Error("expected false when not favorited")
		}
	})

	t.Run("returns true when favorited", func(t *testing.T) {
		err := repo.Favorite(ctx, userID, articleID)
		if err != nil {
			t.Fatalf("failed to favorite: %v", err)
		}

		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !isFavorited {
			t.Error("expected true when favorited")
		}
	})

	t.Run("returns false for zero user ID", func(t *testing.T) {
		isFavorited, err := repo.IsFavorited(ctx, 0, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFavorited {
			t.Error("expected false for zero user ID")
		}
	})

	t.Run("returns false for zero article ID", func(t *testing.T) {
		isFavorited, err := repo.IsFavorited(ctx, userID, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFavorited {
			t.Error("expected false for zero article ID")
		}
	})
}

func TestGetFavoritesCount(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users and article
	user1ID := createFavoriteTestUser(t, db, "user1@example.com", "user1")
	user2ID := createFavoriteTestUser(t, db, "user2@example.com", "user2")
	user3ID := createFavoriteTestUser(t, db, "user3@example.com", "user3")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("returns 0 when no favorites", func(t *testing.T) {
		count, err := repo.GetFavoritesCount(ctx, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 favorites, got %d", count)
		}
	})

	t.Run("returns correct count after favorites", func(t *testing.T) {
		// Multiple users favorite the article
		err := repo.Favorite(ctx, user1ID, articleID)
		if err != nil {
			t.Fatalf("failed to favorite: %v", err)
		}
		err = repo.Favorite(ctx, user2ID, articleID)
		if err != nil {
			t.Fatalf("failed to favorite: %v", err)
		}
		err = repo.Favorite(ctx, user3ID, articleID)
		if err != nil {
			t.Fatalf("failed to favorite: %v", err)
		}

		count, err := repo.GetFavoritesCount(ctx, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 favorites, got %d", count)
		}
	})

	t.Run("count decreases after unfavorite", func(t *testing.T) {
		err := repo.Unfavorite(ctx, user1ID, articleID)
		if err != nil {
			t.Fatalf("failed to unfavorite: %v", err)
		}

		count, err := repo.GetFavoritesCount(ctx, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2 favorites, got %d", count)
		}
	})
}

func TestIsFavoritedBulk(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and articles
	userID := createFavoriteTestUser(t, db, "user@example.com", "user")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	article1ID := createFavoriteTestArticle(t, db, authorID, "article-1", "Article 1")
	article2ID := createFavoriteTestArticle(t, db, authorID, "article-2", "Article 2")
	article3ID := createFavoriteTestArticle(t, db, authorID, "article-3", "Article 3")

	// User favorites article1 and article2 but not article3
	err := repo.Favorite(ctx, userID, article1ID)
	if err != nil {
		t.Fatalf("failed to favorite: %v", err)
	}
	err = repo.Favorite(ctx, userID, article2ID)
	if err != nil {
		t.Fatalf("failed to favorite: %v", err)
	}

	t.Run("returns correct favorite status for multiple articles", func(t *testing.T) {
		result, err := repo.IsFavoritedBulk(ctx, userID, []int64{article1ID, article2ID, article3ID})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if !result[article1ID] {
			t.Error("expected article1 to be favorited")
		}
		if !result[article2ID] {
			t.Error("expected article2 to be favorited")
		}
		if result[article3ID] {
			t.Error("expected article3 to not be favorited")
		}
	})

	t.Run("returns all false for zero user ID", func(t *testing.T) {
		result, err := repo.IsFavoritedBulk(ctx, 0, []int64{article1ID, article2ID})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result[article1ID] {
			t.Error("expected false for article1 with zero user ID")
		}
		if result[article2ID] {
			t.Error("expected false for article2 with zero user ID")
		}
	})

	t.Run("returns empty map for empty article IDs", func(t *testing.T) {
		result, err := repo.IsFavoritedBulk(ctx, userID, []int64{})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})
}
