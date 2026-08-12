import type {ModuleRouteSpec} from '@m8/core';

export const projectDetailsRoute = {
  metadata: {title: 'Project'},
  access: {permission: 'resource-manager.projects.read'},
  queries: {
    project: {
      query: 'resource-manager.projects.get',
      input: {projectId: {$state: '/__runtime/params/projectId'}},
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
          resource: {$state: '/__runtime/queryResults/project/data'},
        },
        children: [],
      },
      properties: {
        type: 'PropertyList',
        props: {
          value: {$state: '/__runtime/queryResults/project/data'},
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
            confirm: {
              title: 'Delete project?',
              message: 'This action cannot be undone.',
              confirmLabel: 'Delete project',
              variant: 'danger',
            },
            params: {
              operation: 'resource-manager.projects.delete',
              input: {
                projectId: {$state: '/__runtime/queryResults/project/data/id'},
                version: {$state: '/__runtime/queryResults/project/data/version'},
              },
            },
          },
        },
        children: [],
      },
    },
  },
} satisfies ModuleRouteSpec;
