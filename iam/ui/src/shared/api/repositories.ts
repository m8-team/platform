import {apiClient} from '@/shared/api/client';
import {env} from '@/shared/config/env';
import {createId, cloneValue, delay, getMockDatabase} from '@/mocks/data';
import type {
  AuditEvent,
  DashboardData,
  EffectiveAccessRow,
  ExplainAccessResult,
  GroupDetail,
  GroupMember,
  GroupSummary,
  ImpactRow,
  OAuthClientDetail,
  Operation,
  PagedResponse,
  PolicyTemplate,
  ResourceBinding,
  RoleDetail,
  SearchResult,
  ServiceAccountDetail,
  SettingsSection,
  SupportGrant,
  TenantDetail,
  TenantSummary,
  TokenInfo,
  UserDetail,
  UserSummary,
} from '@/shared/types/iam';

type LiveModeMapper<T> = () => Promise<T>;
type MockMapper<T> = () => Promise<T>;
type TenantTier = 'active' | 'trial';
type TenantMutationPayload = {
  tenantId: string;
  displayName: string;
  externalRef: string;
  region: string;
  tier: TenantTier;
};

const db = getMockDatabase();

async function withFallback<T>(liveMapper: LiveModeMapper<T>, mockMapper: MockMapper<T>): Promise<T> {
  if (env.apiMode !== 'live') {
    return mockMapper();
  }

  try {
    return await liveMapper();
  } catch (error) {
    if (!env.enableFallbackToMock) {
      throw error;
    }
    return mockMapper();
  }
}

function mapTenantFromLive(raw: Record<string, unknown>): TenantDetail {
  const tenantId = String(raw.tenant_id ?? raw.tenantId ?? '');
  const labels = (raw.labels as Record<string, string> | undefined) ?? {};
  const existing = db.tenants.find((tenant) => tenant.tenantId === tenantId);

  return {
    id: tenantId,
    tenantId,
    name: String(raw.display_name ?? raw.displayName ?? existing?.name ?? tenantId),
    organizationId: existing?.organizationId ?? 'org-1',
    plan: existing?.plan ?? (labels.tier === 'trial' ? 'Trial' : 'Enterprise'),
    region: labels.region ?? existing?.region ?? 'eu-central',
    status: existing?.status ?? (labels.tier === 'trial' ? 'trial' : 'active'),
    tags: existing?.tags ?? Object.values(labels),
    memberCount: existing?.memberCount ?? 0,
    groupCount: existing?.groupCount ?? 0,
    serviceAccountCount: existing?.serviceAccountCount ?? 0,
    oauthClientCount: existing?.oauthClientCount ?? 0,
    updatedAt: String(raw.updated_at ?? raw.updatedAt ?? existing?.updatedAt ?? new Date().toISOString()),
    ownersCount: existing?.ownersCount ?? 1,
    createdAt: String(raw.created_at ?? raw.createdAt ?? existing?.createdAt ?? new Date().toISOString()),
    externalRef: String(raw.external_ref ?? raw.externalRef ?? existing?.externalRef ?? ''),
    description: existing?.description ?? 'Tenant synchronized from IAM.',
    summary: existing?.summary ?? ['Live data from grpc-gateway'],
    integrations: existing?.integrations ?? ['Keycloak', 'SpiceDB'],
    resourceMap: existing?.resourceMap ?? [tenantId],
  };
}

function buildTenantLabels(payload: Pick<TenantMutationPayload, 'region' | 'tier'>): Record<string, string> {
  return {
    region: payload.region,
    tier: payload.tier,
  };
}

function buildTenantPlan(tier: TenantTier): string {
  return tier === 'trial' ? 'Trial' : 'Enterprise';
}

function upsertTenantReferences(tenantId: string, tenantName: string) {
  db.users.forEach((user) => {
    user.memberships = user.memberships.map((membership) =>
      membership.tenantId === tenantId ? {...membership, tenantName} : membership,
    );
  });
  db.groups.forEach((group) => {
    if (group.tenantId === tenantId) {
      group.tenantName = tenantName;
    }
  });
  db.serviceAccounts.forEach((account) => {
    if (account.tenantId === tenantId) {
      account.tenantName = tenantName;
    }
  });
  db.oauthClients.forEach((client) => {
    if (client.tenantId === tenantId) {
      client.tenantName = tenantName;
    }
  });
  db.supportGrants.forEach((grant) => {
    if (grant.tenantId === tenantId) {
      grant.tenantName = tenantName;
    }
  });
}

function buildTenantDetail(existing: TenantDetail | undefined, payload: TenantMutationPayload, updatedAt: string): TenantDetail {
  return {
    id: payload.tenantId,
    tenantId: payload.tenantId,
    name: payload.displayName,
    organizationId: existing?.organizationId ?? 'org-local',
    plan: buildTenantPlan(payload.tier),
    region: payload.region,
    status: payload.tier,
    tags: existing?.tags ?? [payload.region, payload.tier],
    memberCount: existing?.memberCount ?? 0,
    groupCount: existing?.groupCount ?? 0,
    serviceAccountCount: existing?.serviceAccountCount ?? 0,
    oauthClientCount: existing?.oauthClientCount ?? 0,
    updatedAt,
    ownersCount: existing?.ownersCount ?? 1,
    createdAt: existing?.createdAt ?? updatedAt,
    externalRef: payload.externalRef,
    description: existing?.description ?? 'Tenant managed from the IAM Admin UI.',
    summary: existing?.summary ?? [`Primary region: ${payload.region}`, `Plan: ${buildTenantPlan(payload.tier)}`],
    integrations: existing?.integrations ?? ['Keycloak', 'SpiceDB'],
    resourceMap: existing?.resourceMap ?? [payload.tenantId],
  };
}

function deleteTenantReferences(tenantId: string) {
  db.tenants = db.tenants.filter((tenant) => tenant.tenantId !== tenantId);
  db.users = db.users
    .map((user) => ({
      ...user,
      tenantIds: user.tenantIds.filter((id) => id !== tenantId),
      memberships: user.memberships.filter((membership) => membership.tenantId !== tenantId),
      groups: user.groups.filter((group) => group.tenantId !== tenantId),
    }))
    .filter((user) => user.tenantIds.length > 0);
  db.groups = db.groups.filter((group) => group.tenantId !== tenantId);
  db.serviceAccounts = db.serviceAccounts.filter((account) => account.tenantId !== tenantId);
  db.oauthClients = db.oauthClients.filter((client) => client.tenantId !== tenantId);
  db.supportGrants = db.supportGrants.filter((grant) => grant.tenantId !== tenantId);
  db.auditEvents = db.auditEvents.filter((event) => event.tenantId !== tenantId);
  db.operations = db.operations.filter((operation) => operation.tenantId !== tenantId);
  db.resourceBindings = db.resourceBindings.filter((binding) => binding.tenantId !== tenantId);
}

function mapUserFromLive(raw: Record<string, unknown>, tenantName?: string): UserDetail {
  const userId = String(raw.user_id ?? raw.userId ?? '');
  const tenantId = String(raw.tenant_id ?? raw.tenantId ?? env.defaultTenantId);
  const existing = db.users.find((user) => user.userId === userId);

  return {
    id: userId,
    userId,
    tenantIds: existing?.tenantIds ?? [tenantId],
    name: String(raw.display_name ?? raw.displayName ?? existing?.name ?? userId),
    email: String(raw.primary_email ?? raw.primaryEmail ?? existing?.email ?? ''),
    source: existing?.source ?? 'IAM',
    status: String(raw.state ?? '').includes('DISABLED') ? 'disabled' : existing?.status ?? 'active',
    mfaEnabled: existing?.mfaEnabled ?? false,
    lastLoginAt: existing?.lastLoginAt ?? new Date().toISOString(),
    memberships: existing?.memberships ?? [
      {
        id: `${tenantId}:${userId}`,
        tenantId,
        tenantName: tenantName ?? tenantId,
        roleIds: [],
        status: 'active',
      },
    ],
    groups: existing?.groups ?? [],
    accessSummary: existing?.accessSummary ?? [],
    sessions: existing?.sessions ?? [],
    tokens: existing?.tokens ?? [],
    labels: (raw.labels as Record<string, string> | undefined) ?? existing?.labels,
  };
}

function mapGroupFromLive(raw: Record<string, unknown>, tenantName?: string): GroupDetail {
  const groupId = String(raw.group_id ?? raw.groupId ?? '');
  const tenantId = String(raw.tenant_id ?? raw.tenantId ?? env.defaultTenantId);
  const existing = db.groups.find((group) => group.groupId === groupId);

  return {
    id: groupId,
    groupId,
    tenantId,
    tenantName: tenantName ?? existing?.tenantName ?? tenantId,
    name: String(raw.display_name ?? raw.displayName ?? existing?.name ?? groupId),
    description: String(raw.description ?? existing?.description ?? ''),
    dynamic: existing?.dynamic ?? false,
    membersCount: existing?.membersCount ?? 0,
    status: existing?.status ?? 'active',
    updatedAt: String(raw.updated_at ?? raw.updatedAt ?? existing?.updatedAt ?? new Date().toISOString()),
    rules: existing?.rules ?? [],
    members: existing?.members ?? [],
    effectiveGrantCount: existing?.effectiveGrantCount ?? 0,
    effectiveRoleCount: existing?.effectiveRoleCount ?? 0,
  };
}

function mapServiceAccountFromLive(raw: Record<string, unknown>, tenantName?: string): ServiceAccountDetail {
  const serviceAccountId = String(raw.service_account_id ?? raw.serviceAccountId ?? '');
  const tenantId = String(raw.tenant_id ?? raw.tenantId ?? env.defaultTenantId);
  const existing = db.serviceAccounts.find((item) => item.serviceAccountId === serviceAccountId);

  return {
    id: serviceAccountId,
    serviceAccountId,
    tenantId,
    tenantName: tenantName ?? existing?.tenantName ?? tenantId,
    name: String(raw.display_name ?? raw.displayName ?? existing?.name ?? serviceAccountId),
    description: String(raw.description ?? existing?.description ?? ''),
    status: raw.disabled ? 'disabled' : existing?.status ?? 'active',
    keysCount: existing?.keysCount ?? 0,
    apiKeysCount: existing?.apiKeysCount ?? 0,
    lastAuthAt: existing?.lastAuthAt,
    updatedAt: String(raw.updated_at ?? raw.updatedAt ?? existing?.updatedAt ?? new Date().toISOString()),
    ownerTeam: existing?.ownerTeam ?? 'Platform',
    createdAt: String(raw.created_at ?? raw.createdAt ?? existing?.createdAt ?? new Date().toISOString()),
    tokens: existing?.tokens ?? [],
    accessSummary: existing?.accessSummary ?? [],
    asymmetricKeys: existing?.asymmetricKeys ?? [],
    apiKeys: existing?.apiKeys ?? [],
  };
}

function mapOAuthClientFromLive(raw: Record<string, unknown>, tenantName?: string): OAuthClientDetail {
  const clientId = String(raw.oauth_client_id ?? raw.oauthClientId ?? '');
  const tenantId = String(raw.tenant_id ?? raw.tenantId ?? env.defaultTenantId);
  const redirectUris = (raw.redirect_uris as string[] | undefined) ?? [];
  const scopes = (raw.scopes as string[] | undefined) ?? [];
  const existing = db.oauthClients.find((item) => item.clientId === clientId);

  return {
    id: clientId,
    clientId,
    tenantId,
    tenantName: tenantName ?? existing?.tenantName ?? tenantId,
    name: String(raw.display_name ?? raw.displayName ?? existing?.name ?? clientId),
    type: String(raw.client_type ?? raw.clientType ?? existing?.type ?? 'confidential').replace('OAUTH_CLIENT_TYPE_', '').toLowerCase(),
    status: existing?.status ?? 'active',
    redirectUrisCount: redirectUris.length,
    scopesCount: scopes.length,
    secretsCount: existing?.secretsCount ?? 0,
    updatedAt: String(raw.updated_at ?? raw.updatedAt ?? existing?.updatedAt ?? new Date().toISOString()),
    createdAt: String(raw.created_at ?? raw.createdAt ?? existing?.createdAt ?? new Date().toISOString()),
    redirectUris,
    scopes,
    secrets: existing?.secrets ?? [],
    tokens: existing?.tokens ?? [],
  };
}

function mapRoleFromLive(raw: Record<string, unknown>): RoleDetail {
  const roleId = String(raw.role_id ?? raw.roleId ?? '');
  const permissions = Array.isArray(raw.permissions)
    ? raw.permissions.map((item) => {
        const permission = item as Record<string, unknown>;
        return {
          id: String(permission.id ?? ''),
          displayName: String(permission.display_name ?? permission.displayName ?? permission.id ?? ''),
          description: String(permission.description ?? ''),
        };
      })
    : [];
  const existing = db.roles.find((role) => role.roleId === roleId);

  return {
    id: roleId,
    roleId,
    namespace: roleId.split('-')[0] ?? existing?.namespace ?? 'iam',
    name: String(raw.display_name ?? raw.displayName ?? existing?.name ?? roleId),
    description: String(raw.description ?? existing?.description ?? ''),
    permissionsCount: permissions.length,
    system: existing?.system ?? false,
    permissions,
    usedBy: existing?.usedBy ?? [],
  };
}

async function mockList<T>(items: T[]): Promise<PagedResponse<T>> {
  await delay();
  return {items: cloneValue(items), total: items.length};
}

async function listTenantsLive(): Promise<PagedResponse<TenantSummary>> {
  const response = await apiClient.get<{tenants?: Record<string, unknown>[]}>('/api/v1/tenants');
  const items = (response.tenants ?? []).map(mapTenantFromLive);
  return {items, total: items.length};
}

async function listUsersLive(): Promise<PagedResponse<UserSummary>> {
  const tenants = await listTenantsLive();
  const pages = await Promise.all(
    tenants.items.map((tenant) =>
      apiClient.get<{users?: Record<string, unknown>[]}>(
        `/api/v1/tenants/${tenant.tenantId}/users`,
      ),
    ),
  );

  const items = pages.flatMap((page, index) =>
    (page.users ?? []).map((user) => mapUserFromLive(user, tenants.items[index]?.name)),
  );
  return {items, total: items.length};
}

async function listGroupsLive(): Promise<PagedResponse<GroupSummary>> {
  const tenants = await listTenantsLive();
  const pages = await Promise.all(
    tenants.items.map((tenant) =>
      apiClient.get<{groups?: Record<string, unknown>[]}>(
        `/api/v1/tenants/${tenant.tenantId}/groups`,
      ),
    ),
  );

  const items = pages.flatMap((page, index) =>
    (page.groups ?? []).map((group) => mapGroupFromLive(group, tenants.items[index]?.name)),
  );
  return {items, total: items.length};
}

async function listServiceAccountsLive(): Promise<PagedResponse<ServiceAccountDetail>> {
  const tenants = await listTenantsLive();
  const pages = await Promise.all(
    tenants.items.map((tenant) =>
      apiClient.get<{service_accounts?: Record<string, unknown>[]}>(
        `/api/v1/tenants/${tenant.tenantId}/service-accounts`,
      ),
    ),
  );

  const items = pages.flatMap((page, index) =>
    (page.service_accounts ?? []).map((item) =>
      mapServiceAccountFromLive(item, tenants.items[index]?.name),
    ),
  );
  return {items, total: items.length};
}

async function listOAuthClientsLive(): Promise<PagedResponse<OAuthClientDetail>> {
  const tenants = await listTenantsLive();
  const pages = await Promise.all(
    tenants.items.map((tenant) =>
      apiClient.get<{oauth_clients?: Record<string, unknown>[]}>(
        `/api/v1/tenants/${tenant.tenantId}/oauth-clients`,
      ),
    ),
  );

  const items = pages.flatMap((page, index) =>
    (page.oauth_clients ?? []).map((item) =>
      mapOAuthClientFromLive(item, tenants.items[index]?.name),
    ),
  );
  return {items, total: items.length};
}

async function listRolesLive(): Promise<PagedResponse<RoleDetail>> {
  const response = await apiClient.get<{roles?: Record<string, unknown>[]}>('/api/v1/roles');
  const items = (response.roles ?? []).map(mapRoleFromLive);
  return {items, total: items.length};
}

async function listSupportGrantsLive(): Promise<PagedResponse<SupportGrant>> {
  const response = await apiClient.get<{grants?: Record<string, unknown>[]}>(
    `/api/v1/tenants/${env.defaultTenantId}/support-grants`,
  );

  const items = (response.grants ?? []).map((grant) => {
    const subject = (grant.subject as Record<string, unknown> | undefined) ?? {};
    return {
      id: String(grant.support_grant_id ?? ''),
      grantId: String(grant.support_grant_id ?? ''),
      tenantId: String(grant.tenant_id ?? env.defaultTenantId),
      tenantName: db.tenants.find((tenant) => tenant.tenantId === String(grant.tenant_id))?.name ?? env.defaultTenantId,
      subjectId: String(subject.id ?? ''),
      subjectName: db.users.find((user) => user.userId === String(subject.id))?.name ?? String(subject.id ?? ''),
      scope: String(grant.role_id ?? ''),
      incidentId: String(grant.approval_ticket ?? 'INC-live'),
      roleId: String(grant.role_id ?? ''),
      reason: String(grant.reason ?? ''),
      expiresAt: String(grant.expires_at ?? new Date().toISOString()),
      status: String(grant.status ?? '').includes('EXPIRED')
        ? 'expired'
        : String(grant.status ?? '').includes('PENDING')
          ? 'pending'
          : 'active',
      requestedAt: String(grant.created_at ?? new Date().toISOString()),
      approvedAt: grant.approved_at ? String(grant.approved_at) : undefined,
    } satisfies SupportGrant;
  });

  return {items, total: items.length};
}

async function listAuditEventsLive(): Promise<PagedResponse<AuditEvent>> {
  const response = await apiClient.get<{events?: Record<string, unknown>[]}>(
    `/api/v1/tenants/${env.defaultTenantId}/audit-events`,
  );

  const items = (response.events ?? []).map((event) => ({
    id: String(event.audit_event_id ?? ''),
    eventId: String(event.audit_event_id ?? ''),
    tenantId: String(event.tenant_id ?? env.defaultTenantId),
    eventType: String(event.event_type ?? ''),
    actor: String(event.actor ?? ''),
    resource: String((event.resource as Record<string, unknown> | undefined)?.id ?? ''),
    result: 'OK',
    reason: String(event.reason ?? ''),
    occurredAt: String(event.occurred_at ?? new Date().toISOString()),
    payload: event,
  }));

  return {items, total: items.length};
}

async function listOperationsLive(): Promise<PagedResponse<Operation>> {
  const response = await apiClient.get<{operations?: Record<string, unknown>[]}>(
    `/api/v1/tenants/${env.defaultTenantId}/operations`,
  );

  const items: Operation[] = (response.operations ?? []).map((operation) => ({
    id: String(operation.operation_id ?? ''),
    operationId: String(operation.operation_id ?? ''),
    tenantId: String(operation.tenant_id ?? env.defaultTenantId),
    type: String(operation.operation_type ?? ''),
    resource: `${String(operation.resource_type ?? 'resource')}/${String(operation.resource_id ?? '')}`,
    actor: String(operation.correlation_id ?? 'system'),
    status: (
      String(operation.status ?? '').includes('FAILED')
        ? 'failed'
        : String(operation.status ?? '').includes('SUCCEEDED')
          ? 'done'
          : String(operation.status ?? '').includes('RUNNING')
            ? 'running'
            : 'pending'
    ) as Operation['status'],
    startedAt: String(operation.created_at ?? new Date().toISOString()),
    updatedAt: String(operation.updated_at ?? new Date().toISOString()),
    completedAt: operation.completed_at ? String(operation.completed_at) : undefined,
    errorMessage: operation.error_message ? String(operation.error_message) : undefined,
    steps: db.operations.find((item) => item.operationId === String(operation.operation_id))?.steps ?? [],
    logs: db.operations.find((item) => item.operationId === String(operation.operation_id))?.logs ?? [],
  }));

  return {items, total: items.length};
}

async function getResourceBindingsLive(resourceType: string, resourceId: string): Promise<ResourceBinding[]> {
  const response = await apiClient.post<{bindings?: Record<string, unknown>[]}>(
    '/api/v1/graph/resource-subjects:list',
    {
      resource: {
        type: resourceType.toUpperCase() === 'TENANT' ? 'RESOURCE_TYPE_TENANT' : 'RESOURCE_TYPE_PROJECT',
        id: resourceId,
        tenant_id: env.defaultTenantId,
      },
    },
  );

  return (response.bindings ?? []).map((binding) => {
    const subject = (binding.subject as Record<string, unknown> | undefined) ?? {};
    return {
      id: String(binding.binding_id ?? ''),
      resourceType: resourceType as ResourceBinding['resourceType'],
      resourceId,
      tenantId: String((binding.resource as Record<string, unknown> | undefined)?.tenant_id ?? env.defaultTenantId),
      subjectType: mapLiveSubjectType(String(subject.type ?? 'SUBJECT_TYPE_USER_ACCOUNT')),
      subjectId: String(subject.id ?? ''),
      subjectName: db.users.find((user) => user.userId === String(subject.id))?.name ??
        db.groups.find((group) => group.groupId === String(subject.id))?.name ??
        db.serviceAccounts.find((item) => item.serviceAccountId === String(subject.id))?.name ??
        String(subject.id ?? ''),
      roleId: String(binding.role_id ?? ''),
      source: 'direct',
      version: 1,
      expiresAt: binding.expires_at ? String(binding.expires_at) : undefined,
    };
  });
}

async function explainAccessLive(subjectId: string, resourceId: string, permission: string): Promise<ExplainAccessResult> {
  const response = await apiClient.post<Record<string, unknown>>('/api/v1/authz/explain', {
    subject: {type: 'SUBJECT_TYPE_USER_ACCOUNT', id: subjectId, tenant_id: env.defaultTenantId},
    resource: {type: 'RESOURCE_TYPE_PROJECT', id: resourceId, tenant_id: env.defaultTenantId},
    permission,
  });

  return {
    subjectId,
    resourceId,
    permission,
    decision: String(response.decision ?? '').includes('ALLOW') ? 'allow' : 'deny',
    evaluatedAt: new Date().toISOString(),
    summary: String(response.summary ?? ''),
    pathIds: Array.isArray(response.path_ids) ? response.path_ids.map(String) : [],
    steps: [
      {
        id: 'live-1',
        title: 'Runtime explanation',
        details: [String(response.summary ?? 'Decision returned by IAM runtime')],
      },
    ],
  };
}

async function simulateImpactLive(resourceId: string, subjectId: string, roleId: string): Promise<ImpactRow[]> {
  const response = await apiClient.post<{impacts?: Record<string, unknown>[]}>(
    '/api/v1/graph/change-impact:simulate',
    {
      resource: {type: 'RESOURCE_TYPE_PROJECT', id: resourceId, tenant_id: env.defaultTenantId},
      delta: {
        mutations: [
          {
            kind: 'BINDING_MUTATION_KIND_REMOVE',
            binding: {
              binding_id: `${subjectId}:${roleId}:${resourceId}`,
              subject: {type: 'SUBJECT_TYPE_USER_ACCOUNT', id: subjectId, tenant_id: env.defaultTenantId},
              resource: {type: 'RESOURCE_TYPE_PROJECT', id: resourceId, tenant_id: env.defaultTenantId},
              role_id: roleId,
              reason: 'ui simulation',
            },
          },
        ],
      },
    },
  );

  return (response.impacts ?? []).map((impact, index) => ({
    id: `impact-${index}`,
    subjectId: String((impact.subject as Record<string, unknown> | undefined)?.id ?? subjectId),
    subjectName: db.users.find((user) => user.userId === subjectId)?.name ?? subjectId,
    before: 'allow',
    after: 'deny',
    status: 'warning',
    affectedPermissions: (impact.removed_permissions as string[] | undefined) ?? [],
  }));
}

function mapLiveSubjectType(value: string): ResourceBinding['subjectType'] {
  if (value.includes('SERVICE')) {
    return 'serviceAccount';
  }
  if (value.includes('GROUP')) {
    return 'group';
  }
  if (value.includes('FEDERATED')) {
    return 'federatedUser';
  }
  return 'userAccount';
}

export const repositories = {
  async getDashboard(): Promise<DashboardData> {
    await delay();
    return cloneValue(db.dashboard);
  },

  async listTenants(): Promise<PagedResponse<TenantSummary>> {
    return withFallback(
      listTenantsLive,
      () => mockList(db.tenants),
    );
  },

  async getTenant(tenantId: string): Promise<TenantDetail> {
    return withFallback(
      async () => {
        const tenant = await apiClient.get<Record<string, unknown>>(`/api/v1/tenants/${tenantId}`);
        return mapTenantFromLive(tenant);
      },
      async () => {
        await delay();
        const tenant = db.tenants.find((item) => item.tenantId === tenantId);
        if (!tenant) {
          throw new Error(`Tenant ${tenantId} not found`);
        }
        return cloneValue(tenant);
      },
    );
  },

  async createTenant(payload: TenantMutationPayload): Promise<TenantDetail> {
    return withFallback(
      async () => {
        const tenant = await apiClient.post<Record<string, unknown>>('/api/v1/tenants', {
          request_id: createId('req'),
          tenant_id: payload.tenantId,
          display_name: payload.displayName,
          external_ref: payload.externalRef,
          labels: buildTenantLabels(payload),
          performed_by: 'ui-admin',
        });
        return mapTenantFromLive(tenant);
      },
      async () => {
        await delay();
        if (db.tenants.some((item) => item.tenantId === payload.tenantId)) {
          throw new Error(`Tenant ${payload.tenantId} already exists`);
        }
        const tenant = buildTenantDetail(undefined, payload, new Date().toISOString());
        db.tenants.unshift(tenant);
        return cloneValue(tenant);
      },
    );
  },

  async updateTenant(payload: TenantMutationPayload): Promise<TenantDetail> {
    return withFallback(
      async () => {
        const tenant = await apiClient.patch<Record<string, unknown>>(`/api/v1/tenants/${payload.tenantId}`, {
          tenant: {
            tenant_id: payload.tenantId,
            display_name: payload.displayName,
            external_ref: payload.externalRef,
            labels: buildTenantLabels(payload),
          },
          update_mask: 'displayName,externalRef,labels',
          request_id: createId('req'),
          performed_by: 'ui-admin',
        });
        return mapTenantFromLive(tenant);
      },
      async () => {
        await delay();
        const existing = db.tenants.find((item) => item.tenantId === payload.tenantId);
        if (!existing) {
          throw new Error(`Tenant ${payload.tenantId} not found`);
        }
        const tenant = buildTenantDetail(existing, payload, new Date().toISOString());
        Object.assign(existing, tenant);
        upsertTenantReferences(payload.tenantId, payload.displayName);
        return cloneValue(existing);
      },
    );
  },

  async deleteTenant(tenantId: string, reason: string): Promise<{tenantId: string}> {
    return withFallback(
      async () => {
        await apiClient.delete(`/api/v1/tenants/${tenantId}`, {
          request_id: createId('req'),
          reason,
          performed_by: 'ui-admin',
        });
        return {tenantId};
      },
      async () => {
        await delay();
        if (!db.tenants.some((item) => item.tenantId === tenantId)) {
          throw new Error(`Tenant ${tenantId} not found`);
        }
        deleteTenantReferences(tenantId);
        return {tenantId};
      },
    );
  },

  async listUsers(): Promise<PagedResponse<UserSummary>> {
    return withFallback(listUsersLive, () => mockList(db.users));
  },

  async getUser(userId: string): Promise<UserDetail> {
    return withFallback(
      async () => {
        const user = await apiClient.get<Record<string, unknown>>(`/api/v1/users/${userId}`);
        return mapUserFromLive(user);
      },
      async () => {
        await delay();
        const user = db.users.find((item) => item.userId === userId);
        if (!user) {
          throw new Error(`User ${userId} not found`);
        }
        return cloneValue(user);
      },
    );
  },

  async listGroups(): Promise<PagedResponse<GroupSummary>> {
    return withFallback(listGroupsLive, () => mockList(db.groups));
  },

  async getGroup(groupId: string): Promise<GroupDetail> {
    return withFallback(
      async () => {
        const group = await apiClient.get<Record<string, unknown>>(`/api/v1/groups/${groupId}`);
        return mapGroupFromLive(group);
      },
      async () => {
        await delay();
        const group = db.groups.find((item) => item.groupId === groupId);
        if (!group) {
          throw new Error(`Group ${groupId} not found`);
        }
        return cloneValue(group);
      },
    );
  },

  async addGroupMember(groupId: string, member: Omit<GroupMember, 'id' | 'addedAt'>): Promise<GroupDetail> {
    await delay();
    const group = db.groups.find((item) => item.groupId === groupId);
    if (!group) {
      throw new Error(`Group ${groupId} not found`);
    }
    group.members.push({
      ...member,
      id: createId(`${groupId}-member`),
      addedAt: new Date().toISOString(),
    });
    group.membersCount = group.members.length;
    return cloneValue(group);
  },

  async removeGroupMember(groupId: string, subjectId: string): Promise<GroupDetail> {
    await delay();
    const group = db.groups.find((item) => item.groupId === groupId);
    if (!group) {
      throw new Error(`Group ${groupId} not found`);
    }
    group.members = group.members.filter((member) => member.subjectId !== subjectId);
    group.membersCount = group.members.length;
    return cloneValue(group);
  },

  async listServiceAccounts(): Promise<PagedResponse<ServiceAccountDetail>> {
    return withFallback(listServiceAccountsLive, () => mockList(db.serviceAccounts));
  },

  async getServiceAccount(serviceAccountId: string): Promise<ServiceAccountDetail> {
    return withFallback(
      async () => {
        const account = await apiClient.get<Record<string, unknown>>(`/api/v1/service-accounts/${serviceAccountId}`);
        return mapServiceAccountFromLive(account);
      },
      async () => {
        await delay();
        const account = db.serviceAccounts.find((item) => item.serviceAccountId === serviceAccountId);
        if (!account) {
          throw new Error(`Service account ${serviceAccountId} not found`);
        }
        return cloneValue(account);
      },
    );
  },

  async createServiceAccount(payload: {
    tenantId: string;
    displayName: string;
    description: string;
  }): Promise<ServiceAccountDetail> {
    return withFallback(
      async () => {
        const response = await apiClient.post<Record<string, unknown>>(
          `/api/v1/tenants/${payload.tenantId}/service-accounts`,
          {
            request_id: createId('req'),
            tenant_id: payload.tenantId,
            service_account_id: createId('sa'),
            display_name: payload.displayName,
            description: payload.description,
            performed_by: 'ui-admin',
          },
        );
        return mapServiceAccountFromLive(response);
      },
      async () => {
        await delay();
        const tenant = db.tenants.find((item) => item.tenantId === payload.tenantId);
        const account: ServiceAccountDetail = {
          id: createId('sa'),
          serviceAccountId: createId('sa'),
          tenantId: payload.tenantId,
          tenantName: tenant?.name ?? payload.tenantId,
          name: payload.displayName,
          description: payload.description,
          status: 'active',
          keysCount: 0,
          apiKeysCount: 0,
          updatedAt: new Date().toISOString(),
          ownerTeam: 'Platform',
          createdAt: new Date().toISOString(),
          tokens: [],
          accessSummary: [],
          asymmetricKeys: [],
          apiKeys: [],
        };
        db.serviceAccounts.unshift(account);
        return cloneValue(account);
      },
    );
  },

  async listOAuthClients(): Promise<PagedResponse<OAuthClientDetail>> {
    return withFallback(listOAuthClientsLive, () => mockList(db.oauthClients));
  },

  async getOAuthClient(clientId: string): Promise<OAuthClientDetail> {
    return withFallback(
      async () => {
        const client = await apiClient.get<Record<string, unknown>>(`/api/v1/oauth-clients/${clientId}`);
        return mapOAuthClientFromLive(client);
      },
      async () => {
        await delay();
        const client = db.oauthClients.find((item) => item.clientId === clientId);
        if (!client) {
          throw new Error(`OAuth client ${clientId} not found`);
        }
        return cloneValue(client);
      },
    );
  },

  async rotateClientSecret(clientId: string, note: string): Promise<OAuthClientDetail> {
    return withFallback(
      async () => {
        await apiClient.post(`/api/v1/oauth-clients/${clientId}:rotateSecret`, {
          request_id: createId('req'),
          reason: note,
          performed_by: 'ui-admin',
        });
        return this.getOAuthClient(clientId);
      },
      async () => {
        await delay();
        const client = db.oauthClients.find((item) => item.clientId === clientId);
        if (!client) {
          throw new Error(`OAuth client ${clientId} not found`);
        }
        client.secrets.unshift({
          id: createId('secret'),
          name: `rotation-${client.secrets.length + 1}`,
          status: 'active',
          createdAt: new Date().toISOString(),
          note,
        });
        client.secretsCount = client.secrets.length;
        return cloneValue(client);
      },
    );
  },

  async listRoles(): Promise<PagedResponse<RoleDetail>> {
    return withFallback(listRolesLive, () => mockList(db.roles));
  },

  async getRole(roleId: string): Promise<RoleDetail> {
    return withFallback(
      async () => {
        const role = await apiClient.get<Record<string, unknown>>(`/api/v1/roles/${roleId}`);
        return mapRoleFromLive(role);
      },
      async () => {
        await delay();
        const role = db.roles.find((item) => item.roleId === roleId);
        if (!role) {
          throw new Error(`Role ${roleId} not found`);
        }
        return cloneValue(role);
      },
    );
  },

  async listResourceBindings(resourceType: string, resourceId: string): Promise<ResourceBinding[]> {
    return withFallback(
      () => getResourceBindingsLive(resourceType, resourceId),
      async () => {
        await delay();
        return cloneValue(
          db.resourceBindings.filter(
            (binding) => binding.resourceType === resourceType && binding.resourceId === resourceId,
          ),
        );
      },
    );
  },

  async grantAccess(binding: Omit<ResourceBinding, 'id' | 'version'>): Promise<ResourceBinding[]> {
    return withFallback(
      async () => {
        await apiClient.post('/api/v1/authz/bindings:update', {
          request_id: createId('req'),
          resource: {
            type: binding.resourceType === 'tenant' ? 'RESOURCE_TYPE_TENANT' : 'RESOURCE_TYPE_PROJECT',
            id: binding.resourceId,
            tenant_id: binding.tenantId,
          },
          delta: {
            mutations: [
              {
                kind: 'BINDING_MUTATION_KIND_ADD',
                binding: {
                  binding_id: createId('bind'),
                  subject: {
                    type:
                      binding.subjectType === 'group'
                        ? 'SUBJECT_TYPE_GROUP'
                        : binding.subjectType === 'serviceAccount'
                          ? 'SUBJECT_TYPE_SERVICE_ACCOUNT'
                          : 'SUBJECT_TYPE_USER_ACCOUNT',
                    id: binding.subjectId,
                    tenant_id: binding.tenantId,
                  },
                  resource: {
                    type: binding.resourceType === 'tenant' ? 'RESOURCE_TYPE_TENANT' : 'RESOURCE_TYPE_PROJECT',
                    id: binding.resourceId,
                    tenant_id: binding.tenantId,
                  },
                  role_id: binding.roleId,
                  reason: 'granted from ui',
                },
              },
            ],
          },
          reason: 'granted from ui',
          performed_by: 'ui-admin',
        });
        return this.listResourceBindings(binding.resourceType, binding.resourceId);
      },
      async () => {
        await delay();
        const created: ResourceBinding = {...binding, id: createId('bind'), version: 43};
        db.resourceBindings.unshift(created);
        return cloneValue(
          db.resourceBindings.filter(
            (item) => item.resourceType === binding.resourceType && item.resourceId === binding.resourceId,
          ),
        );
      },
    );
  },

  async explainAccess(subjectId: string, resourceId: string, permission: string): Promise<ExplainAccessResult> {
    return withFallback(
      () => explainAccessLive(subjectId, resourceId, permission),
      async () => {
        await delay();
        return {
          subjectId,
          resourceId,
          permission,
          decision: 'allow',
          evaluatedAt: new Date().toISOString(),
          summary: `subject ${subjectId} has ${permission} via project-editor`,
          pathIds: ['bind-demo-ops-project-editor'],
          steps: [
            {id: 'step-1', title: 'Group membership found', details: ['via_group_id = grp-demo-ops']},
            {id: 'step-2', title: 'Binding found', details: ['role_id = project-editor', 'source = direct']},
            {id: 'step-3', title: 'Role expansion', details: ['project-editor -> project.write']},
          ],
        };
      },
    );
  },

  async listEffectiveAccess(): Promise<EffectiveAccessRow[]> {
    await delay();
    return cloneValue(db.effectiveAccess);
  },

  async simulateImpact(resourceId: string, subjectId: string, roleId: string): Promise<ImpactRow[]> {
    return withFallback(
      () => simulateImpactLive(resourceId, subjectId, roleId),
      async () => {
        await delay();
        return [
          {
            id: 'impact-1',
            subjectId,
            subjectName: db.users.find((user) => user.userId === subjectId)?.name ?? subjectId,
            before: 'allow',
            after: 'deny',
            status: 'warning',
            affectedPermissions: ['project.read', 'project.write'],
          },
        ];
      },
    );
  },

  async listPolicyTemplates(): Promise<PagedResponse<PolicyTemplate>> {
    return mockList(db.policyTemplates);
  },

  async getPolicyTemplate(templateId: string): Promise<PolicyTemplate> {
    await delay();
    const template = db.policyTemplates.find((item) => item.templateId === templateId);
    if (!template) {
      throw new Error(`Policy template ${templateId} not found`);
    }
    return cloneValue(template);
  },

  async listSupportGrants(): Promise<PagedResponse<SupportGrant>> {
    return withFallback(listSupportGrantsLive, () => mockList(db.supportGrants));
  },

  async createSupportGrant(payload: {
    subjectId: string;
    tenantId: string;
    roleId: string;
    reason: string;
    incidentId: string;
    expiresAt: string;
  }): Promise<SupportGrant> {
    return withFallback<SupportGrant>(
      async () => {
        const response = await apiClient.post<Record<string, unknown>>(
          `/api/v1/tenants/${payload.tenantId}/support-grants`,
          {
            request_id: createId('req'),
            tenant_id: payload.tenantId,
            subject: {type: 'SUBJECT_TYPE_USER_ACCOUNT', id: payload.subjectId, tenant_id: payload.tenantId},
            resource: {type: 'RESOURCE_TYPE_TENANT', id: payload.tenantId, tenant_id: payload.tenantId},
            role_id: payload.roleId,
            ttl: '3600s',
            reason: payload.reason,
            requested_by: 'ui-admin',
          },
        );
        const grantId = String(response.support_grant_id ?? createId('sg'));
        return {
          id: grantId,
          grantId,
          tenantId: payload.tenantId,
          tenantName: db.tenants.find((tenant) => tenant.tenantId === payload.tenantId)?.name ?? payload.tenantId,
          subjectId: payload.subjectId,
          subjectName: db.users.find((user) => user.userId === payload.subjectId)?.name ?? payload.subjectId,
          scope: payload.roleId,
          incidentId: payload.incidentId,
          roleId: payload.roleId,
          reason: payload.reason,
          expiresAt: payload.expiresAt,
          status: 'pending',
          requestedAt: String(response.created_at ?? new Date().toISOString()),
        };
      },
      async () => {
        await delay();
        const grant: SupportGrant = {
          id: createId('sg'),
          grantId: createId('sg'),
          tenantId: payload.tenantId,
          tenantName: db.tenants.find((tenant) => tenant.tenantId === payload.tenantId)?.name ?? payload.tenantId,
          subjectId: payload.subjectId,
          subjectName: db.users.find((user) => user.userId === payload.subjectId)?.name ?? payload.subjectId,
          scope: payload.roleId,
          incidentId: payload.incidentId,
          roleId: payload.roleId,
          reason: payload.reason,
          expiresAt: payload.expiresAt,
          status: 'pending',
          requestedAt: new Date().toISOString(),
        };
        db.supportGrants.unshift(grant);
        return cloneValue(grant);
      },
    );
  },

  async listAuditEvents(): Promise<PagedResponse<AuditEvent>> {
    return withFallback(listAuditEventsLive, () => mockList(db.auditEvents));
  },

  async getAuditEvent(eventId: string): Promise<AuditEvent> {
    return withFallback(
      async () => {
        const event = await apiClient.get<Record<string, unknown>>(`/api/v1/audit-events/${eventId}`);
        return {
          id: String(event.audit_event_id ?? eventId),
          eventId: String(event.audit_event_id ?? eventId),
          tenantId: String(event.tenant_id ?? env.defaultTenantId),
          eventType: String(event.event_type ?? ''),
          actor: String(event.actor ?? ''),
          resource: String((event.resource as Record<string, unknown> | undefined)?.id ?? ''),
          result: 'OK',
          reason: String(event.reason ?? ''),
          occurredAt: String(event.occurred_at ?? new Date().toISOString()),
          payload: event,
        };
      },
      async () => {
        await delay();
        const event = db.auditEvents.find((item) => item.eventId === eventId);
        if (!event) {
          throw new Error(`Audit event ${eventId} not found`);
        }
        return cloneValue(event);
      },
    );
  },

  async listOperations(): Promise<PagedResponse<Operation>> {
    return withFallback(listOperationsLive, () => mockList(db.operations));
  },

  async getOperation(operationId: string): Promise<Operation> {
    return withFallback(
      async () => {
        const operation = await apiClient.get<Record<string, unknown>>(`/api/v1/operations/${operationId}`);
        const mapped = (await listOperationsLive()).items.find((item) => item.operationId === String(operation.operation_id));
        if (!mapped) {
          throw new Error(`Operation ${operationId} not found`);
        }
        return mapped;
      },
      async () => {
        await delay();
        const operation = db.operations.find((item) => item.operationId === operationId);
        if (!operation) {
          throw new Error(`Operation ${operationId} not found`);
        }
        return cloneValue(operation);
      },
    );
  },

  async listSearchResults(query: string): Promise<PagedResponse<SearchResult>> {
    await delay();
    const normalized = query.trim().toLowerCase();
    const filtered = db.searchResults.filter((item) => {
      return [item.title, item.type, item.context, item.description].some((value) =>
        value.toLowerCase().includes(normalized),
      );
    });
    return {items: cloneValue(filtered), total: filtered.length};
  },

  async listSettings(): Promise<SettingsSection[]> {
    await delay();
    return cloneValue(db.settings);
  },

  async saveSettings(nextSettings: SettingsSection[]): Promise<SettingsSection[]> {
    await delay();
    db.settings.splice(0, db.settings.length, ...cloneValue(nextSettings));
    return cloneValue(db.settings);
  },

  async listSessions(): Promise<PagedResponse<TokenInfo>> {
    await delay();
    const allTokens = db.users.flatMap((user) => user.tokens).concat(db.serviceAccounts.flatMap((item) => item.tokens));
    return {items: cloneValue(allTokens), total: allTokens.length};
  },
};
