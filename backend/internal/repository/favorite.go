package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

// FavoriteRepository defines the interface for article favorite data operations
type FavoriteRepository interface {
	// Favorite adds an article to user's favorites
	Favorite(ctx context.Context, userID, articleID int64) error
	// Unfavorite removes an article from user's favorites
	Unfavorite(ctx context.Context, userID, articleID int64) error
	// IsFavorited checks if a user has favorited an article
	IsFavorited(ctx context.Context, userID, articleID int64) (bool, error)
	// GetFavoritesCount returns the number of favorites for an article
	GetFavoritesCount(ctx context.Context, articleID int64) (int, error)
	// IsFavoritedBulk checks favorite status for multiple articles at once
	IsFavoritedBulk(ctx context.Context, userID int64, articleIDs []int64) (map[int64]bool, error)
}

// SQLiteFavoriteRepository implements FavoriteRepository for SQLite
type SQLiteFavoriteRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLiteFavoriteRepository creates a new SQLite favorite repository
func NewSQLiteFavoriteRepository(db *sql.DB, logger *slog.Logger) *SQLiteFavoriteRepository {
	return &SQLiteFavoriteRepository{
		db:     db,
		logger: logger,
	}
}

// Favorite adds an article to user's favorites
func (r *SQLiteFavoriteRepository) Favorite(ctx context.Context, userID, articleID int64) error {
	query := `
		INSERT INTO favorites (user_id, article_id, created_at)
		VALUES (?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, articleID, now)
	if err != nil {
		// Check if already favorited (unique constraint violation)
		if isUniqueConstraintError(err) {
			// Already favorited is not an error, just a no-op
			r.logger.Debug("article already favorited",
				"user_id", userID,
				"article_id", articleID,
			)
			return nil
		}
		r.logger.Error("failed to favorite article",
			"error", err,
			"user_id", userID,
			"article_id", articleID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	r.logger.Info("article favorited",
		"user_id", userID,
		"article_id", articleID,
	)

	return nil
}

// Unfavorite removes an article from user's favorites
func (r *SQLiteFavoriteRepository) Unfavorite(ctx context.Context, userID, articleID int64) error {
	query := `
		DELETE FROM favorites
		WHERE user_id = ? AND article_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, userID, articleID)
	if err != nil {
		r.logger.Error("failed to unfavorite article",
			"error", err,
			"user_id", userID,
			"article_id", articleID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		// Not favorited is not an error, just a no-op
		r.logger.Debug("article was not favorited",
			"user_id", userID,
			"article_id", articleID,
		)
		return nil
	}

	r.logger.Info("article unfavorited",
		"user_id", userID,
		"article_id", articleID,
	)

	return nil
}

// IsFavorited checks if a user has favorited an article
func (r *SQLiteFavoriteRepository) IsFavorited(ctx context.Context, userID, articleID int64) (bool, error) {
	if userID == 0 || articleID == 0 {
		return false, nil
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM favorites
			WHERE user_id = ? AND article_id = ?
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check favorite status",
			"error", err,
			"user_id", userID,
			"article_id", articleID,
		)
		return false, errors.Join(domain.ErrDatabase, err)
	}

	return exists, nil
}

// GetFavoritesCount returns the number of favorites for an article
func (r *SQLiteFavoriteRepository) GetFavoritesCount(ctx context.Context, articleID int64) (int, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE article_id = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, articleID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to get favorites count",
			"error", err,
			"article_id", articleID,
		)
		return 0, errors.Join(domain.ErrDatabase, err)
	}

	return count, nil
}

// IsFavoritedBulk checks favorite status for multiple articles at once
// Returns a map of articleID -> isFavorited
func (r *SQLiteFavoriteRepository) IsFavoritedBulk(ctx context.Context, userID int64, articleIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool)

	if userID == 0 || len(articleIDs) == 0 {
		// Return all false for empty input
		for _, id := range articleIDs {
			result[id] = false
		}
		return result, nil
	}

	// Initialize all to false
	for _, id := range articleIDs {
		result[id] = false
	}

	// Build query with placeholders
	placeholders := make([]interface{}, len(articleIDs)+1)
	placeholders[0] = userID
	questionMarks := ""
	for i, id := range articleIDs {
		placeholders[i+1] = id
		if i > 0 {
			questionMarks += ", "
		}
		questionMarks += "?"
	}

	query := `
		SELECT article_id
		FROM favorites
		WHERE user_id = ? AND article_id IN (` + questionMarks + `)
	`

	rows, err := r.db.QueryContext(ctx, query, placeholders...)
	if err != nil {
		r.logger.Error("failed to check bulk favorite status",
			"error", err,
			"user_id", userID,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	for rows.Next() {
		var articleID int64
		if err := rows.Scan(&articleID); err != nil {
			r.logger.Error("failed to scan article id", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		result[articleID] = true
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating bulk favorite status", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return result, nil
}
