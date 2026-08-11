import type {M8RouteSpec} from '@/platform/modules/types';

export const projectDetailsRoute = {
  metadata: {title: 'Project'},
  access: {permission: 'resource-manager.projects.read'},
  queries: {
    project: {
      query: 'resource-manager.projects.get',
      input: {projectId: {$param: 'projectId'}},
    },
  },
  page: {
    root: 'page',
    elements: {
      page: {
        type: 'Page',
        props: {},
        children: ['header', 'properties', 'danger-zone'],
      },
      header: {
        type: 'ResourceHeader',
        props: {
          resourceType: 'project',
          resource: {$state: '/queries/project/data'},
        },
        children: [],
      },
      properties: {
        type: 'PropertyList',
        props: {
          value: {$state: '/queries/project/data'},
          fields: [
            {field: 'id', title: 'Project ID'},
            {field: 'organizationId', title: 'Organization'},
            {field: 'status', title: 'Status'},
            {field: 'version', title: 'Version'},
            {field: 'createdAt', title: 'Created'},
          ],
        },
        children: [],
      },
      'danger-zone': {
        type: 'DangerZone',
        props: {
          title: 'Delete project',
          description: 'Permanently delete this project.',
        },
        children: ['delete'],
      },
      delete: {
        type: 'Button',
        props: {label: 'Delete project', view: 'outlined-danger'},
        on: {
          press: {
            action: 'executeOperation',
            params: {
              operation: 'resource-manager.projects.delete',
              input: {
                projectId: {$state: '/queries/project/data/id'},
                version: {$state: '/queries/project/data/version'},
              },
            },
          },
        },
        children: [],
      },
    },
  },
} satisfies M8RouteSpec;
