import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { theme } from '../../lib/theme';
import { ArticleMeta } from './ArticleMeta';
import type { Author } from '../../features/article';

// Mock react-router
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}));

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider theme={theme}>{ui}</MantineProvider>);
};

const mockAuthor: Author = {
  username: 'testuser',
  bio: 'Test bio',
  image: 'https://example.com/avatar.jpg',
  following: false,
};

describe('ArticleMeta', () => {
  it('renders author username', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2024-01-15T10:30:00.000Z" />
    );
    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  it('renders formatted date', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2024-01-15T10:30:00.000Z" />
    );
    expect(screen.getByText(/January 15, 2024/)).toBeInTheDocument();
  });

  it('renders author avatar', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2024-01-15T10:30:00.000Z" />
    );
    const avatar = screen.getByRole('img', { name: 'testuser' });
    expect(avatar).toBeInTheDocument();
    expect(avatar).toHaveAttribute('src', 'https://example.com/avatar.jpg');
  });

  it('renders without author image', () => {
    const authorWithoutImage = { ...mockAuthor, image: null };
    renderWithMantine(
      <ArticleMeta author={authorWithoutImage} createdAt="2024-01-15T10:30:00.000Z" />
    );
    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  it('links to author profile', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2024-01-15T10:30:00.000Z" />
    );
    // Find the link that contains the username text
    const usernameText = screen.getByText('testuser');
    const profileLink = usernameText.closest('a');
    expect(profileLink).toHaveAttribute('href', '/profile/testuser');
  });

  it('renders with small size', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2024-01-15T10:30:00.000Z" size="sm" />
    );
    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  it('renders with medium size (default)', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2024-01-15T10:30:00.000Z" size="md" />
    );
    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  it('handles different date formats', () => {
    renderWithMantine(
      <ArticleMeta author={mockAuthor} createdAt="2023-12-25T00:00:00.000Z" />
    );
    expect(screen.getByText(/December 25, 2023/)).toBeInTheDocument();
  });
});
