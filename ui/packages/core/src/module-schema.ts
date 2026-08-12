import {z} from 'zod';

const contributionSchema = z.object({
  id: z.string().min(1),
}).passthrough();

export const moduleDefinitionSchema = z.strictObject({
  id: z.string().min(1).regex(/^[a-z][a-z0-9-]*$/),
  title: z.string().min(1),
  icon: z.string().optional(),
  routes: z.record(z.string(), z.unknown()).optional(),
  queries: z.array(contributionSchema).optional(),
  operations: z.array(contributionSchema).optional(),
});
