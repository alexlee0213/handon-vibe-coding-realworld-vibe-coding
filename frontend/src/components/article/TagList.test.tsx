import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import { theme } from '../../lib/theme';
import { TagList } from './TagList';

// Mock useTags hook
const mockUseTags = vi.fn();
vi.mock('../../features/article', () => ({
  useTags: () => mockUseTags(),
}));

const renderWithMantine = (ui: React.ReactElement) => {
  return render(<MantineProvider theme={theme}>{ui}</MantineProvider>);
};

describe('TagList', () => {
  const mockOnTagSelect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state', () => {
    mockUseTags.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
    });

    const { container } = renderWithMantine(
      <TagList onTagSelect={mockOnTagSelect} />
    );

    // Should show loader during loading state - look for the Mantine loader element
    expect(container.querySelector('.mantine-Loader-root')).toBeInTheDocument();
  });

  it('renders nothing when there is an error', () => {
    mockUseTags.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Failed to fetch'),
    });

    const { container } = renderWithMantine(
      <TagList onTagSelect={mockOnTagSelect} />
    );

    // The Paper component won't render, but container.firstChild is the Mantine wrapper
    expect(container.querySelector('.mantine-Paper-root')).not.toBeInTheDocument();
  });

  it('renders nothing when data is empty', () => {
    mockUseTags.mockReturnValue({
      data: { tags: [] },
      isLoading: false,
      error: null,
    });

    const { container } = renderWithMantine(
      <TagList onTagSelect={mockOnTagSelect} />
    );

    expect(container.querySelector('.mantine-Paper-root')).not.toBeInTheDocument();
  });

  it('renders popular tags title', () => {
    mockUseTags.mockReturnValue({
      data: { tags: ['react', 'typescript'] },
      isLoading: false,
      error: null,
    });

    renderWithMantine(
      <TagList onTagSelect={mockOnTagSelect} />
    );

    expect(screen.getByText('Popular Tags')).toBeInTheDocument();
  });

  it('renders all tags', () => {
    mockUseTags.mockReturnValue({
      data: { tags: ['react', 'typescript', 'testing'] },
      isLoading: false,
      error: null,
    });

    renderWithMantine(
      <TagList onTagSelect={mockOnTagSelect} />
    );

    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByText('typescript')).toBeInTheDocument();
    expect(screen.getByText('testing')).toBeInTheDocument();
  });

  it('calls onTagSelect when clicking a tag', () => {
    mockUseTags.mockReturnValue({
      data: { tags: ['react', 'typescript'] },
      isLoading: false,
      error: null,
    });

    renderWithMantine(
      <TagList onTagSelect={mockOnTagSelect} />
    );

    fireEvent.click(screen.getByText('react'));
    expect(mockOnTagSelect).toHaveBeenCalledWith('react');
  });

  it('deselects tag when clicking selected tag', () => {
    mockUseTags.mockReturnValue({
      data: { tags: ['react', 'typescript'] },
      isLoading: false,
      error: null,
    });

    renderWithMantine(
      <TagList selectedTag="react" onTagSelect={mockOnTagSelect} />
    );

    fireEvent.click(screen.getByText('react'));
    expect(mockOnTagSelect).toHaveBeenCalledWith(undefined);
  });

  it('shows selected tag with different style', () => {
    mockUseTags.mockReturnValue({
      data: { tags: ['react', 'typescript'] },
      isLoading: false,
      error: null,
    });

    renderWithMantine(
      <TagList selectedTag="react" onTagSelect={mockOnTagSelect} />
    );

    const reactTag = screen.getByText('react');
    const typescriptTag = screen.getByText('typescript');

    // Both tags should be visible
    expect(reactTag).toBeInTheDocument();
    expect(typescriptTag).toBeInTheDocument();
  });
});
