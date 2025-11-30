package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alexlee0213/realworld-conduit/backend/internal/api/handler"
	"github.com/alexlee0213/realworld-conduit/backend/internal/api/middleware"
	"github.com/alexlee0213/realworld-conduit/backend/internal/config"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
	config *config.Config
	db     *sql.DB
}

func NewRouter(cfg *config.Config, logger *slog.Logger) (*Router, error) {
	// Initialize database
	db, err := initDatabase(cfg.Database.URL)
	if err != nil {
		return nil, err
	}

	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
		db:     db,
	}, nil
}

func initDatabase(databaseURL string) (*sql.DB, error) {
	// Parse database URL (sqlite://./data/conduit.db)
	dbPath := strings.TrimPrefix(databaseURL, "sqlite://")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (r *Router) Setup() http.Handler {
	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(r.db, r.logger)
	articleRepo := repository.NewSQLiteArticleRepository(r.db, r.logger)
	commentRepo := repository.NewSQLiteCommentRepository(r.db, r.logger)

	// Initialize services
	authService := service.NewAuthService(
		userRepo,
		r.config.JWT.Secret,
		r.config.JWT.Expiry,
		r.logger,
	)
	articleService := service.NewArticleService(articleRepo, userRepo, r.logger)
	commentService := service.NewCommentService(commentRepo, articleRepo, userRepo, r.logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(authService, r.logger)
	articleHandler := handler.NewArticleHandler(articleService, r.logger)
	commentHandler := handler.NewCommentHandler(commentService, r.logger)

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

	// Article routes (public - with optional auth for favorited status)
	r.mux.Handle("GET /api/articles", optionalAuthMw(http.HandlerFunc(articleHandler.ListArticles)))
	r.mux.Handle("GET /api/articles/{slug}", optionalAuthMw(http.HandlerFunc(articleHandler.GetArticle)))

	// Article routes (authenticated)
	r.mux.Handle("POST /api/articles", authMw(http.HandlerFunc(articleHandler.CreateArticle)))
	r.mux.Handle("PUT /api/articles/{slug}", authMw(http.HandlerFunc(articleHandler.UpdateArticle)))
	r.mux.Handle("DELETE /api/articles/{slug}", authMw(http.HandlerFunc(articleHandler.DeleteArticle)))
	r.mux.Handle("GET /api/articles/feed", authMw(http.HandlerFunc(articleHandler.GetFeed)))

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
	h = middleware.CORS(middleware.DefaultCORSConfig())(h)
	h = middleware.Recover(r.logger)(h)

	return h
}

func (r *Router) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
