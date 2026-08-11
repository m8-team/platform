import type {M8RouteSpec} from '@m8/core';

export const organizationDetailsRoute = {
  metadata: {title: 'Organization'},
  access: {permission: 'resource-manager.organizations.read'},
  queries: {
    organization: {
      query: 'resource-manager.organizations.get',
      input: {organizationId: {$param: 'organizationId'}},
    },
    projects: {
      query: 'resource-manager.projects.list',
      input: {
        organizationId: {$param: 'organizationId'},
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
          resource: {$state: '/queries/organization/data'},
        },
        children: [],
      },
      summary: {
        type: 'PropertyList',
        props: {
          value: {$state: '/queries/organization/data'},
          fields: [
            {field: 'id', title: 'ID'},
            {field: 'status', title: 'Status'},
            {field: 'createdAt', title: 'Created'},
          ],
        },
        children: [],
      },
      'projects-title': {
        type: 'SectionHeader',
        props: {title: 'Projects'},
        children: [],
      },
      projects: {
        type: 'ResourceTable',
        props: {
          resourceType: 'project',
          rows: {$state: '/queries/projects/data/items'},
          loading: {$state: '/queries/projects/loading'},
          rowHrefTemplate: '/resource-manager/projects/{id}',
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
} satisfies M8RouteSpec;
