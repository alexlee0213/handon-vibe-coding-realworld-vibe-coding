import { Stack, Text, Loader, Center, Alert } from '@mantine/core';
import { IconAlertCircle } from '@tabler/icons-react';
import type { Article } from '../../features/article';
import { ArticleCard } from './ArticleCard';

interface ArticleListProps {
  articles: Article[] | undefined;
  isLoading: boolean;
  error: Error | null;
  emptyMessage?: string;
}

export function ArticleList({
  articles,
  isLoading,
  error,
  emptyMessage = 'No articles are here... yet.',
}: ArticleListProps) {
  if (isLoading) {
    return (
      <Center py="xl">
        <Loader size="lg" />
      </Center>
    );
  }

  if (error) {
    return (
      <Alert
        icon={<IconAlertCircle size={16} />}
        title="Error loading articles"
        color="red"
        variant="light"
      >
        {error.message || 'Failed to load articles. Please try again.'}
      </Alert>
    );
  }

  if (!articles || articles.length === 0) {
    return (
      <Text c="dimmed" ta="center" py="xl">
        {emptyMessage}
      </Text>
    );
  }

  return (
    <Stack gap="md">
      {articles.map((article) => (
        <ArticleCard key={article.slug} article={article} />
      ))}
    </Stack>
  );
}
