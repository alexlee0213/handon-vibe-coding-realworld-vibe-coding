import {
  Paper,
  Avatar,
  Text,
  Button,
  Group,
  Stack,
  Skeleton,
} from '@mantine/core';
import { IconPlus, IconCheck, IconSettings } from '@tabler/icons-react';
import { Link } from '@tanstack/react-router';
import type { Profile } from '../../features/profile';
import { useToggleFollow } from '../../features/profile';
import { useUser } from '../../features/auth';

interface ProfileCardProps {
  profile?: Profile;
  isLoading?: boolean;
}

export function ProfileCard({ profile, isLoading }: ProfileCardProps) {
  const currentUser = useUser();
  const { toggleFollow, isPending } = useToggleFollow();

  const isOwnProfile = currentUser?.username === profile?.username;

  if (isLoading || !profile) {
    return <ProfileCardSkeleton />;
  }

  return (
    <Paper
      p="xl"
      radius="md"
      withBorder
      style={{ textAlign: 'center' }}
    >
      <Stack align="center" gap="md">
        <Avatar
          src={profile.image || undefined}
          alt={profile.username}
          size={120}
          radius={120}
        />

        <Text size="xl" fw={700}>
          {profile.username}
        </Text>

        {profile.bio && (
          <Text c="dimmed" size="sm" maw={400}>
            {profile.bio}
          </Text>
        )}

        <Group gap="sm">
          {isOwnProfile ? (
            <Button
              component={Link}
              to="/settings"
              variant="outline"
              leftSection={<IconSettings size={16} />}
            >
              Edit Profile Settings
            </Button>
          ) : currentUser ? (
            <Button
              variant={profile.following ? 'filled' : 'outline'}
              color={profile.following ? 'gray' : 'green'}
              leftSection={
                profile.following ? (
                  <IconCheck size={16} />
                ) : (
                  <IconPlus size={16} />
                )
              }
              onClick={() => toggleFollow(profile.username, profile.following)}
              loading={isPending}
            >
              {profile.following ? 'Following' : 'Follow'} {profile.username}
            </Button>
          ) : null}
        </Group>
      </Stack>
    </Paper>
  );
}

function ProfileCardSkeleton() {
  return (
    <Paper
      p="xl"
      radius="md"
      withBorder
      style={{ textAlign: 'center' }}
    >
      <Stack align="center" gap="md">
        <Skeleton height={120} width={120} circle />
        <Skeleton height={24} width={150} />
        <Skeleton height={40} width={300} />
        <Skeleton height={36} width={180} />
      </Stack>
    </Paper>
  );
}
