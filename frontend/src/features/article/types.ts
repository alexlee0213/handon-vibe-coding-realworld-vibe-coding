// Author profile in article responses
export interface Author {
  username: string;
  bio: string;
  image: string;
  following: boolean;
}

// Article type returned from API
export interface Article {
  slug: string;
  title: string;
  description: string;
  body: string;
  tagList: string[];
  createdAt: string;
  updatedAt: string;
  favorited: boolean;
  favoritesCount: number;
  author: Author;
}

// Single article response wrapper
export interface ArticleResponse {
  article: Article;
}

// Multiple articles response wrapper
export interface ArticlesResponse {
  articles: Article[];
  articlesCount: number;
}

// Tags response wrapper
export interface TagsResponse {
  tags: string[];
}

// Create article request
export interface CreateArticleRequest {
  article: {
    title: string;
    description: string;
    body: string;
    tagList?: string[];
  };
}

// Update article request
export interface UpdateArticleRequest {
  article: {
    title?: string;
    description?: string;
    body?: string;
  };
}

// Query parameters for listing articles
export interface ArticleListParams {
  tag?: string;
  author?: string;
  favorited?: string;
  limit?: number;
  offset?: number;
}

// Query parameters for feed
export interface ArticleFeedParams {
  limit?: number;
  offset?: number;
}
