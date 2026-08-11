import type {M8RouteSpec} from '@m8/core';

export const projectsRoute = {
  metadata: {title: 'Projects'},
  navigation: {
    label: 'Projects',
    icon: 'Folder',
    order: 30,
  },
  access: {permission: 'resource-manager.projects.read'},
  queries: {
    projects: {
      query: 'resource-manager.projects.list',
      input: {
        organizationId: {$state: '/filters/organizationId'},
        search: {$state: '/filters/search'},
        status: {$state: '/filters/status'},
        limit: 50,
      },
    },
  },
  page: {
    state: {
      filters: {
        organizationId: null,
        search: '',
        status: null,
      },
      ui: {
        createProjectOpen: false,
      },
    },
    root: 'page',
    elements: {
      page: {
        type: 'Page',
        props: {},
        children: ['header', 'filters', 'table'],
      },
      header: {
        type: 'PageHeader',
        props: {title: 'Projects'},
        children: ['create'],
      },
      create: {
        type: 'Button',
        props: {label: 'Create project', view: 'action'},
        on: {
          press: {
            action: 'setState',
            params: {
              statePath: '/ui/createProjectOpen',
              value: true,
            },
          },
        },
        children: [],
      },
      filters: {
        type: 'FilterBar',
        props: {
          search: {$bindState: '/filters/search'},
          status: {$bindState: '/filters/status'},
          organizationId: {$bindState: '/filters/organizationId'},
        },
        children: [],
      },
      table: {
        type: 'ResourceTable',
        props: {
          resourceType: 'project',
          rows: {$state: '/queries/projects/data/items'},
          loading: {$state: '/queries/projects/loading'},
          rowHrefTemplate: '/resource-manager/projects/{id}',
          columns: [
            {field: 'name', title: 'Project'},
            {field: 'organizationId', title: 'Organization'},
            {field: 'status', title: 'Status'},
            {field: 'updatedAt', title: 'Updated'},
          ],
        },
        children: [],
      },
    },
  },
} satisfies M8RouteSpec;
