package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestTagDB(t *testing.T) (*sql.DB, func()) {
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
	db.Exec("DROP TABLE IF EXISTS article_tags")
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

	return db, func() {
		db.Close()
	}
}

func createTestUserForTag(t *testing.T, db *sql.DB, username, email string) int64 {
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

func createTestArticleForTag(t *testing.T, db *sql.DB, slug, title string, authorID int64) int64 {
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

func createTestTag(t *testing.T, db *sql.DB, name string) int64 {
	result, err := db.Exec(`INSERT INTO tags (name) VALUES (?)`, name)
	if err != nil {
		t.Fatalf("failed to create test tag: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func linkTagToArticle(t *testing.T, db *sql.DB, articleID, tagID int64) {
	_, err := db.Exec(`INSERT INTO article_tags (article_id, tag_id) VALUES (?, ?)`, articleID, tagID)
	if err != nil {
		t.Fatalf("failed to link tag to article: %v", err)
	}
}

func TestTagRepository_GetAllTags(t *testing.T) {
	db, cleanup := setupTestTagDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteTagRepository(db, logger)

	// Create some tags
	createTestTag(t, db, "go")
	createTestTag(t, db, "javascript")
	createTestTag(t, db, "python")

	t.Run("get all tags", func(t *testing.T) {
		tags, err := repo.GetAllTags(context.Background())
		if err != nil {
			t.Errorf("GetAllTags() error = %v", err)
			return
		}

		if len(tags) != 3 {
			t.Errorf("GetAllTags() count = %v, want 3", len(tags))
		}

		// Tags should be ordered by name
		expected := []string{"go", "javascript", "python"}
		for i, tag := range tags {
			if tag != expected[i] {
				t.Errorf("GetAllTags()[%d] = %v, want %v", i, tag, expected[i])
			}
		}
	})

	t.Run("get all tags when empty", func(t *testing.T) {
		// Clear tags
		db.Exec("DELETE FROM tags")

		tags, err := repo.GetAllTags(context.Background())
		if err != nil {
			t.Errorf("GetAllTags() error = %v", err)
			return
		}

		if len(tags) != 0 {
			t.Errorf("GetAllTags() count = %v, want 0", len(tags))
		}
	})
}

func TestTagRepository_GetTagByID(t *testing.T) {
	db, cleanup := setupTestTagDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteTagRepository(db, logger)

	tagID := createTestTag(t, db, "golang")

	t.Run("get existing tag by ID", func(t *testing.T) {
		tag, err := repo.GetTagByID(context.Background(), tagID)
		if err != nil {
			t.Errorf("GetTagByID() error = %v", err)
			return
		}

		if tag.Name != "golang" {
			t.Errorf("GetTagByID() name = %v, want 'golang'", tag.Name)
		}
	})

	t.Run("get non-existing tag by ID", func(t *testing.T) {
		_, err := repo.GetTagByID(context.Background(), 999999)
		if err == nil {
			t.Error("GetTagByID() expected error for non-existing tag")
		}
	})
}

func TestTagRepository_GetTagByName(t *testing.T) {
	db, cleanup := setupTestTagDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteTagRepository(db, logger)

	createTestTag(t, db, "react")

	t.Run("get existing tag by name", func(t *testing.T) {
		tag, err := repo.GetTagByName(context.Background(), "react")
		if err != nil {
			t.Errorf("GetTagByName() error = %v", err)
			return
		}

		if tag == nil {
			t.Error("GetTagByName() returned nil for existing tag")
			return
		}

		if tag.Name != "react" {
			t.Errorf("GetTagByName() name = %v, want 'react'", tag.Name)
		}
	})

	t.Run("get non-existing tag by name", func(t *testing.T) {
		tag, err := repo.GetTagByName(context.Background(), "nonexistent")
		if err != nil {
			t.Errorf("GetTagByName() error = %v", err)
			return
		}

		if tag != nil {
			t.Error("GetTagByName() expected nil for non-existing tag")
		}
	})
}

func TestTagRepository_GetTagsByArticleID(t *testing.T) {
	db, cleanup := setupTestTagDB(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := NewSQLiteTagRepository(db, logger)

	authorID := createTestUserForTag(t, db, "testuser", "test@example.com")
	articleID := createTestArticleForTag(t, db, "test-article", "Test Article", authorID)

	// Create tags and link to article
	tag1ID := createTestTag(t, db, "backend")
	tag2ID := createTestTag(t, db, "api")
	createTestTag(t, db, "unrelated") // This one won't be linked

	linkTagToArticle(t, db, articleID, tag1ID)
	linkTagToArticle(t, db, articleID, tag2ID)

	t.Run("get tags for article", func(t *testing.T) {
		tags, err := repo.GetTagsByArticleID(context.Background(), articleID)
		if err != nil {
			t.Errorf("GetTagsByArticleID() error = %v", err)
			return
		}

		if len(tags) != 2 {
			t.Errorf("GetTagsByArticleID() count = %v, want 2", len(tags))
		}

		// Tags should be ordered by name: "api", "backend"
		expected := []string{"api", "backend"}
		for i, tag := range tags {
			if tag != expected[i] {
				t.Errorf("GetTagsByArticleID()[%d] = %v, want %v", i, tag, expected[i])
			}
		}
	})

	t.Run("get tags for article with no tags", func(t *testing.T) {
		newArticleID := createTestArticleForTag(t, db, "no-tags", "No Tags", authorID)

		tags, err := repo.GetTagsByArticleID(context.Background(), newArticleID)
		if err != nil {
			t.Errorf("GetTagsByArticleID() error = %v", err)
			return
		}

		if len(tags) != 0 {
			t.Errorf("GetTagsByArticleID() count = %v, want 0", len(tags))
		}
	})

	t.Run("get tags for non-existing article", func(t *testing.T) {
		tags, err := repo.GetTagsByArticleID(context.Background(), 999999)
		if err != nil {
			t.Errorf("GetTagsByArticleID() error = %v", err)
			return
		}

		if len(tags) != 0 {
			t.Errorf("GetTagsByArticleID() count = %v, want 0", len(tags))
		}
	})
}
