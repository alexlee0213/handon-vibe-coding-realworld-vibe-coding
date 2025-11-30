package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
	logger    *slog.Logger
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
		logger:    logger,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, input *domain.CreateUserInput) (*domain.User, string, error) {
	// Validate input
	if err := s.validateRegisterInput(input); err != nil {
		return nil, "", err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, "", errors.Join(domain.ErrDatabase, err)
	}

	// Create user
	user := &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		Username:     strings.TrimSpace(input.Username),
		PasswordHash: string(hashedPassword),
		Bio:          "",
		Image:        "",
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	s.logger.Info("user registered",
		"user_id", user.ID,
		"username", user.Username,
	)

	return user, token, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	// Find user by email
	user, err := s.userRepo.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	s.logger.Info("user logged in",
		"user_id", user.ID,
		"username", user.Username,
	)

	return user, token, nil
}

// GenerateToken creates a new JWT token for the given user ID
func (s *AuthService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error("failed to sign token", "error", err)
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id in token")
	}

	return int64(userIDFloat), nil
}

// GetCurrentUser retrieves the current user by ID
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

// UpdateUser updates user information
func (s *AuthService) UpdateUser(ctx context.Context, userID int64, input *domain.UpdateUserInput) (*domain.User, error) {
	// Get current user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Email != nil {
		user.Email = strings.ToLower(strings.TrimSpace(*input.Email))
	}
	if input.Username != nil {
		user.Username = strings.TrimSpace(*input.Username)
	}
	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", "error", err)
			return nil, errors.Join(domain.ErrDatabase, err)
		}
		user.PasswordHash = string(hashedPassword)
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.Image != nil {
		user.Image = *input.Image
	}

	// Save updates
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("user updated",
		"user_id", user.ID,
		"username", user.Username,
	)

	return user, nil
}

// validateRegisterInput validates registration input
func (s *AuthService) validateRegisterInput(input *domain.CreateUserInput) error {
	validationErrors := domain.NewValidationErrors()

	if strings.TrimSpace(input.Email) == "" {
		validationErrors.Add("email", "email is required")
	}
	if strings.TrimSpace(input.Username) == "" {
		validationErrors.Add("username", "username is required")
	}
	if input.Password == "" {
		validationErrors.Add("password", "password is required")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
