import { api } from '../../lib/api';
import type {
  ArticleResponse,
  ArticlesResponse,
  TagsResponse,
  CreateArticleRequest,
  UpdateArticleRequest,
  ArticleListParams,
  ArticleFeedParams,
} from './types';

// List articles with optional filters
export async function listArticles(params?: ArticleListParams): Promise<ArticlesResponse> {
  const searchParams: Record<string, string> = {};

  if (params?.tag) searchParams.tag = params.tag;
  if (params?.author) searchParams.author = params.author;
  if (params?.favorited) searchParams.favorited = params.favorited;
  if (params?.limit !== undefined) searchParams.limit = String(params.limit);
  if (params?.offset !== undefined) searchParams.offset = String(params.offset);

  const response = await api.get('articles', { searchParams });
  return response.json<ArticlesResponse>();
}

// Get user feed (articles from followed users)
export async function getFeed(params?: ArticleFeedParams): Promise<ArticlesResponse> {
  const searchParams: Record<string, string> = {};

  if (params?.limit !== undefined) searchParams.limit = String(params.limit);
  if (params?.offset !== undefined) searchParams.offset = String(params.offset);

  const response = await api.get('articles/feed', { searchParams });
  return response.json<ArticlesResponse>();
}

// Get single article by slug
export async function getArticle(slug: string): Promise<ArticleResponse> {
  const response = await api.get(`articles/${slug}`);
  return response.json<ArticleResponse>();
}

// Create new article
export async function createArticle(
  data: CreateArticleRequest['article']
): Promise<ArticleResponse> {
  const response = await api.post('articles', {
    json: { article: data },
  });
  return response.json<ArticleResponse>();
}

// Update existing article
export async function updateArticle(
  slug: string,
  data: UpdateArticleRequest['article']
): Promise<ArticleResponse> {
  // Filter out undefined values
  const filteredData = Object.fromEntries(
    Object.entries(data).filter(([, value]) => value !== undefined)
  );

  const response = await api.put(`articles/${slug}`, {
    json: { article: filteredData },
  });
  return response.json<ArticleResponse>();
}

// Delete article
export async function deleteArticle(slug: string): Promise<void> {
  await api.delete(`articles/${slug}`);
}

// Get all tags
export async function getTags(): Promise<TagsResponse> {
  const response = await api.get('tags');
  return response.json<TagsResponse>();
}
