package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/alexlee0213/realworld-conduit/backend/internal/domain"
	"github.com/alexlee0213/realworld-conduit/backend/internal/repository"
	"github.com/alexlee0213/realworld-conduit/backend/internal/util"
)

// ArticleService handles article business logic
type ArticleService struct {
	articleRepo repository.ArticleRepository
	userRepo    repository.UserRepository
	logger      *slog.Logger
}

// NewArticleService creates a new ArticleService instance
func NewArticleService(
	articleRepo repository.ArticleRepository,
	userRepo repository.UserRepository,
	logger *slog.Logger,
) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// CreateArticle creates a new article
func (s *ArticleService) CreateArticle(ctx context.Context, authorID int64, input *domain.CreateArticleInput) (*domain.Article, error) {
	// Validate input
	if err := s.validateCreateArticleInput(input); err != nil {
		return nil, err
	}

	// Generate unique slug
	baseSlug := util.GenerateSlug(input.Title)
	slug := util.GenerateUniqueSlug(input.Title, func(slug string) bool {
		return s.articleRepo.SlugExists(ctx, slug)
	})

	article := &domain.Article{
		Slug:        slug,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Body:        input.Body,
		AuthorID:    authorID,
	}

	if err := s.articleRepo.CreateArticle(ctx, article, input.TagList); err != nil {
		return nil, err
	}

	// Load tags
	article.TagList = input.TagList
	if article.TagList == nil {
		article.TagList = []string{}
	}

	s.logger.Info("article created",
		"article_id", article.ID,
		"slug", article.Slug,
		"author_id", authorID,
		"base_slug", baseSlug,
	)

	return article, nil
}

// GetArticleBySlug retrieves an article by its slug
func (s *ArticleService) GetArticleBySlug(ctx context.Context, slug string, currentUserID *int64) (*domain.Article, error) {
	article, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Load author information
	author, err := s.userRepo.GetUserByID(ctx, article.AuthorID)
	if err != nil {
		s.logger.Error("failed to get article author", "error", err, "author_id", article.AuthorID)
		return nil, err
	}
	article.Author = author

	return article, nil
}

// UpdateArticle updates an existing article
// Only the author can update the article (explicit authorization check)
func (s *ArticleService) UpdateArticle(ctx context.Context, slug string, authorID int64, input *domain.UpdateArticleInput) (*domain.Article, error) {
	// Get the article
	article, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// EXPLICIT AUTHORIZATION CHECK: Only the author can update
	if article.AuthorID != authorID {
		s.logger.Warn("unauthorized article update attempt",
			"article_id", article.ID,
			"author_id", article.AuthorID,
			"attempted_by", authorID,
		)
		return nil, domain.ErrForbidden
	}

	// Apply updates
	if input.Title != nil {
		newTitle := strings.TrimSpace(*input.Title)
		article.Title = newTitle
		// Regenerate slug if title changed
		article.Slug = util.GenerateUniqueSlug(newTitle, func(candidateSlug string) bool {
			// Allow the same slug if it's the article's current slug
			if candidateSlug == slug {
				return false
			}
			return s.articleRepo.SlugExists(ctx, candidateSlug)
		})
	}
	if input.Description != nil {
		article.Description = strings.TrimSpace(*input.Description)
	}
	if input.Body != nil {
		article.Body = *input.Body
	}

	if err := s.articleRepo.UpdateArticle(ctx, article); err != nil {
		return nil, err
	}

	// Load author information
	author, err := s.userRepo.GetUserByID(ctx, article.AuthorID)
	if err != nil {
		s.logger.Error("failed to get article author", "error", err, "author_id", article.AuthorID)
		return nil, err
	}
	article.Author = author

	s.logger.Info("article updated",
		"article_id", article.ID,
		"slug", article.Slug,
		"updated_by", authorID,
	)

	return article, nil
}

// DeleteArticle deletes an article
// Only the author can delete the article (explicit authorization check)
func (s *ArticleService) DeleteArticle(ctx context.Context, slug string, authorID int64) error {
	// Get the article
	article, err := s.articleRepo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}

	// EXPLICIT AUTHORIZATION CHECK: Only the author can delete
	if article.AuthorID != authorID {
		s.logger.Warn("unauthorized article delete attempt",
			"article_id", article.ID,
			"author_id", article.AuthorID,
			"attempted_by", authorID,
		)
		return domain.ErrForbidden
	}

	if err := s.articleRepo.DeleteArticle(ctx, article.ID); err != nil {
		return err
	}

	s.logger.Info("article deleted",
		"article_id", article.ID,
		"slug", slug,
		"deleted_by", authorID,
	)

	return nil
}

// ListArticles retrieves a list of articles with optional filters
func (s *ArticleService) ListArticles(ctx context.Context, params *domain.ArticleListParams, currentUserID *int64) ([]*domain.Article, int, error) {
	if params == nil {
		params = domain.DefaultArticleListParams()
	}

	// Apply defaults if not set
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	articles, total, err := s.articleRepo.ListArticles(ctx, params, currentUserID)
	if err != nil {
		return nil, 0, err
	}

	// Load author information for each article
	for _, article := range articles {
		author, err := s.userRepo.GetUserByID(ctx, article.AuthorID)
		if err != nil {
			s.logger.Error("failed to get article author", "error", err, "author_id", article.AuthorID)
			continue
		}
		article.Author = author
	}

	return articles, total, nil
}

// GetFeed retrieves articles from followed users
func (s *ArticleService) GetFeed(ctx context.Context, userID int64, params *domain.ArticleFeedParams) ([]*domain.Article, int, error) {
	if params == nil {
		params = domain.DefaultArticleFeedParams()
	}

	// Apply defaults if not set
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	articles, total, err := s.articleRepo.GetFeed(ctx, userID, params)
	if err != nil {
		return nil, 0, err
	}

	// Load author information for each article
	for _, article := range articles {
		author, err := s.userRepo.GetUserByID(ctx, article.AuthorID)
		if err != nil {
			s.logger.Error("failed to get article author", "error", err, "author_id", article.AuthorID)
			continue
		}
		article.Author = author
	}

	return articles, total, nil
}

// GetAllTags retrieves all unique tags
func (s *ArticleService) GetAllTags(ctx context.Context) ([]string, error) {
	return s.articleRepo.GetAllTags(ctx)
}

// validateCreateArticleInput validates article creation input
func (s *ArticleService) validateCreateArticleInput(input *domain.CreateArticleInput) error {
	validationErrors := domain.NewValidationErrors()

	if strings.TrimSpace(input.Title) == "" {
		validationErrors.Add("title", "can't be blank")
	}
	if strings.TrimSpace(input.Description) == "" {
		validationErrors.Add("description", "can't be blank")
	}
	if strings.TrimSpace(input.Body) == "" {
		validationErrors.Add("body", "can't be blank")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
