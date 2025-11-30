package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/service"
)

// Test setup for profile handler tests
type profileTestSetup struct {
	handler        *ProfileHandler
	profileService *service.ProfileService
	authService    *service.AuthService
	db             *sql.DB
}

func setupProfileTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			bio TEXT DEFAULT '',
			image TEXT DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_username ON users(username);
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Create follows table
	_, err = db.Exec(`
		CREATE TABLE follows (
			follower_id INTEGER NOT NULL,
			following_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id),
			FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
			CHECK (follower_id != following_id)
		);
		CREATE INDEX idx_follows_follower_id ON follows(follower_id);
		CREATE INDEX idx_follows_following_id ON follows(following_id);
	`)
	if err != nil {
		t.Fatalf("failed to create follows table: %v", err)
	}

	return db
}

func newTestProfileHandler(t *testing.T) *profileTestSetup {
	t.Helper()
	db := setupProfileTestDB(t)
	logger := newTestLogger()
	userRepo := repository.NewSQLiteUserRepository(db, logger)
	followRepo := repository.NewSQLiteFollowRepository(db, logger)
	authService := service.NewAuthService(userRepo, "test-jwt-secret", 24*time.Hour, logger)
	profileService := service.NewProfileService(userRepo, followRepo, logger)
	profileHandler := NewProfileHandler(profileService, logger)

	return &profileTestSetup{
		handler:        profileHandler,
		profileService: profileService,
		authService:    authService,
		db:             db,
	}
}

// =============================================================================
// TDD: GET /api/profiles/:username Tests
// =============================================================================

func TestGetProfileHandler(t *testing.T) {
	t.Run("successfully gets profile without authentication", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		// Create a user
		ctx := context.Background()
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "profile@example.com",
			Username: "profileuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/profiles/profileuser", nil)
		req.SetPathValue("username", "profileuser")
		w := httptest.NewRecorder()

		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		profile, ok := response["profile"].(map[string]interface{})
		if !ok {
			t.Fatal("expected profile object in response")
		}

		if profile["username"] != "profileuser" {
			t.Errorf("expected username profileuser, got %v", profile["username"])
		}
		if profile["following"] != false {
			t.Errorf("expected following false (no auth), got %v", profile["following"])
		}
	})

	t.Run("successfully gets profile with authentication - not following", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create target user
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "target@example.com",
			Username: "targetuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register target user: %v", err)
		}

		// Create current user
		currentUser, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "current@example.com",
			Username: "currentuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register current user: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/profiles/targetuser", nil)
		req.SetPathValue("username", "targetuser")
		ctx = context.WithValue(req.Context(), UserIDContextKey, currentUser.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		profile, ok := response["profile"].(map[string]interface{})
		if !ok {
			t.Fatal("expected profile object in response")
		}

		if profile["username"] != "targetuser" {
			t.Errorf("expected username targetuser, got %v", profile["username"])
		}
		if profile["following"] != false {
			t.Errorf("expected following false, got %v", profile["following"])
		}
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/profiles/nonexistent", nil)
		req.SetPathValue("username", "nonexistent")
		w := httptest.NewRecorder()

		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for empty username", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/profiles/", nil)
		req.SetPathValue("username", "")
		w := httptest.NewRecorder()

		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

// =============================================================================
// TDD: POST /api/profiles/:username/follow Tests
// =============================================================================

func TestFollowUserHandler(t *testing.T) {
	t.Run("successfully follows a user", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create target user
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "follow-target@example.com",
			Username: "followtarget",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register target user: %v", err)
		}

		// Create follower user
		follower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "follower@example.com",
			Username: "follower",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register follower user: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/profiles/followtarget/follow", nil)
		req.SetPathValue("username", "followtarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		profile, ok := response["profile"].(map[string]interface{})
		if !ok {
			t.Fatal("expected profile object in response")
		}

		if profile["username"] != "followtarget" {
			t.Errorf("expected username followtarget, got %v", profile["username"])
		}
		if profile["following"] != true {
			t.Errorf("expected following true, got %v", profile["following"])
		}
	})

	t.Run("returns 401 without authentication", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/profiles/someuser/follow", nil)
		req.SetPathValue("username", "someuser")
		w := httptest.NewRecorder()

		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns error when trying to follow self", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create user who will try to follow themselves
		selfUser, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "self@example.com",
			Username: "selfuser",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/profiles/selfuser/follow", nil)
		req.SetPathValue("username", "selfuser")
		ctx = context.WithValue(req.Context(), UserIDContextKey, selfUser.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d: %s", http.StatusUnprocessableEntity, w.Code, w.Body.String())
		}
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create follower user
		follower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "follower2@example.com",
			Username: "follower2",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register follower user: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/profiles/nonexistent/follow", nil)
		req.SetPathValue("username", "nonexistent")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("following same user again is idempotent", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create target user
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "target-idem@example.com",
			Username: "targetidem",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register target user: %v", err)
		}

		// Create follower user
		follower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "follower-idem@example.com",
			Username: "followeridem",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register follower user: %v", err)
		}

		// Follow once
		req := httptest.NewRequest(http.MethodPost, "/api/profiles/targetidem/follow", nil)
		req.SetPathValue("username", "targetidem")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("first follow: expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Follow again (should be idempotent)
		req = httptest.NewRequest(http.MethodPost, "/api/profiles/targetidem/follow", nil)
		req.SetPathValue("username", "targetidem")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("second follow: expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

// =============================================================================
// TDD: DELETE /api/profiles/:username/follow Tests
// =============================================================================

func TestUnfollowUserHandler(t *testing.T) {
	t.Run("successfully unfollows a user", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create target user
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "unfollow-target@example.com",
			Username: "unfollowtarget",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register target user: %v", err)
		}

		// Create follower user
		follower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "unfollower@example.com",
			Username: "unfollower",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register follower user: %v", err)
		}

		// First, follow the user
		_, err = setup.profileService.FollowUser(ctx, follower.ID, "unfollowtarget")
		if err != nil {
			t.Fatalf("failed to follow user: %v", err)
		}

		// Then, unfollow
		req := httptest.NewRequest(http.MethodDelete, "/api/profiles/unfollowtarget/follow", nil)
		req.SetPathValue("username", "unfollowtarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UnfollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		profile, ok := response["profile"].(map[string]interface{})
		if !ok {
			t.Fatal("expected profile object in response")
		}

		if profile["username"] != "unfollowtarget" {
			t.Errorf("expected username unfollowtarget, got %v", profile["username"])
		}
		if profile["following"] != false {
			t.Errorf("expected following false after unfollow, got %v", profile["following"])
		}
	})

	t.Run("returns 401 without authentication", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		req := httptest.NewRequest(http.MethodDelete, "/api/profiles/someuser/follow", nil)
		req.SetPathValue("username", "someuser")
		w := httptest.NewRecorder()

		setup.handler.UnfollowUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create unfollower user
		unfollower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "unfollower2@example.com",
			Username: "unfollower2",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register unfollower user: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/profiles/nonexistent/follow", nil)
		req.SetPathValue("username", "nonexistent")
		ctx = context.WithValue(req.Context(), UserIDContextKey, unfollower.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UnfollowUser(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("unfollowing when not following is idempotent", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create target user
		_, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "target-unfoll-idem@example.com",
			Username: "targetunfollidem",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register target user: %v", err)
		}

		// Create unfollower user (not following)
		unfollower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "unfollower-idem@example.com",
			Username: "unfolloweridem",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register unfollower user: %v", err)
		}

		// Unfollow without ever following (should be idempotent)
		req := httptest.NewRequest(http.MethodDelete, "/api/profiles/targetunfollidem/follow", nil)
		req.SetPathValue("username", "targetunfollidem")
		ctx = context.WithValue(req.Context(), UserIDContextKey, unfollower.ID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		setup.handler.UnfollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})
}

// =============================================================================
// Integration: Follow/Unfollow Flow Tests
// =============================================================================

func TestFollowUnfollowFlow(t *testing.T) {
	t.Run("complete follow and unfollow flow", func(t *testing.T) {
		setup := newTestProfileHandler(t)
		defer setup.db.Close()

		ctx := context.Background()

		// Create target user with bio and image
		targetInput := &domain.CreateUserInput{
			Email:    "complete-target@example.com",
			Username: "completetarget",
			Password: "password123",
		}
		_, _, err := setup.authService.Register(ctx, targetInput)
		if err != nil {
			t.Fatalf("failed to register target user: %v", err)
		}

		// Update target user's bio (simulate having profile info)
		_, err = setup.db.Exec(`UPDATE users SET bio = ?, image = ? WHERE username = ?`,
			"I am the target user", "https://example.com/avatar.png", "completetarget")
		if err != nil {
			t.Fatalf("failed to update user bio: %v", err)
		}

		// Create follower user
		follower, _, err := setup.authService.Register(ctx, &domain.CreateUserInput{
			Email:    "complete-follower@example.com",
			Username: "completefollower",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("failed to register follower user: %v", err)
		}

		// Step 1: Get profile without auth - following should be false
		req := httptest.NewRequest(http.MethodGet, "/api/profiles/completetarget", nil)
		req.SetPathValue("username", "completetarget")
		w := httptest.NewRecorder()
		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("step 1: expected status %d, got %d", http.StatusOK, w.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)
		profile := resp["profile"].(map[string]interface{})
		if profile["following"] != false {
			t.Errorf("step 1: expected following false, got %v", profile["following"])
		}
		if profile["bio"] != "I am the target user" {
			t.Errorf("step 1: expected bio, got %v", profile["bio"])
		}

		// Step 2: Get profile with auth - following should be false (not followed yet)
		req = httptest.NewRequest(http.MethodGet, "/api/profiles/completetarget", nil)
		req.SetPathValue("username", "completetarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("step 2: expected status %d, got %d", http.StatusOK, w.Code)
		}
		json.NewDecoder(w.Body).Decode(&resp)
		profile = resp["profile"].(map[string]interface{})
		if profile["following"] != false {
			t.Errorf("step 2: expected following false, got %v", profile["following"])
		}

		// Step 3: Follow the user
		req = httptest.NewRequest(http.MethodPost, "/api/profiles/completetarget/follow", nil)
		req.SetPathValue("username", "completetarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		setup.handler.FollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("step 3: expected status %d, got %d", http.StatusOK, w.Code)
		}
		json.NewDecoder(w.Body).Decode(&resp)
		profile = resp["profile"].(map[string]interface{})
		if profile["following"] != true {
			t.Errorf("step 3: expected following true, got %v", profile["following"])
		}

		// Step 4: Get profile again - following should be true
		req = httptest.NewRequest(http.MethodGet, "/api/profiles/completetarget", nil)
		req.SetPathValue("username", "completetarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("step 4: expected status %d, got %d", http.StatusOK, w.Code)
		}
		json.NewDecoder(w.Body).Decode(&resp)
		profile = resp["profile"].(map[string]interface{})
		if profile["following"] != true {
			t.Errorf("step 4: expected following true, got %v", profile["following"])
		}

		// Step 5: Unfollow the user
		req = httptest.NewRequest(http.MethodDelete, "/api/profiles/completetarget/follow", nil)
		req.SetPathValue("username", "completetarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		setup.handler.UnfollowUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("step 5: expected status %d, got %d", http.StatusOK, w.Code)
		}
		json.NewDecoder(w.Body).Decode(&resp)
		profile = resp["profile"].(map[string]interface{})
		if profile["following"] != false {
			t.Errorf("step 5: expected following false, got %v", profile["following"])
		}

		// Step 6: Get profile again - following should be false
		req = httptest.NewRequest(http.MethodGet, "/api/profiles/completetarget", nil)
		req.SetPathValue("username", "completetarget")
		ctx = context.WithValue(req.Context(), UserIDContextKey, follower.ID)
		req = req.WithContext(ctx)
		w = httptest.NewRecorder()
		setup.handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("step 6: expected status %d, got %d", http.StatusOK, w.Code)
		}
		json.NewDecoder(w.Body).Decode(&resp)
		profile = resp["profile"].(map[string]interface{})
		if profile["following"] != false {
			t.Errorf("step 6: expected following false, got %v", profile["following"])
		}
	})
}
