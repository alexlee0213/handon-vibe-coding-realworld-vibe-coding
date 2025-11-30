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

// setupCommentTestDB creates a test database with all required tables
func setupCommentTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Drop existing tables for clean state
	db.Exec("DROP TABLE IF EXISTS comments")
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

	// Create comments table
	_, err = db.Exec(`
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			body TEXT NOT NULL,
			article_id INTEGER NOT NULL,
			author_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
			FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create comments table: %v", err)
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

func newCommentTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

func newTestCommentService(t *testing.T) (*CommentService, *sql.DB) {
	t.Helper()
	db := setupCommentTestDB(t)
	logger := newCommentTestLogger()
	commentRepo := repository.NewSQLiteCommentRepository(db, logger)
	articleRepo := repository.NewSQLiteArticleRepository(db, logger)
	userRepo := repository.NewSQLiteUserRepository(db, logger)

	commentService := NewCommentService(commentRepo, articleRepo, userRepo, logger)
	return commentService, db
}

// createCommentTestUser creates a test user and returns the user ID
func createCommentTestUser(t *testing.T, db *sql.DB, username, email string) int64 {
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

// createCommentTestArticle creates a test article and returns the slug
func createCommentTestArticle(t *testing.T, db *sql.DB, authorID int64, slug, title string) string {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO articles (slug, title, description, body, author_id)
		VALUES (?, ?, 'Description', 'Body', ?)
	`, slug, title, authorID)
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}
	return slug
}

// =============================================================================
// CreateComment Tests
// =============================================================================

func TestCommentService_CreateComment(t *testing.T) {
	t.Run("successfully creates a comment", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		commenterID := createCommentTestUser(t, db, "commenter", "commenter@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		input := &domain.CreateCommentInput{
			Body: "This is a test comment",
		}

		comment, err := service.CreateComment(ctx, slug, commenterID, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if comment == nil {
			t.Fatal("expected comment, got nil")
		}
		if comment.Body != "This is a test comment" {
			t.Errorf("expected body 'This is a test comment', got '%s'", comment.Body)
		}
		if comment.Author == nil {
			t.Error("expected author to be loaded")
		}
	})

	t.Run("fails with empty body", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		input := &domain.CreateCommentInput{
			Body: "",
		}

		_, err := service.CreateComment(ctx, slug, authorID, input)
		if err == nil {
			t.Fatal("expected error for empty body")
		}
		_, ok := err.(*domain.ValidationErrors)
		if !ok {
			t.Fatalf("expected ValidationErrors, got %T", err)
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		userID := createCommentTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		input := &domain.CreateCommentInput{
			Body: "Test comment",
		}

		_, err := service.CreateComment(ctx, "non-existent-slug", userID, input)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})

	t.Run("trims whitespace from body", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		input := &domain.CreateCommentInput{
			Body: "  Trimmed comment  ",
		}

		comment, err := service.CreateComment(ctx, slug, authorID, input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if comment.Body != "Trimmed comment" {
			t.Errorf("expected body 'Trimmed comment', got '%s'", comment.Body)
		}
	})
}

// =============================================================================
// GetCommentsByArticleSlug Tests
// =============================================================================

func TestCommentService_GetCommentsByArticleSlug(t *testing.T) {
	t.Run("returns comments for article", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		// Create comments
		for i := 0; i < 3; i++ {
			input := &domain.CreateCommentInput{
				Body: "Comment " + string(rune('A'+i)),
			}
			service.CreateComment(ctx, slug, authorID, input)
		}

		comments, err := service.GetCommentsByArticleSlug(ctx, slug)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(comments) != 3 {
			t.Errorf("expected 3 comments, got %d", len(comments))
		}
	})

	t.Run("returns empty list for article without comments", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		comments, err := service.GetCommentsByArticleSlug(ctx, slug)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(comments) != 0 {
			t.Errorf("expected 0 comments, got %d", len(comments))
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		ctx := context.Background()

		_, err := service.GetCommentsByArticleSlug(ctx, "non-existent-slug")
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})
}

// =============================================================================
// DeleteComment Tests
// =============================================================================

func TestCommentService_DeleteComment(t *testing.T) {
	t.Run("successfully deletes own comment", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		input := &domain.CreateCommentInput{
			Body: "To be deleted",
		}
		comment, _ := service.CreateComment(ctx, slug, authorID, input)

		err := service.DeleteComment(ctx, slug, comment.ID, authorID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify deletion
		comments, _ := service.GetCommentsByArticleSlug(ctx, slug)
		if len(comments) != 0 {
			t.Error("expected comment to be deleted")
		}
	})

	t.Run("fails when non-author tries to delete", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		otherUserID := createCommentTestUser(t, db, "other", "other@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		input := &domain.CreateCommentInput{
			Body: "Protected comment",
		}
		comment, _ := service.CreateComment(ctx, slug, authorID, input)

		err := service.DeleteComment(ctx, slug, comment.ID, otherUserID)
		if err != domain.ErrForbidden {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("fails for non-existent article", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		userID := createCommentTestUser(t, db, "user", "user@example.com")
		ctx := context.Background()

		err := service.DeleteComment(ctx, "non-existent-slug", 1, userID)
		if err != domain.ErrArticleNotFound {
			t.Errorf("expected ErrArticleNotFound, got %v", err)
		}
	})

	t.Run("fails for non-existent comment", func(t *testing.T) {
		service, db := newTestCommentService(t)
		defer db.Close()

		authorID := createCommentTestUser(t, db, "author", "author@example.com")
		slug := createCommentTestArticle(t, db, authorID, "test-article", "Test Article")
		ctx := context.Background()

		err := service.DeleteComment(ctx, slug, 9999, authorID)
		if err != domain.ErrCommentNotFound {
			t.Errorf("expected ErrCommentNotFound, got %v", err)
		}
	})
}
