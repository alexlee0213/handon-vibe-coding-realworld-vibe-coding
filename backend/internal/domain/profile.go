package domain

import "time"

// Profile represents a user profile with follow status
// Note: ProfileResponse in user.go is the API response format
// This Profile struct is the domain model for profile-related operations
type Profile struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

// Follow represents a follow relationship between two users
type Follow struct {
	FollowerID  int64     `json:"follower_id"`
	FollowingID int64     `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewProfileFromUser creates a Profile from a User with the given following status
func NewProfileFromUser(user *User, following bool) *Profile {
	return &Profile{
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		Following: following,
	}
}
