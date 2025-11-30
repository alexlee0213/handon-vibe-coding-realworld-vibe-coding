import { test, expect } from '@playwright/test';

/**
 * Authentication E2E Tests
 * Covers: 회원가입 → 로그인 → 로그아웃
 */

// Generate unique test user for each run
const testUser = {
  username: `testuser${Date.now()}`,
  email: `test${Date.now()}@example.com`,
  password: 'password123',
};

test.describe('Authentication Flow', () => {
  test('should display home page with Global Feed', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('heading', { name: 'conduit' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Global Feed' })).toBeVisible();
  });

  test('should navigate to register page', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/register');
    await expect(page.getByRole('heading', { name: 'Sign up' })).toBeVisible();
  });

  test('should navigate to login page', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Sign in' }).click();
    await expect(page).toHaveURL('/login');
    await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible();
  });

  test('should show validation errors for empty registration form', async ({ page }) => {
    await page.goto('/register');
    await page.getByRole('button', { name: 'Sign up' }).click();
    // Form should show validation errors
    await expect(page.getByText(/username/i)).toBeVisible();
  });

  test('should register a new user', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel('Username').fill(testUser.username);
    await page.getByLabel('Email').fill(testUser.email);
    await page.getByLabel('Password').fill(testUser.password);
    await page.getByRole('button', { name: 'Sign up' }).click();

    // Should redirect to home page after successful registration
    await expect(page).toHaveURL('/', { timeout: 10000 });
    // Should show user's name in navigation
    await expect(page.getByRole('link', { name: testUser.username })).toBeVisible({ timeout: 5000 });
  });

  test('should login with existing user', async ({ page }) => {
    // First register a new user
    await page.goto('/register');
    const loginUser = {
      username: `logintest${Date.now()}`,
      email: `logintest${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.getByLabel('Username').fill(loginUser.username);
    await page.getByLabel('Email').fill(loginUser.email);
    await page.getByLabel('Password').fill(loginUser.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Logout first
    await page.getByRole('link', { name: 'Settings' }).click();
    await expect(page).toHaveURL('/settings');
    await page.getByRole('button', { name: /logout/i }).click();
    await expect(page).toHaveURL('/');

    // Now login
    await page.getByRole('link', { name: 'Sign in' }).click();
    await page.getByLabel('Email').fill(loginUser.email);
    await page.getByLabel('Password').fill(loginUser.password);
    await page.getByRole('button', { name: 'Sign in' }).click();

    await expect(page).toHaveURL('/', { timeout: 10000 });
    await expect(page.getByRole('link', { name: loginUser.username })).toBeVisible({ timeout: 5000 });
  });

  test('should show error for invalid login credentials', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel('Email').fill('nonexistent@example.com');
    await page.getByLabel('Password').fill('wrongpassword');
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Should show error message
    await expect(page.getByText(/invalid|email or password|unauthorized/i)).toBeVisible({ timeout: 5000 });
  });

  test('should access settings page when logged in', async ({ page }) => {
    // Register and stay logged in
    await page.goto('/register');
    const settingsUser = {
      username: `settings${Date.now()}`,
      email: `settings${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.getByLabel('Username').fill(settingsUser.username);
    await page.getByLabel('Email').fill(settingsUser.email);
    await page.getByLabel('Password').fill(settingsUser.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Navigate to settings
    await page.getByRole('link', { name: 'Settings' }).click();
    await expect(page).toHaveURL('/settings');
    await expect(page.getByRole('heading', { name: /settings/i })).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    // Register and stay logged in
    await page.goto('/register');
    const logoutUser = {
      username: `logout${Date.now()}`,
      email: `logout${Date.now()}@example.com`,
      password: 'password123',
    };

    await page.getByLabel('Username').fill(logoutUser.username);
    await page.getByLabel('Email').fill(logoutUser.email);
    await page.getByLabel('Password').fill(logoutUser.password);
    await page.getByRole('button', { name: 'Sign up' }).click();
    await expect(page).toHaveURL('/', { timeout: 10000 });

    // Go to settings and logout
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('button', { name: /logout/i }).click();

    // Should redirect to home and show sign in/sign up links
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('link', { name: 'Sign in' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Sign up' })).toBeVisible();
  });
});
