package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

// TagRepository defines the interface for tag data operations
type TagRepository interface {
	GetAllTags(ctx context.Context) ([]string, error)
	GetTagByID(ctx context.Context, id int64) (*domain.Tag, error)
	GetTagByName(ctx context.Context, name string) (*domain.Tag, error)
	GetTagsByArticleID(ctx context.Context, articleID int64) ([]string, error)
}

// SQLiteTagRepository implements TagRepository for SQLite
type SQLiteTagRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLiteTagRepository creates a new SQLite tag repository
func NewSQLiteTagRepository(db *sql.DB, logger *slog.Logger) *SQLiteTagRepository {
	return &SQLiteTagRepository{
		db:     db,
		logger: logger,
	}
}

// GetAllTags retrieves all unique tags from the database
func (r *SQLiteTagRepository) GetAllTags(ctx context.Context) ([]string, error) {
	query := `SELECT name FROM tags ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
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

// GetTagByID retrieves a tag by its ID
func (r *SQLiteTagRepository) GetTagByID(ctx context.Context, id int64) (*domain.Tag, error) {
	query := `SELECT id, name FROM tags WHERE id = ?`

	tag := &domain.Tag{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDatabase // No specific tag error, using general database error
		}
		r.logger.Error("failed to get tag by id", "error", err, "tag_id", id)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return tag, nil
}

// GetTagByName retrieves a tag by its name
func (r *SQLiteTagRepository) GetTagByName(ctx context.Context, name string) (*domain.Tag, error) {
	query := `SELECT id, name FROM tags WHERE name = ?`

	tag := &domain.Tag{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Tag not found, return nil without error
		}
		r.logger.Error("failed to get tag by name", "error", err, "tag_name", name)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return tag, nil
}

// GetTagsByArticleID retrieves all tags for an article
func (r *SQLiteTagRepository) GetTagsByArticleID(ctx context.Context, articleID int64) ([]string, error) {
	query := `
		SELECT t.name
		FROM tags t
		INNER JOIN article_tags at ON t.id = at.tag_id
		WHERE at.article_id = ?
		ORDER BY t.name
	`

	rows, err := r.db.QueryContext(ctx, query, articleID)
	if err != nil {
		r.logger.Error("failed to get tags by article id", "error", err, "article_id", articleID)
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
