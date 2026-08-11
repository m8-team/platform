import {defineQuery} from '@m8/query';
import {z} from 'zod';

import {resourceManagerApi} from '../api/client';

const organizationSchema = z.object({
  id: z.string(),
  name: z.string(),
  parentId: z.string().nullable(),
  status: z.enum(['active', 'suspended', 'deleting']),
  createdAt: z.string(),
});

export const listOrganizationsQuery = defineQuery({
  id: 'resource-manager.organizations.list',
  input: z.object({
    search: z.string().optional(),
    parentId: z.string().optional(),
    limit: z.number().default(50),
    cursor: z.string().optional(),
  }),
  output: z.object({
    items: z.array(organizationSchema),
    nextCursor: z.string().nullable(),
  }),
  queryKey: input => ['resource-manager', 'organizations', input],
  execute: async ({input, signal}) => {
    const result = await resourceManagerApi.organizations.list(input, {signal});
    return {
      items: result.items,
      nextCursor: result.nextCursor,
    };
  },
});

export const getOrganizationQuery = defineQuery({
  id: 'resource-manager.organizations.get',
  input: z.object({
    organizationId: z.string(),
  }),
  output: organizationSchema,
  queryKey: input => [
    'resource-manager',
    'organizations',
    input.organizationId,
  ],
  execute: ({input, signal}) =>
    resourceManagerApi.organizations.get(input.organizationId, {signal}),
});
