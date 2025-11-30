package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// ArticleHandler handles article-related HTTP requests
type ArticleHandler struct {
	articleService *service.ArticleService
	logger         *slog.Logger
}

// NewArticleHandler creates a new ArticleHandler instance
func NewArticleHandler(articleService *service.ArticleService, logger *slog.Logger) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
		logger:         logger,
	}
}

// CreateArticleRequest represents the create article request body
type CreateArticleRequest struct {
	Article struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Body        string   `json:"body"`
		TagList     []string `json:"tagList,omitempty"`
	} `json:"article"`
}

// UpdateArticleRequest represents the update article request body
type UpdateArticleRequest struct {
	Article struct {
		Title       *string `json:"title,omitempty"`
		Description *string `json:"description,omitempty"`
		Body        *string `json:"body,omitempty"`
	} `json:"article"`
}

// ArticleResponse represents a single article response
type ArticleResponse struct {
	Article ArticleResponseBody `json:"article"`
}

// ArticlesResponse represents a list of articles response
type ArticlesResponse struct {
	Articles      []ArticleResponseBody `json:"articles"`
	ArticlesCount int                   `json:"articlesCount"`
}

// ArticleResponseBody represents the article data in responses
type ArticleResponseBody struct {
	Slug           string              `json:"slug"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	Body           string              `json:"body"`
	TagList        []string            `json:"tagList"`
	CreatedAt      string              `json:"createdAt"`
	UpdatedAt      string              `json:"updatedAt"`
	Favorited      bool                `json:"favorited"`
	FavoritesCount int                 `json:"favoritesCount"`
	Author         ProfileResponseBody `json:"author"`
}

// ProfileResponseBody represents the author profile in article responses
type ProfileResponseBody struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

// TagsResponse represents the tags list response
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// CreateArticle handles POST /api/articles
func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	var req CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode create article request", "error", err)
		h.writeError(w, http.StatusUnprocessableEntity, "body", "invalid request body")
		return
	}

	input := &domain.CreateArticleInput{
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
		TagList:     req.Article.TagList,
	}

	article, err := h.articleService.CreateArticle(r.Context(), userID, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeArticleResponse(w, http.StatusCreated, article)
}

// GetArticle handles GET /api/articles/{slug}
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := h.extractSlugFromPath(r.URL.Path, "/api/articles/")
	if slug == "" {
		h.writeError(w, http.StatusNotFound, "article", "article not found")
		return
	}

	// Get optional current user ID for favorited status
	var currentUserID *int64
	if userID, ok := r.Context().Value(UserIDContextKey).(int64); ok {
		currentUserID = &userID
	}

	article, err := h.articleService.GetArticleBySlug(r.Context(), slug, currentUserID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeArticleResponse(w, http.StatusOK, article)
}

// UpdateArticle handles PUT /api/articles/{slug}
func (h *ArticleHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	slug := h.extractSlugFromPath(r.URL.Path, "/api/articles/")
	if slug == "" {
		h.writeError(w, http.StatusNotFound, "article", "article not found")
		return
	}

	var req UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode update article request", "error", err)
		h.writeError(w, http.StatusUnprocessableEntity, "body", "invalid request body")
		return
	}

	input := &domain.UpdateArticleInput{
		Title:       req.Article.Title,
		Description: req.Article.Description,
		Body:        req.Article.Body,
	}

	article, err := h.articleService.UpdateArticle(r.Context(), slug, userID, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeArticleResponse(w, http.StatusOK, article)
}

// DeleteArticle handles DELETE /api/articles/{slug}
func (h *ArticleHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	slug := h.extractSlugFromPath(r.URL.Path, "/api/articles/")
	if slug == "" {
		h.writeError(w, http.StatusNotFound, "article", "article not found")
		return
	}

	err := h.articleService.DeleteArticle(r.Context(), slug, userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListArticles handles GET /api/articles
func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	// Get optional current user ID for favorited status
	var currentUserID *int64
	if userID, ok := r.Context().Value(UserIDContextKey).(int64); ok {
		currentUserID = &userID
	}

	// Parse query parameters
	params := &domain.ArticleListParams{
		Tag:       r.URL.Query().Get("tag"),
		Author:    r.URL.Query().Get("author"),
		Favorited: r.URL.Query().Get("favorited"),
		Limit:     h.parseIntParam(r.URL.Query().Get("limit"), 20),
		Offset:    h.parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	articles, total, err := h.articleService.ListArticles(r.Context(), params, currentUserID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeArticlesResponse(w, http.StatusOK, articles, total)
}

// GetFeed handles GET /api/articles/feed
func (h *ArticleHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	// Parse query parameters
	params := &domain.ArticleFeedParams{
		Limit:  h.parseIntParam(r.URL.Query().Get("limit"), 20),
		Offset: h.parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	articles, total, err := h.articleService.GetFeed(r.Context(), userID, params)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeArticlesResponse(w, http.StatusOK, articles, total)
}

// GetTags handles GET /api/tags
func (h *ArticleHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.articleService.GetAllTags(r.Context())
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	resp := TagsResponse{Tags: tags}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// extractSlugFromPath extracts the slug from the URL path
func (h *ArticleHandler) extractSlugFromPath(path, prefix string) string {
	slug := strings.TrimPrefix(path, prefix)
	// Remove any trailing slashes or additional path segments
	if idx := strings.Index(slug, "/"); idx != -1 {
		slug = slug[:idx]
	}
	return strings.TrimSpace(slug)
}

// parseIntParam parses an integer query parameter with a default value
func (h *ArticleHandler) parseIntParam(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// writeArticleResponse writes a single article response
func (h *ArticleHandler) writeArticleResponse(w http.ResponseWriter, status int, article *domain.Article) {
	resp := ArticleResponse{
		Article: h.toArticleResponseBody(article),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// writeArticlesResponse writes a list of articles response
func (h *ArticleHandler) writeArticlesResponse(w http.ResponseWriter, status int, articles []*domain.Article, total int) {
	articleBodies := make([]ArticleResponseBody, 0, len(articles))
	for _, article := range articles {
		articleBodies = append(articleBodies, h.toArticleResponseBody(article))
	}

	resp := ArticlesResponse{
		Articles:      articleBodies,
		ArticlesCount: total,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// toArticleResponseBody converts a domain article to response body
func (h *ArticleHandler) toArticleResponseBody(article *domain.Article) ArticleResponseBody {
	tagList := article.TagList
	if tagList == nil {
		tagList = []string{}
	}

	body := ArticleResponseBody{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        tagList,
		CreatedAt:      article.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt:      article.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		Favorited:      article.Favorited,
		FavoritesCount: article.FavoritesCount,
	}

	// Add author profile if available
	if article.Author != nil {
		body.Author = ProfileResponseBody{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: false, // TODO: Implement following status
		}
	}

	return body
}

// writeError writes an error response
func (h *ArticleHandler) writeError(w http.ResponseWriter, status int, field string, message string) {
	resp := ErrorResponse{
		Errors: map[string][]string{
			field: {message},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// handleServiceError handles service layer errors and writes appropriate HTTP responses
func (h *ArticleHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *domain.ValidationErrors:
		// Convert ValidationErrors to RealWorld API format
		errorsMap := make(map[string][]string)
		for _, ve := range e.Errors {
			errorsMap[ve.Field] = append(errorsMap[ve.Field], ve.Message)
		}
		resp := ErrorResponse{
			Errors: errorsMap,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(resp)
	default:
		if err == domain.ErrArticleNotFound {
			h.writeError(w, http.StatusNotFound, "article", "article not found")
		} else if err == domain.ErrArticleAlreadyExists {
			h.writeError(w, http.StatusUnprocessableEntity, "slug", "has already been taken")
		} else if err == domain.ErrForbidden {
			h.writeError(w, http.StatusForbidden, "article", "you are not authorized to perform this action")
		} else if err == domain.ErrUnauthorized {
			h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		} else {
			h.logger.Error("unexpected error", "error", err)
			h.writeError(w, http.StatusInternalServerError, "server", "internal server error")
		}
	}
}
