import {defineOperation} from '@m8/operation';
import {z} from 'zod';

import {resourceManagerApi} from '../api/client';

export const deleteProjectOperation = defineOperation({
  id: 'resource-manager.projects.delete',
  mode: 'long-running',
  input: z.object({
    projectId: z.string(),
    version: z.string(),
  }),
  output: z.object({
    operationId: z.string(),
  }),
  destructive: true,
  confirmation: {
    title: 'Delete project?',
    description: 'This action cannot be undone.',
    confirmLabel: 'Delete project',
  },
  requiredPermission: 'resource-manager.projects.delete',
  execute: async ({input, signal}) => {
    const result = await resourceManagerApi.projects.delete(input, {signal});
    return {operationId: result.operationId};
  },
  completion: {invalidate: [
    'resource-manager.projects.list',
    'resource-manager.projects.get',
  ]},
});
