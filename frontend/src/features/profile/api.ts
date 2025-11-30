import { api } from '../../lib/api';
import type { ProfileResponse } from './types';

// Get user profile by username
export async function getProfile(username: string): Promise<ProfileResponse> {
  const response = await api.get(`profiles/${username}`);
  return response.json<ProfileResponse>();
}

// Follow a user
export async function followUser(username: string): Promise<ProfileResponse> {
  const response = await api.post(`profiles/${username}/follow`);
  return response.json<ProfileResponse>();
}

// Unfollow a user
export async function unfollowUser(username: string): Promise<ProfileResponse> {
  const response = await api.delete(`profiles/${username}/follow`);
  return response.json<ProfileResponse>();
}
