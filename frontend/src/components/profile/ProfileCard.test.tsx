import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '../../test/utils';
import { ProfileCard } from './ProfileCard';
import type { Profile } from '../../features/profile';

// Mock the profile hooks
const mockToggleFollow = vi.fn();
vi.mock('../../features/profile', () => ({
  useToggleFollow: () => ({
    toggleFollow: mockToggleFollow,
    isPending: false,
  }),
}));

// Mock the auth hooks
const mockUseUser = vi.fn();
vi.mock('../../features/auth', () => ({
  useUser: () => mockUseUser(),
}));

const mockProfile: Profile = {
  username: 'testuser',
  bio: 'Test bio description',
  image: 'https://example.com/avatar.png',
  following: false,
};

describe('ProfileCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseUser.mockReturnValue(null);
  });

  it('renders profile username', () => {
    render(<ProfileCard profile={mockProfile} />);
    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  it('renders profile bio', () => {
    render(<ProfileCard profile={mockProfile} />);
    expect(screen.getByText('Test bio description')).toBeInTheDocument();
  });

  it('renders profile avatar', () => {
    render(<ProfileCard profile={mockProfile} />);
    const avatar = screen.getByAltText('testuser');
    expect(avatar).toBeInTheDocument();
  });

  it('shows skeleton when loading', () => {
    render(<ProfileCard isLoading={true} />);
    // Should not show username when loading
    expect(screen.queryByText('testuser')).not.toBeInTheDocument();
  });

  it('shows skeleton when profile is undefined', () => {
    render(<ProfileCard profile={undefined} />);
    // Should not crash and show skeleton
    expect(screen.queryByText('testuser')).not.toBeInTheDocument();
  });

  it('shows Edit Profile Settings button for own profile', () => {
    mockUseUser.mockReturnValue({ username: 'testuser' });
    render(<ProfileCard profile={mockProfile} />);
    expect(screen.getByText('Edit Profile Settings')).toBeInTheDocument();
  });

  it('hides follow button for own profile', () => {
    mockUseUser.mockReturnValue({ username: 'testuser' });
    render(<ProfileCard profile={mockProfile} />);
    expect(screen.queryByText(/Follow testuser/)).not.toBeInTheDocument();
  });

  it('shows Follow button for other user when logged in', () => {
    mockUseUser.mockReturnValue({ username: 'otheruser' });
    render(<ProfileCard profile={mockProfile} />);
    expect(screen.getByText('Follow testuser')).toBeInTheDocument();
  });

  it('shows Following button when already following', () => {
    mockUseUser.mockReturnValue({ username: 'otheruser' });
    const followingProfile = { ...mockProfile, following: true };
    render(<ProfileCard profile={followingProfile} />);
    expect(screen.getByText('Following testuser')).toBeInTheDocument();
  });

  it('hides follow button when not logged in', () => {
    mockUseUser.mockReturnValue(null);
    render(<ProfileCard profile={mockProfile} />);
    expect(screen.queryByText(/Follow testuser/)).not.toBeInTheDocument();
    expect(screen.queryByText('Edit Profile Settings')).not.toBeInTheDocument();
  });

  it('handles empty bio', () => {
    const profileWithNoBio = { ...mockProfile, bio: '' };
    render(<ProfileCard profile={profileWithNoBio} />);
    expect(screen.getByText('testuser')).toBeInTheDocument();
    // Bio should not be rendered
    expect(screen.queryByText('Test bio description')).not.toBeInTheDocument();
  });
});
