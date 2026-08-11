import type {ModuleRouteSpec} from '@m8/core';

export const organizationDetailsRoute = {
  metadata: {title: 'Organization'},
  access: {permission: 'resource-manager.organizations.read'},
  queries: {
    organization: {
      query: 'resource-manager.organizations.get',
      input: {organizationId: {$state: '/__runtime/params/organizationId'}},
    },
    projects: {
      query: 'resource-manager.projects.list',
      input: {
        organizationId: {$state: '/__runtime/params/organizationId'},
        limit: 50,
      },
    },
  },
  page: {
    root: 'page',
    elements: {
      page: {
        type: 'Page',
        props: {},
        children: ['header', 'summary', 'projects-title', 'projects'],
      },
      header: {
        type: 'ResourceHeader',
        props: {
          resourceType: 'organization',
          resource: {$state: '/__runtime/queries/organization/data'},
        },
        children: [],
      },
      summary: {
        type: 'PropertyList',
        props: {
          value: {$state: '/__runtime/queries/organization/data'},
          fields: [
            {field: 'id', title: 'ID'},
            {field: 'status', title: 'Status'},
            {field: 'createdAt', title: 'Created'},
          ],
        },
        children: [],
      },
      'projects-title': {
        type: 'Heading',
        props: {text: 'Projects', level: '2'},
        children: [],
      },
      projects: {
        type: 'ResourceTable',
        props: {
          resourceType: 'project',
          rows: {$state: '/__runtime/queries/projects/data/items'},
          loading: {$state: '/__runtime/queries/projects/fetching'},
          columns: [
            {field: 'name', title: 'Name'},
            {field: 'status', title: 'Status'},
            {field: 'updatedAt', title: 'Updated'},
          ],
        },
        children: [],
      },
    },
  },
} satisfies ModuleRouteSpec;
