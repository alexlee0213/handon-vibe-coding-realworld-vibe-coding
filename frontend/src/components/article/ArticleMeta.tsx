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
      <Link to="/profile/$username" params={{ username: author.username }}>
        <Avatar
          src={author.image || undefined}
          alt={author.username}
          radius="xl"
          size={avatarSize}
        />
      </Link>
      <Stack gap={0}>
        <Link
          to="/profile/$username"
          params={{ username: author.username }}
          style={{ textDecoration: 'none' }}
        >
          <Text size={textSize} fw={500} c="brand">
            {author.username}
          </Text>
        </Link>
        <Text size="xs" c="dimmed">
          {formattedDate}
        </Text>
      </Stack>
    </Group>
  );
}
