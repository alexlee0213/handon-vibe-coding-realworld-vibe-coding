import { createFileRoute, useNavigate } from '@tanstack/react-router';
import {
  Container,
  Tabs,
  Alert,
  Stack,
} from '@mantine/core';
import { IconAlertCircle } from '@tabler/icons-react';
import { useProfile } from '../../features/profile';
import { useArticles } from '../../features/article';
import { ProfileCard } from '../../components/profile';
import { ArticleList } from '../../components/article';

export const Route = createFileRoute('/profile/$username')({
  component: ProfilePage,
});

function ProfilePage() {
  const { username } = Route.useParams();
  const navigate = useNavigate();
  const { data: profileData, isLoading: profileLoading, error: profileError } = useProfile(username);
  const { data: articlesData, isLoading: articlesLoading } = useArticles({ author: username, limit: 10 });

  if (profileError) {
    return (
      <Container size="md" py="xl">
        <Alert
          icon={<IconAlertCircle size={16} />}
          title="Error loading profile"
          color="red"
          variant="light"
        >
          {profileError?.message || 'Profile not found'}
        </Alert>
      </Container>
    );
  }

  return (
    <Container size="md" py="xl">
      <Stack gap="xl">
        <ProfileCard
          profile={profileData?.profile}
          isLoading={profileLoading}
        />

        <Tabs
          defaultValue="my-articles"
          onChange={(value) => {
            if (value === 'favorited') {
              navigate({
                to: '/profile/$username/favorites',
                params: { username },
              });
            }
          }}
        >
          <Tabs.List>
            <Tabs.Tab value="my-articles">My Articles</Tabs.Tab>
            <Tabs.Tab value="favorited">Favorited Articles</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="my-articles" pt="md">
            <ArticleList
              articles={articlesData?.articles}
              isLoading={articlesLoading}
              error={null}
              emptyMessage={`${username} hasn't written any articles yet.`}
            />
          </Tabs.Panel>
        </Tabs>
      </Stack>
    </Container>
  );
}
