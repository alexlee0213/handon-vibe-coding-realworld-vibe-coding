// Profile type returned from API
export interface Profile {
  username: string;
  bio: string;
  image: string;
  following: boolean;
}

// Profile response wrapper
export interface ProfileResponse {
  profile: Profile;
}
