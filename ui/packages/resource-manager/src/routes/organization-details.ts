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
          resource: {$state: '/__runtime/queryResults/organization/data'},
        },
        children: [],
      },
      summary: {
        type: 'PropertyList',
        props: {
          value: {$state: '/__runtime/queryResults/organization/data'},
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
          rows: {$state: '/__runtime/queryResults/projects/data/items'},
          loading: {$state: '/__runtime/queryResults/projects/fetching'},
          detailPath: '/resource-manager/projects/{id}',
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
