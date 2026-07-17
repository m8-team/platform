import {createRootRoute, createRoute, createRouter} from '@tanstack/react-router'

import App, {
  ResourceManagerOverviewPage,
  ResourceOrganizationDetailsPage,
  ResourceOrganizationsPage,
  ResourceProjectDetailsPage,
  ResourceProjectsPage,
  ResourceWorkspaceDetailsPage,
  ResourceWorkspacesPage,
} from './App'
import {commerceIntelligenceRoutes} from './modules/commerce-intelligence/routes'

const rootRoute = createRootRoute({
  component: App,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: ResourceProjectsPage,
})

const resourceManagerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager',
  component: ResourceManagerOverviewPage,
})

const resourceOrganizationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/organizations',
  component: ResourceOrganizationsPage,
})

const resourceOrganizationDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/organizations/$organizationId',
  component: ResourceOrganizationDetailsPage,
})

const resourceWorkspacesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/workspaces',
  component: ResourceWorkspacesPage,
})

const resourceWorkspaceDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/workspaces/$workspaceId',
  component: ResourceWorkspaceDetailsPage,
})

const resourceProjectsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/projects',
  component: ResourceProjectsPage,
})

const resourceProjectDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/projects/$projectId',
  component: ResourceProjectDetailsPage,
})

const commerceRoutes = commerceIntelligenceRoutes.map((route) =>
  createRoute({
    getParentRoute: () => rootRoute,
    path: route.path,
    component: route.component,
  }),
)

const routeTree = rootRoute.addChildren([
  indexRoute,
  resourceManagerRoute,
  resourceOrganizationsRoute,
  resourceOrganizationDetailsRoute,
  resourceWorkspacesRoute,
  resourceWorkspaceDetailsRoute,
  resourceProjectsRoute,
  resourceProjectDetailsRoute,
  ...commerceRoutes,
])

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
