export const resourceManagerRoutes = {
  overview: '/resource-manager',
  organizations: {
    list: '/resource-manager/organizations',
    detail: '/resource-manager/organizations/$organizationId',
  },
  workspaces: {
    list: '/resource-manager/workspaces',
    detail: '/resource-manager/workspaces/$workspaceId',
  },
  projects: {
    list: '/resource-manager/projects',
    detail: '/resource-manager/projects/$projectId',
  },
} as const
