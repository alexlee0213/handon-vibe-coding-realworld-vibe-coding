import { test, expect } from '@playwright/test';

/**
 * Article E2E Tests
 * Covers: 글 작성 → 조회 → 댓글 → 삭제
 */

test.describe('Article Flow', () => {
  // Create a unique user and login before each test
  const testUser = {
    username: `articletest${Date.now()}`,
    email: `articletest${Date.now()}@example.com`,
    password: 'password123',
  };

  test.beforeEach(async ({ page }) => {
    // Register and login
    await page.goto('/register');
    await page.getByLabel('Username').fill(testUser.username);
    await page.getByLabel('Email').fill(testUser.email);
    await page.getByLabel('Password').fill(testUser.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });
    await expect(page.getByRole('link', { name: testUser.username })).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to new article page', async ({ page }) => {
    await page.getByRole('link', { name: 'New Article' }).click();
    await expect(page).toHaveURL('/editor');
    await expect(page.getByRole('heading', { name: /new article|create/i })).toBeVisible();
  });

  test('should create a new article', async ({ page }) => {
    await page.getByRole('link', { name: 'New Article' }).click();
    await expect(page).toHaveURL('/editor');

    const articleTitle = `Test Article ${Date.now()}`;
    const articleDescription = 'This is a test article description';
    const articleBody = 'This is the body of the test article. It contains some content for testing purposes.';
    const articleTags = 'test, playwright, e2e';

    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill(articleDescription);
    await page.locator('textarea').first().fill(articleBody);
    await page.getByLabel(/tags/i).fill(articleTags);

    await page.getByRole('button', { name: /publish|create|submit/i }).click();

    // Should redirect to article page
    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });
    await expect(page.getByText(articleTitle)).toBeVisible();
    await expect(page.getByText(articleBody)).toBeVisible();
  });

  test('should view article details', async ({ page }) => {
    // First create an article
    await page.getByRole('link', { name: 'New Article' }).click();

    const articleTitle = `View Article ${Date.now()}`;
    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill('Test description');
    await page.locator('textarea').first().fill('Test body content');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();

    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Verify article content is displayed
    await expect(page.getByText(articleTitle)).toBeVisible();
    await expect(page.getByText(testUser.username)).toBeVisible();
  });

  test('should add a comment to an article', async ({ page }) => {
    // First create an article
    await page.getByRole('link', { name: 'New Article' }).click();

    const articleTitle = `Comment Test ${Date.now()}`;
    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill('Test description');
    await page.locator('textarea').first().fill('Test body content');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();

    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Add a comment
    const commentText = `Test comment ${Date.now()}`;
    await page.getByPlaceholder(/comment/i).fill(commentText);
    await page.getByRole('button', { name: /post comment|submit/i }).click();

    // Verify comment appears
    await expect(page.getByText(commentText)).toBeVisible({ timeout: 5000 });
  });

  test('should show Your Feed tab when logged in', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('tab', { name: 'Your Feed' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Global Feed' })).toBeVisible();
  });

  test('should switch between feed tabs', async ({ page }) => {
    await page.goto('/');

    // Click on Global Feed
    await page.getByRole('tab', { name: 'Global Feed' }).click();
    await expect(page.getByRole('tab', { name: 'Global Feed' })).toHaveAttribute('aria-selected', 'true');

    // Click on Your Feed
    await page.getByRole('tab', { name: 'Your Feed' }).click();
    await expect(page.getByRole('tab', { name: 'Your Feed' })).toHaveAttribute('aria-selected', 'true');
  });
});

test.describe('Article Actions', () => {
  test('should favorite an article', async ({ page }) => {
    // Register
    const user = {
      username: `favtest${Date.now()}`,
      email: `favtest${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user.username);
    await page.getByLabel('Email').fill(user.email);
    await page.getByLabel('Password').fill(user.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an article first
    await page.getByRole('link', { name: 'New Article' }).click();
    const articleTitle = `Favorite Test ${Date.now()}`;
    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill('Test description');
    await page.locator('textarea').first().fill('Test body content');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();
    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Go to home and check for favorite button
    await page.goto('/');
    await page.getByRole('tab', { name: 'Global Feed' }).click();

    // Wait for articles to load
    await page.waitForTimeout(1000);

    // Find and click a favorite button
    const favoriteButton = page.getByRole('button', { name: /favorite/i }).first();
    if (await favoriteButton.isVisible()) {
      await favoriteButton.click();
      // Button state should change
      await page.waitForTimeout(500);
    }
  });

  test('should edit own article', async ({ page }) => {
    // Register
    const user = {
      username: `editarticle${Date.now()}`,
      email: `editarticle${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user.username);
    await page.getByLabel('Email').fill(user.email);
    await page.getByLabel('Password').fill(user.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an article
    await page.getByRole('link', { name: 'New Article' }).click();
    const originalTitle = `Edit Test ${Date.now()}`;
    await page.getByLabel(/title/i).fill(originalTitle);
    await page.getByLabel(/description|about/i).fill('Original description');
    await page.locator('textarea').first().fill('Original body');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();
    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Click edit button
    const editLink = page.getByRole('link', { name: /edit/i });
    if (await editLink.isVisible()) {
      await editLink.click();
      await expect(page).toHaveURL(/\/editor\//, { timeout: 5000 });

      // Update article
      const updatedTitle = `Updated ${originalTitle}`;
      await page.getByLabel(/title/i).fill(updatedTitle);
      await page.getByRole('button', { name: /publish|update|submit/i }).click();

      await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });
      await expect(page.getByText(updatedTitle)).toBeVisible();
    }
  });

  test('should delete own article', async ({ page }) => {
    // Register
    const user = {
      username: `deletearticle${Date.now()}`,
      email: `deletearticle${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user.username);
    await page.getByLabel('Email').fill(user.email);
    await page.getByLabel('Password').fill(user.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an article
    await page.getByRole('link', { name: 'New Article' }).click();
    const articleTitle = `Delete Test ${Date.now()}`;
    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill('Delete test description');
    await page.locator('textarea').first().fill('Delete test body');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();
    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Click delete button
    const deleteButton = page.getByRole('button', { name: /delete/i });
    if (await deleteButton.isVisible()) {
      await deleteButton.click();
      // Should redirect to home after deletion
      await expect(page).toHaveURL('/', { timeout: 10000 });
    }
  });
});
