package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexlee0213/realworld-conduit/backend/internal/api/handler"
	"github.com/alexlee0213/realworld-conduit/backend/internal/api/middleware"
	"github.com/alexlee0213/realworld-conduit/backend/internal/config"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite driver (for development)
)

// DatabaseType represents the type of database being used
type DatabaseType string

const (
	DatabaseTypeSQLite   DatabaseType = "sqlite"
	DatabaseTypePostgres DatabaseType = "postgres"
)

type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
	config *config.Config
	db     *sql.DB
	dbType DatabaseType
}

func NewRouter(cfg *config.Config, logger *slog.Logger) (*Router, error) {
	// Initialize database
	db, dbType, err := initDatabase(cfg.Database.URL, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("database initialized", "type", dbType, "url_prefix", maskDatabaseURL(cfg.Database.URL))

	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
		db:     db,
		dbType: dbType,
	}, nil
}

// maskDatabaseURL hides sensitive parts of the database URL for logging
func maskDatabaseURL(url string) string {
	if strings.HasPrefix(url, "postgres://") {
		// Mask credentials in postgres URL
		parts := strings.SplitN(url, "@", 2)
		if len(parts) == 2 {
			return "postgres://****@" + parts[1]
		}
		return "postgres://..."
	}
	return url
}

func initDatabase(databaseURL string, logger *slog.Logger) (*sql.DB, DatabaseType, error) {
	// Detect database type from URL
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return initPostgresDatabase(databaseURL, logger)
	}

	// Default to SQLite for development
	return initSQLiteDatabase(databaseURL, logger)
}

func initPostgresDatabase(databaseURL string, logger *slog.Logger) (*sql.DB, DatabaseType, error) {
	logger.Debug("connecting to PostgreSQL database")

	// pgx/stdlib uses "pgx" as driver name
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, DatabaseTypePostgres, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Configure connection pool for production
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, DatabaseTypePostgres, fmt.Errorf("failed to ping postgres: %w", err)
	}

	logger.Debug("PostgreSQL connection established")

	// Run migrations
	if err := runPostgresMigrations(db, logger); err != nil {
		db.Close()
		return nil, DatabaseTypePostgres, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, DatabaseTypePostgres, nil
}

// runPostgresMigrations runs database migrations for PostgreSQL
func runPostgresMigrations(db *sql.DB, logger *slog.Logger) error {
	logger.Info("running PostgreSQL migrations")

	// Find migrations directory
	migrationsPath, err := findMigrationsPath()
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	logger.Debug("migrations path found", "path", migrationsPath)

	// Create postgres driver for migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	logger.Info("migrations completed", "version", version, "dirty", dirty)

	return nil
}

// findMigrationsPath finds the PostgreSQL migrations directory
func findMigrationsPath() (string, error) {
	// Try common paths
	paths := []string{
		"db/migrations_postgres",       // From project root
		"./db/migrations_postgres",     // Relative
		"../db/migrations_postgres",    // From cmd/server
		"../../db/migrations_postgres", // From deeper directories
		"/app/db/migrations_postgres",  // Docker container path
	}

	// Also check relative to executable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		paths = append(paths,
			filepath.Join(execDir, "db/migrations_postgres"),
			filepath.Join(execDir, "../db/migrations_postgres"),
		)
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found, tried: %v", paths)
}

func initSQLiteDatabase(databaseURL string, logger *slog.Logger) (*sql.DB, DatabaseType, error) {
	logger.Debug("connecting to SQLite database")

	// Parse database URL (supports both sqlite:// and sqlite3:// prefixes)
	// golang-migrate uses sqlite3://, so we support both for consistency
	dbPath := databaseURL
	if strings.HasPrefix(dbPath, "sqlite3://") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite3://")
	} else if strings.HasPrefix(dbPath, "sqlite://") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite://")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, DatabaseTypeSQLite, fmt.Errorf("failed to open sqlite connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, DatabaseTypeSQLite, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	logger.Debug("SQLite connection established", "path", dbPath)
	return db, DatabaseTypeSQLite, nil
}

func (r *Router) Setup() http.Handler {
	// Initialize repositories based on database type
	var userRepo repository.UserRepository
	var articleRepo repository.ArticleRepository
	var commentRepo repository.CommentRepository
	var followRepo repository.FollowRepository

	switch r.dbType {
	case DatabaseTypePostgres:
		r.logger.Info("using PostgreSQL repositories")
		userRepo = repository.NewPostgresUserRepository(r.db, r.logger)
		articleRepo = repository.NewPostgresArticleRepository(r.db, r.logger)
		commentRepo = repository.NewPostgresCommentRepository(r.db, r.logger)
		followRepo = repository.NewPostgresFollowRepository(r.db, r.logger)
	default:
		r.logger.Info("using SQLite repositories")
		userRepo = repository.NewSQLiteUserRepository(r.db, r.logger)
		articleRepo = repository.NewSQLiteArticleRepository(r.db, r.logger)
		commentRepo = repository.NewSQLiteCommentRepository(r.db, r.logger)
		followRepo = repository.NewSQLiteFollowRepository(r.db, r.logger)
	}

	// Initialize services
	authService := service.NewAuthService(
		userRepo,
		r.config.JWT.Secret,
		r.config.JWT.Expiry,
		r.logger,
	)
	articleService := service.NewArticleService(articleRepo, userRepo, r.logger)
	commentService := service.NewCommentService(commentRepo, articleRepo, userRepo, r.logger)
	profileService := service.NewProfileService(userRepo, followRepo, r.logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(authService, r.logger)
	articleHandler := handler.NewArticleHandler(articleService, r.logger)
	commentHandler := handler.NewCommentHandler(commentService, r.logger)
	profileHandler := handler.NewProfileHandler(profileService, r.logger)

	// Health check
	r.mux.HandleFunc("GET /health", healthHandler.Health)

	// API info endpoint
	r.mux.HandleFunc("GET /api/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "RealWorld Conduit API"}`))
	})

	// User routes (public)
	r.mux.HandleFunc("POST /api/users", userHandler.Register)
	r.mux.HandleFunc("POST /api/users/login", userHandler.Login)

	// User routes (authenticated)
	authMw := middleware.Auth(authService)
	optionalAuthMw := middleware.OptionalAuth(authService)
	r.mux.Handle("GET /api/user", authMw(http.HandlerFunc(userHandler.GetCurrentUser)))
	r.mux.Handle("PUT /api/user", authMw(http.HandlerFunc(userHandler.UpdateUser)))

	// Profile routes (public - with optional auth for following status)
	r.mux.Handle("GET /api/profiles/{username}", optionalAuthMw(http.HandlerFunc(profileHandler.GetProfile)))

	// Profile routes (authenticated)
	r.mux.Handle("POST /api/profiles/{username}/follow", authMw(http.HandlerFunc(profileHandler.FollowUser)))
	r.mux.Handle("DELETE /api/profiles/{username}/follow", authMw(http.HandlerFunc(profileHandler.UnfollowUser)))

	// Article routes (public - with optional auth for favorited status)
	r.mux.Handle("GET /api/articles", optionalAuthMw(http.HandlerFunc(articleHandler.ListArticles)))
	r.mux.Handle("GET /api/articles/{slug}", optionalAuthMw(http.HandlerFunc(articleHandler.GetArticle)))

	// Article routes (authenticated)
	r.mux.Handle("POST /api/articles", authMw(http.HandlerFunc(articleHandler.CreateArticle)))
	r.mux.Handle("PUT /api/articles/{slug}", authMw(http.HandlerFunc(articleHandler.UpdateArticle)))
	r.mux.Handle("DELETE /api/articles/{slug}", authMw(http.HandlerFunc(articleHandler.DeleteArticle)))
	r.mux.Handle("GET /api/articles/feed", authMw(http.HandlerFunc(articleHandler.GetFeed)))

	// Favorite routes (authenticated)
	r.mux.Handle("POST /api/articles/{slug}/favorite", authMw(http.HandlerFunc(articleHandler.FavoriteArticle)))
	r.mux.Handle("DELETE /api/articles/{slug}/favorite", authMw(http.HandlerFunc(articleHandler.UnfavoriteArticle)))

	// Tags route (public)
	r.mux.HandleFunc("GET /api/tags", articleHandler.GetTags)

	// Comment routes (public - with optional auth)
	r.mux.Handle("GET /api/articles/{slug}/comments", optionalAuthMw(http.HandlerFunc(commentHandler.GetComments)))

	// Comment routes (authenticated)
	r.mux.Handle("POST /api/articles/{slug}/comments", authMw(http.HandlerFunc(commentHandler.CreateComment)))
	r.mux.Handle("DELETE /api/articles/{slug}/comments/{id}", authMw(http.HandlerFunc(commentHandler.DeleteComment)))

	// Apply middleware chain
	var h http.Handler = r.mux
	h = middleware.Logging(r.logger)(h)

	// Configure CORS with origins from config
	corsConfig := middleware.CORSConfig{
		AllowedOrigins:   r.config.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
	}
	h = middleware.CORS(corsConfig)(h)
	h = middleware.Security()(h)
	h = middleware.Recover(r.logger)(h)

	return h
}

func (r *Router) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
