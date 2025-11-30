import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { theme } from '../../lib/theme';
import { Header } from './Header';

// Mock react-router
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}));

// Mock auth hooks
const mockIsAuthenticated = vi.fn();
const mockUser = vi.fn();
const mockLogout = vi.fn();

vi.mock('../../features/auth', () => ({
  useIsAuthenticated: () => mockIsAuthenticated(),
  useUser: () => mockUser(),
  useLogout: () => mockLogout,
}));

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider theme={theme}>{ui}</MantineProvider>);
};

describe('Header', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('when not authenticated', () => {
    beforeEach(() => {
      mockIsAuthenticated.mockReturnValue(false);
      mockUser.mockReturnValue(null);
    });

    it('renders conduit logo', () => {
      renderWithMantine(<Header />);
      expect(screen.getByText('conduit')).toBeInTheDocument();
    });

    it('renders Home link', () => {
      renderWithMantine(<Header />);
      expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument();
    });

    it('renders Sign in link', () => {
      renderWithMantine(<Header />);
      expect(screen.getByRole('link', { name: 'Sign in' })).toBeInTheDocument();
    });

    it('renders Sign up link', () => {
      renderWithMantine(<Header />);
      expect(screen.getByRole('link', { name: 'Sign up' })).toBeInTheDocument();
    });

    it('does not render New Article link', () => {
      renderWithMantine(<Header />);
      expect(screen.queryByRole('link', { name: /new article/i })).not.toBeInTheDocument();
    });

    it('logo links to home', () => {
      renderWithMantine(<Header />);
      const logo = screen.getByText('conduit').closest('a');
      expect(logo).toHaveAttribute('href', '/');
    });
  });

  describe('when authenticated', () => {
    const mockUserData = {
      username: 'testuser',
      email: 'test@example.com',
      image: 'https://example.com/avatar.jpg',
      bio: 'Test bio',
    };

    beforeEach(() => {
      mockIsAuthenticated.mockReturnValue(true);
      mockUser.mockReturnValue(mockUserData);
    });

    it('renders conduit logo', () => {
      renderWithMantine(<Header />);
      expect(screen.getByText('conduit')).toBeInTheDocument();
    });

    it('renders Home link', () => {
      renderWithMantine(<Header />);
      expect(screen.getByRole('link', { name: 'Home' })).toBeInTheDocument();
    });

    it('renders New Article link', () => {
      renderWithMantine(<Header />);
      expect(screen.getByRole('link', { name: /new article/i })).toBeInTheDocument();
    });

    it('renders username', () => {
      renderWithMantine(<Header />);
      expect(screen.getByText('testuser')).toBeInTheDocument();
    });

    it('renders user avatar', () => {
      renderWithMantine(<Header />);
      const avatar = screen.getByRole('img', { name: 'testuser' });
      expect(avatar).toBeInTheDocument();
    });

    it('does not render Sign in link', () => {
      renderWithMantine(<Header />);
      expect(screen.queryByRole('link', { name: 'Sign in' })).not.toBeInTheDocument();
    });

    it('does not render Sign up link', () => {
      renderWithMantine(<Header />);
      expect(screen.queryByRole('link', { name: 'Sign up' })).not.toBeInTheDocument();
    });

    it('has user menu button', () => {
      renderWithMantine(<Header />);
      // The user button should be present and clickable
      const userButton = screen.getByText('testuser').closest('button');
      expect(userButton).toBeInTheDocument();
    });

    it('renders without user image', () => {
      mockUser.mockReturnValue({ ...mockUserData, image: null });
      renderWithMantine(<Header />);
      expect(screen.getByText('testuser')).toBeInTheDocument();
    });
  });
});
