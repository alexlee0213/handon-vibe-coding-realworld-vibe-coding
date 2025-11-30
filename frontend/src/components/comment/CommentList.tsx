import { Stack, Text, Loader, Center, Alert } from '@mantine/core';
import { IconAlertCircle } from '@tabler/icons-react';
import type { Comment } from '../../features/comment';
import { CommentCard } from './CommentCard';

interface CommentListProps {
  comments?: Comment[];
  isLoading?: boolean;
  error?: Error | null;
  currentUsername?: string;
  onDeleteComment?: (commentId: number) => void;
  deletingCommentId?: number | null;
}

export function CommentList({
  comments,
  isLoading,
  error,
  currentUsername,
  onDeleteComment,
  deletingCommentId,
}: CommentListProps) {
  if (isLoading) {
    return (
      <Center py="xl">
        <Loader size="md" />
      </Center>
    );
  }

  if (error) {
    return (
      <Alert
        icon={<IconAlertCircle size={16} />}
        title="Error loading comments"
        color="red"
        variant="light"
      >
        {error.message || 'Failed to load comments'}
      </Alert>
    );
  }

  if (!comments || comments.length === 0) {
    return (
      <Text c="dimmed" ta="center" py="lg">
        No comments yet. Be the first to comment!
      </Text>
    );
  }

  return (
    <Stack gap="md">
      {comments.map((comment) => (
        <CommentCard
          key={comment.id}
          comment={comment}
          isAuthor={currentUsername === comment.author.username}
          onDelete={
            onDeleteComment ? () => onDeleteComment(comment.id) : undefined
          }
          isDeleting={deletingCommentId === comment.id}
        />
      ))}
    </Stack>
  );
}
