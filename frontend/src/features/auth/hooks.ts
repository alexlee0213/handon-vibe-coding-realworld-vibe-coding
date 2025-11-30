import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { useAuthStore } from './store';
import * as authApi from './api';
import type { RegisterRequest, LoginRequest, UpdateUserRequest } from './types';

// Query keys
export const authKeys = {
  all: ['auth'] as const,
  currentUser: () => [...authKeys.all, 'currentUser'] as const,
};

// Hook to fetch current user
export function useCurrentUser() {
  const { token, setUser, logout } = useAuthStore();

  return useQuery({
    queryKey: authKeys.currentUser(),
    queryFn: async () => {
      const response = await authApi.getCurrentUser();
      setUser(response.user);
      return response.user;
    },
    enabled: !!token,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
    meta: {
      onError: () => {
        logout();
      },
    },
  });
}

// Hook for registration
export function useRegister() {
  const { setUser } = useAuthStore();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: RegisterRequest['user']) => authApi.register(data),
    onSuccess: (response) => {
      setUser(response.user);
      queryClient.setQueryData(authKeys.currentUser(), response.user);
      navigate({ to: '/' });
    },
  });
}

// Hook for login
export function useLogin() {
  const { setUser } = useAuthStore();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: LoginRequest['user']) => authApi.login(data),
    onSuccess: (response) => {
      setUser(response.user);
      queryClient.setQueryData(authKeys.currentUser(), response.user);
      navigate({ to: '/' });
    },
  });
}

// Hook for logout
export function useLogout() {
  const { logout } = useAuthStore();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return () => {
    logout();
    queryClient.removeQueries({ queryKey: authKeys.currentUser() });
    navigate({ to: '/' });
  };
}

// Hook for updating user
export function useUpdateUser() {
  const { setUser } = useAuthStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateUserRequest['user']) => authApi.updateUser(data),
    onSuccess: (response) => {
      setUser(response.user);
      queryClient.setQueryData(authKeys.currentUser(), response.user);
    },
  });
}
