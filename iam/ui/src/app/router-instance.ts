import {createRootRoute, createRoute, createRouter, redirect} from '@tanstack/react-router';

import {AppShell} from '@/layouts/app-shell';
import {
  AccessExplainPage,
  AccessExplorerPage,
  AccessSimulatePage,
  AuditEventPage,
  AuditPage,
  DashboardPage,
  GroupDetailPage,
  GroupsPage,
  OAuthClientDetailPage,
  OAuthClientsPage,
  OperationDetailPage,
  OperationsPage,
  PoliciesPage,
  PolicyDetailPage,
  ResourceAccessPage,
  RoleDetailPage,
  RolesPage,
  SearchPage,
  ServiceAccountsPage,
  ServiceAccountsOverviewRoutePage,
  ServiceAccountsKeysRoutePage,
  SessionsPage,
  SettingsPage,
  SupportAccessPage,
  TenantAccessRoutePage,
  TenantAuditRoutePage,
  TenantGroupsRoutePage,
  TenantMembersRoutePage,
  TenantOAuthClientsRoutePage,
  TenantOverviewRoutePage,
  TenantServiceAccountsRoutePage,
  TenantsPage,
  UserProfilePage,
  UsersPage,
} from '@/pages/screens';

const rootRoute = createRootRoute({
  component: AppShell,
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({to: '/dashboard'});
  },
});

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/dashboard',
  component: DashboardPage,
});

const tenantsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants',
  component: TenantsPage,
});

const tenantOverviewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId',
  component: TenantOverviewRoutePage,
});

const tenantMembersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId/members',
  component: TenantMembersRoutePage,
});

const tenantGroupsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId/groups',
  component: TenantGroupsRoutePage,
});

const tenantServiceAccountsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId/service-accounts',
  component: TenantServiceAccountsRoutePage,
});

const tenantOAuthClientsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId/oauth-clients',
  component: TenantOAuthClientsRoutePage,
});

const tenantAccessRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId/access',
  component: TenantAccessRoutePage,
});

const tenantAuditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tenants/$tenantId/audit',
  component: TenantAuditRoutePage,
});

const usersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/users',
  component: UsersPage,
});

const userRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/users/$userId',
  component: UserProfilePage,
});

const groupsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/groups',
  component: GroupsPage,
});

const groupRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/groups/$groupId',
  component: GroupDetailPage,
});

const serviceAccountsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/service-accounts',
  component: ServiceAccountsPage,
});

const serviceAccountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/service-accounts/$serviceAccountId',
  component: ServiceAccountsOverviewRoutePage,
});

const serviceAccountKeysRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/service-accounts/$serviceAccountId/keys',
  component: ServiceAccountsKeysRoutePage,
});

const oauthClientsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/oauth-clients',
  component: OAuthClientsPage,
});

const oauthClientRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/oauth-clients/$clientId',
  component: OAuthClientDetailPage,
});

const rolesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/roles',
  component: RolesPage,
});

const roleRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/roles/$roleId',
  component: RoleDetailPage,
});

const resourceAccessRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/access/resources/$resourceType/$resourceId',
  component: ResourceAccessPage,
});

const accessExplorerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/access/explorer',
  component: AccessExplorerPage,
});

const accessExplainRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/access/explain',
  component: AccessExplainPage,
});

const accessSimulateRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/access/simulate',
  component: AccessSimulatePage,
});

const policiesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/policies',
  component: PoliciesPage,
});

const policyRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/policies/$templateId',
  component: PolicyDetailPage,
});

const supportAccessRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/support-access',
  component: SupportAccessPage,
});

const sessionsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/sessions',
  component: SessionsPage,
});

const auditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/audit',
  component: AuditPage,
});

const auditEventRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/audit/$eventId',
  component: AuditEventPage,
});

const operationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/operations',
  component: OperationsPage,
});

const operationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/operations/$operationId',
  component: OperationDetailPage,
});

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
});

const searchRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/search',
  component: SearchPage,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  dashboardRoute,
  tenantsRoute,
  tenantOverviewRoute,
  tenantMembersRoute,
  tenantGroupsRoute,
  tenantServiceAccountsRoute,
  tenantOAuthClientsRoute,
  tenantAccessRoute,
  tenantAuditRoute,
  usersRoute,
  userRoute,
  groupsRoute,
  groupRoute,
  serviceAccountsRoute,
  serviceAccountRoute,
  serviceAccountKeysRoute,
  oauthClientsRoute,
  oauthClientRoute,
  rolesRoute,
  roleRoute,
  resourceAccessRoute,
  accessExplorerRoute,
  accessExplainRoute,
  accessSimulateRoute,
  policiesRoute,
  policyRoute,
  supportAccessRoute,
  sessionsRoute,
  auditRoute,
  auditEventRoute,
  operationsRoute,
  operationRoute,
  settingsRoute,
  searchRoute,
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
