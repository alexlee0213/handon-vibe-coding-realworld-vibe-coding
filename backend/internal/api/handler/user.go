package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// UserIDContextKey is the context key for the authenticated user ID
type contextKey string

const UserIDContextKey contextKey = "userID"

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(authService *service.AuthService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	User struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

// UpdateUserRequest represents the update user request body
type UpdateUserRequest struct {
	User struct {
		Email    *string `json:"email,omitempty"`
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
		Bio      *string `json:"bio,omitempty"`
		Image    *string `json:"image,omitempty"`
	} `json:"user"`
}

// UserResponse represents the user response body
type UserResponse struct {
	User UserResponseBody `json:"user"`
}

// UserResponseBody represents the user data in responses
type UserResponseBody struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// ErrorResponse represents an error response body
type ErrorResponse struct {
	Errors map[string][]string `json:"errors"`
}

// Register handles POST /api/users
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode register request", "error", err)
		h.writeError(w, http.StatusUnprocessableEntity, "body", "invalid request body")
		return
	}

	input := &domain.CreateUserInput{
		Username: req.User.Username,
		Email:    req.User.Email,
		Password: req.User.Password,
	}

	user, token, err := h.authService.Register(r.Context(), input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeUserResponse(w, http.StatusCreated, user, token)
}

// Login handles POST /api/users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode login request", "error", err)
		h.writeError(w, http.StatusUnprocessableEntity, "body", "invalid request body")
		return
	}

	user, token, err := h.authService.Login(r.Context(), req.User.Email, req.User.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeUserResponse(w, http.StatusOK, user, token)
}

// GetCurrentUser handles GET /api/user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Generate a fresh token for the response
	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		h.writeError(w, http.StatusInternalServerError, "server", "internal server error")
		return
	}

	h.writeUserResponse(w, http.StatusOK, user, token)
}

// UpdateUser handles PUT /api/user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDContextKey).(int64)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("failed to decode update user request", "error", err)
		h.writeError(w, http.StatusUnprocessableEntity, "body", "invalid request body")
		return
	}

	input := &domain.UpdateUserInput{
		Email:    req.User.Email,
		Username: req.User.Username,
		Password: req.User.Password,
		Bio:      req.User.Bio,
		Image:    req.User.Image,
	}

	user, err := h.authService.UpdateUser(r.Context(), userID, input)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Generate a fresh token for the response
	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		h.logger.Error("failed to generate token", "error", err)
		h.writeError(w, http.StatusInternalServerError, "server", "internal server error")
		return
	}

	h.writeUserResponse(w, http.StatusOK, user, token)
}

// GetUserIDFromContext retrieves the user ID from context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int64)
	return userID, ok
}

// writeUserResponse writes a user response
func (h *UserHandler) writeUserResponse(w http.ResponseWriter, status int, user *domain.User, token string) {
	resp := UserResponse{
		User: UserResponseBody{
			Email:    user.Email,
			Token:    token,
			Username: user.Username,
			Bio:      user.Bio,
			Image:    user.Image,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response
func (h *UserHandler) writeError(w http.ResponseWriter, status int, field string, message string) {
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
func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
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
		if err == domain.ErrUserNotFound {
			h.writeError(w, http.StatusNotFound, "user", "user not found")
		} else if err == domain.ErrEmailAlreadyTaken {
			h.writeError(w, http.StatusUnprocessableEntity, "email", "has already been taken")
		} else if err == domain.ErrUsernameAlreadyTaken {
			h.writeError(w, http.StatusUnprocessableEntity, "username", "has already been taken")
		} else if err == domain.ErrInvalidCredentials {
			h.writeError(w, http.StatusUnprocessableEntity, "email or password", "is invalid")
		} else {
			h.logger.Error("unexpected error", "error", err)
			h.writeError(w, http.StatusInternalServerError, "server", "internal server error")
		}
	}
}
