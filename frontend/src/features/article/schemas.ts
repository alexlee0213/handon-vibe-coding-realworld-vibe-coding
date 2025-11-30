import { z } from 'zod';

// Create article form schema
export const createArticleSchema = z.object({
  title: z
    .string()
    .min(1, 'Title is required')
    .max(200, 'Title must be at most 200 characters'),
  description: z
    .string()
    .min(1, 'Description is required')
    .max(500, 'Description must be at most 500 characters'),
  body: z
    .string()
    .min(1, 'Article body is required'),
  tagList: z
    .string()
    .optional()
    .transform((val) => {
      if (!val) return [];
      return val
        .split(',')
        .map((tag) => tag.trim())
        .filter((tag) => tag.length > 0);
    }),
});

export type CreateArticleFormValues = z.infer<typeof createArticleSchema>;

// Update article form schema (all fields optional)
export const updateArticleSchema = z.object({
  title: z
    .string()
    .max(200, 'Title must be at most 200 characters')
    .optional()
    .or(z.literal('')),
  description: z
    .string()
    .max(500, 'Description must be at most 500 characters')
    .optional()
    .or(z.literal('')),
  body: z
    .string()
    .optional()
    .or(z.literal('')),
});

export type UpdateArticleFormValues = z.infer<typeof updateArticleSchema>;

// Tag input validation
export const tagSchema = z
  .string()
  .min(1, 'Tag cannot be empty')
  .max(30, 'Tag must be at most 30 characters')
  .regex(/^[a-zA-Z0-9-]+$/, 'Tag can only contain letters, numbers, and hyphens');

// Article slug validation
export const slugSchema = z
  .string()
  .min(1, 'Slug is required')
  .regex(/^[a-z0-9-]+$/, 'Invalid slug format');
