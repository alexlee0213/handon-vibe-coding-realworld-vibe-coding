// User type returned from API
export interface User {
  email: string;
  token: string;
  username: string;
  bio: string;
  image: string;
}

// API response wrapper
export interface UserResponse {
  user: User;
}

// Registration request
export interface RegisterRequest {
  user: {
    username: string;
    email: string;
    password: string;
  };
}

// Login request
export interface LoginRequest {
  user: {
    email: string;
    password: string;
  };
}

// Update user request
export interface UpdateUserRequest {
  user: {
    email?: string;
    username?: string;
    password?: string;
    bio?: string;
    image?: string;
  };
}

// API error response
export interface ApiError {
  errors: Record<string, string[]>;
}

// Auth state for Zustand store
export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

// Auth actions for Zustand store
export interface AuthActions {
  setUser: (user: User) => void;
  setToken: (token: string) => void;
  logout: () => void;
  initialize: () => void;
}
