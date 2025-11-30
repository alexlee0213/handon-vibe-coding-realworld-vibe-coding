import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, AuthState, AuthActions } from './types';

const TOKEN_KEY = 'token';

type AuthStore = AuthState & AuthActions;

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      // State
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: true,

      // Actions
      setUser: (user: User) => {
        set({
          user,
          token: user.token,
          isAuthenticated: true,
          isLoading: false,
        });
        // Also set in localStorage for API interceptor
        localStorage.setItem(TOKEN_KEY, user.token);
      },

      setToken: (token: string) => {
        set({ token });
        localStorage.setItem(TOKEN_KEY, token);
      },

      logout: () => {
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          isLoading: false,
        });
        localStorage.removeItem(TOKEN_KEY);
      },

      initialize: () => {
        const token = localStorage.getItem(TOKEN_KEY);
        if (token) {
          set({ token, isLoading: true });
        } else {
          set({ isLoading: false });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // If we have a token after rehydration, sync with localStorage
          if (state.token) {
            localStorage.setItem(TOKEN_KEY, state.token);
          }
          state.isLoading = false;
        }
      },
    }
  )
);

// Selector hooks for common use cases
export const useUser = () => useAuthStore((state) => state.user);
export const useToken = () => useAuthStore((state) => state.token);
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated);
export const useIsLoading = () => useAuthStore((state) => state.isLoading);
