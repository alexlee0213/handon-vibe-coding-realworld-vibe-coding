import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '../../test/utils';
import { CommentList } from './CommentList';
import type { Comment } from '../../features/comment';

const mockComments: Comment[] = [
  {
    id: 1,
    body: 'First comment',
    createdAt: '2024-01-15T10:30:00Z',
    updatedAt: '2024-01-15T10:30:00Z',
    author: {
      username: 'user1',
      bio: '',
      image: '',
      following: false,
    },
  },
  {
    id: 2,
    body: 'Second comment',
    createdAt: '2024-01-16T10:30:00Z',
    updatedAt: '2024-01-16T10:30:00Z',
    author: {
      username: 'user2',
      bio: '',
      image: '',
      following: false,
    },
  },
];

describe('CommentList', () => {
  it('renders loading state', () => {
    render(<CommentList isLoading={true} />);

    // Mantine Loader uses a span element with class mantine-Loader-root
    expect(document.querySelector('.mantine-Loader-root')).toBeInTheDocument();
  });

  it('renders error state', () => {
    const error = new Error('Failed to load comments');
    render(<CommentList error={error} />);

    expect(screen.getByText('Error loading comments')).toBeInTheDocument();
    expect(screen.getByText('Failed to load comments')).toBeInTheDocument();
  });

  it('renders empty state when no comments', () => {
    render(<CommentList comments={[]} />);

    expect(screen.getByText(/no comments yet/i)).toBeInTheDocument();
  });

  it('renders list of comments', () => {
    render(<CommentList comments={mockComments} />);

    expect(screen.getByText('First comment')).toBeInTheDocument();
    expect(screen.getByText('Second comment')).toBeInTheDocument();
    expect(screen.getByText('user1')).toBeInTheDocument();
    expect(screen.getByText('user2')).toBeInTheDocument();
  });

  it('shows delete button only for current user comments', () => {
    const onDeleteComment = vi.fn();
    render(
      <CommentList
        comments={mockComments}
        currentUsername="user1"
        onDeleteComment={onDeleteComment}
      />
    );

    const deleteButtons = screen.getAllByRole('button', { name: /delete comment/i });
    expect(deleteButtons).toHaveLength(1);
  });

  it('calls onDeleteComment with correct comment id', () => {
    const onDeleteComment = vi.fn();
    render(
      <CommentList
        comments={mockComments}
        currentUsername="user1"
        onDeleteComment={onDeleteComment}
      />
    );

    const deleteButton = screen.getByRole('button', { name: /delete comment/i });
    fireEvent.click(deleteButton);

    expect(onDeleteComment).toHaveBeenCalledWith(1);
  });
});
