import type {ModuleRouteSpec} from '@m8/core';

export const overviewRoute = {
  metadata: {title: 'Resource Manager'},
  navigation: {
    label: 'Overview',
    icon: 'House',
    order: 10,
  },
  access: {permission: 'resource-manager.overview.read'},
  page: {
    root: 'page',
    elements: {
      page: {
        type: 'Page',
        props: {},
        children: ['header', 'content'],
      },
      header: {
        type: 'PageHeader',
        props: {
          title: 'Resource Manager',
          description: 'Organizations, workspaces, projects and resources.',
        },
        children: [],
      },
      content: {
        type: 'Grid',
        props: {columns: 2, gap: 'l'},
        children: ['organizations', 'projects'],
      },
      organizations: {
        type: 'NavigationCard',
        props: {
          title: 'Organizations',
          description: 'Manage organizations and hierarchy.',
          icon: 'Buildings',
          href: '/resource-manager/organizations',
        },
        children: [],
      },
      projects: {
        type: 'NavigationCard',
        props: {
          title: 'Projects',
          description: 'Manage platform projects.',
          icon: 'Folder',
          href: '/resource-manager/projects',
        },
        children: [],
      },
    },
  },
} satisfies ModuleRouteSpec;
