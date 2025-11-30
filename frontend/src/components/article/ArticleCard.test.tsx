import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { theme } from '../../lib/theme';
import { ArticleCard } from './ArticleCard';
import type { Article } from '../../features/article';

// Mock react-router
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}));

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider theme={theme}>{ui}</MantineProvider>);
};

const mockArticle: Article = {
  slug: 'test-article-slug',
  title: 'Test Article Title',
  description: 'This is a test article description',
  body: 'Article body content',
  tagList: ['react', 'testing', 'typescript'],
  createdAt: '2024-01-15T10:30:00.000Z',
  updatedAt: '2024-01-15T10:30:00.000Z',
  favorited: false,
  favoritesCount: 5,
  author: {
    username: 'testauthor',
    bio: 'Test bio',
    image: 'https://example.com/avatar.jpg',
    following: false,
  },
};

describe('ArticleCard', () => {
  it('renders article title', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    expect(screen.getByText('Test Article Title')).toBeInTheDocument();
  });

  it('renders article description', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    expect(screen.getByText('This is a test article description')).toBeInTheDocument();
  });

  it('renders author username', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    expect(screen.getByText('testauthor')).toBeInTheDocument();
  });

  it('renders article tags', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByText('testing')).toBeInTheDocument();
    expect(screen.getByText('typescript')).toBeInTheDocument();
  });

  it('renders Read more link', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    expect(screen.getByText('Read more...')).toBeInTheDocument();
  });

  it('renders formatted date', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    // Date should be formatted as "Month Day, Year"
    expect(screen.getByText(/January 15, 2024/)).toBeInTheDocument();
  });

  it('renders with favorited state', () => {
    const favoritedArticle = { ...mockArticle, favorited: true, favoritesCount: 10 };
    renderWithMantine(<ArticleCard article={favoritedArticle} />);
    // The favorite button should be present
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('renders with empty tagList', () => {
    const articleWithoutTags = { ...mockArticle, tagList: [] };
    renderWithMantine(<ArticleCard article={articleWithoutTags} />);
    expect(screen.getByText('Test Article Title')).toBeInTheDocument();
  });

  it('links to article detail page', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    const links = screen.getAllByRole('link');
    const articleLinks = links.filter(link => link.getAttribute('href')?.includes('/article/'));
    expect(articleLinks.length).toBeGreaterThan(0);
  });

  it('links to author profile', () => {
    renderWithMantine(<ArticleCard article={mockArticle} />);
    // Find the link that contains the author username
    const usernameText = screen.getByText('testauthor');
    const profileLink = usernameText.closest('a');
    expect(profileLink).toHaveAttribute('href', '/profile/testauthor');
  });
});
