import type { Author } from '../article';

// Comment type returned from API
export interface Comment {
  id: number;
  body: string;
  createdAt: string;
  updatedAt: string;
  author: Author;
}

// Single comment response wrapper
export interface CommentResponse {
  comment: Comment;
}

// Multiple comments response wrapper
export interface CommentsResponse {
  comments: Comment[];
}

// Create comment request
export interface CreateCommentRequest {
  comment: {
    body: string;
  };
}
