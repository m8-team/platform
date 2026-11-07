import {createRootRoute, createRoute, createRouter} from '@tanstack/react-router'

import App, {
  ResourceManagerOverviewPage,
  ResourceOrganizationDetailsPage,
  ResourceOrganizationStructurePage,
  ResourceOrganizationsPage,
  ResourceOrganizationGroupsPage,
  ResourceOrganizationRelationshipsPage,
  ResourceCatalogRoutePage,
  ResourceProjectDetailsPage,
  ResourceProjectsPage,
  ResourceWorkspaceDetailsPage,
  ResourceWorkspacesPage,
} from './App'
import {commerceIntelligenceRoutes} from './modules/commerce-intelligence/routes'
import {
  SettingsProjectPage, SettingsModulesPage, SettingsIntegrationsPage,
} from './modules/settings/SettingsProjectModulesIntegrations'
import {SettingsWebhooksPage,SettingsApiTokensPage} from './modules/settings/SettingsWebhooksTokens'
import {ObservabilityMetricsPage,ObservabilityLogsPage,ObservabilityLogDetailsPage} from './modules/observability/ObservabilityMetricsLogs'
import {ObservabilityTracesPage,ObservabilityTraceDetailsPage} from './modules/observability/ObservabilityTraces'
import {ObservabilityAlertsPage,ObservabilityAlertDetailsPage,ObservabilitySloPage} from './modules/observability/ObservabilityAlertsSlo'
import {AuditEventsPage,AuditEventDetailsPage} from './modules/audit/AuditEvents'
import {AuditExportsPage} from './modules/audit/AuditExports'
import {
  GatewayApiServicesPage,GatewayRoutesPage,GatewayConsumersPage,
} from './modules/gateway/GatewayInventory'
import {
  GatewayPoliciesPage,GatewayRateLimitsPage,GatewayYamlPage,
} from './modules/gateway/GatewayGovernance'
import {
  LongRunningOperationsPage,
  LongRunningOperationDetailsPage,
} from './modules/long-running-operations/LongRunningOperations'
import {
  QuotaOverviewPage,
  QuotaPoliciesPage,
  QuotaReservationsPage,
  QuotaRequestsPage,
  QuotaReconciliationPage,
} from './modules/quota-manager/QuotaManager'

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

const resourceOrganizationStructureRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/organizations/$organizationId/structure',
  component: ResourceOrganizationStructurePage,
})

const resourceOrganizationGroupsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/groups',
  component: ResourceOrganizationGroupsPage,
})

const resourceOrganizationGroupDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/groups/$groupId',
  component: ResourceOrganizationGroupsPage,
})

const resourceOrganizationRelationshipsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/relationships',
  component: ResourceOrganizationRelationshipsPage,
})

const resourceCatalogRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/catalog',
  component: ResourceCatalogRoutePage,
})

const resourceCatalogDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/resource-manager/catalog/$kind/$resourceId',
  component: ResourceCatalogRoutePage,
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

const longRunningOperationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/platform-operations/long-running',
  component: LongRunningOperationsPage,
})

const longRunningOperationDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/platform-operations/long-running/$operationId',
  component: LongRunningOperationDetailsPage,
})

const gatewayPages = [
  {path: '/gateway', component: GatewayApiServicesPage},
  {path: '/gateway/services', component: GatewayApiServicesPage},
  {path: '/gateway/routes', component: GatewayRoutesPage},
  {path: '/gateway/consumers', component: GatewayConsumersPage},
  {path: '/gateway/policies', component: GatewayPoliciesPage},
  {path: '/gateway/rate-limits', component: GatewayRateLimitsPage},
  {path: '/gateway/yaml', component: GatewayYamlPage},
] as const

const gatewayRoutes = gatewayPages.map(page =>
  createRoute({
    getParentRoute: () => rootRoute,
    path: page.path,
    component: page.component,
  }),
)

const auditPages = [
  {path: '/audit', component: AuditEventsPage},
  {path: '/audit/events', component: AuditEventsPage},
  {path: '/audit/exports', component: AuditExportsPage},
] as const

const auditRoutes = auditPages.map(page =>
  createRoute({
    getParentRoute: () => rootRoute,
    path: page.path,
    component: page.component,
  }),
)

const auditEventDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/audit/events/$eventId',
  component: AuditEventDetailsPage,
})

const observabilityPages = [
  {path: '/observability', component: ObservabilityMetricsPage},
  {path: '/observability/metrics', component: ObservabilityMetricsPage},
  {path: '/observability/logs', component: ObservabilityLogsPage},
  {path: '/observability/traces', component: ObservabilityTracesPage},
  {path: '/observability/alerts', component: ObservabilityAlertsPage},
  {path: '/observability/slo', component: ObservabilitySloPage},
] as const

const observabilityRoutes = observabilityPages.map(page =>
  createRoute({
    getParentRoute: () => rootRoute,
    path: page.path,
    component: page.component,
  }),
)
const observabilityLogDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/observability/logs/$logId',
  component: ObservabilityLogDetailsPage,
})
const observabilityTraceDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/observability/traces/$traceId',
  component: ObservabilityTraceDetailsPage,
})
const observabilityAlertDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/observability/alerts/$incidentId',
  component: ObservabilityAlertDetailsPage,
})

const settingsPages = [
  {path: '/settings', component: SettingsProjectPage},
  {path: '/settings/project', component: SettingsProjectPage},
  {path: '/settings/modules', component: SettingsModulesPage},
  {path: '/settings/integrations', component: SettingsIntegrationsPage},
  {path: '/settings/webhooks', component: SettingsWebhooksPage},
  {path: '/settings/api-tokens', component: SettingsApiTokensPage},
] as const

const settingsRoutes = settingsPages.map(page =>
  createRoute({
    getParentRoute: () => rootRoute,
    path: page.path,
    component: page.component,
  }),
)

const quotaRoutes = [
  {path: '/quota-manager', component: QuotaOverviewPage},
  {path: '/quota-manager/overview', component: QuotaOverviewPage},
  {path: '/quota-manager/policies', component: QuotaPoliciesPage},
  {path: '/quota-manager/reservations', component: QuotaReservationsPage},
  {path: '/quota-manager/requests', component: QuotaRequestsPage},
  {path: '/quota-manager/reconciliation', component: QuotaReconciliationPage},
] as const

const quotaManagerRoutes = quotaRoutes.map((route) =>
  createRoute({
    getParentRoute: () => rootRoute,
    path: route.path,
    component: route.component,
  }),
)

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
  resourceOrganizationStructureRoute,
  resourceOrganizationGroupsRoute,
  resourceOrganizationGroupDetailRoute,
  resourceOrganizationRelationshipsRoute,
  resourceCatalogRoute,
  resourceCatalogDetailRoute,
  resourceWorkspacesRoute,
  resourceWorkspaceDetailsRoute,
  resourceProjectsRoute,
  resourceProjectDetailsRoute,
  ...commerceRoutes,
  ...quotaManagerRoutes,
  ...settingsRoutes,
  ...observabilityRoutes,
  observabilityLogDetailsRoute,
  observabilityTraceDetailsRoute,
  observabilityAlertDetailsRoute,
  ...auditRoutes,
  auditEventDetailsRoute,
  ...gatewayRoutes,
  longRunningOperationsRoute,
  longRunningOperationDetailsRoute,
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
