import {createRootRoute, createRoute, createRouter} from '@tanstack/react-router'

import App from './App'
import {commerceIntelligenceRoutes} from './modules/commerce-intelligence/routes'
import {ResourceOrganizationsPage} from './modules/resource-manager/pages/OrganizationsPage'
import {ResourceManagerOverviewPage} from './modules/resource-manager/pages/ResourceManagerOverviewPage'
import {
  ResourceOrganizationDetailsPage,
  ResourceWorkspaceDetailsPage,
} from './modules/resource-manager/pages/ResourcePlaceholderPage'
import {
  ResourceProjectDetailsPage,
  ResourceProjectsPage,
} from './modules/resource-manager/pages/ResourceProjectsPage'
import {ResourceWorkspacesPage} from './modules/resource-manager/pages/WorkspacesPage'
import {resourceManagerRoutes} from './modules/resource-manager/routes'

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
  path: resourceManagerRoutes.overview,
  component: ResourceManagerOverviewPage,
})

const resourceOrganizationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: resourceManagerRoutes.organizations.list,
  component: ResourceOrganizationsPage,
})

const resourceOrganizationDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: resourceManagerRoutes.organizations.detail,
  component: ResourceOrganizationDetailsPage,
})

const resourceWorkspacesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: resourceManagerRoutes.workspaces.list,
  component: ResourceWorkspacesPage,
})

const resourceWorkspaceDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: resourceManagerRoutes.workspaces.detail,
  component: ResourceWorkspaceDetailsPage,
})

const resourceProjectsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: resourceManagerRoutes.projects.list,
  component: ResourceProjectsPage,
})

const resourceProjectDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: resourceManagerRoutes.projects.detail,
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
