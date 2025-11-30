package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestCommentDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Drop existing tables
	db.Exec("DROP TABLE IF EXISTS comments")
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

	return db, func() {
		db.Close()
	}
}

func createTestUserForComment(t *testing.T, db *sql.DB, username, email string) int64 {
	result, err := db.Exec(`
		INSERT INTO users (email, username, password_hash)
		VALUES (?, ?, 'hashedpassword')
	`, email, username)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func createTestArticle(t *testing.T, db *sql.DB, slug, title string, authorID int64) int64 {
	result, err := db.Exec(`
		INSERT INTO articles (slug, title, description, body, author_id)
		VALUES (?, ?, 'description', 'body', ?)
	`, slug, title, authorID)
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func TestCommentRepository_CreateComment(t *testing.T) {
	db, cleanup := setupTestCommentDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteCommentRepository(db, logger)

	authorID := createTestUserForComment(t, db, "testuser", "test@example.com")
	articleID := createTestArticle(t, db, "test-article", "Test Article", authorID)

	t.Run("successfully creates a comment", func(t *testing.T) {
		comment := &domain.Comment{
			Body:      "This is a test comment",
			ArticleID: articleID,
			AuthorID:  authorID,
		}

		err := repo.CreateComment(context.Background(), comment)
		if err != nil {
			t.Errorf("CreateComment() error = %v", err)
			return
		}

		if comment.ID == 0 {
			t.Error("CreateComment() did not set comment ID")
		}

		if comment.CreatedAt.IsZero() {
			t.Error("CreateComment() did not set created_at")
		}

		if comment.UpdatedAt.IsZero() {
			t.Error("CreateComment() did not set updated_at")
		}
	})
}

func TestCommentRepository_GetCommentByID(t *testing.T) {
	db, cleanup := setupTestCommentDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteCommentRepository(db, logger)

	authorID := createTestUserForComment(t, db, "testuser", "test@example.com")
	articleID := createTestArticle(t, db, "test-article", "Test Article", authorID)

	// Create a test comment
	comment := &domain.Comment{
		Body:      "Test comment",
		ArticleID: articleID,
		AuthorID:  authorID,
	}
	err := repo.CreateComment(context.Background(), comment)
	if err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}

	t.Run("get existing comment", func(t *testing.T) {
		result, err := repo.GetCommentByID(context.Background(), comment.ID)
		if err != nil {
			t.Errorf("GetCommentByID() error = %v", err)
			return
		}

		if result.Body != "Test comment" {
			t.Errorf("GetCommentByID() body = %v, want 'Test comment'", result.Body)
		}

		if result.ArticleID != articleID {
			t.Errorf("GetCommentByID() article_id = %v, want %v", result.ArticleID, articleID)
		}
	})

	t.Run("get non-existing comment", func(t *testing.T) {
		_, err := repo.GetCommentByID(context.Background(), 999999)
		if err == nil {
			t.Error("GetCommentByID() expected error for non-existing comment")
		}
		if err != domain.ErrCommentNotFound {
			t.Errorf("GetCommentByID() error = %v, want ErrCommentNotFound", err)
		}
	})
}

func TestCommentRepository_GetCommentsByArticleID(t *testing.T) {
	db, cleanup := setupTestCommentDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteCommentRepository(db, logger)

	authorID := createTestUserForComment(t, db, "testuser", "test@example.com")
	articleID := createTestArticle(t, db, "test-article", "Test Article", authorID)

	// Create multiple comments
	for i := 1; i <= 3; i++ {
		comment := &domain.Comment{
			Body:      "Comment " + string(rune('0'+i)),
			ArticleID: articleID,
			AuthorID:  authorID,
		}
		err := repo.CreateComment(context.Background(), comment)
		if err != nil {
			t.Fatalf("failed to create test comment %d: %v", i, err)
		}
	}

	t.Run("get comments for article", func(t *testing.T) {
		comments, err := repo.GetCommentsByArticleID(context.Background(), articleID)
		if err != nil {
			t.Errorf("GetCommentsByArticleID() error = %v", err)
			return
		}

		if len(comments) != 3 {
			t.Errorf("GetCommentsByArticleID() count = %v, want 3", len(comments))
		}
	})

	t.Run("get comments for non-existing article", func(t *testing.T) {
		comments, err := repo.GetCommentsByArticleID(context.Background(), 999999)
		if err != nil {
			t.Errorf("GetCommentsByArticleID() error = %v", err)
			return
		}

		if len(comments) != 0 {
			t.Errorf("GetCommentsByArticleID() count = %v, want 0", len(comments))
		}
	})
}

func TestCommentRepository_DeleteComment(t *testing.T) {
	db, cleanup := setupTestCommentDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteCommentRepository(db, logger)

	authorID := createTestUserForComment(t, db, "testuser", "test@example.com")
	articleID := createTestArticle(t, db, "test-article", "Test Article", authorID)

	// Create a test comment
	comment := &domain.Comment{
		Body:      "To be deleted",
		ArticleID: articleID,
		AuthorID:  authorID,
	}
	err := repo.CreateComment(context.Background(), comment)
	if err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}

	t.Run("delete existing comment", func(t *testing.T) {
		err := repo.DeleteComment(context.Background(), comment.ID)
		if err != nil {
			t.Errorf("DeleteComment() error = %v", err)
			return
		}

		// Verify deletion
		_, err = repo.GetCommentByID(context.Background(), comment.ID)
		if err != domain.ErrCommentNotFound {
			t.Errorf("GetCommentByID() after delete error = %v, want ErrCommentNotFound", err)
		}
	})

	t.Run("delete non-existing comment", func(t *testing.T) {
		err := repo.DeleteComment(context.Background(), 999999)
		if err == nil {
			t.Error("DeleteComment() expected error for non-existing comment")
		}
		if err != domain.ErrCommentNotFound {
			t.Errorf("DeleteComment() error = %v, want ErrCommentNotFound", err)
		}
	})
}
