export interface RequestOptions {
  signal?: AbortSignal;
}

function throwIfAborted(options?: RequestOptions): void {
  options?.signal?.throwIfAborted();
}

export const resourceManagerApi = {
  organizations: {
    async list(input: unknown, options?: RequestOptions) {
      throwIfAborted(options);
      return {
        items: [],
        nextCursor: null,
        input,
      };
    },

    async get(organizationId: string, options?: RequestOptions) {
      throwIfAborted(options);
      return {
        id: organizationId,
        name: `Organization ${organizationId}`,
        parentId: null,
        status: 'active' as const,
        createdAt: new Date().toISOString(),
      };
    },
  },

  projects: {
    async list(input: unknown, options?: RequestOptions) {
      throwIfAborted(options);
      return {
        items: [],
        nextCursor: null,
        input,
      };
    },

    async get(projectId: string, options?: RequestOptions) {
      throwIfAborted(options);
      const now = new Date().toISOString();
      return {
        id: projectId,
        organizationId: 'org-1',
        workspaceId: null,
        name: `Project ${projectId}`,
        description: null,
        status: 'active' as const,
        version: '1',
        createdAt: now,
        updatedAt: now,
      };
    },

    async create(input: unknown, options?: RequestOptions) {
      throwIfAborted(options);
      return {
        operationId: crypto.randomUUID(),
        resourceId: crypto.randomUUID(),
        input,
      };
    },

    async delete(input: unknown, options?: RequestOptions) {
      throwIfAborted(options);
      return {
        operationId: crypto.randomUUID(),
        input,
      };
    },
  },
};
