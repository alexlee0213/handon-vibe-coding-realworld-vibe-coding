import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { theme } from '../../lib/theme';
import { FavoriteButton } from './FavoriteButton';

// Mock useNavigate - must be before component import
const mockNavigate = vi.fn();
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}));

// Mock auth store
const mockIsAuthenticated = vi.fn(() => false);
vi.mock('../../features/auth', () => ({
  useIsAuthenticated: () => mockIsAuthenticated(),
}));

// Mock favorite hooks
const mockFavoriteMutate = vi.fn();
const mockUnfavoriteMutate = vi.fn();
vi.mock('../../features/article', () => ({
  useFavoriteArticle: () => ({
    mutate: mockFavoriteMutate,
    isPending: false,
  }),
  useUnfavoriteArticle: () => ({
    mutate: mockUnfavoriteMutate,
    isPending: false,
  }),
}));

// Custom render with MantineProvider
const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider theme={theme}>{ui}</MantineProvider>);
};

describe('FavoriteButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIsAuthenticated.mockReturnValue(false);
  });

  it('renders unfavorited state correctly', () => {
    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={false}
        favoritesCount={5}
      />
    );

    expect(screen.getByRole('button', { name: /favorite article/i })).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('renders favorited state correctly', () => {
    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={true}
        favoritesCount={10}
      />
    );

    expect(screen.getByRole('button', { name: /unfavorite article/i })).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
  });

  it('hides count when showCount is false', () => {
    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={false}
        favoritesCount={5}
        showCount={false}
      />
    );

    expect(screen.queryByText('5')).not.toBeInTheDocument();
  });

  it('redirects to login when clicked and not authenticated', () => {
    mockIsAuthenticated.mockReturnValue(false);

    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={false}
        favoritesCount={5}
      />
    );

    const button = screen.getByRole('button', { name: /favorite article/i });
    fireEvent.click(button);

    expect(mockNavigate).toHaveBeenCalledWith({ to: '/login' });
    expect(mockFavoriteMutate).not.toHaveBeenCalled();
  });

  it('calls favorite mutation when clicked and not favorited', () => {
    mockIsAuthenticated.mockReturnValue(true);

    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={false}
        favoritesCount={5}
      />
    );

    const button = screen.getByRole('button', { name: /favorite article/i });
    fireEvent.click(button);

    expect(mockFavoriteMutate).toHaveBeenCalledWith('test-article');
    expect(mockUnfavoriteMutate).not.toHaveBeenCalled();
  });

  it('calls unfavorite mutation when clicked and already favorited', () => {
    mockIsAuthenticated.mockReturnValue(true);

    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={true}
        favoritesCount={10}
      />
    );

    const button = screen.getByRole('button', { name: /unfavorite article/i });
    fireEvent.click(button);

    expect(mockUnfavoriteMutate).toHaveBeenCalledWith('test-article');
    expect(mockFavoriteMutate).not.toHaveBeenCalled();
  });

  it('renders with different sizes', () => {
    const { rerender } = renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={false}
        favoritesCount={5}
        size="xs"
      />
    );

    expect(screen.getByRole('button')).toBeInTheDocument();

    rerender(
      <MantineProvider theme={theme}>
        <FavoriteButton
          slug="test-article"
          favorited={false}
          favoritesCount={5}
          size="lg"
        />
      </MantineProvider>
    );

    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('shows zero count correctly', () => {
    renderWithMantine(
      <FavoriteButton
        slug="test-article"
        favorited={false}
        favoritesCount={0}
      />
    );

    expect(screen.getByText('0')).toBeInTheDocument();
  });
});
