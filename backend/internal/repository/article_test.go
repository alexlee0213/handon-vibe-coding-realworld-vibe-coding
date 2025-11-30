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

func setupTestArticleDB(t *testing.T) (*sql.DB, func()) {
	// Use file:memory with shared cache to ensure tables persist across connections
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Drop existing tables (for shared cache cleanup between tests)
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

	return db, func() {
		db.Close()
	}
}

func createTestUser(t *testing.T, db *sql.DB, username, email string) int64 {
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

func TestArticleRepository_CreateArticle(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	// Create a test user first
	authorID := createTestUser(t, db, "testuser", "test@example.com")

	tests := []struct {
		name    string
		article *domain.Article
		tags    []string
		wantErr bool
	}{
		{
			name: "create article without tags",
			article: &domain.Article{
				Slug:        "hello-world",
				Title:       "Hello World",
				Description: "A test article",
				Body:        "This is the body of the article.",
				AuthorID:    authorID,
			},
			tags:    nil,
			wantErr: false,
		},
		{
			name: "create article with tags",
			article: &domain.Article{
				Slug:        "another-article",
				Title:       "Another Article",
				Description: "Another test article",
				Body:        "This is another article body.",
				AuthorID:    authorID,
			},
			tags:    []string{"go", "programming", "tutorial"},
			wantErr: false,
		},
		{
			name: "create article with duplicate slug",
			article: &domain.Article{
				Slug:        "hello-world", // Already exists
				Title:       "Duplicate Title",
				Description: "Duplicate",
				Body:        "Duplicate body",
				AuthorID:    authorID,
			},
			tags:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateArticle(context.Background(), tt.article, tt.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.article.ID == 0 {
				t.Error("CreateArticle() did not set article ID")
			}
		})
	}
}

func TestArticleRepository_GetArticleBySlug(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	authorID := createTestUser(t, db, "testuser", "test@example.com")

	// Create a test article
	article := &domain.Article{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "Test description",
		Body:        "Test body",
		AuthorID:    authorID,
	}
	tags := []string{"test", "golang"}
	err := repo.CreateArticle(context.Background(), article, tags)
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}

	tests := []struct {
		name    string
		slug    string
		wantErr error
	}{
		{
			name:    "get existing article",
			slug:    "test-article",
			wantErr: nil,
		},
		{
			name:    "get non-existing article",
			slug:    "non-existing",
			wantErr: domain.ErrArticleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetArticleBySlug(context.Background(), tt.slug)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("GetArticleBySlug() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetArticleBySlug() unexpected error: %v", err)
				return
			}

			if result.Slug != tt.slug {
				t.Errorf("GetArticleBySlug() slug = %v, want %v", result.Slug, tt.slug)
			}

			// Check that tags are loaded
			if len(result.TagList) != 2 {
				t.Errorf("GetArticleBySlug() tagList length = %v, want 2", len(result.TagList))
			}
		})
	}
}

func TestArticleRepository_UpdateArticle(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	authorID := createTestUser(t, db, "testuser", "test@example.com")

	// Create a test article
	article := &domain.Article{
		Slug:        "original-slug",
		Title:       "Original Title",
		Description: "Original description",
		Body:        "Original body",
		AuthorID:    authorID,
	}
	err := repo.CreateArticle(context.Background(), article, nil)
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}

	// Update the article
	article.Title = "Updated Title"
	article.Slug = "updated-slug"
	article.Description = "Updated description"

	err = repo.UpdateArticle(context.Background(), article)
	if err != nil {
		t.Errorf("UpdateArticle() unexpected error: %v", err)
		return
	}

	// Verify the update
	updated, err := repo.GetArticleBySlug(context.Background(), "updated-slug")
	if err != nil {
		t.Errorf("GetArticleBySlug() after update unexpected error: %v", err)
		return
	}

	if updated.Title != "Updated Title" {
		t.Errorf("UpdateArticle() title = %v, want 'Updated Title'", updated.Title)
	}

	if updated.Description != "Updated description" {
		t.Errorf("UpdateArticle() description = %v, want 'Updated description'", updated.Description)
	}
}

func TestArticleRepository_DeleteArticle(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	authorID := createTestUser(t, db, "testuser", "test@example.com")

	// Create a test article
	article := &domain.Article{
		Slug:        "to-delete",
		Title:       "To Delete",
		Description: "This will be deleted",
		Body:        "Delete me",
		AuthorID:    authorID,
	}
	err := repo.CreateArticle(context.Background(), article, []string{"deletable"})
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}

	// Delete the article
	err = repo.DeleteArticle(context.Background(), article.ID)
	if err != nil {
		t.Errorf("DeleteArticle() unexpected error: %v", err)
		return
	}

	// Verify deletion
	_, err = repo.GetArticleBySlug(context.Background(), "to-delete")
	if err != domain.ErrArticleNotFound {
		t.Errorf("GetArticleBySlug() after delete error = %v, want ErrArticleNotFound", err)
	}
}

func TestArticleRepository_ListArticles(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	// Create test users
	author1ID := createTestUser(t, db, "author1", "author1@example.com")
	author2ID := createTestUser(t, db, "author2", "author2@example.com")

	// Create test articles
	articles := []struct {
		article *domain.Article
		tags    []string
	}{
		{
			article: &domain.Article{
				Slug:        "go-basics",
				Title:       "Go Basics",
				Description: "Learn Go",
				Body:        "Go is great",
				AuthorID:    author1ID,
			},
			tags: []string{"go", "tutorial"},
		},
		{
			article: &domain.Article{
				Slug:        "python-basics",
				Title:       "Python Basics",
				Description: "Learn Python",
				Body:        "Python is cool",
				AuthorID:    author1ID,
			},
			tags: []string{"python", "tutorial"},
		},
		{
			article: &domain.Article{
				Slug:        "rust-basics",
				Title:       "Rust Basics",
				Description: "Learn Rust",
				Body:        "Rust is fast",
				AuthorID:    author2ID,
			},
			tags: []string{"rust", "systems"},
		},
	}

	for _, a := range articles {
		err := repo.CreateArticle(context.Background(), a.article, a.tags)
		if err != nil {
			t.Fatalf("failed to create test article: %v", err)
		}
	}

	tests := []struct {
		name       string
		params     *domain.ArticleListParams
		wantCount  int
		wantTitles []string
	}{
		{
			name: "list all articles",
			params: &domain.ArticleListParams{
				Limit:  20,
				Offset: 0,
			},
			wantCount: 3,
		},
		{
			name: "filter by tag",
			params: &domain.ArticleListParams{
				Tag:    "go",
				Limit:  20,
				Offset: 0,
			},
			wantCount:  1,
			wantTitles: []string{"Go Basics"},
		},
		{
			name: "filter by tutorial tag",
			params: &domain.ArticleListParams{
				Tag:    "tutorial",
				Limit:  20,
				Offset: 0,
			},
			wantCount: 2,
		},
		{
			name: "filter by author",
			params: &domain.ArticleListParams{
				Author: "author1",
				Limit:  20,
				Offset: 0,
			},
			wantCount: 2,
		},
		{
			name: "pagination - limit",
			params: &domain.ArticleListParams{
				Limit:  2,
				Offset: 0,
			},
			wantCount: 2,
		},
		{
			name: "pagination - offset",
			params: &domain.ArticleListParams{
				Limit:  20,
				Offset: 2,
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := repo.ListArticles(context.Background(), tt.params, nil)
			if err != nil {
				t.Errorf("ListArticles() unexpected error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("ListArticles() count = %v, want %v", len(result), tt.wantCount)
			}

			// For non-paginated queries, total should be >= count
			if tt.params.Offset == 0 && tt.params.Limit >= 20 && total < tt.wantCount {
				t.Errorf("ListArticles() total = %v, want >= %v", total, tt.wantCount)
			}

			if len(tt.wantTitles) > 0 {
				for i, title := range tt.wantTitles {
					if i < len(result) && result[i].Title != title {
						t.Errorf("ListArticles() result[%d].Title = %v, want %v", i, result[i].Title, title)
					}
				}
			}
		})
	}
}

func TestArticleRepository_SlugExists(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	authorID := createTestUser(t, db, "testuser", "test@example.com")

	// Create a test article
	article := &domain.Article{
		Slug:        "existing-slug",
		Title:       "Existing Title",
		Description: "Existing description",
		Body:        "Existing body",
		AuthorID:    authorID,
	}
	err := repo.CreateArticle(context.Background(), article, nil)
	if err != nil {
		t.Fatalf("failed to create test article: %v", err)
	}

	tests := []struct {
		name   string
		slug   string
		exists bool
	}{
		{
			name:   "existing slug",
			slug:   "existing-slug",
			exists: true,
		},
		{
			name:   "non-existing slug",
			slug:   "non-existing-slug",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.SlugExists(context.Background(), tt.slug)
			if result != tt.exists {
				t.Errorf("SlugExists(%q) = %v, want %v", tt.slug, result, tt.exists)
			}
		})
	}
}

func TestArticleRepository_GetArticleTags(t *testing.T) {
	db, cleanup := setupTestArticleDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteArticleRepository(db, logger)

	authorID := createTestUser(t, db, "testuser", "test@example.com")

	// Create articles with tags
	article1 := &domain.Article{
		Slug:        "article-1",
		Title:       "Article 1",
		Description: "Desc 1",
		Body:        "Body 1",
		AuthorID:    authorID,
	}
	err := repo.CreateArticle(context.Background(), article1, []string{"go", "tutorial"})
	if err != nil {
		t.Fatalf("failed to create test article 1: %v", err)
	}

	article2 := &domain.Article{
		Slug:        "article-2",
		Title:       "Article 2",
		Description: "Desc 2",
		Body:        "Body 2",
		AuthorID:    authorID,
	}
	err = repo.CreateArticle(context.Background(), article2, []string{"go", "programming"})
	if err != nil {
		t.Fatalf("failed to create test article 2: %v", err)
	}

	// Get all tags
	tags, err := repo.GetAllTags(context.Background())
	if err != nil {
		t.Errorf("GetAllTags() unexpected error: %v", err)
		return
	}

	if len(tags) != 3 {
		t.Errorf("GetAllTags() count = %v, want 3 (go, tutorial, programming)", len(tags))
	}
}
