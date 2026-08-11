import type {ModuleRouteSpec} from '@m8/core';

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
        createProjectError: false,
      },
      form: {createProject: {name: '', organizationId: '', description: ''}},
    },
    root: 'page',
    elements: {
      page: {
        type: 'Page',
        props: {},
        children: ['header', 'create-form', 'filters', 'table'],
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
      'create-form': {
        type: 'Card',
        visible: {$state: '/ui/createProjectOpen'},
        props: {title: 'Create project', titleLevel: '2'},
        children: ['create-name', 'create-organization', 'create-description', 'create-submit', 'create-cancel'],
      },
      'create-name': {
        type: 'TextInput',
        props: {label: 'Name', value: {$bindState: '/form/createProject/name'}},
        children: [],
      },
      'create-organization': {
        type: 'TextInput',
        props: {label: 'Organization ID', value: {$bindState: '/form/createProject/organizationId'}},
        children: [],
      },
      'create-description': {
        type: 'TextInput',
        props: {label: 'Description', value: {$bindState: '/form/createProject/description'}},
        children: [],
      },
      'create-submit': {
        type: 'Button',
        props: {label: 'Create', view: 'action'},
        on: {press: {
          action: 'executeOperation',
          params: {
            operation: 'resource-manager.projects.create',
            input: {$state: '/form/createProject'},
          },
          onSuccess: {set: {'/ui/createProjectOpen': false}},
          onError: {set: {'/ui/createProjectError': true}},
        }},
        children: [],
      },
      'create-cancel': {
        type: 'Button',
        props: {label: 'Cancel', view: 'flat'},
        on: {press: {action: 'setState', params: {statePath: '/ui/createProjectOpen', value: false}}},
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
          rows: {$state: '/__runtime/queries/projects/data/items'},
          loading: {$state: '/__runtime/queries/projects/fetching'},
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
} satisfies ModuleRouteSpec;
