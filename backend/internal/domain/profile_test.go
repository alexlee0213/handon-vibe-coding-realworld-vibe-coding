package domain

import (
	"testing"
	"time"
)

func TestNewProfileFromUser(t *testing.T) {
	user := &User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Bio:          "Test bio",
		Image:        "https://example.com/image.png",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Run("creates profile with following=true", func(t *testing.T) {
		profile := NewProfileFromUser(user, true)

		if profile.Username != user.Username {
			t.Errorf("expected username %s, got %s", user.Username, profile.Username)
		}
		if profile.Bio != user.Bio {
			t.Errorf("expected bio %s, got %s", user.Bio, profile.Bio)
		}
		if profile.Image != user.Image {
			t.Errorf("expected image %s, got %s", user.Image, profile.Image)
		}
		if !profile.Following {
			t.Error("expected following to be true")
		}
	})

	t.Run("creates profile with following=false", func(t *testing.T) {
		profile := NewProfileFromUser(user, false)

		if profile.Username != user.Username {
			t.Errorf("expected username %s, got %s", user.Username, profile.Username)
		}
		if profile.Following {
			t.Error("expected following to be false")
		}
	})

	t.Run("handles empty bio and image", func(t *testing.T) {
		userWithEmptyFields := &User{
			ID:       2,
			Email:    "empty@example.com",
			Username: "emptyuser",
			Bio:      "",
			Image:    "",
		}

		profile := NewProfileFromUser(userWithEmptyFields, false)

		if profile.Bio != "" {
			t.Errorf("expected empty bio, got %s", profile.Bio)
		}
		if profile.Image != "" {
			t.Errorf("expected empty image, got %s", profile.Image)
		}
	})
}

func TestFollowStruct(t *testing.T) {
	t.Run("creates follow relationship", func(t *testing.T) {
		now := time.Now()
		follow := Follow{
			FollowerID:  1,
			FollowingID: 2,
			CreatedAt:   now,
		}

		if follow.FollowerID != 1 {
			t.Errorf("expected follower_id 1, got %d", follow.FollowerID)
		}
		if follow.FollowingID != 2 {
			t.Errorf("expected following_id 2, got %d", follow.FollowingID)
		}
		if !follow.CreatedAt.Equal(now) {
			t.Errorf("expected created_at %v, got %v", now, follow.CreatedAt)
		}
	})
}
