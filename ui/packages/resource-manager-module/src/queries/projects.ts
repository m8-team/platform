import {defineQuery} from '@m8/module-sdk';
import {z} from 'zod';

import {resourceManagerApi} from '../api/client';

export const projectSchema = z.object({
  id: z.string(),
  organizationId: z.string(),
  workspaceId: z.string().nullable(),
  name: z.string(),
  description: z.string().nullable(),
  status: z.enum(['active', 'suspended', 'deleting']),
  version: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const listProjectsQuery = defineQuery({
  id: 'resource-manager.projects.list',
  input: z.object({
    organizationId: z.string().optional(),
    search: z.string().optional(),
    status: z.enum(['active', 'suspended', 'deleting']).optional(),
    limit: z.number().default(50),
    cursor: z.string().optional(),
  }),
  output: z.object({
    items: z.array(projectSchema),
    nextCursor: z.string().nullable(),
  }),
  queryKey: input => ['resource-manager', 'projects', input],
  execute: async ({input, signal}) => {
    const result = await resourceManagerApi.projects.list(input, {signal});
    return {
      items: result.items,
      nextCursor: result.nextCursor,
    };
  },
});

export const getProjectQuery = defineQuery({
  id: 'resource-manager.projects.get',
  input: z.object({
    projectId: z.string(),
  }),
  output: projectSchema,
  queryKey: input => ['resource-manager', 'projects', input.projectId],
  execute: ({input, signal}) =>
    resourceManagerApi.projects.get(input.projectId, {signal}),
});
