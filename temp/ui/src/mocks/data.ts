import type {
  ApiKey,
  AsymmetricKey,
  AuditEvent,
  DashboardData,
  EffectiveAccessRow,
  GroupDetail,
  OAuthClientDetail,
  Operation,
  PolicyTemplate,
  ResourceBinding,
  RoleDetail,
  SearchResult,
  ServiceAccountDetail,
  SettingsSection,
  SupportGrant,
  TenantDetail,
  TokenInfo,
  UserDetail,
} from '@/shared/types/iam';

type MockDatabase = {
  dashboard: DashboardData;
  tenants: TenantDetail[];
  users: UserDetail[];
  groups: GroupDetail[];
  serviceAccounts: ServiceAccountDetail[];
  oauthClients: OAuthClientDetail[];
  roles: RoleDetail[];
  resourceBindings: ResourceBinding[];
  effectiveAccess: EffectiveAccessRow[];
  policyTemplates: PolicyTemplate[];
  supportGrants: SupportGrant[];
  auditEvents: AuditEvent[];
  operations: Operation[];
  searchResults: SearchResult[];
  settings: SettingsSection[];
};

const now = '2026-04-04T14:00:00Z';

const apiKeys: ApiKey[] = [
  {
    id: 'ak-demo-1',
    prefix: 'ak_live_9x',
    status: 'active',
    createdAt: '2026-03-03T11:44:00Z',
    lastUsedAt: '2026-04-03T09:20:00Z',
  },
];

const asymmetricKeys: AsymmetricKey[] = [
  {
    id: 'key-demo-11',
    algorithm: 'RSA-2048',
    status: 'active',
    createdAt: '2026-03-01T10:00:00Z',
    lastUsedAt: '2026-04-03T12:28:00Z',
  },
  {
    id: 'key-demo-09',
    algorithm: 'RSA-2048',
    status: 'active',
    createdAt: '2026-01-15T10:00:00Z',
    lastUsedAt: '2026-04-02T17:11:00Z',
  },
];

const tokens: TokenInfo[] = [
  {
    id: 'rt-demo-11',
    client: 'Admin Portal',
    type: 'refresh',
    protectionLevel: 'SECURE_KEY_DPOP',
    expiresAt: '2026-05-01T00:00:00Z',
    lastUsedAt: '2026-04-03T12:31:00Z',
    status: 'active',
  },
  {
    id: 'at-demo-11',
    client: 'Admin Portal',
    type: 'access',
    protectionLevel: 'SESSION_BOUND',
    expiresAt: '2026-04-04T15:10:00Z',
    lastUsedAt: '2026-04-04T14:01:00Z',
    status: 'active',
  },
];

const users: UserDetail[] = [
  {
    id: 'user-demo-admin',
    userId: 'user-demo-admin',
    tenantIds: ['tenant-demo', 'tenant-sandbox'],
    name: 'Ivan Petrov',
    email: 'ivan@acme.io',
    source: 'SSO',
    status: 'active',
    mfaEnabled: true,
    lastLoginAt: '2026-04-03T12:31:00Z',
    memberships: [
      {id: 'mship-demo-admin', tenantId: 'tenant-demo', tenantName: 'Acme Production', roleIds: ['tenant-admin'], status: 'active'},
      {id: 'mship-sandbox-owner', tenantId: 'tenant-sandbox', tenantName: 'Acme Sandbox', roleIds: ['tenant-owner'], status: 'active'},
    ],
    groups: [],
    accessSummary: [],
    sessions: [
      {
        id: 'session-demo-1',
        client: 'Admin Portal',
        device: 'Safari / macOS',
        ipAddress: '10.20.30.40',
        lastSeenAt: '2026-04-03T12:31:00Z',
        protectionLevel: 'MFA',
        status: 'active',
      },
    ],
    tokens,
    labels: {persona: 'platform-admin'},
  },
  {
    id: 'user-demo-analyst',
    userId: 'user-demo-analyst',
    tenantIds: ['tenant-demo'],
    name: 'Anna Volkova',
    email: 'anna@acme.io',
    source: 'Local',
    status: 'active',
    mfaEnabled: false,
    lastLoginAt: '2026-04-03T11:03:00Z',
    memberships: [
      {id: 'mship-demo-analyst', tenantId: 'tenant-demo', tenantName: 'Acme Production', roleIds: ['project-viewer'], status: 'active'},
    ],
    groups: [],
    accessSummary: [],
    sessions: [],
    tokens: [],
    labels: {persona: 'analyst'},
  },
  {
    id: 'user-demo-support',
    userId: 'user-demo-support',
    tenantIds: ['tenant-demo'],
    name: 'Support Engineer',
    email: 'support@m8.team',
    source: 'Internal',
    status: 'active',
    mfaEnabled: true,
    lastLoginAt: '2026-04-03T12:28:00Z',
    memberships: [
      {id: 'mship-demo-support', tenantId: 'tenant-demo', tenantName: 'Acme Production', roleIds: ['support-operator'], status: 'active'},
    ],
    groups: [],
    accessSummary: [],
    sessions: [],
    tokens: [],
    labels: {persona: 'support'},
  },
];

const groups: GroupDetail[] = [
  {
    id: 'grp-demo-finance',
    groupId: 'grp-demo-finance',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    name: 'Finance Admins',
    description: 'Finance admins with access to invoices and exports',
    dynamic: false,
    membersCount: 2,
    status: 'active',
    updatedAt: now,
    rules: ['manual assignments only'],
    members: [
      {id: 'grp-demo-finance:user-demo-analyst', subjectId: 'user-demo-analyst', subjectType: 'userAccount', displayName: 'Anna Volkova', addedAt: '2026-03-10T10:11:00Z'},
      {id: 'grp-demo-finance:sa-demo-bot', subjectId: 'sa-demo-bot', subjectType: 'serviceAccount', displayName: 'Demo Bot', addedAt: '2026-03-10T10:12:00Z'},
    ],
    effectiveGrantCount: 8,
    effectiveRoleCount: 2,
  },
  {
    id: 'grp-demo-ops',
    groupId: 'grp-demo-ops',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    name: 'Operations',
    description: 'Operations engineers and automation',
    dynamic: false,
    membersCount: 2,
    status: 'active',
    updatedAt: now,
    rules: ['manual assignments only', 'must be reviewed quarterly'],
    members: [
      {id: 'grp-demo-ops:user-demo-admin', subjectId: 'user-demo-admin', subjectType: 'userAccount', displayName: 'Ivan Petrov', addedAt: '2026-03-10T10:11:00Z'},
      {id: 'grp-demo-ops:sa-demo-bot', subjectId: 'sa-demo-bot', subjectType: 'serviceAccount', displayName: 'Demo Bot', addedAt: '2026-03-10T10:12:00Z'},
    ],
    effectiveGrantCount: 14,
    effectiveRoleCount: 3,
  },
];

const serviceAccounts: ServiceAccountDetail[] = [
  {
    id: 'sa-demo-bot',
    serviceAccountId: 'sa-demo-bot',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    name: 'billing-worker',
    description: 'Worker for invoice sync and exports',
    status: 'active',
    keysCount: 2,
    apiKeysCount: 1,
    lastAuthAt: '2026-04-03T12:28:00Z',
    updatedAt: now,
    ownerTeam: 'Billing',
    createdAt: '2026-02-02T09:00:00Z',
    tokens,
    accessSummary: [],
    asymmetricKeys,
    apiKeys,
  },
  {
    id: 'sa-sandbox-ci',
    serviceAccountId: 'sa-sandbox-ci',
    tenantId: 'tenant-sandbox',
    tenantName: 'Acme Sandbox',
    name: 'sandbox-ci',
    description: 'CI automation account',
    status: 'paused',
    keysCount: 1,
    apiKeysCount: 0,
    lastAuthAt: '2026-04-02T19:01:00Z',
    updatedAt: '2026-04-03T10:02:00Z',
    ownerTeam: 'Platform',
    createdAt: '2026-01-20T10:00:00Z',
    tokens: [],
    accessSummary: [],
    asymmetricKeys: [asymmetricKeys[0]],
    apiKeys: [],
  },
];

const oauthClients: OAuthClientDetail[] = [
  {
    id: 'client-demo-admin-ui',
    clientId: 'client-demo-admin-ui',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    name: 'Admin Portal',
    type: 'confidential',
    status: 'active',
    redirectUrisCount: 2,
    scopesCount: 4,
    secretsCount: 2,
    updatedAt: now,
    createdAt: '2026-02-12T12:00:00Z',
    redirectUris: ['https://admin.acme.io/callback', 'http://localhost:3000/callback'],
    scopes: ['openid', 'profile', 'email', 'offline_access'],
    secrets: [
      {id: 'secret-demo-1', name: 'prod-secret-1', status: 'active', createdAt: '2026-02-12T12:01:00Z', expiresAt: '2026-10-01T00:00:00Z'},
      {id: 'secret-demo-2', name: 'prod-secret-2', status: 'active', createdAt: '2026-03-20T13:10:00Z', expiresAt: '2026-10-01T00:00:00Z', note: 'Routine rotation'},
    ],
    tokens,
  },
  {
    id: 'client-demo-cli',
    clientId: 'client-demo-cli',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    name: 'Demo CLI',
    type: 'public',
    status: 'paused',
    redirectUrisCount: 1,
    scopesCount: 2,
    secretsCount: 0,
    updatedAt: '2026-04-03T10:07:00Z',
    createdAt: '2026-01-10T12:00:00Z',
    redirectUris: ['http://localhost:3001/callback'],
    scopes: ['openid', 'email'],
    secrets: [],
    tokens: [],
  },
];

const roles: RoleDetail[] = [
  {
    id: 'tenant-owner',
    roleId: 'tenant-owner',
    namespace: 'tenant',
    name: 'Tenant Owner',
    description: 'Full tenant management access.',
    permissionsCount: 3,
    system: true,
    permissions: [
      {id: 'tenant.manage', displayName: 'Manage tenant'},
      {id: 'project.read', displayName: 'Read project'},
      {id: 'project.write', displayName: 'Write project'},
    ],
    usedBy: ['tenant-demo / user-demo-admin', 'tenant-sandbox / user-sandbox-owner'],
  },
  {
    id: 'tenant-admin',
    roleId: 'tenant-admin',
    namespace: 'tenant',
    name: 'Tenant Admin',
    description: 'Administrative access for tenant resources.',
    permissionsCount: 2,
    system: true,
    permissions: [
      {id: 'tenant.manage', displayName: 'Manage tenant'},
      {id: 'project.read', displayName: 'Read project'},
    ],
    usedBy: ['tenant-demo / user-demo-admin'],
  },
  {
    id: 'project-editor',
    roleId: 'project-editor',
    namespace: 'project',
    name: 'Project Editor',
    description: 'Read-write project access.',
    permissionsCount: 2,
    system: false,
    permissions: [
      {id: 'project.read', displayName: 'Read project'},
      {id: 'project.write', displayName: 'Write project'},
    ],
    usedBy: ['project-demo-infra / grp-demo-ops', 'project-demo-infra / sa-demo-bot'],
  },
  {
    id: 'project-viewer',
    roleId: 'project-viewer',
    namespace: 'project',
    name: 'Project Viewer',
    description: 'Read-only project access.',
    permissionsCount: 1,
    system: false,
    permissions: [{id: 'project.read', displayName: 'Read project'}],
    usedBy: ['project-demo-analytics / user-demo-analyst'],
  },
];

const resourceBindings: ResourceBinding[] = [
  {
    id: 'bind-demo-admin-tenant',
    resourceType: 'tenant',
    resourceId: 'tenant-demo',
    tenantId: 'tenant-demo',
    subjectType: 'userAccount',
    subjectId: 'user-demo-admin',
    subjectName: 'Ivan Petrov',
    roleId: 'tenant-admin',
    source: 'direct',
    version: 42,
  },
  {
    id: 'bind-demo-ops-project-editor',
    resourceType: 'project',
    resourceId: 'project-demo-infra',
    tenantId: 'tenant-demo',
    subjectType: 'group',
    subjectId: 'grp-demo-ops',
    subjectName: 'Operations',
    roleId: 'project-editor',
    source: 'group',
    version: 42,
  },
  {
    id: 'bind-demo-bot-project-editor',
    resourceType: 'project',
    resourceId: 'project-demo-infra',
    tenantId: 'tenant-demo',
    subjectType: 'serviceAccount',
    subjectId: 'sa-demo-bot',
    subjectName: 'billing-worker',
    roleId: 'project-editor',
    source: 'direct',
    version: 42,
  },
];

const effectiveAccess: EffectiveAccessRow[] = [
  {
    id: 'eff-1',
    subjectId: 'sa-demo-bot',
    subjectName: 'billing-worker',
    subjectType: 'serviceAccount',
    resourceType: 'project',
    resourceId: 'project-demo-infra',
    roleId: 'project-editor',
    permission: 'project.write',
    source: 'direct',
    decision: 'allow',
  },
  {
    id: 'eff-2',
    subjectId: 'user-demo-analyst',
    subjectName: 'Anna Volkova',
    subjectType: 'userAccount',
    resourceType: 'project',
    resourceId: 'project-demo-analytics',
    roleId: 'project-viewer',
    permission: 'project.read',
    source: 'direct',
    decision: 'allow',
  },
];

const tenants: TenantDetail[] = [
  {
    id: 'tenant-demo',
    tenantId: 'tenant-demo',
    name: 'Acme Production',
    organizationId: 'org-1',
    plan: 'Enterprise',
    region: 'eu-central',
    status: 'active',
    tags: ['gold', 'sso', 'scim'],
    memberCount: 412,
    groupCount: 12,
    serviceAccountCount: 24,
    oauthClientCount: 8,
    updatedAt: now,
    ownersCount: 3,
    createdAt: '2026-01-12T09:00:00Z',
    externalRef: 'crm-demo-001',
    description: 'Production tenant for Acme.',
    summary: [
      'Uses SSO + SCIM',
      'Billing account bound: bill-7',
      'Primary region: eu-central',
    ],
    integrations: ['Keycloak', 'SpiceDB', 'SCIM bridge'],
    resourceMap: ['tenant-demo', 'project/project-demo-infra', 'project/project-demo-analytics', 'support-case/case-demo-001', 'env/prod', 'env/staging'],
  },
  {
    id: 'tenant-sandbox',
    tenantId: 'tenant-sandbox',
    name: 'Acme Sandbox',
    organizationId: 'org-1',
    plan: 'Trial',
    region: 'us-east',
    status: 'trial',
    tags: ['trial', 'sandbox'],
    memberCount: 24,
    groupCount: 4,
    serviceAccountCount: 3,
    oauthClientCount: 2,
    updatedAt: '2026-04-03T11:52:00Z',
    ownersCount: 1,
    createdAt: '2026-02-08T09:00:00Z',
    externalRef: 'crm-sandbox-002',
    description: 'Sandbox tenant for integration tests and rehearsals.',
    summary: ['Reduced SLA', 'No SCIM', 'Shorter token policies'],
    integrations: ['Keycloak'],
    resourceMap: ['tenant-sandbox', 'project/project-sandbox-api', 'env/dev'],
  },
];

const supportGrants: SupportGrant[] = [
  {
    id: 'sg-11',
    grantId: 'sg-11',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    subjectId: 'user-demo-support',
    subjectName: 'Alex Support',
    scope: 'readonly',
    incidentId: 'INC-1901',
    roleId: 'support-operator',
    reason: 'customer reports billing issue',
    expiresAt: '2026-04-04T23:00:00Z',
    status: 'active',
    requestedAt: '2026-04-04T18:00:00Z',
    approvedAt: '2026-04-04T18:05:00Z',
  },
  {
    id: 'sg-12',
    grantId: 'sg-12',
    tenantId: 'tenant-demo',
    tenantName: 'Acme Production',
    subjectId: 'user-demo-support',
    subjectName: 'Nina Support',
    scope: 'limited-admin',
    incidentId: 'INC-1888',
    roleId: 'support-operator',
    reason: 'expired support window',
    expiresAt: '2026-04-03T18:00:00Z',
    status: 'expired',
    requestedAt: '2026-04-03T14:00:00Z',
    approvedAt: '2026-04-03T14:10:00Z',
  },
];

const auditEvents: AuditEvent[] = [
  {
    id: 'evt-8d0d2c9e',
    eventId: 'evt-8d0d2c9e',
    tenantId: 'tenant-demo',
    eventType: 'binding.added',
    actor: 'userAccount/u-admin',
    resource: 'project/project-demo-infra',
    result: 'OK',
    reason: 'analytics onboarding',
    occurredAt: '2026-04-03T12:41:02Z',
    payload: {
      action: 'ADD',
      role_id: 'project-viewer',
      subject: {type: 'userAccount', id: 'user-demo-admin'},
      reason: 'analytics onboarding',
    },
  },
  {
    id: 'evt-rotate-secret',
    eventId: 'evt-rotate-secret',
    tenantId: 'tenant-demo',
    eventType: 'oauth.secret.rotated',
    actor: 'userAccount/u-admin',
    resource: 'oauthClient/client-demo-admin-ui',
    result: 'OK',
    occurredAt: '2026-04-03T12:39:00Z',
    payload: {secret_name: 'prod-secret-2', note: 'Routine rotation'},
  },
];

const operations: Operation[] = [
  {
    id: 'op-11',
    operationId: 'op-11',
    tenantId: 'tenant-demo',
    type: 'updateAccessBindings',
    resource: 'project/project-demo-infra',
    actor: 'u-admin',
    status: 'running',
    startedAt: '2026-04-03T12:40:00Z',
    updatedAt: '2026-04-03T12:40:12Z',
    steps: [
      {id: 'step-1', title: 'Validate request', status: 'done'},
      {id: 'step-2', title: 'Write canonical bindings', status: 'done'},
      {id: 'step-3', title: 'Publish domain event', status: 'done'},
      {id: 'step-4', title: 'Update access graph projections', status: 'running'},
      {id: 'step-5', title: 'Complete audit trail', status: 'pending'},
    ],
    logs: [
      '12:40:11 validated input',
      '12:40:11 bindings committed version=42',
      '12:40:11 event published topic=iam.authz.relationships.v1',
      '12:40:12 access graph rebuild started',
    ],
  },
  {
    id: 'op-sa-seed-demo-bot',
    operationId: 'op-sa-seed-demo-bot',
    tenantId: 'tenant-demo',
    type: 'createServiceAccount',
    resource: 'serviceAccount/sa-demo-bot',
    actor: 'u-admin',
    status: 'done',
    startedAt: '2026-04-03T12:20:00Z',
    updatedAt: '2026-04-03T12:22:00Z',
    completedAt: '2026-04-03T12:22:00Z',
    steps: [
      {id: 'step-1', title: 'Provision client', status: 'done'},
      {id: 'step-2', title: 'Persist metadata', status: 'done'},
      {id: 'step-3', title: 'Publish event', status: 'done'},
    ],
    logs: ['12:20:00 created Keycloak client', '12:20:01 persisted service account'],
  },
];

const policyTemplates: PolicyTemplate[] = [
  {
    id: 'support.session.v1',
    templateId: 'support.session.v1',
    name: 'Temporary support session',
    scope: 'tenant',
    status: 'active',
    parameters: ['expires_at', 'incident_id', 'scope'],
    description: 'Temporary support access with mandatory expiration and incident reason.',
    generatedBindings: ['support-operator', 'project-viewer', 'audit.read'],
  },
  {
    id: 'finance.readonly.v1',
    templateId: 'finance.readonly.v1',
    name: 'Finance readonly baseline',
    scope: 'billing',
    status: 'active',
    parameters: ['tenant_id'],
    description: 'Read-only finance operations baseline.',
    generatedBindings: ['billing.reader'],
  },
];

const searchResults: SearchResult[] = [
  {
    id: 'result-role',
    type: 'Role',
    title: 'project-editor',
    context: 'Roles catalog',
    description: 'Read-write project access',
    href: '/roles/project-editor',
  },
  {
    id: 'result-group',
    type: 'Group',
    title: 'Operations',
    context: 'tenant-demo',
    description: 'Operations engineers and automation',
    href: '/groups/grp-demo-ops',
  },
  {
    id: 'result-audit',
    type: 'Audit Event',
    title: 'binding.added',
    context: 'evt-8d0d2c9e',
    description: 'Analytics onboarding access grant',
    href: '/audit/evt-8d0d2c9e',
  },
];

const settings: SettingsSection[] = [
  {
    id: 'token-policies',
    title: 'Token Policies',
    description: 'Global defaults for access and refresh token handling.',
    fields: [
      {id: 'access-token-ttl', label: 'Access token TTL', value: '1h', type: 'select', options: ['15m', '30m', '1h', '4h']},
      {id: 'refresh-token-ttl', label: 'Refresh token TTL', value: '30d', type: 'select', options: ['7d', '14d', '30d', '90d']},
      {id: 'revoke-on-password-reset', label: 'Revoke on password reset', value: true, type: 'switch'},
      {id: 'require-dpop', label: 'Require DPoP for public apps', value: true, type: 'switch'},
    ],
  },
  {
    id: 'audit-retention',
    title: 'Audit Retention',
    description: 'Audit trail storage and notification defaults.',
    fields: [
      {id: 'audit-retention-days', label: 'Audit retention', value: '365 days', type: 'select', options: ['90 days', '180 days', '365 days']},
      {id: 'secret-reminder', label: 'Secret rotation reminder', value: '30 days before expiry', type: 'select', options: ['7 days before expiry', '14 days before expiry', '30 days before expiry']},
    ],
  },
];

const dashboard: DashboardData = {
  metrics: [
    {id: 'subjects', title: 'Subjects', value: '12 482', delta: '+3.2%', tone: 'success', description: 'Users, groups and service accounts across current org.'},
    {id: 'sessions', title: 'Active Sessions', value: '1 245', delta: '+64', tone: 'info', description: 'Interactive sessions in the last 24 hours.'},
    {id: 'changes', title: 'Access Changes 24h', value: '86', delta: '+12', tone: 'warning', description: 'Direct binding and role updates.'},
    {id: 'failed-checks', title: 'Failed Checks', value: '14', delta: '-2', tone: 'danger', description: 'Denied or conditional runtime checks.'},
  ],
  quickActions: [
    {id: 'create-user', title: 'Create User', href: '/users'},
    {id: 'create-group', title: 'Create Group', href: '/groups'},
    {id: 'create-sa', title: 'Create Service Account', href: '/service-accounts'},
    {id: 'grant-access', title: 'Grant Access', href: '/access/resources/project/project-demo-infra'},
    {id: 'support-access', title: 'Support Access', href: '/support-access'},
    {id: 'explain-access', title: 'Explain Access', href: '/access/explain'},
  ],
  recentActivity: [
    {id: 'act-1', time: '12:41', title: 'Added project.viewer to user-demo-admin', description: 'project/project-demo-infra', tone: 'success', href: '/audit/evt-8d0d2c9e'},
    {id: 'act-2', time: '12:39', title: 'Rotated secret for client-demo-admin-ui', description: 'Routine rotation', tone: 'info', href: '/oauth-clients/client-demo-admin-ui'},
    {id: 'act-3', time: '12:33', title: 'Revoked refresh token for client-demo-cli', description: 'Security cleanup', tone: 'warning', href: '/sessions'},
  ],
};

const database: MockDatabase = {
  dashboard,
  tenants,
  users,
  groups,
  serviceAccounts,
  oauthClients,
  roles,
  resourceBindings,
  effectiveAccess,
  policyTemplates,
  supportGrants,
  auditEvents,
  operations,
  searchResults,
  settings,
};

export function getMockDatabase(): MockDatabase {
  return database;
}

export function cloneValue<T>(value: T): T {
  return structuredClone(value);
}

export function delay(ms = 180): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export function createId(prefix: string): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}
