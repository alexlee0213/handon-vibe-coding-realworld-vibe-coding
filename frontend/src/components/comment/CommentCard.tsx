import { Link } from '@tanstack/react-router';
import {
  Card,
  Text,
  Group,
  Avatar,
  Stack,
  ActionIcon,
} from '@mantine/core';
import { IconTrash } from '@tabler/icons-react';
import type { Comment } from '../../features/comment';

interface CommentCardProps {
  comment: Comment;
  isAuthor: boolean;
  onDelete?: () => void;
  isDeleting?: boolean;
}

export function CommentCard({
  comment,
  isAuthor,
  onDelete,
  isDeleting,
}: CommentCardProps) {
  const formattedDate = new Date(comment.createdAt).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  return (
    <Card withBorder radius="md" p={0}>
      <Stack gap={0}>
        {/* Comment body */}
        <Text p="md" style={{ whiteSpace: 'pre-wrap' }}>
          {comment.body}
        </Text>

        {/* Comment footer with author info and delete button */}
        <Group
          justify="space-between"
          p="sm"
          bg="gray.0"
          style={{ borderTop: '1px solid var(--mantine-color-gray-2)' }}
        >
          <Group gap="xs">
            <Link to={`/profile/${comment.author.username}`}>
              <Avatar
                src={comment.author.image || undefined}
                alt={comment.author.username}
                radius="xl"
                size="sm"
              />
            </Link>
            <Text
              component={Link}
              to={`/profile/${comment.author.username}`}
              size="sm"
              c="brand"
              style={{ textDecoration: 'none' }}
            >
              {comment.author.username}
            </Text>
            <Text size="xs" c="dimmed">
              {formattedDate}
            </Text>
          </Group>

          {isAuthor && onDelete && (
            <ActionIcon
              variant="subtle"
              color="red"
              size="sm"
              onClick={onDelete}
              loading={isDeleting}
              aria-label="Delete comment"
            >
              <IconTrash size={16} />
            </ActionIcon>
          )}
        </Group>
      </Stack>
    </Card>
  );
}
