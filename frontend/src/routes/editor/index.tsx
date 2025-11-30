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
} from '@mantine/core';
import { useForm, zodResolver } from '@mantine/form';
import { IconAlertCircle } from '@tabler/icons-react';
import { useCreateArticle, createArticleSchema, type CreateArticleFormValues } from '../../features/article';
import { AuthGuard } from '../../components/auth/AuthGuard';

export const Route = createFileRoute('/editor/')({
  component: NewArticlePage,
});

function NewArticlePage() {
  return (
    <AuthGuard>
      <NewArticleForm />
    </AuthGuard>
  );
}

function NewArticleForm() {
  const createArticle = useCreateArticle();

  const form = useForm<CreateArticleFormValues>({
    mode: 'uncontrolled',
    initialValues: {
      title: '',
      description: '',
      body: '',
      tagList: '',
    },
    validate: zodResolver(createArticleSchema),
  });

  const handleSubmit = (values: CreateArticleFormValues) => {
    // Parse through zod schema to transform tagList string to array
    const parsed = createArticleSchema.parse(values);
    createArticle.mutate({
      title: parsed.title,
      description: parsed.description,
      body: parsed.body,
      tagList: parsed.tagList,
    });
  };

  return (
    <Container size="md" py="xl">
      <Title order={2} mb="lg" ta="center">
        New Article
      </Title>

      <Paper withBorder p="xl" radius="md">
        {createArticle.error && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            title="Error creating article"
            color="red"
            variant="light"
            mb="md"
          >
            {createArticle.error.message || 'Failed to create article. Please try again.'}
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

            <TextInput
              label="Enter tags"
              placeholder="Enter tags separated by commas"
              description="e.g., programming, javascript, tutorial"
              key={form.key('tagList')}
              {...form.getInputProps('tagList')}
            />

            <Button
              type="submit"
              size="lg"
              loading={createArticle.isPending}
            >
              Publish Article
            </Button>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
}
