import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import * as profileApi from './api';
import type { Profile, ProfileResponse } from './types';
import { articleKeys } from '../article/hooks';

// Query keys factory
export const profileKeys = {
  all: ['profiles'] as const,
  detail: (username: string) => [...profileKeys.all, username] as const,
};

// Hook to get profile by username
export function useProfile(username: string) {
  return useQuery({
    queryKey: profileKeys.detail(username),
    queryFn: () => profileApi.getProfile(username),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!username,
  });
}

// Hook to follow a user
export function useFollowUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (username: string) => profileApi.followUser(username),
    onMutate: async (username) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: profileKeys.detail(username) });

      // Snapshot the previous value
      const previousProfile = queryClient.getQueryData<ProfileResponse>(
        profileKeys.detail(username)
      );

      // Optimistically update
      if (previousProfile) {
        queryClient.setQueryData<ProfileResponse>(profileKeys.detail(username), {
          profile: { ...previousProfile.profile, following: true },
        });
      }

      return { previousProfile };
    },
    onError: (_err, username, context) => {
      // Rollback on error
      if (context?.previousProfile) {
        queryClient.setQueryData(
          profileKeys.detail(username),
          context.previousProfile
        );
      }
    },
    onSuccess: (data, username) => {
      // Update with server response
      queryClient.setQueryData(profileKeys.detail(username), data);
      // Invalidate article lists to update author following status
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
      queryClient.invalidateQueries({ queryKey: articleKeys.feed() });
    },
  });
}

// Hook to unfollow a user
export function useUnfollowUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (username: string) => profileApi.unfollowUser(username),
    onMutate: async (username) => {
      await queryClient.cancelQueries({ queryKey: profileKeys.detail(username) });

      const previousProfile = queryClient.getQueryData<ProfileResponse>(
        profileKeys.detail(username)
      );

      if (previousProfile) {
        queryClient.setQueryData<ProfileResponse>(profileKeys.detail(username), {
          profile: { ...previousProfile.profile, following: false },
        });
      }

      return { previousProfile };
    },
    onError: (_err, username, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(
          profileKeys.detail(username),
          context.previousProfile
        );
      }
    },
    onSuccess: (data, username) => {
      queryClient.setQueryData(profileKeys.detail(username), data);
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
      queryClient.invalidateQueries({ queryKey: articleKeys.feed() });
    },
  });
}

// Helper hook for toggling follow status
export function useToggleFollow() {
  const followMutation = useFollowUser();
  const unfollowMutation = useUnfollowUser();

  return {
    toggleFollow: (username: string, isFollowing: boolean) => {
      if (isFollowing) {
        unfollowMutation.mutate(username);
      } else {
        followMutation.mutate(username);
      }
    },
    isPending: followMutation.isPending || unfollowMutation.isPending,
  };
}

// Helper to update profile in cache
export function useOptimisticProfileUpdate() {
  const queryClient = useQueryClient();

  return {
    updateProfileInCache: (username: string, updater: (profile: Profile) => Profile) => {
      queryClient.setQueryData<ProfileResponse>(
        profileKeys.detail(username),
        (old) => {
          if (!old) return old;
          return { profile: updater(old.profile) };
        }
      );
    },
  };
}
