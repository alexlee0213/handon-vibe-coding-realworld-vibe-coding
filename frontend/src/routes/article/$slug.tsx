import { createFileRoute, Link } from '@tanstack/react-router';
import {
  Container,
  Title,
  Text,
  Stack,
  Group,
  Button,
  Badge,
  Paper,
  Divider,
  Loader,
  Center,
  Alert,
  Box,
} from '@mantine/core';
import { IconEdit, IconTrash, IconAlertCircle } from '@tabler/icons-react';
import { useArticle, useDeleteArticle } from '../../features/article';
import { useUser, useIsAuthenticated } from '../../features/auth';
import { ArticleMeta } from '../../components/article';

export const Route = createFileRoute('/article/$slug')({
  component: ArticlePage,
});

function ArticlePage() {
  const { slug } = Route.useParams();
  const { data, isLoading, error } = useArticle(slug);
  const deleteArticle = useDeleteArticle();
  const user = useUser();
  const isAuthenticated = useIsAuthenticated();

  if (isLoading) {
    return (
      <Center py="xl">
        <Loader size="lg" />
      </Center>
    );
  }

  if (error || !data) {
    return (
      <Container size="md" py="xl">
        <Alert
          icon={<IconAlertCircle size={16} />}
          title="Error loading article"
          color="red"
          variant="light"
        >
          {error?.message || 'Article not found'}
        </Alert>
      </Container>
    );
  }

  const article = data.article;
  const isAuthor = isAuthenticated && user?.username === article.author.username;

  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete this article?')) {
      deleteArticle.mutate(slug);
    }
  };

  return (
    <>
      {/* Article Banner */}
      <Box bg="dark.7" py="xl">
        <Container size="lg">
          <Stack gap="md">
            <Title order={1} c="white">
              {article.title}
            </Title>
            <Group justify="space-between" align="flex-start">
              <ArticleMeta author={article.author} createdAt={article.createdAt} />
              {isAuthor && (
                <Group gap="xs">
                  <Button
                    component={Link}
                    to="/editor/$slug"
                    params={{ slug: article.slug }}
                    variant="outline"
                    color="gray"
                    size="xs"
                    leftSection={<IconEdit size={14} />}
                  >
                    Edit Article
                  </Button>
                  <Button
                    variant="outline"
                    color="red"
                    size="xs"
                    leftSection={<IconTrash size={14} />}
                    onClick={handleDelete}
                    loading={deleteArticle.isPending}
                  >
                    Delete Article
                  </Button>
                </Group>
              )}
            </Group>
          </Stack>
        </Container>
      </Box>

      {/* Article Content */}
      <Container size="lg" py="xl">
        <Paper p="xl" radius="md">
          <Text style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
            {article.body}
          </Text>

          {article.tagList && article.tagList.length > 0 && (
            <>
              <Divider my="lg" />
              <Group gap={6}>
                {article.tagList.map((tag) => (
                  <Badge key={tag} variant="outline" color="gray" size="sm">
                    {tag}
                  </Badge>
                ))}
              </Group>
            </>
          )}
        </Paper>

        <Divider my="xl" />

        {/* Author info at bottom */}
        <Paper p="lg" radius="md" withBorder>
          <Group justify="center">
            <ArticleMeta author={article.author} createdAt={article.createdAt} size="md" />
          </Group>
        </Paper>
      </Container>
    </>
  );
}
