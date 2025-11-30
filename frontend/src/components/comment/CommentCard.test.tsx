import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '../../test/utils';
import { CommentCard } from './CommentCard';
import type { Comment } from '../../features/comment';

const mockComment: Comment = {
  id: 1,
  body: 'This is a test comment',
  createdAt: '2024-01-15T10:30:00Z',
  updatedAt: '2024-01-15T10:30:00Z',
  author: {
    username: 'testuser',
    bio: 'Test bio',
    image: 'https://example.com/avatar.png',
    following: false,
  },
};

describe('CommentCard', () => {
  it('renders comment body', () => {
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={false}
      />
    );

    expect(screen.getByText('This is a test comment')).toBeInTheDocument();
  });

  it('renders author username', () => {
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={false}
      />
    );

    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  it('renders formatted date', () => {
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={false}
      />
    );

    expect(screen.getByText('January 15, 2024')).toBeInTheDocument();
  });

  it('shows delete button when user is author', () => {
    const onDelete = vi.fn();
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={true}
        onDelete={onDelete}
      />
    );

    const deleteButton = screen.getByRole('button', { name: /delete comment/i });
    expect(deleteButton).toBeInTheDocument();
  });

  it('hides delete button when user is not author', () => {
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={false}
      />
    );

    expect(screen.queryByRole('button', { name: /delete comment/i })).not.toBeInTheDocument();
  });

  it('calls onDelete when delete button is clicked', () => {
    const onDelete = vi.fn();
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={true}
        onDelete={onDelete}
      />
    );

    const deleteButton = screen.getByRole('button', { name: /delete comment/i });
    fireEvent.click(deleteButton);

    expect(onDelete).toHaveBeenCalledTimes(1);
  });

  it('renders author avatar', () => {
    render(
      <CommentCard
        comment={mockComment}
        isAuthor={false}
      />
    );

    const avatar = screen.getByAltText('testuser');
    expect(avatar).toBeInTheDocument();
    expect(avatar).toHaveAttribute('src', 'https://example.com/avatar.png');
  });
});
