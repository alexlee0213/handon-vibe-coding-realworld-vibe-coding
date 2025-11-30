import { z } from 'zod';

// Registration form schema
export const registerSchema = z.object({
  username: z
    .string()
    .min(1, 'Username is required')
    .min(3, 'Username must be at least 3 characters')
    .max(20, 'Username must be at most 20 characters'),
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters'),
});

export type RegisterFormValues = z.infer<typeof registerSchema>;

// Login form schema
export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address'),
  password: z
    .string()
    .min(1, 'Password is required'),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

// Update user form schema
export const updateUserSchema = z.object({
  email: z
    .string()
    .email('Invalid email address')
    .optional()
    .or(z.literal('')),
  username: z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .max(20, 'Username must be at most 20 characters')
    .optional()
    .or(z.literal('')),
  password: z
    .string()
    .min(8, 'Password must be at least 8 characters')
    .optional()
    .or(z.literal('')),
  bio: z.string().optional(),
  image: z.string().url('Invalid URL').optional().or(z.literal('')),
});

export type UpdateUserFormValues = z.infer<typeof updateUserSchema>;
