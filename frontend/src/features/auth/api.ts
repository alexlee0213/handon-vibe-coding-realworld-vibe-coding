import { api } from '../../lib/api';
import type {
  UserResponse,
  RegisterRequest,
  LoginRequest,
  UpdateUserRequest,
} from './types';

// Register a new user
export async function register(data: RegisterRequest['user']): Promise<UserResponse> {
  const response = await api.post('users', {
    json: { user: data },
  });
  return response.json<UserResponse>();
}

// Login with email and password
export async function login(data: LoginRequest['user']): Promise<UserResponse> {
  const response = await api.post('users/login', {
    json: { user: data },
  });
  return response.json<UserResponse>();
}

// Get current user
export async function getCurrentUser(): Promise<UserResponse> {
  const response = await api.get('user');
  return response.json<UserResponse>();
}

// Update current user
export async function updateUser(data: UpdateUserRequest['user']): Promise<UserResponse> {
  // Filter out empty strings and undefined values
  const filteredData = Object.fromEntries(
    Object.entries(data).filter(([, value]) => value !== undefined && value !== '')
  );

  const response = await api.put('user', {
    json: { user: filteredData },
  });
  return response.json<UserResponse>();
}
