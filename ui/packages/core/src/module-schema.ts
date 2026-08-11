import {z} from 'zod';

export const moduleManifestSchema = z.object({
  id: z.string().min(1).regex(/^[a-z][a-z0-9-]*$/),
  title: z.string().min(1),
  icon: z.string().optional(),
  order: z.number().int().optional(),
  dependencies: z
    .object({
      required: z.array(z.string()).optional(),
      optional: z.array(z.string()).optional(),
    })
    .optional(),
});

/** @deprecated Metadata-only schema. Use moduleManifestSchema. */
export const moduleSchema = moduleManifestSchema;
