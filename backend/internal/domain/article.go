package domain

import (
	"time"
)

// Article represents a blog article in the system
type Article struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Body        string    `json:"body"`
	AuthorID    int64     `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Related data (populated by queries)
	Author         *User    `json:"author,omitempty"`
	TagList        []string `json:"tagList"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
}

// ArticleResponse represents the article data returned to clients (RealWorld API format)
type ArticleResponse struct {
	Slug           string           `json:"slug"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Body           string           `json:"body"`
	TagList        []string         `json:"tagList"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
	Favorited      bool             `json:"favorited"`
	FavoritesCount int              `json:"favoritesCount"`
	Author         *ProfileResponse `json:"author"`
}

// ToArticleResponse converts an Article to ArticleResponse
func (a *Article) ToArticleResponse(author *ProfileResponse) *ArticleResponse {
	tagList := a.TagList
	if tagList == nil {
		tagList = []string{}
	}
	return &ArticleResponse{
		Slug:           a.Slug,
		Title:          a.Title,
		Description:    a.Description,
		Body:           a.Body,
		TagList:        tagList,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		Favorited:      a.Favorited,
		FavoritesCount: a.FavoritesCount,
		Author:         author,
	}
}

// CreateArticleInput represents the input for creating a new article
type CreateArticleInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList,omitempty"`
}

// UpdateArticleInput represents the input for updating an article
type UpdateArticleInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Body        *string `json:"body,omitempty"`
}

// ArticleListParams represents parameters for listing articles
type ArticleListParams struct {
	Tag       string // Filter by tag
	Author    string // Filter by author username
	Favorited string // Filter by username who favorited
	Limit     int    // Number of articles to return (default 20)
	Offset    int    // Number of articles to skip (default 0)
}

// DefaultArticleListParams returns default list parameters
func DefaultArticleListParams() *ArticleListParams {
	return &ArticleListParams{
		Limit:  20,
		Offset: 0,
	}
}

// ArticleFeedParams represents parameters for the user feed
type ArticleFeedParams struct {
	Limit  int // Number of articles to return (default 20)
	Offset int // Number of articles to skip (default 0)
}

// DefaultArticleFeedParams returns default feed parameters
func DefaultArticleFeedParams() *ArticleFeedParams {
	return &ArticleFeedParams{
		Limit:  20,
		Offset: 0,
	}
}
