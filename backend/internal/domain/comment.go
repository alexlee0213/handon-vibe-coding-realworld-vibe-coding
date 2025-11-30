package domain

import (
	"time"
)

// Comment represents a comment on an article
type Comment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	ArticleID int64     `json:"article_id"`
	AuthorID  int64     `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Related data (populated by queries)
	Author *User `json:"author,omitempty"`
}

// CommentResponse represents the comment data returned to clients (RealWorld API format)
type CommentResponse struct {
	ID        int64            `json:"id"`
	Body      string           `json:"body"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Author    *ProfileResponse `json:"author"`
}

// ToCommentResponse converts a Comment to CommentResponse
func (c *Comment) ToCommentResponse(author *ProfileResponse) *CommentResponse {
	return &CommentResponse{
		ID:        c.ID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Author:    author,
	}
}

// CreateCommentInput represents the input for creating a new comment
type CreateCommentInput struct {
	Body string `json:"body"`
}

// Validate validates the comment input
func (i *CreateCommentInput) Validate() *ValidationErrors {
	errors := NewValidationErrors()

	if i.Body == "" {
		errors.Add("body", "can't be blank")
	}

	return errors
}
