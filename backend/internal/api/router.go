package api

import (
	"log/slog"
	"net/http"

	"github.com/alexlee0213/realworld-conduit/backend/internal/api/handler"
	"github.com/alexlee0213/realworld-conduit/backend/internal/api/middleware"
)

type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewRouter(logger *slog.Logger) *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
	}
}

func (r *Router) Setup() http.Handler {
	// Health check
	healthHandler := handler.NewHealthHandler()
	r.mux.HandleFunc("GET /health", healthHandler.Health)

	// API routes placeholder
	r.mux.HandleFunc("GET /api/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "RealWorld Conduit API"}`))
	})

	// Apply middleware chain
	var h http.Handler = r.mux
	h = middleware.Logging(r.logger)(h)
	h = middleware.CORS(middleware.DefaultCORSConfig())(h)
	h = middleware.Recover(r.logger)(h)

	return h
}
