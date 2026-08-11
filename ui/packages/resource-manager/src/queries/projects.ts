import {defineQuery} from '@m8/query';
import {z} from 'zod';

import {resourceManagerApi} from '../api/client';

const MAX_PAGE_SIZE = 1000;

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
  href: z.string().optional(),
  deleteInput: z.object({projectId: z.string(), version: z.string()}).optional(),
});

export const listProjectsQuery = defineQuery({
  id: 'resource-manager.projects.list',
  input: z.object({
    organizationId: z.string().optional(),
    search: z.string().optional(),
    status: z.enum(['active', 'suspended', 'deleting']).optional(),
    limit: z.number().int().positive().max(MAX_PAGE_SIZE).default(50),
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
      items: result.items.map(item => ({
        ...item,
        href: `/resource-manager/projects/${encodeURIComponent(item.id)}`,
      })),
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
  execute: async ({input, signal}) => {
    const project = await resourceManagerApi.projects.get(input.projectId, {signal});
    return {
      ...project,
      deleteInput: {projectId: project.id, version: project.version},
    };
  },
});
