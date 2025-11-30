package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Bio          string    `json:"bio"`
	Image        string    `json:"image"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserResponse represents the user data returned to clients (RealWorld API format)
type UserResponse struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// ProfileResponse represents a public user profile (RealWorld API format)
type ProfileResponse struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

// ToUserResponse converts a User to UserResponse with the given token
func (u *User) ToUserResponse(token string) *UserResponse {
	return &UserResponse{
		Email:    u.Email,
		Token:    token,
		Username: u.Username,
		Bio:      u.Bio,
		Image:    u.Image,
	}
}

// ToProfileResponse converts a User to ProfileResponse
func (u *User) ToProfileResponse(following bool) *ProfileResponse {
	return &ProfileResponse{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: following,
	}
}

// CreateUserInput represents the input for creating a new user
type CreateUserInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Bio      *string `json:"bio,omitempty"`
	Image    *string `json:"image,omitempty"`
}
