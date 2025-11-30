package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

// PostgresCommentRepository implements CommentRepository for PostgreSQL
type PostgresCommentRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewPostgresCommentRepository creates a new PostgreSQL comment repository
func NewPostgresCommentRepository(db *sql.DB, logger *slog.Logger) *PostgresCommentRepository {
	return &PostgresCommentRepository{
		db:     db,
		logger: logger,
	}
}

// CreateComment inserts a new comment into the database
func (r *PostgresCommentRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (body, article_id, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		comment.Body,
		comment.ArticleID,
		comment.AuthorID,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)

	if err != nil {
		r.logger.Error("failed to create comment",
			"error", err,
			"article_id", comment.ArticleID,
			"author_id", comment.AuthorID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	r.logger.Info("comment created",
		"comment_id", comment.ID,
		"article_id", comment.ArticleID,
		"author_id", comment.AuthorID,
	)

	return nil
}

// GetCommentByID retrieves a comment by its ID
func (r *PostgresCommentRepository) GetCommentByID(ctx context.Context, id int64) (*domain.Comment, error) {
	query := `
		SELECT id, body, article_id, author_id, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	comment := &domain.Comment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.Body,
		&comment.ArticleID,
		&comment.AuthorID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCommentNotFound
		}
		r.logger.Error("failed to get comment by id",
			"error", err,
			"comment_id", id,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return comment, nil
}

// GetCommentsByArticleID retrieves all comments for an article
func (r *PostgresCommentRepository) GetCommentsByArticleID(ctx context.Context, articleID int64) ([]*domain.Comment, error) {
	query := `
		SELECT id, body, article_id, author_id, created_at, updated_at
		FROM comments
		WHERE article_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, articleID)
	if err != nil {
		r.logger.Error("failed to get comments by article id",
			"error", err,
			"article_id", articleID,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		comment := &domain.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.Body,
			&comment.ArticleID,
			&comment.AuthorID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan comment", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating comments", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	if comments == nil {
		comments = []*domain.Comment{}
	}

	return comments, nil
}

// DeleteComment removes a comment from the database
func (r *PostgresCommentRepository) DeleteComment(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		r.logger.Error("failed to delete comment", "error", err, "comment_id", id)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return domain.ErrCommentNotFound
	}

	r.logger.Info("comment deleted", "comment_id", id)

	return nil
}
