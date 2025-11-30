package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
)

// CommentService handles comment business logic
type CommentService struct {
	commentRepo repository.CommentRepository
	articleRepo repository.ArticleRepository
	userRepo    repository.UserRepository
	logger      *slog.Logger
}

// NewCommentService creates a new CommentService instance
func NewCommentService(
	commentRepo repository.CommentRepository,
	articleRepo repository.ArticleRepository,
	userRepo repository.UserRepository,
	logger *slog.Logger,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// CreateComment creates a new comment on an article
func (s *CommentService) CreateComment(ctx context.Context, slug string, authorID int64, input *domain.CreateCommentInput) (*domain.Comment, error) {
	// Validate input
	if validationErrors := input.Validate(); validationErrors.HasErrors() {
		return nil, validationErrors
	}

	// Get the article by slug to verify it exists
	article, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	comment := &domain.Comment{
		Body:      strings.TrimSpace(input.Body),
		ArticleID: article.ID,
		AuthorID:  authorID,
	}

	if err := s.commentRepo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	// Load author information
	author, err := s.userRepo.GetUserByID(ctx, authorID)
	if err != nil {
		s.logger.Error("failed to get comment author", "error", err, "author_id", authorID)
		return nil, err
	}
	comment.Author = author

	s.logger.Info("comment created",
		"comment_id", comment.ID,
		"article_slug", slug,
		"author_id", authorID,
	)

	return comment, nil
}

// GetCommentsByArticleSlug retrieves all comments for an article
func (s *CommentService) GetCommentsByArticleSlug(ctx context.Context, slug string) ([]*domain.Comment, error) {
	// Get the article by slug to verify it exists and get its ID
	article, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetCommentsByArticleID(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	// Load author information for each comment
	for _, comment := range comments {
		author, err := s.userRepo.GetUserByID(ctx, comment.AuthorID)
		if err != nil {
			s.logger.Error("failed to get comment author", "error", err, "author_id", comment.AuthorID)
			continue
		}
		comment.Author = author
	}

	return comments, nil
}

// DeleteComment deletes a comment
// Only the comment author can delete the comment (explicit authorization check)
func (s *CommentService) DeleteComment(ctx context.Context, slug string, commentID int64, userID int64) error {
	// Get the article by slug to verify it exists
	_, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}

	// Get the comment
	comment, err := s.commentRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	// EXPLICIT AUTHORIZATION CHECK: Only the author can delete
	if comment.AuthorID != userID {
		s.logger.Warn("unauthorized comment delete attempt",
			"comment_id", commentID,
			"author_id", comment.AuthorID,
			"attempted_by", userID,
		)
		return domain.ErrForbidden
	}

	if err := s.commentRepo.DeleteComment(ctx, commentID); err != nil {
		return err
	}

	s.logger.Info("comment deleted",
		"comment_id", commentID,
		"article_slug", slug,
		"deleted_by", userID,
	)

	return nil
}
