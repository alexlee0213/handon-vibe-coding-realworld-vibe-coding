import { Link } from '@tanstack/react-router';
import {
  Card,
  Text,
  Group,
  Badge,
  Stack,
} from '@mantine/core';
import type { Article } from '../../features/article';
import { ArticleMeta } from './ArticleMeta';
import { FavoriteButton } from './FavoriteButton';

interface ArticleCardProps {
  article: Article;
}

export function ArticleCard({ article }: ArticleCardProps) {
  return (
    <Card withBorder radius="md" p="md">
      <Stack gap="sm">
        <Group justify="space-between" align="flex-start">
          <ArticleMeta author={article.author} createdAt={article.createdAt} />
          <FavoriteButton
            slug={article.slug}
            favorited={article.favorited}
            favoritesCount={article.favoritesCount}
          />
        </Group>

        <Stack gap={4}>
          <Text
            component={Link}
            to="/article/$slug"
            params={{ slug: article.slug }}
            size="lg"
            fw={600}
            style={{ textDecoration: 'none', color: 'inherit' }}
          >
            {article.title}
          </Text>
          <Text size="sm" c="dimmed" lineClamp={2}>
            {article.description}
          </Text>
        </Stack>

        <Group justify="space-between" align="center">
          <Text
            component={Link}
            to="/article/$slug"
            params={{ slug: article.slug }}
            size="xs"
            c="dimmed"
            style={{ textDecoration: 'none' }}
          >
            Read more...
          </Text>
          <Group gap={4}>
            {(article.tagList || []).map((tag) => (
              <Badge
                key={tag}
                variant="outline"
                color="gray"
                size="xs"
                radius="sm"
              >
                {tag}
              </Badge>
            ))}
          </Group>
        </Group>
      </Stack>
    </Card>
  );
}
