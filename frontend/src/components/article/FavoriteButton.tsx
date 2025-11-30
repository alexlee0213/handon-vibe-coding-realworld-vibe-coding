import { useNavigate } from '@tanstack/react-router';
import { ActionIcon, Group, Text, Tooltip } from '@mantine/core';
import { IconHeart, IconHeartFilled } from '@tabler/icons-react';
import { useFavoriteArticle, useUnfavoriteArticle } from '../../features/article';
import { useIsAuthenticated } from '../../features/auth';

interface FavoriteButtonProps {
  slug: string;
  favorited: boolean;
  favoritesCount: number;
  showCount?: boolean;
  size?: 'xs' | 'sm' | 'md' | 'lg';
}

export function FavoriteButton({
  slug,
  favorited,
  favoritesCount,
  showCount = true,
  size = 'sm',
}: FavoriteButtonProps) {
  const navigate = useNavigate();
  const isAuthenticated = useIsAuthenticated();
  const favoriteArticle = useFavoriteArticle();
  const unfavoriteArticle = useUnfavoriteArticle();

  const isLoading = favoriteArticle.isPending || unfavoriteArticle.isPending;

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (!isAuthenticated) {
      navigate({ to: '/login' });
      return;
    }

    if (favorited) {
      unfavoriteArticle.mutate(slug);
    } else {
      favoriteArticle.mutate(slug);
    }
  };

  const iconSize = size === 'xs' ? 12 : size === 'sm' ? 14 : size === 'md' ? 16 : 20;

  return (
    <Tooltip label={favorited ? 'Unfavorite article' : 'Favorite article'} withArrow>
      <Group gap={4} wrap="nowrap">
        <ActionIcon
          variant={favorited ? 'filled' : 'outline'}
          color="brand"
          size={size}
          onClick={handleClick}
          loading={isLoading}
          aria-label={favorited ? 'Unfavorite article' : 'Favorite article'}
        >
          {favorited ? (
            <IconHeartFilled size={iconSize} />
          ) : (
            <IconHeart size={iconSize} />
          )}
        </ActionIcon>
        {showCount && (
          <Text size="xs" c="dimmed">
            {favoritesCount}
          </Text>
        )}
      </Group>
    </Tooltip>
  );
}
