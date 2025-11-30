import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAuthStore } from './store';
import type { User } from './types';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
    // Reset store state
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
    });
  });

  const mockUser: User = {
    email: 'test@example.com',
    token: 'test-token-123',
    username: 'testuser',
    bio: 'Test bio',
    image: 'https://example.com/avatar.png',
  };

  describe('setUser', () => {
    it('should set user and update authentication state', () => {
      const { setUser } = useAuthStore.getState();

      setUser(mockUser);

      const state = useAuthStore.getState();
      expect(state.user).toEqual(mockUser);
      expect(state.token).toBe(mockUser.token);
      expect(state.isAuthenticated).toBe(true);
      expect(state.isLoading).toBe(false);
    });

    it('should save token to localStorage', () => {
      const { setUser } = useAuthStore.getState();

      setUser(mockUser);

      expect(localStorageMock.setItem).toHaveBeenCalledWith('token', mockUser.token);
    });
  });

  describe('logout', () => {
    it('should clear user and authentication state', () => {
      // First set a user
      useAuthStore.getState().setUser(mockUser);

      // Then logout
      useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });

    it('should remove token from localStorage', () => {
      useAuthStore.getState().setUser(mockUser);
      localStorageMock.setItem.mockClear();

      useAuthStore.getState().logout();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
    });
  });

  describe('setToken', () => {
    it('should set token and save to localStorage', () => {
      const { setToken } = useAuthStore.getState();
      const newToken = 'new-token-456';

      setToken(newToken);

      const state = useAuthStore.getState();
      expect(state.token).toBe(newToken);
      expect(localStorageMock.setItem).toHaveBeenCalledWith('token', newToken);
    });
  });

  describe('selector hooks', () => {
    it('useUser should return current user', () => {
      useAuthStore.getState().setUser(mockUser);
      expect(useAuthStore.getState().user).toEqual(mockUser);
    });

    it('useIsAuthenticated should return authentication status', () => {
      expect(useAuthStore.getState().isAuthenticated).toBe(false);

      useAuthStore.getState().setUser(mockUser);
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
    });
  });
});
