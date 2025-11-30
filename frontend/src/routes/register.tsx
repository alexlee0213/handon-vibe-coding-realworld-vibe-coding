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
import { useRegister, useIsAuthenticated, registerSchema, type RegisterFormValues } from '../features/auth';

export const Route = createFileRoute('/register')({
  component: RegisterPage,
});

function RegisterPage() {
  const isAuthenticated = useIsAuthenticated();
  const registerMutation = useRegister();

  const form = useForm<RegisterFormValues>({
    mode: 'uncontrolled',
    initialValues: {
      username: '',
      email: '',
      password: '',
    },
    validate: zodResolver(registerSchema),
  });

  // Redirect if already authenticated
  if (isAuthenticated) {
    return <Navigate to="/" />;
  }

  const handleSubmit = (values: RegisterFormValues) => {
    registerMutation.mutate(values);
  };

  // Extract error message from API response
  const getErrorMessage = () => {
    const error = registerMutation.error as { response?: { json?: () => Promise<{ errors?: Record<string, string[]> }> } } | null;
    if (error?.response?.json) {
      return 'Registration failed. Please check your information.';
    }
    return 'An error occurred. Please try again.';
  };

  return (
    <Container size="xs" py="xl">
      <Paper radius="md" p="xl" withBorder>
        <Stack align="center" gap="xs" mb="lg">
          <Title order={2}>Sign up</Title>
          <Text c="dimmed" size="sm">
            <Anchor component={Link} to="/login">
              Have an account?
            </Anchor>
          </Text>
        </Stack>

        {registerMutation.error && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            color="red"
            mb="md"
          >
            {getErrorMessage()}
          </Alert>
        )}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Username"
              placeholder="Your username"
              required
              key={form.key('username')}
              {...form.getInputProps('username')}
            />

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
              loading={registerMutation.isPending}
            >
              Sign up
            </Button>
          </Stack>
        </form>
      </Paper>
    </Container>
  );
}
