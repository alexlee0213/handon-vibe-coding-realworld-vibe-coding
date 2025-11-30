package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

// ArticleRepository defines the interface for article data operations
type ArticleRepository interface {
	CreateArticle(ctx context.Context, article *domain.Article, tags []string) error
	GetArticleByID(ctx context.Context, id int64) (*domain.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*domain.Article, error)
	UpdateArticle(ctx context.Context, article *domain.Article) error
	DeleteArticle(ctx context.Context, id int64) error
	ListArticles(ctx context.Context, params *domain.ArticleListParams, currentUserID *int64) ([]*domain.Article, int, error)
	GetFeed(ctx context.Context, userID int64, params *domain.ArticleFeedParams) ([]*domain.Article, int, error)
	SlugExists(ctx context.Context, slug string) bool
	GetAllTags(ctx context.Context) ([]string, error)
	FavoriteArticle(ctx context.Context, articleID, userID int64) error
	UnfavoriteArticle(ctx context.Context, articleID, userID int64) error
}

// SQLiteArticleRepository implements ArticleRepository for SQLite
type SQLiteArticleRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLiteArticleRepository creates a new SQLite article repository
func NewSQLiteArticleRepository(db *sql.DB, logger *slog.Logger) *SQLiteArticleRepository {
	return &SQLiteArticleRepository{
		db:     db,
		logger: logger,
	}
}

// CreateArticle inserts a new article with tags into the database
func (r *SQLiteArticleRepository) CreateArticle(ctx context.Context, article *domain.Article, tags []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}
	defer tx.Rollback()

	now := time.Now()
	article.CreatedAt = now
	article.UpdatedAt = now

	// Insert article
	result, err := tx.ExecContext(ctx, `
		INSERT INTO articles (slug, title, description, body, author_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, article.Slug, article.Title, article.Description, article.Body,
		article.AuthorID, article.CreatedAt, article.UpdatedAt)

	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrArticleAlreadyExists
		}
		r.logger.Error("failed to create article",
			"error", err,
			"slug", article.Slug,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("failed to get last insert id", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}
	article.ID = id

	// Insert tags if provided
	if len(tags) > 0 {
		for _, tagName := range tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			// Get or create tag
			tagID, err := r.getOrCreateTag(ctx, tx, tagName)
			if err != nil {
				return err
			}

			// Link article to tag
			_, err = tx.ExecContext(ctx, `
				INSERT OR IGNORE INTO article_tags (article_id, tag_id)
				VALUES (?, ?)
			`, article.ID, tagID)
			if err != nil {
				r.logger.Error("failed to link article to tag",
					"error", err,
					"article_id", article.ID,
					"tag_id", tagID,
				)
				return errors.Join(domain.ErrDatabase, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	article.TagList = tags

	r.logger.Info("article created",
		"article_id", article.ID,
		"slug", article.Slug,
		"author_id", article.AuthorID,
	)

	return nil
}

// getOrCreateTag returns the ID of an existing tag or creates a new one
func (r *SQLiteArticleRepository) getOrCreateTag(ctx context.Context, tx *sql.Tx, name string) (int64, error) {
	// Try to get existing tag
	var tagID int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, name).Scan(&tagID)
	if err == nil {
		return tagID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		r.logger.Error("failed to query tag", "error", err, "name", name)
		return 0, errors.Join(domain.ErrDatabase, err)
	}

	// Create new tag
	result, err := tx.ExecContext(ctx, `INSERT INTO tags (name) VALUES (?)`, name)
	if err != nil {
		r.logger.Error("failed to create tag", "error", err, "name", name)
		return 0, errors.Join(domain.ErrDatabase, err)
	}

	tagID, err = result.LastInsertId()
	if err != nil {
		r.logger.Error("failed to get tag id", "error", err)
		return 0, errors.Join(domain.ErrDatabase, err)
	}

	return tagID, nil
}

// GetArticleByID retrieves an article by its ID
func (r *SQLiteArticleRepository) GetArticleByID(ctx context.Context, id int64) (*domain.Article, error) {
	article := &domain.Article{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, title, description, body, author_id, created_at, updated_at
		FROM articles
		WHERE id = ?
	`, id).Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.Description,
		&article.Body,
		&article.AuthorID,
		&article.CreatedAt,
		&article.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrArticleNotFound
		}
		r.logger.Error("failed to get article by id", "error", err, "id", id)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	// Load tags
	tags, err := r.getArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.TagList = tags

	// Load favorites count
	article.FavoritesCount, err = r.getFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// GetArticleBySlug retrieves an article by its slug
func (r *SQLiteArticleRepository) GetArticleBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	article := &domain.Article{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, title, description, body, author_id, created_at, updated_at
		FROM articles
		WHERE slug = ?
	`, slug).Scan(
		&article.ID,
		&article.Slug,
		&article.Title,
		&article.Description,
		&article.Body,
		&article.AuthorID,
		&article.CreatedAt,
		&article.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrArticleNotFound
		}
		r.logger.Error("failed to get article by slug", "error", err, "slug", slug)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	// Load tags
	tags, err := r.getArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.TagList = tags

	// Load favorites count
	article.FavoritesCount, err = r.getFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// getArticleTags retrieves all tags for an article
func (r *SQLiteArticleRepository) getArticleTags(ctx context.Context, articleID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.name
		FROM tags t
		INNER JOIN article_tags at ON t.id = at.tag_id
		WHERE at.article_id = ?
		ORDER BY t.name
	`, articleID)
	if err != nil {
		r.logger.Error("failed to get article tags", "error", err, "article_id", articleID)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			r.logger.Error("failed to scan tag", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating tags", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	if tags == nil {
		tags = []string{}
	}

	return tags, nil
}

// getFavoritesCount returns the number of favorites for an article
func (r *SQLiteArticleRepository) getFavoritesCount(ctx context.Context, articleID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM favorites WHERE article_id = ?
	`, articleID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to get favorites count", "error", err, "article_id", articleID)
		return 0, errors.Join(domain.ErrDatabase, err)
	}
	return count, nil
}

// UpdateArticle updates an existing article in the database
func (r *SQLiteArticleRepository) UpdateArticle(ctx context.Context, article *domain.Article) error {
	article.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, `
		UPDATE articles
		SET slug = ?, title = ?, description = ?, body = ?, updated_at = ?
		WHERE id = ?
	`, article.Slug, article.Title, article.Description, article.Body,
		article.UpdatedAt, article.ID)

	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrArticleAlreadyExists
		}
		r.logger.Error("failed to update article",
			"error", err,
			"article_id", article.ID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return domain.ErrArticleNotFound
	}

	r.logger.Info("article updated",
		"article_id", article.ID,
		"slug", article.Slug,
	)

	return nil
}

// DeleteArticle removes an article from the database
func (r *SQLiteArticleRepository) DeleteArticle(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM articles WHERE id = ?`, id)
	if err != nil {
		r.logger.Error("failed to delete article", "error", err, "id", id)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return domain.ErrArticleNotFound
	}

	r.logger.Info("article deleted", "article_id", id)

	return nil
}

// ListArticles retrieves articles with optional filters
func (r *SQLiteArticleRepository) ListArticles(ctx context.Context, params *domain.ArticleListParams, currentUserID *int64) ([]*domain.Article, int, error) {
	// Build query
	query := `
		SELECT DISTINCT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at
		FROM articles a
		LEFT JOIN users u ON a.author_id = u.id
	`

	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM articles a
		LEFT JOIN users u ON a.author_id = u.id
	`

	var conditions []string
	var args []interface{}

	// Filter by tag
	if params.Tag != "" {
		query = `
			SELECT DISTINCT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at
			FROM articles a
			LEFT JOIN users u ON a.author_id = u.id
			INNER JOIN article_tags at ON a.id = at.article_id
			INNER JOIN tags t ON at.tag_id = t.id
		`
		countQuery = `
			SELECT COUNT(DISTINCT a.id)
			FROM articles a
			LEFT JOIN users u ON a.author_id = u.id
			INNER JOIN article_tags at ON a.id = at.article_id
			INNER JOIN tags t ON at.tag_id = t.id
		`
		conditions = append(conditions, "t.name = ?")
		args = append(args, params.Tag)
	}

	// Filter by author
	if params.Author != "" {
		conditions = append(conditions, "u.username = ?")
		args = append(args, params.Author)
	}

	// Filter by favorited
	if params.Favorited != "" {
		query = `
			SELECT DISTINCT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at
			FROM articles a
			LEFT JOIN users u ON a.author_id = u.id
			INNER JOIN favorites f ON a.id = f.article_id
			INNER JOIN users fu ON f.user_id = fu.id
		`
		countQuery = `
			SELECT COUNT(DISTINCT a.id)
			FROM articles a
			LEFT JOIN users u ON a.author_id = u.id
			INNER JOIN favorites f ON a.id = f.article_id
			INNER JOIN users fu ON f.user_id = fu.id
		`
		conditions = append(conditions, "fu.username = ?")
		args = append(args, params.Favorited)
	}

	// Add WHERE clause if conditions exist
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count articles", "error", err)
		return nil, 0, errors.Join(domain.ErrDatabase, err)
	}

	// Add ordering and pagination
	query += " ORDER BY a.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, params.Limit, params.Offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list articles", "error", err)
		return nil, 0, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var articles []*domain.Article
	for rows.Next() {
		article := &domain.Article{}
		err := rows.Scan(
			&article.ID,
			&article.Slug,
			&article.Title,
			&article.Description,
			&article.Body,
			&article.AuthorID,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan article", "error", err)
			return nil, 0, errors.Join(domain.ErrDatabase, err)
		}

		// Load tags for each article
		article.TagList, err = r.getArticleTags(ctx, article.ID)
		if err != nil {
			return nil, 0, err
		}

		// Load favorites count
		article.FavoritesCount, err = r.getFavoritesCount(ctx, article.ID)
		if err != nil {
			return nil, 0, err
		}

		// Check if current user has favorited this article
		if currentUserID != nil {
			article.Favorited, err = r.isArticleFavoritedByUser(ctx, article.ID, *currentUserID)
			if err != nil {
				return nil, 0, err
			}
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating articles", "error", err)
		return nil, 0, errors.Join(domain.ErrDatabase, err)
	}

	if articles == nil {
		articles = []*domain.Article{}
	}

	return articles, total, nil
}

// isArticleFavoritedByUser checks if a user has favorited an article
func (r *SQLiteArticleRepository) isArticleFavoritedByUser(ctx context.Context, articleID, userID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM favorites WHERE article_id = ? AND user_id = ?
	`, articleID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.logger.Error("failed to check favorite", "error", err)
		return false, errors.Join(domain.ErrDatabase, err)
	}
	return true, nil
}

// GetFeed retrieves articles from followed users
func (r *SQLiteArticleRepository) GetFeed(ctx context.Context, userID int64, params *domain.ArticleFeedParams) ([]*domain.Article, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM articles a
		INNER JOIN follows f ON a.author_id = f.following_id
		WHERE f.follower_id = ?
	`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count feed articles", "error", err)
		return nil, 0, errors.Join(domain.ErrDatabase, err)
	}

	// Get articles
	query := `
		SELECT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at
		FROM articles a
		INNER JOIN follows f ON a.author_id = f.following_id
		WHERE f.follower_id = ?
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, params.Limit, params.Offset)
	if err != nil {
		r.logger.Error("failed to get feed", "error", err)
		return nil, 0, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var articles []*domain.Article
	for rows.Next() {
		article := &domain.Article{}
		err := rows.Scan(
			&article.ID,
			&article.Slug,
			&article.Title,
			&article.Description,
			&article.Body,
			&article.AuthorID,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan article", "error", err)
			return nil, 0, errors.Join(domain.ErrDatabase, err)
		}

		// Load tags
		article.TagList, err = r.getArticleTags(ctx, article.ID)
		if err != nil {
			return nil, 0, err
		}

		// Load favorites count
		article.FavoritesCount, err = r.getFavoritesCount(ctx, article.ID)
		if err != nil {
			return nil, 0, err
		}

		// Check if current user has favorited
		article.Favorited, err = r.isArticleFavoritedByUser(ctx, article.ID, userID)
		if err != nil {
			return nil, 0, err
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating feed articles", "error", err)
		return nil, 0, errors.Join(domain.ErrDatabase, err)
	}

	if articles == nil {
		articles = []*domain.Article{}
	}

	return articles, total, nil
}

// SlugExists checks if a slug already exists in the database
func (r *SQLiteArticleRepository) SlugExists(ctx context.Context, slug string) bool {
	var exists int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM articles WHERE slug = ?`, slug).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		r.logger.Error("failed to check slug existence", "error", err, "slug", slug)
		return false
	}
	return true
}

// GetAllTags retrieves all unique tags from the database
func (r *SQLiteArticleRepository) GetAllTags(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM tags ORDER BY name`)
	if err != nil {
		r.logger.Error("failed to get all tags", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			r.logger.Error("failed to scan tag", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating tags", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	if tags == nil {
		tags = []string{}
	}

	return tags, nil
}

// FavoriteArticle adds a favorite relationship between a user and an article
func (r *SQLiteArticleRepository) FavoriteArticle(ctx context.Context, articleID, userID int64) error {
	// Check if already favorited
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM favorites WHERE article_id = ? AND user_id = ?
	`, articleID, userID).Scan(&exists)
	if err == nil {
		return domain.ErrArticleAlreadyFavorited
	}
	if !errors.Is(err, sql.ErrNoRows) {
		r.logger.Error("failed to check favorite exists", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	// Insert favorite
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO favorites (article_id, user_id, created_at)
		VALUES (?, ?, ?)
	`, articleID, userID, time.Now())
	if err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrArticleAlreadyFavorited
		}
		r.logger.Error("failed to favorite article",
			"error", err,
			"article_id", articleID,
			"user_id", userID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	r.logger.Info("article favorited",
		"article_id", articleID,
		"user_id", userID,
	)

	return nil
}

// UnfavoriteArticle removes a favorite relationship between a user and an article
func (r *SQLiteArticleRepository) UnfavoriteArticle(ctx context.Context, articleID, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM favorites WHERE article_id = ? AND user_id = ?
	`, articleID, userID)
	if err != nil {
		r.logger.Error("failed to unfavorite article",
			"error", err,
			"article_id", articleID,
			"user_id", userID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return domain.ErrArticleNotFavorited
	}

	r.logger.Info("article unfavorited",
		"article_id", articleID,
		"user_id", userID,
	)

	return nil
}
