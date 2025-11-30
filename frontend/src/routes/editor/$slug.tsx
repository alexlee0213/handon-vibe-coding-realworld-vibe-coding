import { useEffect } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import {
  Container,
  Title,
  Paper,
  TextInput,
  Textarea,
  Button,
  Stack,
  Alert,
  Loader,
  Center,
} from '@mantine/core';
import { useForm, zodResolver } from '@mantine/form';
import { IconAlertCircle } from '@tabler/icons-react';
import {
  useArticle,
  useUpdateArticle,
  updateArticleSchema,
  type UpdateArticleFormValues
} from '../../features/article';
import { AuthGuard } from '../../components/auth/AuthGuard';

export const Route = createFileRoute('/editor/$slug')({
  component: EditArticlePage,
});

function EditArticlePage() {
  return (
    <AuthGuard>
      <EditArticleForm />
    </AuthGuard>
  );
}

function EditArticleForm() {
  const { slug } = Route.useParams();
  const { data, isLoading, error } = useArticle(slug);
  const updateArticle = useUpdateArticle(slug);

  const form = useForm<UpdateArticleFormValues>({
    mode: 'uncontrolled',
    initialValues: {
      title: '',
      description: '',
      body: '',
    },
    validate: zodResolver(updateArticleSchema),
  });

  // Populate form when article loads
  useEffect(() => {
    if (data?.article) {
      form.setValues({
        title: data.article.title,
        description: data.article.description,
        body: data.article.body,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  const handleSubmit = (values: UpdateArticleFormValues) => {
    // Only include fields that have changed
    const updates: UpdateArticleFormValues = {};
    if (values.title && values.title !== data?.article.title) {
      updates.title = values.title;
    }
    if (values.description && values.description !== data?.article.description) {
      updates.description = values.description;
    }
    if (values.body && values.body !== data?.article.body) {
      updates.body = values.body;
    }

    // Only submit if there are changes
    if (Object.keys(updates).length > 0) {
      updateArticle.mutate(updates);
    }
  };

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

  return (
    <Container size="md" py="xl">
      <Title order={2} mb="lg" ta="center">
        Edit Article
      </Title>

      <Paper withBorder p="xl" radius="md">
        {updateArticle.error && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            title="Error updating article"
            color="red"
            variant="light"
            mb="md"
          >
            {updateArticle.error.message || 'Failed to update article. Please try again.'}
          </Alert>
        )}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Article Title"
              placeholder="How to train your dragon"
              size="lg"
              key={form.key('title')}
              {...form.getInputProps('title')}
            />

            <TextInput
              label="What's this article about?"
              placeholder="Ever wonder how?"
              key={form.key('description')}
              {...form.getInputProps('description')}
            />

            <Textarea
              label="Write your article (in markdown)"
              placeholder="Share your knowledge..."
              minRows={10}
              autosize
              key={form.key('body')}
              {...form.getInputProps('body')}
            />

            <Button
              type="submit"
              size="lg"
              loading={updateArticle.isPending}
            >
              Update Article
            </Button>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
}
