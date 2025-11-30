import { createFileRoute } from '@tanstack/react-router';
import {
  Container,
  Title,
  TextInput,
  PasswordInput,
  Textarea,
  Button,
  Paper,
  Stack,
  Alert,
} from '@mantine/core';
import { useForm, zodResolver } from '@mantine/form';
import { IconAlertCircle, IconCheck } from '@tabler/icons-react';
import { AuthGuard } from '../components/auth';
import {
  useUser,
  useUpdateUser,
  useLogout,
  updateUserSchema,
  type UpdateUserFormValues,
} from '../features/auth';

export const Route = createFileRoute('/settings')({
  component: SettingsPage,
});

function SettingsPage() {
  return (
    <AuthGuard>
      <SettingsForm />
    </AuthGuard>
  );
}

function SettingsForm() {
  const user = useUser();
  const updateMutation = useUpdateUser();
  const logout = useLogout();

  const form = useForm<UpdateUserFormValues>({
    mode: 'uncontrolled',
    initialValues: {
      image: user?.image || '',
      username: user?.username || '',
      bio: user?.bio || '',
      email: user?.email || '',
      password: '',
    },
    validate: zodResolver(updateUserSchema),
  });

  const handleSubmit = (values: UpdateUserFormValues) => {
    // Only send non-empty values
    const updateData: UpdateUserFormValues = {};
    if (values.email && values.email !== user?.email) {
      updateData.email = values.email;
    }
    if (values.username && values.username !== user?.username) {
      updateData.username = values.username;
    }
    if (values.bio !== undefined && values.bio !== user?.bio) {
      updateData.bio = values.bio;
    }
    if (values.image && values.image !== user?.image) {
      updateData.image = values.image;
    }
    if (values.password) {
      updateData.password = values.password;
    }

    if (Object.keys(updateData).length > 0) {
      updateMutation.mutate(updateData);
    }
  };

  return (
    <Container size="sm" py="xl">
      <Paper radius="md" p="xl" withBorder>
        <Title order={2} ta="center" mb="lg">
          Your Settings
        </Title>

        {updateMutation.isSuccess && (
          <Alert
            icon={<IconCheck size={16} />}
            color="green"
            mb="md"
          >
            Settings updated successfully
          </Alert>
        )}

        {updateMutation.error && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            color="red"
            mb="md"
          >
            Failed to update settings
          </Alert>
        )}

        <form onSubmit={form.onSubmit(handleSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Profile Picture URL"
              placeholder="URL of profile picture"
              key={form.key('image')}
              {...form.getInputProps('image')}
            />

            <TextInput
              label="Username"
              placeholder="Your username"
              key={form.key('username')}
              {...form.getInputProps('username')}
            />

            <Textarea
              label="Bio"
              placeholder="Short bio about you"
              rows={4}
              key={form.key('bio')}
              {...form.getInputProps('bio')}
            />

            <TextInput
              label="Email"
              placeholder="your@email.com"
              key={form.key('email')}
              {...form.getInputProps('email')}
            />

            <PasswordInput
              label="New Password"
              placeholder="Leave blank to keep current"
              key={form.key('password')}
              {...form.getInputProps('password')}
            />

            <Button
              type="submit"
              fullWidth
              loading={updateMutation.isPending}
            >
              Update Settings
            </Button>
          </Stack>
        </form>

        <Button
          variant="outline"
          color="red"
          fullWidth
          mt="xl"
          onClick={logout}
        >
          Or click here to logout
        </Button>
      </Paper>
    </Container>
  );
}
