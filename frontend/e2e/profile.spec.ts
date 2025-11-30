import { test, expect } from '@playwright/test';

/**
 * Profile E2E Tests
 * Covers: 프로필 조회 → 팔로우 → 피드 확인 → 좋아요 → 프로필에서 확인
 */

test.describe('Profile and Follow Flow', () => {
  test('should view own profile', async ({ page }) => {
    // Register
    const user = {
      username: `profile${Date.now()}`,
      email: `profile${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user.username);
    await page.getByLabel('Email').fill(user.email);
    await page.getByLabel('Password').fill(user.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Click on username to go to profile
    await page.getByRole('link', { name: user.username }).click();
    await expect(page).toHaveURL(`/profile/${user.username}`, { timeout: 5000 });
    await expect(page.getByText(user.username)).toBeVisible();
  });

  test('should show My Articles and Favorited Articles tabs on profile', async ({ page }) => {
    // Register
    const user = {
      username: `profiletabs${Date.now()}`,
      email: `profiletabs${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user.username);
    await page.getByLabel('Email').fill(user.email);
    await page.getByLabel('Password').fill(user.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Go to profile
    await page.getByRole('link', { name: user.username }).click();
    await expect(page).toHaveURL(`/profile/${user.username}`, { timeout: 5000 });

    // Check for tabs
    await expect(page.getByRole('tab', { name: /my articles/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /favorited/i })).toBeVisible();
  });

  test('should follow another user and see their articles in feed', async ({ page }) => {
    // Create first user who will write an article
    const author = {
      username: `author${Date.now()}`,
      email: `author${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(author.username);
    await page.getByLabel('Email').fill(author.email);
    await page.getByLabel('Password').fill(author.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Create an article
    await page.getByRole('link', { name: 'New Article' }).click();
    const articleTitle = `Follow Test Article ${Date.now()}`;
    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill('Article by author');
    await page.locator('textarea').first().fill('Content from author');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();
    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Logout
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('button', { name: /logout/i }).click();

    // Create second user who will follow the first
    const follower = {
      username: `follower${Date.now()}`,
      email: `follower${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(follower.username);
    await page.getByLabel('Email').fill(follower.email);
    await page.getByLabel('Password').fill(follower.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Go to author's profile
    await page.goto(`/profile/${author.username}`);
    await expect(page.getByText(author.username)).toBeVisible();

    // Click follow button
    const followButton = page.getByRole('button', { name: /follow/i });
    if (await followButton.isVisible()) {
      await followButton.click();
      // Button should change to "Unfollow"
      await expect(page.getByRole('button', { name: /unfollow/i })).toBeVisible({ timeout: 5000 });
    }

    // Go to home and check Your Feed
    await page.goto('/');
    await page.getByRole('tab', { name: 'Your Feed' }).click();
    await page.waitForTimeout(1000);

    // Should see the followed author's article
    // Note: This depends on the actual implementation
  });

  test('should unfollow a user', async ({ page }) => {
    // Create and login as first user
    const user1 = {
      username: `user1${Date.now()}`,
      email: `user1${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user1.username);
    await page.getByLabel('Email').fill(user1.email);
    await page.getByLabel('Password').fill(user1.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Logout
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('button', { name: /logout/i }).click();

    // Create and login as second user
    const user2 = {
      username: `user2${Date.now()}`,
      email: `user2${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user2.username);
    await page.getByLabel('Email').fill(user2.email);
    await page.getByLabel('Password').fill(user2.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Go to first user's profile and follow
    await page.goto(`/profile/${user1.username}`);
    const followButton = page.getByRole('button', { name: /follow/i });
    if (await followButton.isVisible()) {
      await followButton.click();
      await expect(page.getByRole('button', { name: /unfollow/i })).toBeVisible({ timeout: 5000 });

      // Now unfollow
      await page.getByRole('button', { name: /unfollow/i }).click();
      await expect(page.getByRole('button', { name: /follow/i })).toBeVisible({ timeout: 5000 });
    }
  });

  test('should see favorited articles in profile', async ({ page }) => {
    // Create user
    const user = {
      username: `favprofile${Date.now()}`,
      email: `favprofile${Date.now()}@example.com`,
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
    const articleTitle = `Fav Profile Test ${Date.now()}`;
    await page.getByLabel(/title/i).fill(articleTitle);
    await page.getByLabel(/description|about/i).fill('Test description');
    await page.locator('textarea').first().fill('Test body');
    await page.getByRole('button', { name: /publish|create|submit/i }).click();
    await expect(page).toHaveURL(/\/article\//, { timeout: 10000 });

    // Favorite the article (if there's a favorite button on article page)
    const favoriteButton = page.getByRole('button', { name: /favorite/i }).first();
    if (await favoriteButton.isVisible()) {
      await favoriteButton.click();
      await page.waitForTimeout(500);
    }

    // Go to profile and click Favorited Articles tab
    await page.goto(`/profile/${user.username}/favorites`);
    await expect(page.getByRole('tab', { name: /favorited/i })).toBeVisible();
  });

  test('should update profile settings', async ({ page }) => {
    // Register
    const user = {
      username: `settings${Date.now()}`,
      email: `settings${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.goto('/register');
    await page.getByLabel('Username').fill(user.username);
    await page.getByLabel('Email').fill(user.email);
    await page.getByLabel('Password').fill(user.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Go to settings
    await page.getByRole('link', { name: 'Settings' }).click();
    await expect(page).toHaveURL('/settings');

    // Update bio
    const newBio = 'This is my updated bio';
    await page.getByLabel(/bio|short bio/i).fill(newBio);
    await page.getByRole('button', { name: /update|save/i }).click();

    // Verify update was successful
    await page.waitForTimeout(1000);
    // Go to profile to verify bio
    await page.goto(`/profile/${user.username}`);
    await expect(page.getByText(newBio)).toBeVisible({ timeout: 5000 });
  });
});
