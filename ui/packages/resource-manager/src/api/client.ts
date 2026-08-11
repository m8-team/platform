export interface RequestOptions {
  signal?: AbortSignal;
}

export interface OrganizationResource {
  readonly id: string;
  readonly name: string;
  readonly parentId: string | null;
  readonly status: 'active' | 'suspended' | 'deleting';
  readonly createdAt: string;
}

export interface ProjectResource {
  readonly id: string;
  readonly organizationId: string;
  readonly workspaceId: string | null;
  readonly name: string;
  readonly description: string | null;
  readonly status: 'active' | 'suspended' | 'deleting';
  readonly version: string;
  readonly createdAt: string;
  readonly updatedAt: string;
}

function throwIfAborted(options?: RequestOptions): void {
  options?.signal?.throwIfAborted();
}

export const resourceManagerApi = {
  organizations: {
    async list(input: unknown, options?: RequestOptions) {
      throwIfAborted(options);
      const items: OrganizationResource[] = [];
      return {
        items,
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
      const items: ProjectResource[] = [];
      return {
        items,
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
