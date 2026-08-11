import {defineOperation} from '@m8/operation';
import {z} from 'zod';

import {resourceManagerApi} from '../api/client';

export const createProjectOperation = defineOperation({
  id: 'resource-manager.projects.create',
  mode: 'long-running',
  input: z.object({
    organizationId: z.string(),
    workspaceId: z.string().optional(),
    name: z.string().min(1).max(128),
    description: z.string().max(1024).optional(),
  }),
  output: z.object({
    operationId: z.string(),
    resourceId: z.string().optional(),
  }),
  requiredPermission: 'resource-manager.projects.create',
  execute: async ({input, signal}) => {
    const result = await resourceManagerApi.projects.create(input, {signal});
    return {
      operationId: result.operationId,
      resourceId: result.resourceId,
    };
  },
  invalidate: ['resource-manager.projects.list'],
});
