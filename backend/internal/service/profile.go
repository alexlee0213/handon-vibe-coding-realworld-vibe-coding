package service

import (
	"context"
	"log/slog"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
)

// ProfileService handles profile business logic
type ProfileService struct {
	userRepo   repository.UserRepository
	followRepo repository.FollowRepository
	logger     *slog.Logger
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(
	userRepo repository.UserRepository,
	followRepo repository.FollowRepository,
	logger *slog.Logger,
) *ProfileService {
	return &ProfileService{
		userRepo:   userRepo,
		followRepo: followRepo,
		logger:     logger,
	}
}

// GetProfileByUsername retrieves a user's profile by username
// currentUserID is optional - if provided, the following status will be included
func (s *ProfileService) GetProfileByUsername(ctx context.Context, username string, currentUserID *int64) (*domain.Profile, error) {
	// Get the user
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Check if current user is following this user
	following := false
	if currentUserID != nil && *currentUserID != 0 {
		following, err = s.followRepo.IsFollowing(ctx, *currentUserID, user.ID)
		if err != nil {
			s.logger.Error("failed to check follow status",
				"error", err,
				"follower_id", *currentUserID,
				"following_id", user.ID,
			)
			// Don't fail the request, just log the error
			following = false
		}
	}

	return domain.NewProfileFromUser(user, following), nil
}

// FollowUser makes the current user follow the target user
func (s *ProfileService) FollowUser(ctx context.Context, followerID int64, username string) (*domain.Profile, error) {
	// Get the target user
	targetUser, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Prevent self-follow
	if followerID == targetUser.ID {
		s.logger.Warn("attempted self-follow",
			"user_id", followerID,
		)
		return nil, domain.ErrValidation
	}

	// Create follow relationship
	if err := s.followRepo.FollowUser(ctx, followerID, targetUser.ID); err != nil {
		return nil, err
	}

	s.logger.Info("user followed",
		"follower_id", followerID,
		"following_username", username,
		"following_id", targetUser.ID,
	)

	// Return profile with following=true
	return domain.NewProfileFromUser(targetUser, true), nil
}

// UnfollowUser makes the current user unfollow the target user
func (s *ProfileService) UnfollowUser(ctx context.Context, followerID int64, username string) (*domain.Profile, error) {
	// Get the target user
	targetUser, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Remove follow relationship
	if err := s.followRepo.UnfollowUser(ctx, followerID, targetUser.ID); err != nil {
		return nil, err
	}

	s.logger.Info("user unfollowed",
		"follower_id", followerID,
		"following_username", username,
		"following_id", targetUser.ID,
	)

	// Return profile with following=false
	return domain.NewProfileFromUser(targetUser, false), nil
}
