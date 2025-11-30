package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// ProfileHandler handles profile-related HTTP requests
type ProfileHandler struct {
	profileService *service.ProfileService
	logger         *slog.Logger
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(profileService *service.ProfileService, logger *slog.Logger) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		logger:         logger,
	}
}

// ProfileResponse represents the profile response body
type ProfileResponse struct {
	Profile ProfileResponseBody `json:"profile"`
}

// Note: ProfileResponseBody is defined in article.go and reused here

// GetProfile handles GET /api/profiles/:username
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		h.writeError(w, http.StatusBadRequest, "username", "username is required")
		return
	}

	// Get current user ID (optional)
	var currentUserID *int64
	if userID, ok := GetUserIDFromContext(r.Context()); ok {
		currentUserID = &userID
	}

	profile, err := h.profileService.GetProfileByUsername(r.Context(), username, currentUserID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeProfileResponse(w, http.StatusOK, profile)
}

// FollowUser handles POST /api/profiles/:username/follow
func (h *ProfileHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		h.writeError(w, http.StatusBadRequest, "username", "username is required")
		return
	}

	// Get current user ID (required)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	profile, err := h.profileService.FollowUser(r.Context(), userID, username)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeProfileResponse(w, http.StatusOK, profile)
}

// UnfollowUser handles DELETE /api/profiles/:username/follow
func (h *ProfileHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		h.writeError(w, http.StatusBadRequest, "username", "username is required")
		return
	}

	// Get current user ID (required)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "token", "authorization required")
		return
	}

	profile, err := h.profileService.UnfollowUser(r.Context(), userID, username)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeProfileResponse(w, http.StatusOK, profile)
}

// writeProfileResponse writes a profile response
func (h *ProfileHandler) writeProfileResponse(w http.ResponseWriter, status int, profile *domain.Profile) {
	resp := ProfileResponse{
		Profile: ProfileResponseBody{
			Username:  profile.Username,
			Bio:       profile.Bio,
			Image:     profile.Image,
			Following: profile.Following,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response
func (h *ProfileHandler) writeError(w http.ResponseWriter, status int, field string, message string) {
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
func (h *ProfileHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *domain.ValidationErrors:
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
			h.writeError(w, http.StatusNotFound, "profile", "profile not found")
		} else if err == domain.ErrValidation {
			h.writeError(w, http.StatusUnprocessableEntity, "profile", "cannot follow yourself")
		} else {
			h.logger.Error("unexpected error", "error", err)
			h.writeError(w, http.StatusInternalServerError, "server", "internal server error")
		}
	}
}
