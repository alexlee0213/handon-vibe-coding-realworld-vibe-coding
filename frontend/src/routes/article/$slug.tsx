import { useState } from 'react';
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
import { useComments, useCreateComment, useDeleteComment } from '../../features/comment';
import { ArticleMeta } from '../../components/article';
import { CommentForm, CommentList } from '../../components/comment';

export const Route = createFileRoute('/article/$slug')({
  component: ArticlePage,
});

function ArticlePage() {
  const { slug } = Route.useParams();
  const { data, isLoading, error } = useArticle(slug);
  const deleteArticle = useDeleteArticle();
  const user = useUser();
  const isAuthenticated = useIsAuthenticated();

  // Comment state and hooks
  const [deletingCommentId, setDeletingCommentId] = useState<number | null>(null);
  const commentsQuery = useComments(slug);
  const createComment = useCreateComment(slug);
  const deleteComment = useDeleteComment(slug);

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

  const handleCreateComment = (data: { body: string }) => {
    createComment.mutate(data);
  };

  const handleDeleteComment = (commentId: number) => {
    setDeletingCommentId(commentId);
    deleteComment.mutate(commentId, {
      onSettled: () => setDeletingCommentId(null),
    });
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

        <Divider my="xl" />

        {/* Comments Section */}
        <Stack gap="lg">
          <Title order={3}>Comments</Title>

          <CommentForm
            onSubmit={handleCreateComment}
            isSubmitting={createComment.isPending}
          />

          <CommentList
            comments={commentsQuery.data?.comments}
            isLoading={commentsQuery.isLoading}
            error={commentsQuery.error}
            currentUsername={user?.username}
            onDeleteComment={handleDeleteComment}
            deletingCommentId={deletingCommentId}
          />
        </Stack>
      </Container>
    </>
  );
}
