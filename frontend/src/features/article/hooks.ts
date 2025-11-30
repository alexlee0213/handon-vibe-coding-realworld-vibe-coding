import { useMutation, useQuery, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import * as articleApi from './api';
import type {
  Article,
  CreateArticleRequest,
  UpdateArticleRequest,
  ArticleListParams,
  ArticleFeedParams,
} from './types';

// Query keys factory
export const articleKeys = {
  all: ['articles'] as const,
  lists: () => [...articleKeys.all, 'list'] as const,
  list: (params?: ArticleListParams) => [...articleKeys.lists(), params] as const,
  feed: () => [...articleKeys.all, 'feed'] as const,
  feedPaginated: (params?: ArticleFeedParams) => [...articleKeys.feed(), params] as const,
  details: () => [...articleKeys.all, 'detail'] as const,
  detail: (slug: string) => [...articleKeys.details(), slug] as const,
  tags: () => ['tags'] as const,
};

// Hook to list articles with filters
export function useArticles(params?: ArticleListParams) {
  return useQuery({
    queryKey: articleKeys.list(params),
    queryFn: () => articleApi.listArticles(params),
    staleTime: 1 * 60 * 1000, // 1 minute
  });
}

// Hook for infinite scrolling articles list
export function useInfiniteArticles(params?: Omit<ArticleListParams, 'offset'>) {
  const limit = params?.limit ?? 10;

  return useInfiniteQuery({
    queryKey: [...articleKeys.list(params), 'infinite'],
    queryFn: ({ pageParam = 0 }) =>
      articleApi.listArticles({ ...params, limit, offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      const totalFetched = allPages.reduce((sum, page) => sum + page.articles.length, 0);
      if (totalFetched >= lastPage.articlesCount) {
        return undefined;
      }
      return totalFetched;
    },
  });
}

// Hook to get user feed
export function useFeed(params?: ArticleFeedParams) {
  return useQuery({
    queryKey: articleKeys.feedPaginated(params),
    queryFn: () => articleApi.getFeed(params),
    staleTime: 30 * 1000, // 30 seconds (more fresh for feed)
  });
}

// Hook for infinite scrolling feed
export function useInfiniteFeed(params?: Omit<ArticleFeedParams, 'offset'>) {
  const limit = params?.limit ?? 10;

  return useInfiniteQuery({
    queryKey: [...articleKeys.feed(), 'infinite'],
    queryFn: ({ pageParam = 0 }) =>
      articleApi.getFeed({ ...params, limit, offset: pageParam }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      const totalFetched = allPages.reduce((sum, page) => sum + page.articles.length, 0);
      if (totalFetched >= lastPage.articlesCount) {
        return undefined;
      }
      return totalFetched;
    },
  });
}

// Hook to get single article
export function useArticle(slug: string) {
  return useQuery({
    queryKey: articleKeys.detail(slug),
    queryFn: () => articleApi.getArticle(slug),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!slug,
  });
}

// Hook to create article
export function useCreateArticle() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (data: CreateArticleRequest['article']) => articleApi.createArticle(data),
    onSuccess: (response) => {
      // Invalidate articles list to refetch
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
      queryClient.invalidateQueries({ queryKey: articleKeys.feed() });
      // Cache the new article
      queryClient.setQueryData(articleKeys.detail(response.article.slug), response);
      // Navigate to the new article
      navigate({ to: '/article/$slug', params: { slug: response.article.slug } });
    },
  });
}

// Hook to update article
export function useUpdateArticle(slug: string) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (data: UpdateArticleRequest['article']) => articleApi.updateArticle(slug, data),
    onSuccess: (response) => {
      // Update the cached article
      queryClient.setQueryData(articleKeys.detail(response.article.slug), response);
      // If slug changed, remove old cache entry
      if (response.article.slug !== slug) {
        queryClient.removeQueries({ queryKey: articleKeys.detail(slug) });
      }
      // Invalidate lists to reflect changes
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
      // Navigate to the article (possibly with new slug)
      navigate({ to: '/article/$slug', params: { slug: response.article.slug } });
    },
  });
}

// Hook to delete article
export function useDeleteArticle() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (slug: string) => articleApi.deleteArticle(slug),
    onSuccess: (_, slug) => {
      // Remove the article from cache
      queryClient.removeQueries({ queryKey: articleKeys.detail(slug) });
      // Invalidate lists
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
      queryClient.invalidateQueries({ queryKey: articleKeys.feed() });
      // Navigate to home
      navigate({ to: '/' });
    },
  });
}

// Hook to get all tags
export function useTags() {
  return useQuery({
    queryKey: articleKeys.tags(),
    queryFn: articleApi.getTags,
    staleTime: 10 * 60 * 1000, // 10 minutes (tags don't change often)
  });
}

// Helper to optimistically update article in cache
export function useOptimisticArticleUpdate() {
  const queryClient = useQueryClient();

  return {
    updateArticleInCache: (slug: string, updater: (article: Article) => Article) => {
      // Update in detail cache
      queryClient.setQueryData(articleKeys.detail(slug), (old: { article: Article } | undefined) => {
        if (!old) return old;
        return { article: updater(old.article) };
      });

      // Update in list caches
      queryClient.setQueriesData(
        { queryKey: articleKeys.lists() },
        (old: { articles: Article[]; articlesCount: number } | undefined) => {
          if (!old) return old;
          return {
            ...old,
            articles: old.articles.map((a) => (a.slug === slug ? updater(a) : a)),
          };
        }
      );
    },
  };
}
