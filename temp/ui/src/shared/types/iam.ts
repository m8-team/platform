export type EntityStatus =
  | 'active'
  | 'paused'
  | 'suspended'
  | 'disabled'
  | 'pending'
  | 'running'
  | 'done'
  | 'failed'
  | 'expired'
  | 'revoked'
  | 'trial';

export type Severity = 'info' | 'success' | 'warning' | 'danger';
export type SubjectType = 'userAccount' | 'serviceAccount' | 'group' | 'federatedUser';
export type ResourceType = 'tenant' | 'project' | 'environment' | 'billing' | 'secret' | 'supportCase';
export type OperationStatus = 'pending' | 'running' | 'done' | 'failed';

export interface AppContextSelection {
  tenantId: string;
  organizationId: string;
  environment: string;
  region: string;
}

export interface BreadcrumbItem {
  label: string;
  href?: string;
}

export interface SummaryMetric {
  id: string;
  title: string;
  value: string;
  delta?: string;
  tone?: Severity;
  description?: string;
}

export interface QuickActionItem {
  id: string;
  title: string;
  description?: string;
  href?: string;
}

export interface ActivityItem {
  id: string;
  time: string;
  title: string;
  description: string;
  tone?: Severity;
  href?: string;
}

export interface TenantSummary {
  id: string;
  tenantId: string;
  name: string;
  organizationId: string;
  plan: string;
  region: string;
  status: EntityStatus;
  tags: string[];
  memberCount: number;
  groupCount: number;
  serviceAccountCount: number;
  oauthClientCount: number;
  updatedAt: string;
}

export interface TenantDetail extends TenantSummary {
  ownersCount: number;
  createdAt: string;
  externalRef?: string;
  description: string;
  summary: string[];
  integrations: string[];
  resourceMap: string[];
}

export interface UserSummary {
  id: string;
  userId: string;
  tenantIds: string[];
  name: string;
  email: string;
  source: string;
  status: EntityStatus;
  mfaEnabled: boolean;
  lastLoginAt: string;
  labels?: Record<string, string>;
}

export interface UserMembership {
  id: string;
  tenantId: string;
  tenantName: string;
  roleIds: string[];
  status: EntityStatus;
}

export interface SessionInfo {
  id: string;
  client: string;
  device: string;
  ipAddress: string;
  lastSeenAt: string;
  protectionLevel: string;
  status: EntityStatus;
}

export interface TokenInfo {
  id: string;
  client: string;
  type: 'access' | 'refresh';
  protectionLevel: string;
  expiresAt: string;
  lastUsedAt: string;
  status: EntityStatus;
}

export interface UserDetail extends UserSummary {
  memberships: UserMembership[];
  groups: GroupSummary[];
  accessSummary: EffectiveAccessRow[];
  sessions: SessionInfo[];
  tokens: TokenInfo[];
}

export interface GroupSummary {
  id: string;
  groupId: string;
  tenantId: string;
  tenantName: string;
  name: string;
  description: string;
  dynamic: boolean;
  membersCount: number;
  status: EntityStatus;
  updatedAt: string;
}

export interface GroupMember {
  id: string;
  subjectId: string;
  subjectType: SubjectType;
  displayName: string;
  addedAt: string;
}

export interface GroupDetail extends GroupSummary {
  rules: string[];
  members: GroupMember[];
  effectiveGrantCount: number;
  effectiveRoleCount: number;
}

export interface ApiKey {
  id: string;
  prefix: string;
  status: EntityStatus;
  createdAt: string;
  lastUsedAt: string;
}

export interface AsymmetricKey {
  id: string;
  algorithm: string;
  status: EntityStatus;
  createdAt: string;
  expiresAt?: string;
  lastUsedAt?: string;
}

export interface ServiceAccountSummary {
  id: string;
  serviceAccountId: string;
  tenantId: string;
  tenantName: string;
  name: string;
  description: string;
  status: EntityStatus;
  keysCount: number;
  apiKeysCount: number;
  lastAuthAt?: string;
  updatedAt: string;
}

export interface ServiceAccountDetail extends ServiceAccountSummary {
  ownerTeam: string;
  createdAt: string;
  tokens: TokenInfo[];
  accessSummary: EffectiveAccessRow[];
  asymmetricKeys: AsymmetricKey[];
  apiKeys: ApiKey[];
}

export interface OAuthSecretMeta {
  id: string;
  name: string;
  status: EntityStatus;
  createdAt: string;
  expiresAt?: string;
  note?: string;
}

export interface OAuthClientSummary {
  id: string;
  clientId: string;
  tenantId: string;
  tenantName: string;
  name: string;
  type: string;
  status: EntityStatus;
  redirectUrisCount: number;
  scopesCount: number;
  secretsCount: number;
  updatedAt: string;
}

export interface OAuthClientDetail extends OAuthClientSummary {
  createdAt: string;
  redirectUris: string[];
  scopes: string[];
  secrets: OAuthSecretMeta[];
  tokens: TokenInfo[];
}

export interface Permission {
  id: string;
  displayName: string;
  description?: string;
}

export interface RoleSummary {
  id: string;
  roleId: string;
  namespace: string;
  name: string;
  description: string;
  permissionsCount: number;
  system: boolean;
}

export interface RoleDetail extends RoleSummary {
  permissions: Permission[];
  usedBy: string[];
}

export interface ResourceBinding {
  id: string;
  resourceType: ResourceType;
  resourceId: string;
  tenantId: string;
  subjectType: SubjectType;
  subjectId: string;
  subjectName: string;
  roleId: string;
  source: 'direct' | 'group' | 'policy';
  version: number;
  condition?: string;
  expiresAt?: string;
}

export interface EffectiveAccessRow {
  id: string;
  subjectId: string;
  subjectName: string;
  subjectType: SubjectType;
  resourceType: ResourceType;
  resourceId: string;
  roleId: string;
  permission: string;
  source: 'direct' | 'group' | 'policy';
  decision: 'allow' | 'deny' | 'conditional';
}

export interface ExplainStep {
  id: string;
  title: string;
  details: string[];
}

export interface ExplainAccessResult {
  subjectId: string;
  resourceId: string;
  permission: string;
  decision: 'allow' | 'deny' | 'conditional';
  evaluatedAt: string;
  summary: string;
  pathIds: string[];
  steps: ExplainStep[];
}

export interface ImpactRow {
  id: string;
  subjectId: string;
  subjectName: string;
  before: string;
  after: string;
  status: Severity;
  affectedPermissions: string[];
}

export interface PolicyTemplate {
  id: string;
  templateId: string;
  name: string;
  scope: string;
  status: EntityStatus;
  parameters: string[];
  description: string;
  generatedBindings: string[];
}

export interface SupportGrant {
  id: string;
  grantId: string;
  tenantId: string;
  tenantName: string;
  subjectId: string;
  subjectName: string;
  scope: string;
  incidentId: string;
  roleId: string;
  reason: string;
  expiresAt: string;
  status: EntityStatus;
  requestedAt: string;
  approvedAt?: string;
}

export interface AuditEvent {
  id: string;
  eventId: string;
  tenantId: string;
  eventType: string;
  actor: string;
  resource: string;
  result: string;
  reason?: string;
  occurredAt: string;
  payload: Record<string, unknown>;
}

export interface OperationStep {
  id: string;
  title: string;
  status: OperationStatus;
}

export interface Operation {
  id: string;
  operationId: string;
  tenantId: string;
  type: string;
  resource: string;
  actor: string;
  status: OperationStatus;
  startedAt: string;
  updatedAt: string;
  completedAt?: string;
  errorMessage?: string;
  steps: OperationStep[];
  logs: string[];
}

export interface SearchResult {
  id: string;
  type: string;
  title: string;
  context: string;
  description: string;
  href: string;
}

export interface SettingsSection {
  id: string;
  title: string;
  description?: string;
  fields: Array<{
    id: string;
    label: string;
    value: string | boolean;
    type: 'text' | 'select' | 'switch';
    options?: string[];
  }>;
}

export interface DashboardData {
  metrics: SummaryMetric[];
  quickActions: QuickActionItem[];
  recentActivity: ActivityItem[];
}

export interface PagedResponse<T> {
  items: T[];
  total: number;
}

export interface TenantScopedFilter {
  tenantId?: string;
  query?: string;
}
