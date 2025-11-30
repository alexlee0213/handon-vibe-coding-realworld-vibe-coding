import { createFileRoute } from '@tanstack/react-router';
import { Container, Title, Text, Stack } from '@mantine/core';

export const Route = createFileRoute('/')({
  component: HomePage,
});

function HomePage() {
  return (
    <Container size="lg" py="xl">
      <Stack align="center" gap="md">
        <Title order={1} c="brand">conduit</Title>
        <Text size="xl" c="dimmed">A place to share your knowledge.</Text>
      </Stack>
    </Container>
  );
}
