import type {ModuleRouteSpec} from '@m8/core';

export const organizationsRoute = {
  metadata: {title: 'Organizations'},
  navigation: {
    label: 'Organizations',
    icon: 'Buildings',
    order: 20,
  },
  access: {permission: 'resource-manager.organizations.read'},
  queries: {
    organizations: {
      query: 'resource-manager.organizations.list',
      input: {
        search: {$state: '/filters/search'},
        limit: 50,
      },
    },
  },
  page: {
    state: {
      filters: {search: ''},
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
        props: {title: 'Organizations'},
        children: [],
      },
      filters: {
        type: 'FilterBar',
        props: {
          search: {$bindState: '/filters/search'},
          searchPlaceholder: 'Search organizations',
        },
        children: [],
      },
      table: {
        type: 'ResourceTable',
        props: {
          loading: {$state: '/__runtime/queries/organizations/fetching'},
          rows: {$state: '/__runtime/queries/organizations/data/items'},
          resourceType: 'organization',
          columns: [
            {field: 'name', title: 'Name'},
            {field: 'status', title: 'Status'},
            {field: 'createdAt', title: 'Created'},
          ],
        },
        children: [],
      },
    },
  },
} satisfies ModuleRouteSpec;
