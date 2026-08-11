import {z} from 'zod';

export const moduleSchema =
  z.object({
    id: z
      .string()
      .min(1)
      .regex(
        /^[a-z][a-z0-9-]*$/,
      ),

    title: z
      .string()
      .min(1),

    basePath: z
      .string()
      .startsWith('/'),

    icon: z
      .string()
      .optional(),

    order: z
      .number()
      .int()
      .optional(),

    access: z
      .object({
        permission:
          z.string().optional(),

        feature:
          z.string().optional(),

        editions:
          z.array(z.string()).optional(),
      })
      .optional(),

    dependencies: z
      .object({
        required:
          z.array(z.string()).optional(),

        optional:
          z.array(z.string()).optional(),
      })
      .optional(),
  });
