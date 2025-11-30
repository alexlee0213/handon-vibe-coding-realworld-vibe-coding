import { Link } from '@tanstack/react-router';
import {
  Paper,
  Textarea,
  Group,
  Button,
  Avatar,
  Stack,
  Text,
} from '@mantine/core';
import { useForm, zodResolver } from '@mantine/form';
import { createCommentSchema, type CreateCommentFormValues } from '../../features/comment';
import { useUser, useIsAuthenticated } from '../../features/auth';

interface CommentFormProps {
  onSubmit: (data: CreateCommentFormValues) => void;
  isSubmitting?: boolean;
}

export function CommentForm({ onSubmit, isSubmitting }: CommentFormProps) {
  const user = useUser();
  const isAuthenticated = useIsAuthenticated();

  const form = useForm<CreateCommentFormValues>({
    validate: zodResolver(createCommentSchema),
    initialValues: {
      body: '',
    },
  });

  const handleSubmit = form.onSubmit((values) => {
    onSubmit(values);
    form.reset();
  });

  if (!isAuthenticated) {
    return (
      <Paper withBorder p="lg" radius="md" ta="center">
        <Text size="sm" c="dimmed">
          <Text component={Link} to="/login" c="brand" inherit>
            Sign in
          </Text>{' '}
          or{' '}
          <Text component={Link} to="/register" c="brand" inherit>
            sign up
          </Text>{' '}
          to add comments on this article.
        </Text>
      </Paper>
    );
  }

  return (
    <Paper withBorder radius="md" p={0} component="form" onSubmit={handleSubmit}>
      <Textarea
        placeholder="Write a comment..."
        minRows={3}
        styles={{
          input: {
            border: 'none',
            borderRadius: 'var(--mantine-radius-md) var(--mantine-radius-md) 0 0',
          },
        }}
        {...form.getInputProps('body')}
      />
      <Group
        justify="flex-end"
        p="sm"
        bg="gray.0"
        style={{ borderTop: '1px solid var(--mantine-color-gray-2)' }}
      >
        <Stack gap={4} align="flex-start" style={{ flex: 1 }}>
          {user && (
            <Group gap="xs">
              <Avatar
                src={user.image || undefined}
                alt={user.username}
                radius="xl"
                size="sm"
              />
              <Text size="sm" c="dimmed">
                {user.username}
              </Text>
            </Group>
          )}
        </Stack>
        <Button type="submit" size="sm" loading={isSubmitting}>
          Post Comment
        </Button>
      </Group>
    </Paper>
  );
}
