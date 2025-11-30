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
			author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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

func TestFavoriteArticle(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and article
	userID := createFavoriteTestUser(t, db, "user@example.com", "testuser")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("successfully favorites an article", func(t *testing.T) {
		err := repo.FavoriteArticle(ctx, userID, articleID)
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

	t.Run("favoriting same article again is a no-op", func(t *testing.T) {
		err := repo.FavoriteArticle(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error for duplicate favorite, got %v", err)
		}

		// Verify still favorited
		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !isFavorited {
			t.Error("expected article to still be favorited")
		}
	})
}

func TestUnfavoriteArticle(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and article
	userID := createFavoriteTestUser(t, db, "user@example.com", "testuser")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("successfully unfavorites an article", func(t *testing.T) {
		// First, favorite the article
		err := repo.FavoriteArticle(ctx, userID, articleID)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Then unfavorite
		err = repo.UnfavoriteArticle(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify the favorite relationship is removed
		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFavorited {
			t.Error("expected article to not be favorited")
		}
	})

	t.Run("unfavoriting non-favorited article is a no-op", func(t *testing.T) {
		// Create new article that hasn't been favorited
		newArticleID := createFavoriteTestArticle(t, db, authorID, "another-article", "Another Article")

		err := repo.UnfavoriteArticle(ctx, userID, newArticleID)
		if err != nil {
			t.Errorf("expected no error for unfavoriting non-favorited article, got %v", err)
		}
	})
}

func TestIsFavorited(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test users and article
	userID := createFavoriteTestUser(t, db, "user@example.com", "testuser")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	articleID := createFavoriteTestArticle(t, db, authorID, "test-article", "Test Article")

	t.Run("returns false for non-favorited article", func(t *testing.T) {
		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if isFavorited {
			t.Error("expected article to not be favorited")
		}
	})

	t.Run("returns true for favorited article", func(t *testing.T) {
		// Favorite the article
		err := repo.FavoriteArticle(ctx, userID, articleID)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		isFavorited, err := repo.IsFavorited(ctx, userID, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !isFavorited {
			t.Error("expected article to be favorited")
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

	t.Run("returns 0 for article with no favorites", func(t *testing.T) {
		count, err := repo.GetFavoritesCount(ctx, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("returns correct count after favorites", func(t *testing.T) {
		// Add favorites from multiple users
		repo.FavoriteArticle(ctx, user1ID, articleID)
		repo.FavoriteArticle(ctx, user2ID, articleID)
		repo.FavoriteArticle(ctx, user3ID, articleID)

		count, err := repo.GetFavoritesCount(ctx, articleID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 3 {
			t.Errorf("expected count 3, got %d", count)
		}
	})

	t.Run("returns 0 for zero article ID", func(t *testing.T) {
		count, err := repo.GetFavoritesCount(ctx, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if count != 0 {
			t.Errorf("expected count 0 for zero article ID, got %d", count)
		}
	})
}

func TestIsFavoritedBulk(t *testing.T) {
	db := setupFavoriteTestDB(t)
	defer db.Close()

	repo := NewSQLiteFavoriteRepository(db, newTestLogger())
	ctx := context.Background()

	// Create test user and articles
	userID := createFavoriteTestUser(t, db, "user@example.com", "testuser")
	authorID := createFavoriteTestUser(t, db, "author@example.com", "author")
	article1ID := createFavoriteTestArticle(t, db, authorID, "article-1", "Article 1")
	article2ID := createFavoriteTestArticle(t, db, authorID, "article-2", "Article 2")
	article3ID := createFavoriteTestArticle(t, db, authorID, "article-3", "Article 3")

	// Favorite articles 1 and 3 (not 2)
	repo.FavoriteArticle(ctx, userID, article1ID)
	repo.FavoriteArticle(ctx, userID, article3ID)

	t.Run("returns correct status for multiple articles", func(t *testing.T) {
		articleIDs := []int64{article1ID, article2ID, article3ID}
		result, err := repo.IsFavoritedBulk(ctx, userID, articleIDs)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if !result[article1ID] {
			t.Error("expected article 1 to be favorited")
		}
		if result[article2ID] {
			t.Error("expected article 2 to not be favorited")
		}
		if !result[article3ID] {
			t.Error("expected article 3 to be favorited")
		}
	})

	t.Run("returns all false for zero user ID", func(t *testing.T) {
		articleIDs := []int64{article1ID, article2ID}
		result, err := repo.IsFavoritedBulk(ctx, 0, articleIDs)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		for _, id := range articleIDs {
			if result[id] {
				t.Errorf("expected false for article %d with zero user ID", id)
			}
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
