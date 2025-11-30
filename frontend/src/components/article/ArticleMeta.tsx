import { Link } from '@tanstack/react-router';
import { Group, Avatar, Text, Stack } from '@mantine/core';
import type { Author } from '../../features/article';

interface ArticleMetaProps {
  author: Author;
  createdAt: string;
  size?: 'sm' | 'md';
}

export function ArticleMeta({ author, createdAt, size = 'md' }: ArticleMetaProps) {
  const avatarSize = size === 'sm' ? 'sm' : 'md';
  const textSize = size === 'sm' ? 'xs' : 'sm';

  const formattedDate = new Date(createdAt).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  return (
    <Group gap="xs">
      <Link to={`/profile/${author.username}`}>
        <Avatar
          src={author.image || undefined}
          alt={author.username}
          radius="xl"
          size={avatarSize}
        />
      </Link>
      <Stack gap={0}>
        <Text
          component={Link}
          to={`/profile/${author.username}`}
          size={textSize}
          fw={500}
          c="brand"
          style={{ textDecoration: 'none' }}
        >
          {author.username}
        </Text>
        <Text size="xs" c="dimmed">
          {formattedDate}
        </Text>
      </Stack>
    </Group>
  );
}
