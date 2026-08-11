export interface RequestOptions {
  signal?: AbortSignal;
}

export const resourceManagerApi = {
  organizations: {
    async list(input: unknown, _options?: RequestOptions) {
      return {
        items: [],
        nextCursor: null,
        input,
      };
    },

    async get(organizationId: string, _options?: RequestOptions) {
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
    async list(input: unknown, _options?: RequestOptions) {
      return {
        items: [],
        nextCursor: null,
        input,
      };
    },

    async get(projectId: string, _options?: RequestOptions) {
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

    async create(input: unknown, _options?: RequestOptions) {
      return {
        operationId: crypto.randomUUID(),
        resourceId: crypto.randomUUID(),
        input,
      };
    },

    async delete(input: unknown, _options?: RequestOptions) {
      return {
        operationId: crypto.randomUUID(),
        input,
      };
    },
  },
};
