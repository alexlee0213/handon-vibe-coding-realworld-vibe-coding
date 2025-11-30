import { Paper, Title, Group, Badge, Loader } from '@mantine/core';
import { useTags } from '../../features/article';

interface TagListProps {
  selectedTag?: string;
  onTagSelect: (tag: string | undefined) => void;
}

export function TagList({ selectedTag, onTagSelect }: TagListProps) {
  const { data, isLoading, error } = useTags();

  if (isLoading) {
    return (
      <Paper withBorder p="md" radius="md">
        <Loader size="sm" />
      </Paper>
    );
  }

  if (error || !data) {
    return null;
  }

  const tags = data.tags;

  if (tags.length === 0) {
    return null;
  }

  return (
    <Paper withBorder p="md" radius="md">
      <Title order={5} mb="sm">
        Popular Tags
      </Title>
      <Group gap={6}>
        {tags.map((tag) => (
          <Badge
            key={tag}
            variant={selectedTag === tag ? 'filled' : 'light'}
            color={selectedTag === tag ? 'brand' : 'gray'}
            size="md"
            radius="sm"
            style={{ cursor: 'pointer' }}
            onClick={() => onTagSelect(selectedTag === tag ? undefined : tag)}
          >
            {tag}
          </Badge>
        ))}
      </Group>
    </Paper>
  );
}
