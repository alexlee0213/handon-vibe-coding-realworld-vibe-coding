import { createFileRoute, Link, Navigate } from '@tanstack/react-router';
import {
  Container,
  Title,
  Text,
  TextInput,
  PasswordInput,
  Button,
  Paper,
  Stack,
  Anchor,
  Alert,
} from '@mantine/core';
import { useForm, zodResolver } from '@mantine/form';
import { IconAlertCircle } from '@tabler/icons-react';
import { useLogin, useIsAuthenticated, loginSchema, type LoginFormValues } from '../features/auth';

export const Route = createFileRoute('/login')({
  component: LoginPage,
});

function LoginPage() {
  const isAuthenticated = useIsAuthenticated();
  const loginMutation = useLogin();

  const form = useForm<LoginFormValues>({
    mode: 'uncontrolled',
    initialValues: {
      email: '',
      password: '',
    },
    validate: zodResolver(loginSchema),
  });

  // Redirect if already authenticated
  if (isAuthenticated) {
    return <Navigate to="/" />;
  }

  const handleSubmit = (values: LoginFormValues) => {
    loginMutation.mutate(values);
  };

  return (
    <Container size="xs" py="xl">
      <Paper radius="md" p="xl" withBorder>
        <Stack align="center" gap="xs" mb="lg">
          <Title order={2}>Sign in</Title>
          <Text c="dimmed" size="sm">
            <Anchor component={Link} to="/register">
              Need an account?
            </Anchor>
          </Text>
        </Stack>

        {loginMutation.error && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            color="red"
            mb="md"
          >
            Invalid email or password
          </Alert>
        )}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Email"
              placeholder="your@email.com"
              required
              key={form.key('email')}
              {...form.getInputProps('email')}
            />

            <PasswordInput
              label="Password"
              placeholder="Your password"
              required
              key={form.key('password')}
              {...form.getInputProps('password')}
            />

            <Button
              type="submit"
              fullWidth
              loading={loginMutation.isPending}
            >
              Sign in
            </Button>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
}
