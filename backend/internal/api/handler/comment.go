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

// CommentHandler handles comment-related HTTP requests
type CommentHandler struct {
	commentService *service.CommentService
	logger         *slog.Logger
}

// NewCommentHandler creates a new CommentHandler instance
func NewCommentHandler(commentService *service.CommentService, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		logger:         logger,
	}
}

// CreateCommentRequest represents the create comment request body
type CreateCommentRequest struct {
	Comment struct {
		Body string `json:"body"`
	} `json:"comment"`
}

// CommentResponse represents a single comment response
type CommentResponse struct {
	Comment CommentResponseBody `json:"comment"`
}

// CommentsResponse represents a list of comments response
type CommentsResponse struct {
	Comments []CommentResponseBody `json:"comments"`
}

// CommentResponseBody represents the comment data in responses
type CommentResponseBody struct {
	ID        int64               `json:"id"`
	Body      string              `json:"body"`
	CreatedAt string              `json:"createdAt"`
	UpdatedAt string              `json:"updatedAt"`
	Author    ProfileResponseBody `json:"author"`
}

// GetComments handles GET /api/articles/{slug}/comments
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	slug := h.extractSlugFromPath(r.URL.Path)
	if slug == "" {
		h.writeError(w, http.StatusNotFound, "article", "article not found")
		return
	}

	comments, err := h.commentService.GetCommentsByArticleSlug(r.Context(), slug)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeCommentsResponse(w, http.StatusOK, comments)
}

// CreateComment handles POST /api/articles/{slug}/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	slug := h.extractSlugFromPath(r.URL.Path)
	if slug == "" {
		h.writeError(w, http.StatusNotFound, "article", "article not found")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode create comment request", "error", err)
		h.writeError(w, http.StatusUnprocessableEntity, "body", "invalid request body")
		return
	}

	input := &domain.CreateCommentInput{
		Body: req.Comment.Body,
	}

	comment, err := h.commentService.CreateComment(r.Context(), slug, userID, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeCommentResponse(w, http.StatusCreated, comment)
}

// DeleteComment handles DELETE /api/articles/{slug}/comments/{id}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	slug, commentID := h.extractSlugAndCommentID(r.URL.Path)
	if slug == "" || commentID == 0 {
		h.writeError(w, http.StatusNotFound, "comment", "comment not found")
		return
	}

	err := h.commentService.DeleteComment(r.Context(), slug, commentID, userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// extractSlugFromPath extracts the article slug from paths like /api/articles/{slug}/comments
func (h *CommentHandler) extractSlugFromPath(path string) string {
	// Path format: /api/articles/{slug}/comments
	const prefix = "/api/articles/"
	const suffix = "/comments"

	path = strings.TrimPrefix(path, prefix)

	// Find the first slash after the slug
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[:idx]
	}
	return ""
}

// extractSlugAndCommentID extracts slug and comment ID from paths like /api/articles/{slug}/comments/{id}
func (h *CommentHandler) extractSlugAndCommentID(path string) (string, int64) {
	// Path format: /api/articles/{slug}/comments/{id}
	const prefix = "/api/articles/"

	path = strings.TrimPrefix(path, prefix)
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[1] != "comments" {
		return "", 0
	}

	slug := parts[0]
	commentIDStr := parts[2]

	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		return "", 0
	}

	return slug, commentID
}

// writeCommentResponse writes a single comment response
func (h *CommentHandler) writeCommentResponse(w http.ResponseWriter, status int, comment *domain.Comment) {
	resp := CommentResponse{
		Comment: h.toCommentResponseBody(comment),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// writeCommentsResponse writes a list of comments response
func (h *CommentHandler) writeCommentsResponse(w http.ResponseWriter, status int, comments []*domain.Comment) {
	commentBodies := make([]CommentResponseBody, 0, len(comments))
	for _, comment := range comments {
		commentBodies = append(commentBodies, h.toCommentResponseBody(comment))
	}

	resp := CommentsResponse{
		Comments: commentBodies,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// toCommentResponseBody converts a domain comment to response body
func (h *CommentHandler) toCommentResponseBody(comment *domain.Comment) CommentResponseBody {
	body := CommentResponseBody{
		ID:        comment.ID,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt: comment.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
	}

	// Add author profile if available
	if comment.Author != nil {
		body.Author = ProfileResponseBody{
			Username:  comment.Author.Username,
			Bio:       comment.Author.Bio,
			Image:     comment.Author.Image,
			Following: false, // TODO: Implement following status
		}
	}

	return body
}

// writeError writes an error response
func (h *CommentHandler) writeError(w http.ResponseWriter, status int, field string, message string) {
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
func (h *CommentHandler) handleServiceError(w http.ResponseWriter, err error) {
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
		} else if err == domain.ErrCommentNotFound {
			h.writeError(w, http.StatusNotFound, "comment", "comment not found")
		} else if err == domain.ErrForbidden {
			h.writeError(w, http.StatusForbidden, "comment", "you are not authorized to perform this action")
		} else if err == domain.ErrUnauthorized {
			h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		} else {
			h.logger.Error("unexpected error", "error", err)
			h.writeError(w, http.StatusInternalServerError, "server", "internal server error")
		}
	}
}
