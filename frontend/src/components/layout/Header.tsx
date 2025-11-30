import { Link } from '@tanstack/react-router';
import {
  Group,
  Text,
  Button,
  Avatar,
  Menu,
  UnstyledButton,
} from '@mantine/core';
import {
  IconSettings,
  IconLogout,
  IconPencil,
  IconUser,
} from '@tabler/icons-react';
import { useIsAuthenticated, useUser, useLogout } from '../../features/auth';

export function Header() {
  const isAuthenticated = useIsAuthenticated();
  const user = useUser();
  const logout = useLogout();

  return (
    <Group justify="space-between" h="100%" px="md">
      <Link to="/" style={{ textDecoration: 'none' }}>
        <Text
          size="xl"
          fw={700}
          c="brand"
          style={{ fontFamily: 'Titillium Web, sans-serif' }}
        >
          conduit
        </Text>
      </Link>

      <Group gap="sm">
        <Button
          component={Link}
          to="/"
          variant="subtle"
          color="gray"
        >
          Home
        </Button>

        {isAuthenticated && user ? (
          <>
            <Button
              component={Link}
              to="/editor"
              variant="subtle"
              color="gray"
              leftSection={<IconPencil size={16} />}
            >
              New Article
            </Button>

            <Menu shadow="md" width={200}>
              <Menu.Target>
                <UnstyledButton>
                  <Group gap="xs">
                    <Avatar
                      src={user.image || undefined}
                      alt={user.username}
                      radius="xl"
                      size="sm"
                    />
                    <Text size="sm">{user.username}</Text>
                  </Group>
                </UnstyledButton>
              </Menu.Target>

              <Menu.Dropdown>
                <Menu.Item
                  component={Link}
                  to={`/profile/${user.username}`}
                  leftSection={<IconUser size={14} />}
                >
                  Profile
                </Menu.Item>
                <Menu.Item
                  component={Link}
                  to="/settings"
                  leftSection={<IconSettings size={14} />}
                >
                  Settings
                </Menu.Item>
                <Menu.Divider />
                <Menu.Item
                  color="red"
                  leftSection={<IconLogout size={14} />}
                  onClick={logout}
                >
                  Log out
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </>
        ) : (
          <>
            <Button
              component={Link}
              to="/login"
              variant="subtle"
              color="gray"
            >
              Sign in
            </Button>
            <Button
              component={Link}
              to="/register"
              variant="subtle"
              color="gray"
            >
              Sign up
            </Button>
          </>
        )}
      </Group>
    </Group>
  );
}
