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

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// Test helpers for article tests
func setupArticleTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Drop existing tables for clean state
	db.Exec("DROP TABLE IF EXISTS article_tags")
	db.Exec("DROP TABLE IF EXISTS tags")
	db.Exec("DROP TABLE IF EXISTS favorites")
	db.Exec("DROP TABLE IF EXISTS articles")
	db.Exec("DROP TABLE IF EXISTS follows")
	db.Exec("DROP TABLE IF EXISTS users")

	// Create all required tables
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

		CREATE TABLE articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			body TEXT NOT NULL,
			author_id INTEGER NOT NULL,
			favorites_count INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_articles_slug ON articles(slug);
		CREATE INDEX idx_articles_author_id ON articles(author_id);
		CREATE INDEX idx_articles_created_at ON articles(created_at DESC);

		CREATE TABLE tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);
		CREATE INDEX idx_tags_name ON tags(name);

		CREATE TABLE article_tags (
			article_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (article_id, tag_id),
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		);

		CREATE TABLE favorites (
			user_id INTEGER NOT NULL,
			article_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, article_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
		);

		CREATE TABLE follows (
			follower_id INTEGER NOT NULL,
			followed_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, followed_id),
			FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (followed_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return db
}

func newArticleTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

type articleTestSetup struct {
	handler        *ArticleHandler
	articleService *service.ArticleService
	authService    *service.AuthService
	db             *sql.DB
}

func newTestArticleHandler(t *testing.T) *articleTestSetup {
	t.Helper()
	db := setupArticleTestDB(t)
	logger := newArticleTestLogger()
	userRepo := repository.NewSQLiteUserRepository(db, logger)
	articleRepo := repository.NewSQLiteArticleRepository(db, logger)
	favoriteRepo := repository.NewSQLiteFavoriteRepository(db, logger)
	authService := service.NewAuthService(userRepo, "test-jwt-secret", 24*time.Hour, logger)
	articleService := service.NewArticleService(articleRepo, userRepo, favoriteRepo, logger)
	articleHandler := NewArticleHandler(articleService, logger)

	return &articleTestSetup{
		handler:        articleHandler,
		articleService: articleService,
		authService:    authService,
		db:             db,
	}
}

// Helper to create a test user
func createTestUser(t *testing.T, setup *articleTestSetup, email, username, password string) (*domain.User, string) {
	t.Helper()
	ctx := context.Background()
	user, token, err := setup.authService.Register(ctx, &domain.CreateUserInput{
		Email:    email,
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user, token
}

// Helper to create a test article
func createTestArticle(t *testing.T, setup *articleTestSetup, userID int64, title, description, body string, tags []string) *domain.Article {
	t.Helper()
	ctx := context.Background()
	article, err := setup.articleService.CreateArticle(ctx, userID, &domain.CreateArticleInput{
		Title:       title,
		Description: description,
		Body:        body,
		TagList:     tags,
	})
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}
	return article
}

// =============================================================================
// TDD: POST /api/articles (Create Article) Tests
// =============================================================================

func TestCreateArticleHandler(t *testing.T) {
	t.Run("successfully creates an article", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")

		body := `{"article":{"title":"How to train your dragon","description":"Ever wonder how?","body":"You have to believe","tagList":["reactjs","angularjs","dragons"]}}`
		req := httptest.NewRequest(http.MethodPost, "/api/articles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.CreateArticle(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		article, ok := response["article"].(map[string]interface{})
		if !ok {
			t.Fatal("expected article object in response")
		}

		if article["title"] != "How to train your dragon" {
			t.Errorf("expected title 'How to train your dragon', got %v", article["title"])
		}
		if article["slug"] == nil || article["slug"] == "" {
			t.Error("expected slug in response")
		}
		if article["author"] == nil {
			t.Error("expected author in response")
		}
	})

	t.Run("returns error without authentication", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		body := `{"article":{"title":"Test","description":"Test","body":"Test"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/articles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.CreateArticle(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")

		body := `invalid json`
		req := httptest.NewRequest(http.MethodPost, "/api/articles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.CreateArticle(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("returns error for missing required fields", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")

		body := `{"article":{"title":""}}`
		req := httptest.NewRequest(http.MethodPost, "/api/articles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.CreateArticle(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})
}

// =============================================================================
// TDD: GET /api/articles/{slug} (Get Article) Tests
// =============================================================================

func TestGetArticleHandler(t *testing.T) {
	t.Run("successfully gets an article by slug", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		article := createTestArticle(t, setup, user.ID, "Test Article", "Test description", "Test body", []string{"test"})

		req := httptest.NewRequest(http.MethodGet, "/api/articles/"+article.Slug, nil)
		w := httptest.NewRecorder()

		setup.handler.GetArticle(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		respArticle, ok := response["article"].(map[string]interface{})
		if !ok {
			t.Fatal("expected article object in response")
		}

		if respArticle["slug"] != article.Slug {
			t.Errorf("expected slug %s, got %v", article.Slug, respArticle["slug"])
		}
	})

	t.Run("returns 404 for non-existent article", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/articles/non-existent-slug", nil)
		w := httptest.NewRecorder()

		setup.handler.GetArticle(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns 404 for empty slug", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/articles/", nil)
		w := httptest.NewRecorder()

		setup.handler.GetArticle(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// =============================================================================
// TDD: PUT /api/articles/{slug} (Update Article) Tests
// =============================================================================

func TestUpdateArticleHandler(t *testing.T) {
	t.Run("successfully updates article title", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		article := createTestArticle(t, setup, user.ID, "Original Title", "Description", "Body", nil)

		body := `{"article":{"title":"Updated Title"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/articles/"+article.Slug, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UpdateArticle(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		respArticle := response["article"].(map[string]interface{})
		if respArticle["title"] != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got %v", respArticle["title"])
		}
	})

	t.Run("returns error when not authenticated", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		body := `{"article":{"title":"Updated Title"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/articles/some-slug", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		setup.handler.UpdateArticle(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns error when not the author", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		author, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		otherUser, _ := createTestUser(t, setup, "other@example.com", "other", "password123")
		article := createTestArticle(t, setup, author.ID, "Author's Article", "Description", "Body", nil)

		body := `{"article":{"title":"Hacked Title"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/articles/"+article.Slug, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), UserIDContextKey, otherUser.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UpdateArticle(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("returns 404 for non-existent article", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")

		body := `{"article":{"title":"Updated Title"}}`
		req := httptest.NewRequest(http.MethodPut, "/api/articles/non-existent-slug", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UpdateArticle(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// =============================================================================
// TDD: DELETE /api/articles/{slug} (Delete Article) Tests
// =============================================================================

func TestDeleteArticleHandler(t *testing.T) {
	t.Run("successfully deletes article", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		article := createTestArticle(t, setup, user.ID, "To Delete", "Description", "Body", nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/articles/"+article.Slug, nil)
		ctx := context.WithValue(req.Context(), UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.DeleteArticle(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	})

	t.Run("returns error when not authenticated", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodDelete, "/api/articles/some-slug", nil)
		w := httptest.NewRecorder()

		setup.handler.DeleteArticle(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns error when not the author", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		author, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		otherUser, _ := createTestUser(t, setup, "other@example.com", "other", "password123")
		article := createTestArticle(t, setup, author.ID, "Author's Article", "Description", "Body", nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/articles/"+article.Slug, nil)
		ctx := context.WithValue(req.Context(), UserIDContextKey, otherUser.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.DeleteArticle(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

// =============================================================================
// TDD: GET /api/articles (List Articles) Tests
// =============================================================================

func TestListArticlesHandler(t *testing.T) {
	t.Run("lists all articles", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		createTestArticle(t, setup, user.ID, "Article 1", "Desc 1", "Body 1", nil)
		createTestArticle(t, setup, user.ID, "Article 2", "Desc 2", "Body 2", nil)

		req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
		w := httptest.NewRecorder()

		setup.handler.ListArticles(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		articles, ok := response["articles"].([]interface{})
		if !ok {
			t.Fatal("expected articles array in response")
		}

		if len(articles) != 2 {
			t.Errorf("expected 2 articles, got %d", len(articles))
		}

		count := response["articlesCount"].(float64)
		if count != 2 {
			t.Errorf("expected articlesCount 2, got %v", count)
		}
	})

	t.Run("filters articles by tag", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		createTestArticle(t, setup, user.ID, "Go Article", "Desc", "Body", []string{"go", "programming"})
		createTestArticle(t, setup, user.ID, "Python Article", "Desc", "Body", []string{"python", "programming"})

		req := httptest.NewRequest(http.MethodGet, "/api/articles?tag=go", nil)
		w := httptest.NewRecorder()

		setup.handler.ListArticles(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		articles := response["articles"].([]interface{})
		if len(articles) != 1 {
			t.Errorf("expected 1 article with 'go' tag, got %d", len(articles))
		}
	})

	t.Run("filters articles by author", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user1, _ := createTestUser(t, setup, "author1@example.com", "author1", "password123")
		user2, _ := createTestUser(t, setup, "author2@example.com", "author2", "password123")
		createTestArticle(t, setup, user1.ID, "User1 Article", "Desc", "Body", nil)
		createTestArticle(t, setup, user2.ID, "User2 Article", "Desc", "Body", nil)

		req := httptest.NewRequest(http.MethodGet, "/api/articles?author=author1", nil)
		w := httptest.NewRecorder()

		setup.handler.ListArticles(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		articles := response["articles"].([]interface{})
		if len(articles) != 1 {
			t.Errorf("expected 1 article by author1, got %d", len(articles))
		}
	})

	t.Run("respects pagination limit", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		for i := 0; i < 5; i++ {
			createTestArticle(t, setup, user.ID, "Article", "Desc", "Body", nil)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/articles?limit=2", nil)
		w := httptest.NewRecorder()

		setup.handler.ListArticles(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		articles := response["articles"].([]interface{})
		if len(articles) != 2 {
			t.Errorf("expected 2 articles with limit=2, got %d", len(articles))
		}
	})
}

// =============================================================================
// TDD: GET /api/tags (Get Tags) Tests
// =============================================================================

func TestGetTagsHandler(t *testing.T) {
	t.Run("returns empty list when no articles", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
		w := httptest.NewRecorder()

		setup.handler.GetTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		tags := response["tags"].([]interface{})
		if len(tags) != 0 {
			t.Errorf("expected 0 tags, got %d", len(tags))
		}
	})

	t.Run("returns all unique tags", func(t *testing.T) {
		setup := newTestArticleHandler(t)
		defer setup.db.Close()

		user, _ := createTestUser(t, setup, "author@example.com", "author", "password123")
		createTestArticle(t, setup, user.ID, "Article 1", "Desc", "Body", []string{"go", "programming"})
		createTestArticle(t, setup, user.ID, "Article 2", "Desc", "Body", []string{"python", "programming"})

		req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
		w := httptest.NewRecorder()

		setup.handler.GetTags(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		tags := response["tags"].([]interface{})
		if len(tags) != 3 { // go, programming, python
			t.Errorf("expected 3 unique tags, got %d: %v", len(tags), tags)
		}
	})
}
