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

// PostgresUserRepository implements UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sql.DB, logger *slog.Logger) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUser inserts a new user into the database
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, bio, image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Bio,
		user.Image,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		if isPostgresUniqueConstraintError(err) {
			if strings.Contains(err.Error(), "email") {
				return domain.ErrEmailAlreadyTaken
			}
			if strings.Contains(err.Error(), "username") {
				return domain.ErrUsernameAlreadyTaken
			}
			return domain.ErrUserAlreadyExists
		}
		r.logger.Error("failed to create user",
			"error", err,
			"email", user.Email,
			"username", user.Username,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	r.logger.Info("user created",
		"user_id", user.ID,
		"username", user.Username,
	)

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, bio, image, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Bio,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("failed to get user by id",
			"error", err,
			"user_id", id,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, bio, image, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Bio,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("failed to get user by email",
			"error", err,
			"email", email,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by their username
func (r *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, bio, image, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Bio,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		r.logger.Error("failed to get user by username",
			"error", err,
			"username", username,
		)
		return nil, errors.Join(domain.ErrDatabase, err)
	}

	return user, nil
}

// UpdateUser updates an existing user in the database
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, password_hash = $3, bio = $4, image = $5, updated_at = $6
		WHERE id = $7
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Bio,
		user.Image,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		if isPostgresUniqueConstraintError(err) {
			if strings.Contains(err.Error(), "email") {
				return domain.ErrEmailAlreadyTaken
			}
			if strings.Contains(err.Error(), "username") {
				return domain.ErrUsernameAlreadyTaken
			}
			return domain.ErrUserAlreadyExists
		}
		r.logger.Error("failed to update user",
			"error", err,
			"user_id", user.ID,
		)
		return errors.Join(domain.ErrDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	r.logger.Info("user updated",
		"user_id", user.ID,
		"username", user.Username,
	)

	return nil
}

// isPostgresUniqueConstraintError checks if the error is a PostgreSQL unique constraint violation
func isPostgresUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// PostgreSQL unique constraint error codes/messages
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "23505") // PostgreSQL error code for unique violation
}
