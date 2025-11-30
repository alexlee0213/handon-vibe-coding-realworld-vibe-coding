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

	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

func setupCommentTestDB(t *testing.T) (*sql.DB, func()) {
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
	db.Exec("DROP TABLE IF EXISTS article_tags")
	db.Exec("DROP TABLE IF EXISTS favorites")
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

func createCommentTestUser(t *testing.T, db *sql.DB, username, email string) int64 {
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

func createCommentTestArticle(t *testing.T, db *sql.DB, slug, title string, authorID int64) int64 {
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

func createCommentTestComment(t *testing.T, db *sql.DB, body string, articleID, authorID int64) int64 {
	result, err := db.Exec(`
		INSERT INTO comments (body, article_id, author_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, body, articleID, authorID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("failed to create test comment: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func setupCommentHandler(t *testing.T, db *sql.DB) *CommentHandler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userRepo := repository.NewSQLiteUserRepository(db, logger)
	articleRepo := repository.NewSQLiteArticleRepository(db, logger)
	commentRepo := repository.NewSQLiteCommentRepository(db, logger)
	commentService := service.NewCommentService(commentRepo, articleRepo, userRepo, logger)
	return NewCommentHandler(commentService, logger)
}

func TestCommentHandler_GetComments(t *testing.T) {
	db, cleanup := setupCommentTestDB(t)
	defer cleanup()

	handler := setupCommentHandler(t, db)

	authorID := createCommentTestUser(t, db, "testuser", "test@example.com")
	createCommentTestArticle(t, db, "test-article", "Test Article", authorID)
	createCommentTestComment(t, db, "First comment", 1, authorID)
	createCommentTestComment(t, db, "Second comment", 1, authorID)

	t.Run("get comments successfully", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/articles/test-article/comments", nil)
		w := httptest.NewRecorder()

		handler.GetComments(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GetComments() status = %v, want %v", w.Code, http.StatusOK)
		}

		var resp CommentsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Comments) != 2 {
			t.Errorf("GetComments() count = %v, want 2", len(resp.Comments))
		}
	})

	t.Run("get comments for non-existing article", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/articles/non-existing/comments", nil)
		w := httptest.NewRecorder()

		handler.GetComments(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("GetComments() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestCommentHandler_CreateComment(t *testing.T) {
	db, cleanup := setupCommentTestDB(t)
	defer cleanup()

	handler := setupCommentHandler(t, db)

	authorID := createCommentTestUser(t, db, "testuser", "test@example.com")
	createCommentTestArticle(t, db, "test-article", "Test Article", authorID)

	t.Run("create comment successfully", func(t *testing.T) {
		body := `{"comment": {"body": "This is a test comment"}}`
		req := httptest.NewRequest("POST", "/api/articles/test-article/comments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, authorID))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("CreateComment() status = %v, want %v, body: %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var resp CommentResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Comment.Body != "This is a test comment" {
			t.Errorf("CreateComment() body = %v, want 'This is a test comment'", resp.Comment.Body)
		}
	})

	t.Run("create comment without auth", func(t *testing.T) {
		body := `{"comment": {"body": "This is a test comment"}}`
		req := httptest.NewRequest("POST", "/api/articles/test-article/comments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("CreateComment() status = %v, want %v", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("create comment with empty body", func(t *testing.T) {
		body := `{"comment": {"body": ""}}`
		req := httptest.NewRequest("POST", "/api/articles/test-article/comments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, authorID))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("CreateComment() status = %v, want %v", w.Code, http.StatusUnprocessableEntity)
		}
	})

	t.Run("create comment for non-existing article", func(t *testing.T) {
		body := `{"comment": {"body": "This is a test comment"}}`
		req := httptest.NewRequest("POST", "/api/articles/non-existing/comments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, authorID))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("CreateComment() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	db, cleanup := setupCommentTestDB(t)
	defer cleanup()

	handler := setupCommentHandler(t, db)

	authorID := createCommentTestUser(t, db, "testuser", "test@example.com")
	otherUserID := createCommentTestUser(t, db, "otheruser", "other@example.com")
	createCommentTestArticle(t, db, "test-article", "Test Article", authorID)
	commentID := createCommentTestComment(t, db, "Test comment", 1, authorID)

	t.Run("delete comment successfully", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/articles/test-article/comments/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, authorID))
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("DeleteComment() status = %v, want %v", w.Code, http.StatusNoContent)
		}
	})

	t.Run("delete comment by non-author fails", func(t *testing.T) {
		// Create a new comment for this test
		newCommentID := createCommentTestComment(t, db, "Another comment", 1, authorID)

		req := httptest.NewRequest("DELETE", "/api/articles/test-article/comments/"+string(rune('0'+newCommentID)), nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, otherUserID))
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("DeleteComment() status = %v, want %v", w.Code, http.StatusForbidden)
		}
	})

	t.Run("delete non-existing comment", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/articles/test-article/comments/9999", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserIDContextKey, authorID))
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("DeleteComment() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})

	t.Run("delete comment without auth", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/articles/test-article/comments/1", nil)
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("DeleteComment() status = %v, want %v", w.Code, http.StatusUnauthorized)
		}
	})

	// Suppress unused variable warning
	_ = commentID
}
