import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query';

import {repositories} from '@/shared/api/repositories';

export const queryKeys = {
  dashboard: ['dashboard'] as const,
  tenants: ['tenants'] as const,
  tenant: (tenantId: string) => ['tenants', tenantId] as const,
  users: ['users'] as const,
  user: (userId: string) => ['users', userId] as const,
  groups: ['groups'] as const,
  group: (groupId: string) => ['groups', groupId] as const,
  serviceAccounts: ['serviceAccounts'] as const,
  serviceAccount: (serviceAccountId: string) => ['serviceAccounts', serviceAccountId] as const,
  oauthClients: ['oauthClients'] as const,
  oauthClient: (clientId: string) => ['oauthClients', clientId] as const,
  roles: ['roles'] as const,
  role: (roleId: string) => ['roles', roleId] as const,
  accessBindings: (resourceType: string, resourceId: string) =>
    ['accessBindings', resourceType, resourceId] as const,
  effectiveAccess: ['effectiveAccess'] as const,
  explainAccess: (subjectId: string, resourceId: string, permission: string) =>
    ['explainAccess', subjectId, resourceId, permission] as const,
  impact: (resourceId: string, subjectId: string, roleId: string) =>
    ['impact', resourceId, subjectId, roleId] as const,
  policyTemplates: ['policyTemplates'] as const,
  policyTemplate: (templateId: string) => ['policyTemplates', templateId] as const,
  supportGrants: ['supportGrants'] as const,
  auditEvents: ['auditEvents'] as const,
  auditEvent: (eventId: string) => ['auditEvents', eventId] as const,
  operations: ['operations'] as const,
  operation: (operationId: string) => ['operations', operationId] as const,
  settings: ['settings'] as const,
  sessions: ['sessions'] as const,
  search: (query: string) => ['search', query] as const,
};

export function useDashboardQuery() {
  return useQuery({queryKey: queryKeys.dashboard, queryFn: repositories.getDashboard});
}

export function useTenantsQuery() {
  return useQuery({queryKey: queryKeys.tenants, queryFn: repositories.listTenants});
}

export function useTenantQuery(tenantId: string) {
  return useQuery({
    queryKey: queryKeys.tenant(tenantId),
    queryFn: () => repositories.getTenant(tenantId),
    enabled: Boolean(tenantId),
  });
}

export function useUsersQuery() {
  return useQuery({queryKey: queryKeys.users, queryFn: repositories.listUsers});
}

export function useUserQuery(userId: string) {
  return useQuery({
    queryKey: queryKeys.user(userId),
    queryFn: () => repositories.getUser(userId),
    enabled: Boolean(userId),
  });
}

export function useGroupsQuery() {
  return useQuery({queryKey: queryKeys.groups, queryFn: repositories.listGroups});
}

export function useGroupQuery(groupId: string) {
  return useQuery({
    queryKey: queryKeys.group(groupId),
    queryFn: () => repositories.getGroup(groupId),
    enabled: Boolean(groupId),
  });
}

export function useServiceAccountsQuery() {
  return useQuery({queryKey: queryKeys.serviceAccounts, queryFn: repositories.listServiceAccounts});
}

export function useServiceAccountQuery(serviceAccountId: string) {
  return useQuery({
    queryKey: queryKeys.serviceAccount(serviceAccountId),
    queryFn: () => repositories.getServiceAccount(serviceAccountId),
    enabled: Boolean(serviceAccountId),
  });
}

export function useOAuthClientsQuery() {
  return useQuery({queryKey: queryKeys.oauthClients, queryFn: repositories.listOAuthClients});
}

export function useOAuthClientQuery(clientId: string) {
  return useQuery({
    queryKey: queryKeys.oauthClient(clientId),
    queryFn: () => repositories.getOAuthClient(clientId),
    enabled: Boolean(clientId),
  });
}

export function useRolesQuery() {
  return useQuery({queryKey: queryKeys.roles, queryFn: repositories.listRoles});
}

export function useRoleQuery(roleId: string) {
  return useQuery({
    queryKey: queryKeys.role(roleId),
    queryFn: () => repositories.getRole(roleId),
    enabled: Boolean(roleId),
  });
}

export function useAccessBindingsQuery(resourceType: string, resourceId: string, enabled = true) {
  return useQuery({
    queryKey: queryKeys.accessBindings(resourceType, resourceId),
    queryFn: () => repositories.listResourceBindings(resourceType, resourceId),
    enabled: enabled && Boolean(resourceType) && Boolean(resourceId),
  });
}

export function useEffectiveAccessQuery() {
  return useQuery({queryKey: queryKeys.effectiveAccess, queryFn: repositories.listEffectiveAccess});
}

export function useExplainAccessQuery(
  subjectId: string,
  resourceId: string,
  permission: string,
  enabled: boolean,
) {
  return useQuery({
    queryKey: queryKeys.explainAccess(subjectId, resourceId, permission),
    queryFn: () => repositories.explainAccess(subjectId, resourceId, permission),
    enabled: enabled && Boolean(subjectId) && Boolean(resourceId) && Boolean(permission),
  });
}

export function useImpactSimulationQuery(
  resourceId: string,
  subjectId: string,
  roleId: string,
  enabled: boolean,
) {
  return useQuery({
    queryKey: queryKeys.impact(resourceId, subjectId, roleId),
    queryFn: () => repositories.simulateImpact(resourceId, subjectId, roleId),
    enabled: enabled && Boolean(resourceId) && Boolean(subjectId) && Boolean(roleId),
  });
}

export function usePolicyTemplatesQuery() {
  return useQuery({queryKey: queryKeys.policyTemplates, queryFn: repositories.listPolicyTemplates});
}

export function usePolicyTemplateQuery(templateId: string) {
  return useQuery({
    queryKey: queryKeys.policyTemplate(templateId),
    queryFn: () => repositories.getPolicyTemplate(templateId),
    enabled: Boolean(templateId),
  });
}

export function useSupportGrantsQuery() {
  return useQuery({queryKey: queryKeys.supportGrants, queryFn: repositories.listSupportGrants});
}

export function useAuditEventsQuery() {
  return useQuery({queryKey: queryKeys.auditEvents, queryFn: repositories.listAuditEvents});
}

export function useAuditEventQuery(eventId: string) {
  return useQuery({
    queryKey: queryKeys.auditEvent(eventId),
    queryFn: () => repositories.getAuditEvent(eventId),
    enabled: Boolean(eventId),
  });
}

export function useOperationsQuery() {
  return useQuery({queryKey: queryKeys.operations, queryFn: repositories.listOperations});
}

export function useOperationQuery(operationId: string) {
  return useQuery({
    queryKey: queryKeys.operation(operationId),
    queryFn: () => repositories.getOperation(operationId),
    enabled: Boolean(operationId),
  });
}

export function useSettingsQuery() {
  return useQuery({queryKey: queryKeys.settings, queryFn: repositories.listSettings});
}

export function useSessionsQuery() {
  return useQuery({queryKey: queryKeys.sessions, queryFn: repositories.listSessions});
}

export function useSearchResultsQuery(query: string) {
  return useQuery({
    queryKey: queryKeys.search(query),
    queryFn: () => repositories.listSearchResults(query),
    enabled: query.trim().length > 0,
  });
}

export function useAddGroupMemberMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({groupId, subjectId, subjectType, displayName}: {
      groupId: string;
      subjectId: string;
      subjectType: 'userAccount' | 'serviceAccount' | 'group' | 'federatedUser';
      displayName: string;
    }) =>
      repositories.addGroupMember(groupId, {
        subjectId,
        subjectType,
        displayName,
      }),
    onSuccess: (group) => {
      void queryClient.invalidateQueries({queryKey: queryKeys.groups});
      void queryClient.invalidateQueries({queryKey: queryKeys.group(group.groupId)});
      void queryClient.invalidateQueries({queryKey: queryKeys.users});
      void queryClient.invalidateQueries({queryKey: queryKeys.effectiveAccess});
    },
  });
}

export function useRemoveGroupMemberMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({groupId, subjectId}: {groupId: string; subjectId: string}) =>
      repositories.removeGroupMember(groupId, subjectId),
    onSuccess: (group) => {
      void queryClient.invalidateQueries({queryKey: queryKeys.groups});
      void queryClient.invalidateQueries({queryKey: queryKeys.group(group.groupId)});
      void queryClient.invalidateQueries({queryKey: queryKeys.users});
      void queryClient.invalidateQueries({queryKey: queryKeys.effectiveAccess});
    },
  });
}

export function useCreateServiceAccountMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: repositories.createServiceAccount,
    onSuccess: (account) => {
      void queryClient.invalidateQueries({queryKey: queryKeys.serviceAccounts});
      void queryClient.invalidateQueries({queryKey: queryKeys.tenant(account.tenantId)});
    },
  });
}

export function useRotateSecretMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({clientId, note}: {clientId: string; note: string}) =>
      repositories.rotateClientSecret(clientId, note),
    onSuccess: (client) => {
      void queryClient.invalidateQueries({queryKey: queryKeys.oauthClients});
      void queryClient.invalidateQueries({queryKey: queryKeys.oauthClient(client.clientId)});
    },
  });
}

export function useGrantAccessMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: repositories.grantAccess,
    onSuccess: (_, variables) => {
      void queryClient.invalidateQueries({
        queryKey: queryKeys.accessBindings(variables.resourceType, variables.resourceId),
      });
      void queryClient.invalidateQueries({queryKey: queryKeys.effectiveAccess});
    },
  });
}

export function useCreateSupportGrantMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: repositories.createSupportGrant,
    onSuccess: () => {
      void queryClient.invalidateQueries({queryKey: queryKeys.supportGrants});
      void queryClient.invalidateQueries({queryKey: queryKeys.operations});
    },
  });
}

export function useSaveSettingsMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: repositories.saveSettings,
    onSuccess: () => {
      void queryClient.invalidateQueries({queryKey: queryKeys.settings});
    },
  });
}
