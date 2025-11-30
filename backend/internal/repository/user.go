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

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
}

// SQLiteUserRepository implements UserRepository for SQLite
type SQLiteUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLiteUserRepository creates a new SQLite user repository
func NewSQLiteUserRepository(db *sql.DB, logger *slog.Logger) *SQLiteUserRepository {
	return &SQLiteUserRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUser inserts a new user into the database
func (r *SQLiteUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, bio, image, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Bio,
		user.Image,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
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

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("failed to get last insert id", "error", err)
		return errors.Join(domain.ErrDatabase, err)
	}

	user.ID = id

	r.logger.Info("user created",
		"user_id", user.ID,
		"username", user.Username,
	)

	return nil
}

// GetUserByID retrieves a user by their ID
func (r *SQLiteUserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, bio, image, created_at, updated_at
		FROM users
		WHERE id = ?
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
func (r *SQLiteUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, bio, image, created_at, updated_at
		FROM users
		WHERE email = ?
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
func (r *SQLiteUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, bio, image, created_at, updated_at
		FROM users
		WHERE username = ?
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
func (r *SQLiteUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = ?, username = ?, password_hash = ?, bio = ?, image = ?, updated_at = ?
		WHERE id = ?
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
		if isUniqueConstraintError(err) {
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

// isUniqueConstraintError checks if the error is a SQLite unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "unique constraint")
}
