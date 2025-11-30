import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { Container, Title, Text, Stack, Grid, Tabs, Box } from '@mantine/core';
import { useArticles, useFeed } from '../features/article';
import { useIsAuthenticated } from '../features/auth';
import { ArticleList, TagList } from '../components/article';

export const Route = createFileRoute('/')({
  component: HomePage,
});

function HomePage() {
  const isAuthenticated = useIsAuthenticated();
  const [activeTab, setActiveTab] = useState<string | null>(
    isAuthenticated ? 'feed' : 'global'
  );
  const [selectedTag, setSelectedTag] = useState<string | undefined>();

  // When tag is selected, show tag tab
  const currentTab = selectedTag ? 'tag' : activeTab;

  const feedQuery = useFeed({ limit: 10 }, { enabled: isAuthenticated });
  const globalQuery = useArticles({
    limit: 10,
    tag: selectedTag
  });

  const handleTagSelect = (tag: string | undefined) => {
    setSelectedTag(tag);
    if (tag) {
      setActiveTab('tag');
    }
  };

  const handleTabChange = (value: string | null) => {
    setActiveTab(value);
    if (value !== 'tag') {
      setSelectedTag(undefined);
    }
  };

  return (
    <>
      {/* Banner */}
      <Box bg="brand" py="xl">
        <Container size="lg">
          <Stack align="center" gap="xs">
            <Title order={1} c="white" style={{ fontFamily: 'Titillium Web, sans-serif' }}>
              conduit
            </Title>
            <Text size="xl" c="white" opacity={0.9}>
              A place to share your knowledge.
            </Text>
          </Stack>
        </Container>
      </Box>

      {/* Main Content */}
      <Container size="lg" py="xl">
        <Grid>
          <Grid.Col span={{ base: 12, md: 9 }}>
            <Tabs value={currentTab} onChange={handleTabChange}>
              <Tabs.List mb="md">
                {isAuthenticated && (
                  <Tabs.Tab value="feed">Your Feed</Tabs.Tab>
                )}
                <Tabs.Tab value="global">Global Feed</Tabs.Tab>
                {selectedTag && (
                  <Tabs.Tab value="tag">#{selectedTag}</Tabs.Tab>
                )}
              </Tabs.List>

              {isAuthenticated && (
                <Tabs.Panel value="feed">
                  <ArticleList
                    articles={feedQuery.data?.articles}
                    isLoading={feedQuery.isLoading}
                    error={feedQuery.error}
                    emptyMessage="No articles are here... yet. Follow some users to see their articles here."
                  />
                </Tabs.Panel>
              )}

              <Tabs.Panel value="global">
                <ArticleList
                  articles={globalQuery.data?.articles}
                  isLoading={globalQuery.isLoading}
                  error={globalQuery.error}
                />
              </Tabs.Panel>

              {selectedTag && (
                <Tabs.Panel value="tag">
                  <ArticleList
                    articles={globalQuery.data?.articles}
                    isLoading={globalQuery.isLoading}
                    error={globalQuery.error}
                    emptyMessage={`No articles tagged with "${selectedTag}" yet.`}
                  />
                </Tabs.Panel>
              )}
            </Tabs>
          </Grid.Col>

          <Grid.Col span={{ base: 12, md: 3 }}>
            <TagList selectedTag={selectedTag} onTagSelect={handleTagSelect} />
          </Grid.Col>
        </Grid>
      </Container>
    </>
  );
}
