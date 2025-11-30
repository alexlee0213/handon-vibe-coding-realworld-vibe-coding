import { api } from '../../lib/api';
import type { CommentResponse, CommentsResponse, CreateCommentRequest } from './types';

// Get comments for an article
export async function getComments(slug: string): Promise<CommentsResponse> {
  const response = await api.get(`articles/${slug}/comments`);
  return response.json<CommentsResponse>();
}

// Create a comment on an article
export async function createComment(
  slug: string,
  data: CreateCommentRequest['comment']
): Promise<CommentResponse> {
  const response = await api.post(`articles/${slug}/comments`, {
    json: { comment: data },
  });
  return response.json<CommentResponse>();
}

// Delete a comment from an article
export async function deleteComment(slug: string, commentId: number): Promise<void> {
  await api.delete(`articles/${slug}/comments/${commentId}`);
}
