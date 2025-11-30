package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
)

// PostgresFollowRepository implements FollowRepository for PostgreSQL
type PostgresFollowRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewPostgresFollowRepository creates a new PostgreSQL follow repository
func NewPostgresFollowRepository(db *sql.DB, logger *slog.Logger) *PostgresFollowRepository {
	return &PostgresFollowRepository{
		db:     db,
		logger: logger,
	}
}

// FollowUser creates a follow relationship (followerID follows followingID)
func (r *PostgresFollowRepository) FollowUser(ctx context.Context, followerID, followingID int64) error {
	// Prevent self-follow
	if followerID == followingID {
		return domain.ErrValidation
	}

	query := `
		INSERT INTO follows (follower_id, following_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, followerID, followingID, now)
	if err != nil {
		r.logger.Error("failed to follow user",
			"error", err,
			"follower_id", followerID,
			"following_id", followingID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	r.logger.Info("user followed",
		"follower_id", followerID,
		"following_id", followingID,
	)

	return nil
}

// UnfollowUser removes a follow relationship
func (r *PostgresFollowRepository) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
	query := `
		DELETE FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, followerID, followingID)
	if err != nil {
		r.logger.Error("failed to unfollow user",
			"error", err,
			"follower_id", followerID,
			"following_id", followingID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		// Not following is not an error, just a no-op
		r.logger.Debug("user was not following",
			"follower_id", followerID,
			"following_id", followingID,
		)
		return nil
	}

	r.logger.Info("user unfollowed",
		"follower_id", followerID,
		"following_id", followingID,
	)

	return nil
}

// IsFollowing checks if followerID is following followingID
func (r *PostgresFollowRepository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	if followerID == 0 || followingID == 0 {
		return false, nil
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM follows
			WHERE follower_id = $1 AND following_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, followerID, followingID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check follow status",
			"error", err,
			"follower_id", followerID,
			"following_id", followingID,
		)
		return false, errors.Join(domain.ErrDatabase, err)
	}

	return exists, nil
}

// GetFollowers returns all user IDs who follow the given userID
func (r *PostgresFollowRepository) GetFollowers(ctx context.Context, userID int64) ([]int64, error) {
	query := `
		SELECT follower_id
		FROM follows
		WHERE following_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get followers",
			"error", err,
			"user_id", userID,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var followerIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			r.logger.Error("failed to scan follower id", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		followerIDs = append(followerIDs, id)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating followers", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return followerIDs, nil
}

// GetFollowing returns all user IDs that the given userID is following
func (r *PostgresFollowRepository) GetFollowing(ctx context.Context, userID int64) ([]int64, error) {
	query := `
		SELECT following_id
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get following",
			"error", err,
			"user_id", userID,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	var followingIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			r.logger.Error("failed to scan following id", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		followingIDs = append(followingIDs, id)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating following", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return followingIDs, nil
}

// IsFollowingBulk checks follow status for multiple users at once
// Returns a map of followingID -> isFollowing
func (r *PostgresFollowRepository) IsFollowingBulk(ctx context.Context, followerID int64, followingIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool)

	if followerID == 0 || len(followingIDs) == 0 {
		// Return all false for empty input
		for _, id := range followingIDs {
			result[id] = false
		}
		return result, nil
	}

	// Initialize all to false
	for _, id := range followingIDs {
		result[id] = false
	}

	// Build query with PostgreSQL placeholders
	placeholders := make([]interface{}, len(followingIDs)+1)
	placeholders[0] = followerID
	dollarSigns := make([]string, len(followingIDs))
	for i, id := range followingIDs {
		placeholders[i+1] = id
		dollarSigns[i] = fmt.Sprintf("$%d", i+2)
	}

	query := `
		SELECT following_id
		FROM follows
		WHERE follower_id = $1 AND following_id IN (` + strings.Join(dollarSigns, ", ") + `)
	`

	rows, err := r.db.QueryContext(ctx, query, placeholders...)
	if err != nil {
		r.logger.Error("failed to check bulk follow status",
			"error", err,
			"follower_id", followerID,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}
	defer rows.Close()

	for rows.Next() {
		var followingID int64
		if err := rows.Scan(&followingID); err != nil {
			r.logger.Error("failed to scan following id", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		result[followingID] = true
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating bulk follow status", "error", err)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return result, nil
}
