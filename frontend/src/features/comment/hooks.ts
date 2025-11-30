import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import * as commentApi from './api';
import type { Comment, CreateCommentRequest } from './types';

// Query keys factory
export const commentKeys = {
  all: ['comments'] as const,
  lists: () => [...commentKeys.all, 'list'] as const,
  list: (slug: string) => [...commentKeys.lists(), slug] as const,
};

// Hook to get comments for an article
export function useComments(slug: string) {
  return useQuery({
    queryKey: commentKeys.list(slug),
    queryFn: () => commentApi.getComments(slug),
    staleTime: 1 * 60 * 1000, // 1 minute
    enabled: !!slug,
  });
}

// Hook to create a comment
export function useCreateComment(slug: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateCommentRequest['comment']) =>
      commentApi.createComment(slug, data),
    onSuccess: (response) => {
      // Optimistically add the new comment to the cache
      queryClient.setQueryData(
        commentKeys.list(slug),
        (old: { comments: Comment[] } | undefined) => {
          if (!old) return { comments: [response.comment] };
          return { comments: [response.comment, ...old.comments] };
        }
      );
    },
  });
}

// Hook to delete a comment
export function useDeleteComment(slug: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (commentId: number) => commentApi.deleteComment(slug, commentId),
    onSuccess: (_, commentId) => {
      // Remove the deleted comment from cache
      queryClient.setQueryData(
        commentKeys.list(slug),
        (old: { comments: Comment[] } | undefined) => {
          if (!old) return old;
          return {
            comments: old.comments.filter((c) => c.id !== commentId),
          };
        }
      );
    },
  });
}
