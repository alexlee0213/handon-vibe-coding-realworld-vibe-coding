package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors for common domain errors
var (
	// User errors
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrEmailAlreadyTaken    = errors.New("email is already taken")
	ErrUsernameAlreadyTaken = errors.New("username is already taken")
	ErrInvalidCredentials   = errors.New("invalid email or password")

	// Article errors
	ErrArticleNotFound      = errors.New("article not found")
	ErrArticleAlreadyExists = errors.New("article with this slug already exists")

	// Comment errors
	ErrCommentNotFound = errors.New("comment not found")

	// Authorization errors
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	// Validation errors
	ErrValidation = errors.New("validation error")

	// Database errors
	ErrDatabase = errors.New("database error")
)

// ValidationError represents a validation error with field-level details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e.Errors[0].Error())
}

// Add adds a new validation error
func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are any validation errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// NewValidationErrors creates a new ValidationErrors instance
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// IsNotFound checks if the error is a "not found" type error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrArticleNotFound) ||
		errors.Is(err, ErrCommentNotFound)
}

// IsConflict checks if the error is a conflict/duplicate error
func IsConflict(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists) ||
		errors.Is(err, ErrEmailAlreadyTaken) ||
		errors.Is(err, ErrUsernameAlreadyTaken) ||
		errors.Is(err, ErrArticleAlreadyExists)
}

// IsUnauthorized checks if the error is an authorization error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrInvalidCredentials)
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}
