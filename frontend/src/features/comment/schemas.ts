import { z } from 'zod';

export const createCommentSchema = z.object({
  body: z.string().min(1, 'Comment body is required'),
});

export type CreateCommentFormValues = z.infer<typeof createCommentSchema>;
