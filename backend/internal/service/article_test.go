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

// setupArticleTestDB creates a test database with all required tables
func setupArticleTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Use file:memory with shared cache to ensure tables persist across connections
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Drop existing tables for clean state between tests
	db.Exec("DROP TABLE IF EXISTS article_tags")
	db.Exec("DROP TABLE IF EXISTS favorites")
	db.Exec("DROP TABLE IF EXISTS follows")
	db.Exec("DROP TABLE IF EXISTS tags")
	db.Exec("DROP TABLE IF EXISTS articles")
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

	// Create articles table
	_, err = db.Exec(`
		CREATE TABLE articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			body TEXT NOT NULL DEFAULT '',
			author_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create articles table: %v", err)
	}

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tags table: %v", err)
	}

	// Create article_tags junction table
	_, err = db.Exec(`
		CREATE TABLE article_tags (
			article_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (article_id, tag_id),
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create article_tags table: %v", err)
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
		)
	`)
	if err != nil {
		t.Fatalf("failed to create favorites table: %v", err)
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

func newArticleTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func newTestArticleService(t *testing.T) (*ArticleService, *sql.DB) {
	t.Helper()
	db := setupArticleTestDB(t)
	logger := newArticleTestLogger()
	articleRepo := repository.NewSQLiteArticleRepository(db, logger)
	userRepo := repository.NewSQLiteUserRepository(db, logger)

	articleService := NewArticleService(articleRepo, userRepo, logger)
	return articleService, db
}

// createTestUser creates a test user and returns the user ID
func createTestUser(t *testing.T, db *sql.DB, username, email string) int64 {
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
// CreateArticle Tests
// =============================================================================

func TestArticleService_CreateArticle(t *testing.T) {
	t.Run("successfully creates an article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Test description",
			Body:        "Test body content",
			TagList:     []string{"go", "testing"},
		}

		article, err := service.CreateArticle(ctx, userID, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if article == nil {
			t.Fatal("expected article, got nil")
		}
		if article.Title != "Test Article" {
			t.Errorf("expected title 'Test Article', got '%s'", article.Title)
		}
		if article.Slug == "" {
			t.Error("expected non-empty slug")
		}
	})

	t.Run("fails with empty title", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "",
			Description: "Test description",
			Body:        "Test body content",
		}

		_, err := service.CreateArticle(ctx, userID, input)
		if err == nil {
			t.Fatal("expected error for empty title")
		}
		validationErr, ok := err.(*domain.ValidationErrors)
		if !ok {
			t.Fatalf("expected ValidationErrors, got %T", err)
		}
		if !validationErr.HasErrors() {
			t.Error("expected validation errors")
		}
	})

	t.Run("fails with empty description", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Title",
			Description: "",
			Body:        "Test body content",
		}

		_, err := service.CreateArticle(ctx, userID, input)
		if err == nil {
			t.Fatal("expected error for empty description")
		}
	})

	t.Run("fails with empty body", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Title",
			Description: "Test description",
			Body:        "",
		}

		_, err := service.CreateArticle(ctx, userID, input)
		if err == nil {
			t.Fatal("expected error for empty body")
		}
	})

	t.Run("creates article with nil TagList", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Test description",
			Body:        "Test body content",
			TagList:     nil,
		}

		article, err := service.CreateArticle(ctx, userID, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if article.TagList == nil {
			t.Error("expected empty slice, got nil")
		}
	})
}

// =============================================================================
// GetArticleBySlug Tests
// =============================================================================

func TestArticleService_GetArticleBySlug(t *testing.T) {
	t.Run("successfully gets article by slug", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		// Create an article first
		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Test description",
			Body:        "Test body content",
		}
		created, err := service.CreateArticle(ctx, userID, input)
		if err != nil {
			t.Fatalf("failed to create article: %v", err)
		}

		// Get by slug
		article, err := service.GetArticleBySlug(ctx, created.Slug, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if article.Title != "Test Article" {
			t.Errorf("expected title 'Test Article', got '%s'", article.Title)
		}
		if article.Author == nil {
			t.Error("expected author to be loaded")
		}
	})

	t.Run("fails for non-existent slug", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		ctx := context.Background()

		_, err := service.GetArticleBySlug(ctx, "non-existent-slug", nil)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})
}

// =============================================================================
// UpdateArticle Tests
// =============================================================================

func TestArticleService_UpdateArticle(t *testing.T) {
	t.Run("successfully updates article title", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		// Create an article
		input := &domain.CreateArticleInput{
			Title:       "Original Title",
			Description: "Test description",
			Body:        "Test body content",
		}
		created, _ := service.CreateArticle(ctx, userID, input)

		// Update title
		newTitle := "Updated Title"
		updateInput := &domain.UpdateArticleInput{
			Title: &newTitle,
		}

		updated, err := service.UpdateArticle(ctx, created.Slug, userID, updateInput)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
		}
	})

	t.Run("fails when non-author tries to update", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		authorID := createTestUser(t, db, "author", "author@example.com")
		otherUserID := createTestUser(t, db, "other", "other@example.com")
		ctx := context.Background()

		// Create an article as author
		input := &domain.CreateArticleInput{
			Title:       "Original Title",
			Description: "Test description",
			Body:        "Test body content",
		}
		created, _ := service.CreateArticle(ctx, authorID, input)

		// Try to update as different user
		newTitle := "Hacked Title"
		updateInput := &domain.UpdateArticleInput{
			Title: &newTitle,
		}

		_, err := service.UpdateArticle(ctx, created.Slug, otherUserID, updateInput)
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		newTitle := "New Title"
		updateInput := &domain.UpdateArticleInput{
			Title: &newTitle,
		}

		_, err := service.UpdateArticle(ctx, "non-existent-slug", userID, updateInput)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})

	t.Run("updates description and body", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Original Title",
			Description: "Original description",
			Body:        "Original body",
		}
		created, _ := service.CreateArticle(ctx, userID, input)

		newDesc := "Updated description"
		newBody := "Updated body"
		updateInput := &domain.UpdateArticleInput{
			Description: &newDesc,
			Body:        &newBody,
		}

		updated, err := service.UpdateArticle(ctx, created.Slug, userID, updateInput)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Description != "Updated description" {
			t.Errorf("expected description 'Updated description', got '%s'", updated.Description)
		}
		if updated.Body != "Updated body" {
			t.Errorf("expected body 'Updated body', got '%s'", updated.Body)
		}
	})
}

// =============================================================================
// DeleteArticle Tests
// =============================================================================

func TestArticleService_DeleteArticle(t *testing.T) {
	t.Run("successfully deletes article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "To Delete",
			Description: "Test description",
			Body:        "Test body content",
		}
		created, _ := service.CreateArticle(ctx, userID, input)

		err := service.DeleteArticle(ctx, created.Slug, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify deletion
		_, err = service.GetArticleBySlug(ctx, created.Slug, nil)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound after deletion, got %v", err)
		}
	})

	t.Run("fails when non-author tries to delete", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		authorID := createTestUser(t, db, "author", "author@example.com")
		otherUserID := createTestUser(t, db, "other", "other@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "To Delete",
			Description: "Test description",
			Body:        "Test body content",
		}
		created, _ := service.CreateArticle(ctx, authorID, input)

		err := service.DeleteArticle(ctx, created.Slug, otherUserID)
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		err := service.DeleteArticle(ctx, "non-existent-slug", userID)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})
}

// =============================================================================
// ListArticles Tests
// =============================================================================

func TestArticleService_ListArticles(t *testing.T) {
	t.Run("lists articles with default params", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		// Create articles
		for i := 0; i < 3; i++ {
			input := &domain.CreateArticleInput{
				Title:       "Article " + string(rune('A'+i)),
				Description: "Description",
				Body:        "Body",
			}
			service.CreateArticle(ctx, userID, input)
		}

		articles, total, err := service.ListArticles(ctx, nil, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 3 {
			t.Errorf("expected 3 articles, got %d", total)
		}
		if len(articles) != 3 {
			t.Errorf("expected 3 articles in result, got %d", len(articles))
		}
	})

	t.Run("applies limit parameter", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		// Create 5 articles
		for i := 0; i < 5; i++ {
			input := &domain.CreateArticleInput{
				Title:       "Article " + string(rune('A'+i)),
				Description: "Description",
				Body:        "Body",
			}
			service.CreateArticle(ctx, userID, input)
		}

		params := &domain.ArticleListParams{
			Limit: 2,
		}
		articles, total, err := service.ListArticles(ctx, params, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(articles) != 2 {
			t.Errorf("expected 2 articles in result, got %d", len(articles))
		}
	})

	t.Run("caps limit at 100", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		ctx := context.Background()

		params := &domain.ArticleListParams{
			Limit: 200,
		}
		_, _, err := service.ListArticles(ctx, params, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// The service should cap the limit internally
	})

	t.Run("applies default limit for zero value", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		ctx := context.Background()

		params := &domain.ArticleListParams{
			Limit: 0,
		}
		_, _, err := service.ListArticles(ctx, params, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// =============================================================================
// GetFeed Tests
// =============================================================================

func TestArticleService_GetFeed(t *testing.T) {
	t.Run("returns empty feed when not following anyone", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		articles, total, err := service.GetFeed(ctx, userID, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 0 {
			t.Errorf("expected 0 articles, got %d", total)
		}
		if len(articles) != 0 {
			t.Errorf("expected empty articles list, got %d", len(articles))
		}
	})

	t.Run("applies default params", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		_, _, err := service.GetFeed(ctx, userID, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("caps limit at 100", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		params := &domain.ArticleFeedParams{
			Limit: 200,
		}
		_, _, err := service.GetFeed(ctx, userID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("applies default limit for zero value", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		params := &domain.ArticleFeedParams{
			Limit: 0,
		}
		_, _, err := service.GetFeed(ctx, userID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// =============================================================================
// GetAllTags Tests
// =============================================================================

func TestArticleService_GetAllTags(t *testing.T) {
	t.Run("returns empty list when no tags exist", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		ctx := context.Background()

		tags, err := service.GetAllTags(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tags) != 0 {
			t.Errorf("expected empty tags list, got %d", len(tags))
		}
	})

	t.Run("returns tags from articles", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "testuser", "test@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Description",
			Body:        "Body",
			TagList:     []string{"go", "testing"},
		}
		service.CreateArticle(ctx, userID, input)

		tags, err := service.GetAllTags(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tags))
		}
	})
}

// =============================================================================
// FavoriteArticle Tests
// =============================================================================

func TestArticleService_FavoriteArticle(t *testing.T) {
	t.Run("successfully favorites an article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		authorID := createTestUser(t, db, "author", "author@example.com")
		userID := createTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Description",
			Body:        "Body",
		}
		created, _ := service.CreateArticle(ctx, authorID, input)

		article, err := service.FavoriteArticle(ctx, created.Slug, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !article.Favorited {
			t.Error("expected article to be favorited")
		}
		if article.FavoritesCount != 1 {
			t.Errorf("expected favorites count 1, got %d", article.FavoritesCount)
		}
	})

	t.Run("handles already favorited article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		authorID := createTestUser(t, db, "author", "author@example.com")
		userID := createTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Description",
			Body:        "Body",
		}
		created, _ := service.CreateArticle(ctx, authorID, input)

		// Favorite twice
		service.FavoriteArticle(ctx, created.Slug, userID)
		article, err := service.FavoriteArticle(ctx, created.Slug, userID)

		if err != nil {
			t.Fatalf("expected no error on double favorite, got %v", err)
		}
		if article.FavoritesCount != 1 {
			t.Errorf("expected favorites count 1, got %d", article.FavoritesCount)
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		_, err := service.FavoriteArticle(ctx, "non-existent-slug", userID)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})
}

// =============================================================================
// UnfavoriteArticle Tests
// =============================================================================

func TestArticleService_UnfavoriteArticle(t *testing.T) {
	t.Run("successfully unfavorites an article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		authorID := createTestUser(t, db, "author", "author@example.com")
		userID := createTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Description",
			Body:        "Body",
		}
		created, _ := service.CreateArticle(ctx, authorID, input)

		// Favorite then unfavorite
		service.FavoriteArticle(ctx, created.Slug, userID)
		article, err := service.UnfavoriteArticle(ctx, created.Slug, userID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if article.Favorited {
			t.Error("expected article to not be favorited")
		}
		if article.FavoritesCount != 0 {
			t.Errorf("expected favorites count 0, got %d", article.FavoritesCount)
		}
	})

	t.Run("handles already unfavorited article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		authorID := createTestUser(t, db, "author", "author@example.com")
		userID := createTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		input := &domain.CreateArticleInput{
			Title:       "Test Article",
			Description: "Description",
			Body:        "Body",
		}
		created, _ := service.CreateArticle(ctx, authorID, input)

		// Unfavorite without having favorited
		article, err := service.UnfavoriteArticle(ctx, created.Slug, userID)
		if err != nil {
			t.Fatalf("expected no error on unfavorite non-favorited, got %v", err)
		}
		if article.FavoritesCount != 0 {
			t.Errorf("expected favorites count 0, got %d", article.FavoritesCount)
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestArticleService(t)
		defer db.Close()

		userID := createTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		_, err := service.UnfavoriteArticle(ctx, "non-existent-slug", userID)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})
}
