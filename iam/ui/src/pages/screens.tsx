import * as React from 'react';

import {useNavigate, useParams} from '@tanstack/react-router';
import type {ColumnDef} from '@gravity-ui/table/tanstack';
import {ArrowUpRightFromSquare} from '@gravity-ui/icons';
import {
  Button,
  Card,
  Flex,
  Icon,
  Label,
  Select,
  Switch,
  Text,
  TextArea,
  TextInput,
} from '@gravity-ui/uikit';

import {appToaster} from '@/app/providers/app-toaster';
import {useAppUI} from '@/app/providers/app-ui-context';
import {
  useAccessBindingsQuery,
  useAuditEventQuery,
  useAuditEventsQuery,
  useCreateTenantMutation,
  useCreateServiceAccountMutation,
  useDashboardQuery,
  useEffectiveAccessQuery,
  useExplainAccessQuery,
  useGroupQuery,
  useGroupsQuery,
  useImpactSimulationQuery,
  useOAuthClientQuery,
  useOAuthClientsQuery,
  useOperationQuery,
  useOperationsQuery,
  usePolicyTemplateQuery,
  usePolicyTemplatesQuery,
  useRemoveGroupMemberMutation,
  useRoleQuery,
  useRolesQuery,
  useSaveSettingsMutation,
  useSearchResultsQuery,
  useServiceAccountQuery,
  useServiceAccountsQuery,
  useSessionsQuery,
  useSettingsQuery,
  useSupportGrantsQuery,
  useTenantQuery,
  useTenantsQuery,
  useUpdateTenantMutation,
  useUserQuery,
  useUsersQuery,
  useAddGroupMemberMutation,
} from '@/entities/queries';
import {
  CreateServiceAccountWizard,
  DeleteTenantDialog,
  GrantAccessDrawer,
  RotateSecretDialog,
  SupportGrantWizard,
} from '@/features/iam-actions';
import {formatCount, formatDateTime, formatShortDate, titleFromId} from '@/shared/lib/format';
import {
  ActivityFeed,
  DataTableCard,
  DetailTabs,
  EmptyState,
  ErrorState,
  HighlightAlert,
  JsonCodeBlock,
  KeyValueGrid,
  LoadingState,
  MetricGrid,
  OperationTimeline,
  PageHeader,
  PillList,
  ResourceList,
  SectionCard,
  StatusBadge,
  ToneBadge,
} from '@/shared/ui/blocks';
import type {
  AuditEvent,
  EffectiveAccessRow,
  GroupMember,
  ImpactRow,
  Operation,
  ResourceBinding,
  ResourceType,
  RoleDetail,
  SearchResult,
  SettingsSection,
  SupportGrant,
} from '@/shared/types/iam';

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unexpected UI error';
}

function matchesQuery(query: string, values: Array<string | number | undefined>) {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return true;
  }

  return values.some((value) => String(value ?? '').toLowerCase().includes(normalized));
}

function showBulkToast(label: string, count: number) {
  appToaster.add({
    name: `bulk-${label}-${count}`,
    title: label,
    content: `${count} selected`,
    theme: 'info',
  });
}

function TenantCreateAction() {
  const navigate = useNavigate();

  return (
    <Button view="action" onClick={() => navigate({to: '/tenants/create'})}>
      Create Tenant
    </Button>
  );
}

const tenantIdPattern = /^[a-z][a-z0-9-]{2,127}$/;

function TenantFormPage({mode}: {mode: 'create' | 'edit'}) {
  const navigate = useNavigate();
  const params = useParams({strict: false}) as {tenantId?: string};
  const tenantIdFromParams = params.tenantId ?? '';
  const tenantQuery = useTenantQuery(tenantIdFromParams);
  const createMutation = useCreateTenantMutation();
  const updateMutation = useUpdateTenantMutation();
  const mutation = mode === 'create' ? createMutation : updateMutation;

  const [tenantId, setTenantId] = React.useState('');
  const [displayName, setDisplayName] = React.useState('');
  const [externalRef, setExternalRef] = React.useState('');
  const [region, setRegion] = React.useState('eu-central');
  const [tier, setTier] = React.useState<'active' | 'trial'>('active');

  React.useEffect(() => {
    if (mode === 'edit' && tenantQuery.data) {
      setTenantId(tenantQuery.data.tenantId);
      setDisplayName(tenantQuery.data.name);
      setExternalRef(tenantQuery.data.externalRef ?? '');
      setRegion(tenantQuery.data.region);
      setTier(tenantQuery.data.status === 'trial' ? 'trial' : 'active');
    }
  }, [mode, tenantQuery.data]);

  if (mode === 'edit' && tenantQuery.isPending) {
    return <LoadingState title="Loading tenant form" />;
  }

  if (mode === 'edit' && (tenantQuery.error || !tenantQuery.data)) {
    return <ErrorState description={getErrorMessage(tenantQuery.error)} />;
  }

  const canSubmit =
    tenantIdPattern.test(tenantId.trim()) &&
    displayName.trim().length > 1 &&
    region.trim().length > 1;

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Tenant form"
        title={mode === 'create' ? 'Create Tenant' : 'Edit Tenant'}
        description={
          mode === 'create'
            ? 'Provision tenant metadata through the IAM identity facade.'
            : `Update tenant metadata for ${tenantIdFromParams}.`
        }
        actions={
          <Flex gap="2" wrap>
            <Button
              view="flat"
              onClick={() =>
                navigate(
                  mode === 'create'
                    ? {to: '/tenants'}
                    : {to: '/tenants/$tenantId', params: {tenantId: tenantIdFromParams}},
                )
              }
            >
              Cancel
            </Button>
            <Button
              view="action"
              loading={mutation.isPending}
              disabled={!canSubmit}
              onClick={() => {
                mutation.mutate(
                  {
                    tenantId: tenantId.trim(),
                    displayName: displayName.trim(),
                    externalRef: externalRef.trim(),
                    region: region.trim(),
                    tier,
                  },
                  {
                    onSuccess: (tenant) => {
                      appToaster.add({
                        name: `${mode}-tenant-page-${tenant.tenantId}`,
                        title: mode === 'create' ? 'Tenant created' : 'Tenant updated',
                        content: `${tenant.name} (${tenant.tenantId})`,
                        theme: 'success',
                      });
                      navigate({
                        to: '/tenants/$tenantId',
                        params: {tenantId: tenant.tenantId},
                      });
                    },
                  },
                );
              }}
            >
              {mode === 'create' ? 'Create Tenant' : 'Save Changes'}
            </Button>
          </Flex>
        }
      />
      <SectionCard
        title="Tenant metadata"
        description="These fields are sent to the current IAM API and stored in the source of truth."
      >
        <div className="form-grid">
          <TextInput
            label="Tenant ID"
            placeholder="tenant-acme-prod"
            value={tenantId}
            disabled={mode === 'edit'}
            onUpdate={setTenantId}
          />
          <TextInput
            label="Display name"
            placeholder="Acme Production"
            value={displayName}
            onUpdate={setDisplayName}
          />
          <TextInput
            label="External ref"
            placeholder="crm-acme-001"
            value={externalRef}
            onUpdate={setExternalRef}
          />
          <TextInput
            label="Region"
            placeholder="eu-central"
            value={region}
            onUpdate={setRegion}
          />
          <TextInput
            label="Tier"
            placeholder="active or trial"
            value={tier}
            onUpdate={(value) => setTier(value === 'trial' ? 'trial' : 'active')}
          />
          <Text variant="body-1" color="secondary">
            Use `active` or `trial` for tier. Tenant ID must match `[a-z][a-z0-9-]` and be 3-128 chars long.
          </Text>
        </div>
      </SectionCard>
    </div>
  );
}

export function TenantCreatePage() {
  return <TenantFormPage mode="create" />;
}

export function TenantEditPage() {
  return <TenantFormPage mode="edit" />;
}

function AccessBindingsTable({
  bindings,
  onOpen,
}: {
  bindings: ResourceBinding[];
  onOpen?: (binding: ResourceBinding) => void;
}) {
  const navigate = useNavigate();
  const columns: ColumnDef<ResourceBinding>[] = [
    {accessorKey: 'subjectName', header: 'Subject'},
    {accessorKey: 'subjectType', header: 'Type', cell: ({row}) => titleFromId(row.original.subjectType)},
    {accessorKey: 'roleId', header: 'Role'},
    {accessorKey: 'source', header: 'Source'},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Flex gap="2">
          <Button
            size="s"
            view="flat"
            onClick={() => onOpen?.(row.original)}
          >
            Inspect
          </Button>
          <Button
            size="s"
            view="flat"
            onClick={() =>
              navigate({
                to: '/access/explain',
              })
            }
          >
            Explain
          </Button>
        </Flex>
      ),
    },
  ];

  return (
    <DataTableCard
      title="Bindings"
      description="Direct, group and policy-derived runtime grants."
      data={bindings}
      columns={columns}
      emptyTitle="No bindings"
      emptyDescription="This resource has no explicit bindings yet."
    />
  );
}

function EffectiveAccessTable({
  rows,
  title,
  description,
}: {
  rows: EffectiveAccessRow[];
  title: string;
  description?: string;
}) {
  const columns: ColumnDef<EffectiveAccessRow>[] = [
    {accessorKey: 'subjectName', header: 'Subject'},
    {
      accessorKey: 'permission',
      header: 'Permission',
      cell: ({row}) => (
        <Flex gap="2" alignItems="center">
          <Text variant="body-2">{row.original.permission}</Text>
          <ToneBadge tone={row.original.decision === 'allow' ? 'success' : row.original.decision === 'conditional' ? 'warning' : 'danger'}>
            {titleFromId(row.original.decision)}
          </ToneBadge>
        </Flex>
      ),
    },
    {accessorKey: 'roleId', header: 'Role'},
    {
      accessorKey: 'resourceId',
      header: 'Resource',
      cell: ({row}) => `${row.original.resourceType}/${row.original.resourceId}`,
    },
    {accessorKey: 'source', header: 'Source'},
  ];

  return (
    <DataTableCard
      title={title}
      description={description}
      data={rows}
      columns={columns}
      emptyTitle="No effective access"
    />
  );
}

export function DashboardPage() {
  const dashboardQuery = useDashboardQuery();
  const {context} = useAppUI();
  const navigate = useNavigate();

  if (dashboardQuery.isPending) {
    return <LoadingState title="Loading dashboard" />;
  }

  if (dashboardQuery.error || !dashboardQuery.data) {
    return <ErrorState description={getErrorMessage(dashboardQuery.error)} />;
  }

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Enterprise IAM console"
        title="Dashboard"
        description="Unified operational view over identities, bindings, sessions and authorization changes."
        actions={
          <Flex gap="2" wrap>
            <ToneBadge tone="info">{context.environment}</ToneBadge>
            <ToneBadge tone="success">{context.region}</ToneBadge>
          </Flex>
        }
      />
      <Card className="hero-panel">
        <div className="hero-panel__copy">
          <Text variant="header-1" as="div">
            Control plane for tenants, authz and support access
          </Text>
          <Text variant="body-2" color="secondary">
            The UI is already wired to the current grpc-gateway API with mock/live adapters, so you can switch modes without changing screens.
          </Text>
          <Flex gap="2" wrap>
            {dashboardQuery.data.quickActions.map((action) => (
              <Button
                key={action.id}
                view={action.id === 'grant-access' ? 'action' : 'outlined'}
                onClick={() => navigate({to: action.href ?? '/dashboard'})}
              >
                {action.title}
              </Button>
            ))}
          </Flex>
        </div>
        <div className="hero-panel__meta">
          <ToneBadge tone="success">Tenant: {context.tenantId}</ToneBadge>
          <ToneBadge tone="warning">Org: {context.organizationId}</ToneBadge>
        </div>
      </Card>
      <MetricGrid items={dashboardQuery.data.metrics} />
      <div className="two-column-grid">
        <SectionCard title="Quick Actions" description="Common operator workflows from the canvas spec.">
          <div className="quick-action-grid">
            {dashboardQuery.data.quickActions.map((action) => (
              <button
                key={action.id}
                className="quick-action-tile"
                type="button"
                onClick={() => navigate({to: action.href ?? '/dashboard'})}
              >
                <Text variant="subheader-1">{action.title}</Text>
                <Text variant="body-1" color="secondary">
                  Open workflow
                </Text>
              </button>
            ))}
          </div>
        </SectionCard>
        <SectionCard title="Recent Activity" description="Most recent security-impacting actions and changes.">
          <ActivityFeed items={dashboardQuery.data.recentActivity} />
        </SectionCard>
      </div>
    </div>
  );
}

export function TenantsPage() {
  const tenantsQuery = useTenantsQuery();
  const navigate = useNavigate();
  const [query, setQuery] = React.useState('');
  const [statusFilter, setStatusFilter] = React.useState('all');
  const deferredQuery = React.useDeferredValue(query);

  if (tenantsQuery.isPending) {
    return <LoadingState title="Loading tenants" />;
  }

  if (tenantsQuery.error || !tenantsQuery.data) {
    return <ErrorState description={getErrorMessage(tenantsQuery.error)} />;
  }

  const items = tenantsQuery.data.items.filter((tenant) => {
    if (statusFilter !== 'all' && tenant.status !== statusFilter) {
      return false;
    }

    return matchesQuery(deferredQuery, [
      tenant.tenantId,
      tenant.name,
      tenant.organizationId,
      tenant.plan,
      tenant.region,
    ]);
  });

  const metrics = [
    {id: 'all', title: 'Tenants', value: formatCount(tenantsQuery.data.items.length)},
    {
      id: 'active',
      title: 'Active',
      value: formatCount(tenantsQuery.data.items.filter((item) => item.status === 'active').length),
      tone: 'success' as const,
    },
    {
      id: 'trial',
      title: 'Trial',
      value: formatCount(tenantsQuery.data.items.filter((item) => item.status === 'trial').length),
      tone: 'warning' as const,
    },
    {
      id: 'users',
      title: 'Users',
      value: formatCount(tenantsQuery.data.items.reduce((sum, item) => sum + item.memberCount, 0)),
      tone: 'info' as const,
    },
  ];

  const columns: ColumnDef<(typeof items)[number]>[] = [
    {accessorKey: 'tenantId', header: 'Tenant ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'organizationId', header: 'Org ID'},
    {accessorKey: 'plan', header: 'Plan'},
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({row}) => <StatusBadge status={row.original.status} />,
    },
    {accessorKey: 'memberCount', header: 'Users'},
    {accessorKey: 'serviceAccountCount', header: 'Service Accounts'},
    {
      accessorKey: 'updatedAt',
      header: 'Updated',
      cell: ({row}) => formatDateTime(row.original.updatedAt),
    },
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() =>
            navigate({
              to: '/tenants/$tenantId',
              params: {tenantId: row.original.tenantId},
            })
          }
        >
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        title="Tenants"
        description="Tenant inventory with quick operational entry points."
        actions={<TenantCreateAction />}
      />
      <MetricGrid items={metrics} />
      <SectionCard title="Filters" description="Search and operational filters">
        <div className="filter-bar">
          <TextInput
            label="Search"
            placeholder="tenant id / org / region"
            value={query}
            onUpdate={setQuery}
          />
          <Select
            label="Status"
            value={[statusFilter]}
            options={[
              {value: 'all', content: 'All'},
              {value: 'active', content: 'Active'},
              {value: 'trial', content: 'Trial'},
              {value: 'suspended', content: 'Suspended'},
            ]}
            onUpdate={(value) => setStatusFilter(value[0] ?? 'all')}
          />
        </div>
      </SectionCard>
      <DataTableCard
        title="Tenants list"
        description="Operational tenant inventory from the current IAM source of truth."
        data={items}
        columns={columns}
      />
    </div>
  );
}

export function TenantDetailPage({tab}: {tab: 'overview' | 'members' | 'groups' | 'serviceAccounts' | 'oauthClients' | 'access' | 'audit'}) {
  const params = useParams({strict: false}) as {tenantId: string};
  const tenantId = params.tenantId;
  const navigate = useNavigate();
  const tenantQuery = useTenantQuery(tenantId);
  const usersQuery = useUsersQuery();
  const groupsQuery = useGroupsQuery();
  const serviceAccountsQuery = useServiceAccountsQuery();
  const oauthClientsQuery = useOAuthClientsQuery();
  const auditEventsQuery = useAuditEventsQuery();
  const bindingsQuery = useAccessBindingsQuery('tenant', tenantId, tab === 'access');
  const [createOpen, setCreateOpen] = React.useState(false);
  const [supportOpen, setSupportOpen] = React.useState(false);
  const [grantOpen, setGrantOpen] = React.useState(false);
  const [deleteOpen, setDeleteOpen] = React.useState(false);

  if (tenantQuery.isPending) {
    return <LoadingState title="Loading tenant" />;
  }

  if (tenantQuery.error || !tenantQuery.data) {
    return <ErrorState description={getErrorMessage(tenantQuery.error)} />;
  }

  const tenant = tenantQuery.data;
  const members = (usersQuery.data?.items ?? []).filter((user) =>
    user.tenantIds.includes(tenantId),
  );
  const groups = (groupsQuery.data?.items ?? []).filter((group) => group.tenantId === tenantId);
  const serviceAccounts = (serviceAccountsQuery.data?.items ?? []).filter((item) => item.tenantId === tenantId);
  const oauthClients = (oauthClientsQuery.data?.items ?? []).filter((item) => item.tenantId === tenantId);
  const auditEvents = (auditEventsQuery.data?.items ?? []).filter((event) => event.tenantId === tenantId);
  const bindings = bindingsQuery.data ?? [];

  const tabs = [
    {id: 'overview', title: 'Overview'},
    {id: 'members', title: 'Members'},
    {id: 'groups', title: 'Groups'},
    {id: 'serviceAccounts', title: 'Service Accounts'},
    {id: 'oauthClients', title: 'OAuth Clients'},
    {id: 'access', title: 'Access'},
    {id: 'audit', title: 'Audit'},
  ];

  const tabRoutes: Record<string, string> = {
    overview: '/tenants/$tenantId',
    members: '/tenants/$tenantId/members',
    groups: '/tenants/$tenantId/groups',
    serviceAccounts: '/tenants/$tenantId/service-accounts',
    oauthClients: '/tenants/$tenantId/oauth-clients',
    access: '/tenants/$tenantId/access',
    audit: '/tenants/$tenantId/audit',
  };

  const memberColumns: ColumnDef<(typeof members)[number]>[] = [
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'email', header: 'Email'},
    {accessorKey: 'source', header: 'Source'},
    {accessorKey: 'mfaEnabled', header: 'MFA', cell: ({row}) => (row.original.mfaEnabled ? 'On' : 'Off')},
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({row}) => <StatusBadge status={row.original.status} />,
    },
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() => navigate({to: '/users/$userId', params: {userId: row.original.userId}})}
        >
          Open
        </Button>
      ),
    },
  ];

  const groupColumns: ColumnDef<(typeof groups)[number]>[] = [
    {accessorKey: 'groupId', header: 'Group ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'membersCount', header: 'Members'},
    {accessorKey: 'dynamic', header: 'Dynamic', cell: ({row}) => (row.original.dynamic ? 'Yes' : 'No')},
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({row}) => <StatusBadge status={row.original.status} />,
    },
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() => navigate({to: '/groups/$groupId', params: {groupId: row.original.groupId}})}
        >
          Open
        </Button>
      ),
    },
  ];

  const serviceAccountColumns: ColumnDef<(typeof serviceAccounts)[number]>[] = [
    {accessorKey: 'serviceAccountId', header: 'ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'ownerTeam', header: 'Owner team'},
    {accessorKey: 'keysCount', header: 'Keys'},
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({row}) => <StatusBadge status={row.original.status} />,
    },
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() =>
            navigate({
              to: '/service-accounts/$serviceAccountId',
              params: {serviceAccountId: row.original.serviceAccountId},
            })
          }
        >
          Open
        </Button>
      ),
    },
  ];

  const oauthColumns: ColumnDef<(typeof oauthClients)[number]>[] = [
    {accessorKey: 'clientId', header: 'Client ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'scopesCount', header: 'Scopes'},
    {accessorKey: 'redirectUrisCount', header: 'Redirect URIs'},
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({row}) => <StatusBadge status={row.original.status} />,
    },
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() =>
            navigate({
              to: '/oauth-clients/$clientId',
              params: {clientId: row.original.clientId},
            })
          }
        >
          Open
        </Button>
      ),
    },
  ];

  const auditColumns: ColumnDef<AuditEvent>[] = [
    {accessorKey: 'eventType', header: 'Event'},
    {accessorKey: 'actor', header: 'Actor'},
    {accessorKey: 'resource', header: 'Resource'},
    {accessorKey: 'occurredAt', header: 'Occurred', cell: ({row}) => formatDateTime(row.original.occurredAt)},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() => navigate({to: '/audit/$eventId', params: {eventId: row.original.eventId}})}
        >
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Tenant details"
        title={tenant.name}
        description={tenant.description}
        actions={
          <Flex gap="2" wrap>
            <Button
              view="outlined"
              onClick={() =>
                navigate({
                  to: '/tenants/$tenantId/edit',
                  params: {tenantId},
                })
              }
            >
              Edit Tenant
            </Button>
            <Button view="outlined" onClick={() => setCreateOpen(true)}>
              Create SA
            </Button>
            <Button view="outlined" onClick={() => setSupportOpen(true)}>
              Support Access
            </Button>
            <Button view="action" onClick={() => setGrantOpen(true)}>
              Grant Access
            </Button>
            <Button view="outlined" onClick={() => setDeleteOpen(true)}>
              Delete Tenant
            </Button>
          </Flex>
        }
      />
      <DetailTabs
        items={tabs}
        activeTab={tab}
        onSelectTab={(tabId) =>
          navigate({
            to: tabRoutes[tabId] as '/tenants/$tenantId',
            params: {tenantId},
          })
        }
      />
      <SectionCard title="Tenant summary" description="Operational metadata and ownership.">
        <KeyValueGrid
          items={[
            {label: 'Tenant ID', value: tenant.tenantId},
            {label: 'Organization', value: tenant.organizationId},
            {label: 'Plan', value: tenant.plan},
            {label: 'Region', value: tenant.region},
            {label: 'Status', value: <StatusBadge status={tenant.status} />},
            {label: 'Created', value: formatShortDate(tenant.createdAt)},
            {label: 'Owners', value: formatCount(tenant.ownersCount)},
            {label: 'Members', value: formatCount(tenant.memberCount)},
            {label: 'Service Accounts', value: formatCount(tenant.serviceAccountCount)},
          ]}
        />
      </SectionCard>
      {tab === 'overview' ? (
        <div className="two-column-grid">
          <SectionCard title="Highlights" description="Tenant summary points from the seed configuration.">
            <ResourceList items={tenant.summary} />
          </SectionCard>
          <SectionCard title="Integrations" description="Connected systems and sync points.">
            <PillList items={tenant.integrations} />
          </SectionCard>
          <SectionCard title="Resource map" description="Known tenant-scoped resources.">
            <ResourceList items={tenant.resourceMap} />
          </SectionCard>
          <SectionCard title="Tags" description="Quick metadata for operations and support.">
            <PillList items={tenant.tags} />
          </SectionCard>
        </div>
      ) : null}
      {tab === 'members' ? (
        <DataTableCard
          title="Members"
          description="Current tenant members and access sources."
          data={members}
          columns={memberColumns}
          emptyTitle="No tenant members"
        />
      ) : null}
      {tab === 'groups' ? (
        <DataTableCard
          title="Groups"
          description="Tenant-scoped groups and membership bundles."
          data={groups}
          columns={groupColumns}
          emptyTitle="No groups"
        />
      ) : null}
      {tab === 'serviceAccounts' ? (
        <DataTableCard
          title="Service Accounts"
          description="Machine identities provisioned for the tenant."
          data={serviceAccounts}
          columns={serviceAccountColumns}
          emptyTitle="No service accounts"
        />
      ) : null}
      {tab === 'oauthClients' ? (
        <DataTableCard
          title="OAuth Clients"
          description="Interactive and CLI clients configured for the tenant."
          data={oauthClients}
          columns={oauthColumns}
          emptyTitle="No OAuth clients"
        />
      ) : null}
      {tab === 'access' ? <AccessBindingsTable bindings={bindings} /> : null}
      {tab === 'audit' ? (
        <DataTableCard
          title="Tenant audit"
          description="Security and change log scoped to this tenant."
          data={auditEvents}
          columns={auditColumns}
          emptyTitle="No audit events"
        />
      ) : null}
      <CreateServiceAccountWizard
        open={createOpen}
        defaultTenantId={tenantId}
        onClose={() => setCreateOpen(false)}
      />
      <DeleteTenantDialog
        open={deleteOpen}
        tenantId={tenant.tenantId}
        tenantName={tenant.name}
        onClose={() => setDeleteOpen(false)}
        onSuccess={() => navigate({to: '/tenants'})}
      />
      <SupportGrantWizard
        open={supportOpen}
        defaultTenantId={tenantId}
        onClose={() => setSupportOpen(false)}
      />
      <GrantAccessDrawer
        open={grantOpen}
        onClose={() => setGrantOpen(false)}
        tenantId={tenantId}
        resourceType="tenant"
        resourceId={tenantId}
      />
    </div>
  );
}

export function TenantOverviewRoutePage() {
  return <TenantDetailPage tab="overview" />;
}

export function TenantMembersRoutePage() {
  return <TenantDetailPage tab="members" />;
}

export function TenantGroupsRoutePage() {
  return <TenantDetailPage tab="groups" />;
}

export function TenantServiceAccountsRoutePage() {
  return <TenantDetailPage tab="serviceAccounts" />;
}

export function TenantOAuthClientsRoutePage() {
  return <TenantDetailPage tab="oauthClients" />;
}

export function TenantAccessRoutePage() {
  return <TenantDetailPage tab="access" />;
}

export function TenantAuditRoutePage() {
  return <TenantDetailPage tab="audit" />;
}

export function UsersPage() {
  const usersQuery = useUsersQuery();
  const navigate = useNavigate();
  const [query, setQuery] = React.useState('');
  const [statusFilter, setStatusFilter] = React.useState('all');
  const [mfaFilter, setMfaFilter] = React.useState('all');
  const deferredQuery = React.useDeferredValue(query);

  if (usersQuery.isPending) {
    return <LoadingState title="Loading users" />;
  }

  if (usersQuery.error || !usersQuery.data) {
    return <ErrorState description={getErrorMessage(usersQuery.error)} />;
  }

  const items = usersQuery.data.items.filter((user) => {
    if (statusFilter !== 'all' && user.status !== statusFilter) {
      return false;
    }
    if (mfaFilter === 'enabled' && !user.mfaEnabled) {
      return false;
    }
    if (mfaFilter === 'disabled' && user.mfaEnabled) {
      return false;
    }

    return matchesQuery(deferredQuery, [user.userId, user.name, user.email, user.source]);
  });

  const columns: ColumnDef<(typeof items)[number]>[] = [
    {accessorKey: 'userId', header: 'User ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'email', header: 'Email'},
    {accessorKey: 'source', header: 'Source'},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
    {accessorKey: 'mfaEnabled', header: 'MFA', cell: ({row}) => (row.original.mfaEnabled ? 'Enabled' : 'Off')},
    {accessorKey: 'tenantIds', header: 'Tenants', cell: ({row}) => formatCount(row.original.tenantIds.length)},
    {accessorKey: 'lastLoginAt', header: 'Last login', cell: ({row}) => formatDateTime(row.original.lastLoginAt)},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/users/$userId', params: {userId: row.original.userId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        title="Users"
        description="SSO, local and internal identities across tenants."
        actions={<Button view="action" onClick={() => showBulkToast('User invite flow is not implemented yet', 1)}>Add User</Button>}
      />
      <SectionCard title="Filters" description="Search users by identity, source and security posture.">
        <div className="filter-bar">
          <TextInput label="Search" placeholder="email / id / name" value={query} onUpdate={setQuery} />
          <Select
            label="Status"
            value={[statusFilter]}
            options={[
              {value: 'all', content: 'All'},
              {value: 'active', content: 'Active'},
              {value: 'disabled', content: 'Disabled'},
            ]}
            onUpdate={(value) => setStatusFilter(value[0] ?? 'all')}
          />
          <Select
            label="MFA"
            value={[mfaFilter]}
            options={[
              {value: 'all', content: 'All'},
              {value: 'enabled', content: 'Enabled'},
              {value: 'disabled', content: 'Disabled'},
            ]}
            onUpdate={(value) => setMfaFilter(value[0] ?? 'all')}
          />
        </div>
      </SectionCard>
      <DataTableCard
        title="User inventory"
        description="Global identities already synced with IAM."
        data={items}
        columns={columns}
        selectable
        bulkActions={[
          {label: 'Disable', onClick: (rows) => showBulkToast('Disable requested', rows.length)},
          {label: 'Require MFA', onClick: (rows) => showBulkToast('MFA policy requested', rows.length)},
          {label: 'Export CSV', onClick: (rows) => showBulkToast('Export requested', rows.length)},
        ]}
      />
    </div>
  );
}

export function UserProfilePage() {
  const params = useParams({strict: false}) as {userId: string};
  const userQuery = useUserQuery(params.userId);
  const groupsQuery = useGroupsQuery();
  const effectiveAccessQuery = useEffectiveAccessQuery();
  const auditEventsQuery = useAuditEventsQuery();
  const navigate = useNavigate();
  const [tab, setTab] = React.useState<'profile' | 'memberships' | 'groups' | 'access' | 'sessions' | 'audit'>('profile');

  if (userQuery.isPending) {
    return <LoadingState title="Loading user profile" />;
  }

  if (userQuery.error || !userQuery.data) {
    return <ErrorState description={getErrorMessage(userQuery.error)} />;
  }

  const user = userQuery.data;
  const groups = ((groupsQuery.data?.items ?? []) as Array<{
    id?: string;
    groupId: string;
    name: string;
    tenantName: string;
    membersCount: number;
    members?: GroupMember[];
  }>)
    .filter((group) => (group.members ?? []).some((member) => member.subjectId === user.userId))
    .map((group) => ({...group, id: group.id ?? group.groupId}));
  const effectiveAccess = (effectiveAccessQuery.data ?? []).filter((row) => row.subjectId === user.userId);
  const auditEvents = (auditEventsQuery.data?.items ?? []).filter((event) =>
    matchesQuery(user.userId, [event.actor, event.resource, JSON.stringify(event.payload)]),
  );

  const membershipColumns: ColumnDef<(typeof user.memberships)[number]>[] = [
    {accessorKey: 'tenantName', header: 'Tenant'},
    {accessorKey: 'tenantId', header: 'Tenant ID'},
    {accessorKey: 'roleIds', header: 'Roles', cell: ({row}) => row.original.roleIds.join(', ') || '—'},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  const groupColumns: ColumnDef<(typeof groups)[number]>[] = [
    {accessorKey: 'groupId', header: 'Group ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'tenantName', header: 'Tenant'},
    {accessorKey: 'membersCount', header: 'Members'},
  ];

  const sessionColumns: ColumnDef<(typeof user.sessions)[number]>[] = [
    {accessorKey: 'client', header: 'Client'},
    {accessorKey: 'device', header: 'Device'},
    {accessorKey: 'ipAddress', header: 'IP'},
    {accessorKey: 'protectionLevel', header: 'Protection'},
    {accessorKey: 'lastSeenAt', header: 'Last seen', cell: ({row}) => formatDateTime(row.original.lastSeenAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  const tokenColumns: ColumnDef<(typeof user.tokens)[number]>[] = [
    {accessorKey: 'client', header: 'Client'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'protectionLevel', header: 'Protection'},
    {accessorKey: 'expiresAt', header: 'Expires', cell: ({row}) => formatDateTime(row.original.expiresAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  const auditColumns: ColumnDef<AuditEvent>[] = [
    {accessorKey: 'eventType', header: 'Event'},
    {accessorKey: 'actor', header: 'Actor'},
    {accessorKey: 'resource', header: 'Resource'},
    {accessorKey: 'occurredAt', header: 'Occurred', cell: ({row}) => formatDateTime(row.original.occurredAt)},
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="User profile"
        title={user.name}
        description={user.email}
        actions={
          <Flex gap="2" wrap>
            <Button view="outlined" onClick={() => showBulkToast('Disable user requested', 1)}>
              Disable User
            </Button>
            <Button view="outlined" onClick={() => showBulkToast('Reset sessions requested', user.sessions.length)}>
              Reset Sessions
            </Button>
            <Button view="action" onClick={() => navigate({to: '/access/explain'})}>
              Explain Access
            </Button>
          </Flex>
        }
      />
      <SectionCard title="Identity summary" description="Profile metadata and current security state.">
        <KeyValueGrid
          items={[
            {label: 'User ID', value: user.userId},
            {label: 'Email', value: user.email},
            {label: 'Source', value: user.source},
            {label: 'Status', value: <StatusBadge status={user.status} />},
            {label: 'MFA', value: user.mfaEnabled ? 'Enabled' : 'Disabled'},
            {label: 'Last login', value: formatDateTime(user.lastLoginAt)},
          ]}
        />
      </SectionCard>
      <DetailTabs
        items={[
          {id: 'profile', title: 'Profile'},
          {id: 'memberships', title: 'Memberships'},
          {id: 'groups', title: 'Groups'},
          {id: 'access', title: 'Access'},
          {id: 'sessions', title: 'Sessions'},
          {id: 'audit', title: 'Audit'},
        ]}
        activeTab={tab}
        onSelectTab={(nextTab) => setTab(nextTab as typeof tab)}
      />
      {tab === 'profile' ? (
        <SectionCard title="Labels" description="Derived attributes that can drive policy or search.">
          {user.labels ? <PillList items={Object.entries(user.labels).map(([key, value]) => `${key}: ${value}`)} /> : <EmptyState title="No labels" />}
        </SectionCard>
      ) : null}
      {tab === 'memberships' ? (
        <DataTableCard
          title="Memberships"
          description="Tenant-scoped roles and status."
          data={user.memberships}
          columns={membershipColumns}
          emptyTitle="No memberships"
        />
      ) : null}
      {tab === 'groups' ? (
        <DataTableCard
          title="Group memberships"
          description="Derived from current group definitions."
          data={groups}
          columns={groupColumns}
          emptyTitle="No groups"
        />
      ) : null}
      {tab === 'access' ? (
        <EffectiveAccessTable
          rows={effectiveAccess}
          title="Effective access"
          description="Runtime-derived permission rows for this subject."
        />
      ) : null}
      {tab === 'sessions' ? (
        <div className="two-column-grid">
          <DataTableCard title="Sessions" description="Interactive sessions" data={user.sessions} columns={sessionColumns} emptyTitle="No sessions" />
          <DataTableCard title="Tokens" description="Issued tokens" data={user.tokens} columns={tokenColumns} emptyTitle="No tokens" />
        </div>
      ) : null}
      {tab === 'audit' ? (
        <DataTableCard
          title="Audit trail"
          description="Events that reference this identity."
          data={auditEvents}
          columns={auditColumns}
          emptyTitle="No audit references"
        />
      ) : null}
    </div>
  );
}

export function GroupsPage() {
  const groupsQuery = useGroupsQuery();
  const navigate = useNavigate();
  const [query, setQuery] = React.useState('');
  const [tenantFilter, setTenantFilter] = React.useState('all');
  const deferredQuery = React.useDeferredValue(query);

  if (groupsQuery.isPending) {
    return <LoadingState title="Loading groups" />;
  }

  if (groupsQuery.error || !groupsQuery.data) {
    return <ErrorState description={getErrorMessage(groupsQuery.error)} />;
  }

  const items = groupsQuery.data.items.filter((group) => {
    if (tenantFilter !== 'all' && group.tenantId !== tenantFilter) {
      return false;
    }

    return matchesQuery(deferredQuery, [group.groupId, group.name, group.tenantName]);
  });

  const tenantOptions = [
    {value: 'all', content: 'All tenants'},
    ...Array.from(new Set(groupsQuery.data.items.map((group) => group.tenantId))).map((tenantId) => ({
      value: tenantId,
      content: tenantId,
    })),
  ];

  const columns: ColumnDef<(typeof items)[number]>[] = [
    {accessorKey: 'groupId', header: 'Group ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'tenantName', header: 'Tenant'},
    {accessorKey: 'membersCount', header: 'Members'},
    {accessorKey: 'dynamic', header: 'Dynamic', cell: ({row}) => (row.original.dynamic ? 'Yes' : 'No')},
    {accessorKey: 'updatedAt', header: 'Updated', cell: ({row}) => formatDateTime(row.original.updatedAt)},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/groups/$groupId', params: {groupId: row.original.groupId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        title="Groups"
        description="Membership bundles and policy targets scoped by tenant."
        actions={<Button view="action" onClick={() => showBulkToast('Group creation wizard is not in MVP', 1)}>New Group</Button>}
      />
      <SectionCard title="Filters" description="Tenant and name filters.">
        <div className="filter-bar">
          <TextInput label="Search" value={query} onUpdate={setQuery} placeholder="group name / id" />
          <Select label="Tenant" value={[tenantFilter]} options={tenantOptions} onUpdate={(value) => setTenantFilter(value[0] ?? 'all')} />
        </div>
      </SectionCard>
      <DataTableCard title="Groups list" description="Static and dynamic groups." data={items} columns={columns} emptyTitle="No groups found" />
    </div>
  );
}

export function GroupDetailPage() {
  const params = useParams({strict: false}) as {groupId: string};
  const groupQuery = useGroupQuery(params.groupId);
  const usersQuery = useUsersQuery();
  const serviceAccountsQuery = useServiceAccountsQuery();
  const addMemberMutation = useAddGroupMemberMutation();
  const removeMemberMutation = useRemoveGroupMemberMutation();
  const [subjectValue, setSubjectValue] = React.useState('');

  if (groupQuery.isPending) {
    return <LoadingState title="Loading group" />;
  }

  if (groupQuery.error || !groupQuery.data) {
    return <ErrorState description={getErrorMessage(groupQuery.error)} />;
  }

  const group = groupQuery.data;
  const subjectOptions = [
    ...(usersQuery.data?.items ?? [])
      .filter((user) => user.tenantIds.includes(group.tenantId))
      .map((user) => ({
        value: `userAccount:${user.userId}:${user.name}`,
        content: `${user.name} (${user.userId})`,
      })),
    ...(serviceAccountsQuery.data?.items ?? [])
      .filter((account) => account.tenantId === group.tenantId)
      .map((account) => ({
        value: `serviceAccount:${account.serviceAccountId}:${account.name}`,
        content: `${account.name} (${account.serviceAccountId})`,
      })),
  ];

  const memberColumns: ColumnDef<GroupMember>[] = [
    {accessorKey: 'displayName', header: 'Display name'},
    {accessorKey: 'subjectId', header: 'Subject ID'},
    {accessorKey: 'subjectType', header: 'Type', cell: ({row}) => titleFromId(row.original.subjectType)},
    {accessorKey: 'addedAt', header: 'Added', cell: ({row}) => formatDateTime(row.original.addedAt)},
    {
      id: 'actions',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat-danger"
          loading={removeMemberMutation.isPending}
          onClick={() => removeMemberMutation.mutate({groupId: group.groupId, subjectId: row.original.subjectId})}
        >
          Remove
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow={group.tenantName}
        title={group.name}
        description={group.description}
        actions={
          <Flex gap="2" wrap>
            <ToneBadge tone="info">{group.dynamic ? 'Dynamic' : 'Static'}</ToneBadge>
            <StatusBadge status={group.status} />
          </Flex>
        }
      />
      <SectionCard title="Group summary" description="Rules, tenant scope and effective grants.">
        <KeyValueGrid
          items={[
            {label: 'Group ID', value: group.groupId},
            {label: 'Tenant', value: group.tenantName},
            {label: 'Members', value: formatCount(group.membersCount)},
            {label: 'Effective grants', value: formatCount(group.effectiveGrantCount)},
            {label: 'Effective roles', value: formatCount(group.effectiveRoleCount)},
            {label: 'Updated', value: formatDateTime(group.updatedAt)},
          ]}
        />
      </SectionCard>
      <div className="two-column-grid">
        <SectionCard title="Membership editor" description="Add users or service accounts into the group.">
          <div className="form-grid">
            <Select
              label="Subject"
              value={subjectValue ? [subjectValue] : []}
              options={subjectOptions}
              placeholder="Choose subject"
              onUpdate={(value) => setSubjectValue(value[0] ?? '')}
            />
            <Button
              view="action"
              disabled={!subjectValue}
              loading={addMemberMutation.isPending}
              onClick={() => {
                const [subjectType, subjectId, displayName] = subjectValue.split(':');
                addMemberMutation.mutate(
                  {
                    groupId: group.groupId,
                    subjectId,
                    subjectType: subjectType as 'userAccount' | 'serviceAccount',
                    displayName,
                  },
                  {
                    onSuccess: () => {
                      setSubjectValue('');
                    },
                  },
                );
              }}
            >
              Add Member
            </Button>
          </div>
        </SectionCard>
        <SectionCard title="Rules" description="Current membership constraints and notes.">
          <ResourceList items={group.rules} />
        </SectionCard>
      </div>
      <DataTableCard
        title="Members"
        description="Current resolved group members."
        data={group.members}
        columns={memberColumns}
        emptyTitle="No members"
      />
    </div>
  );
}

export function ServiceAccountCreatePage() {
  const navigate = useNavigate();
  const tenantsQuery = useTenantsQuery();
  const mutation = useCreateServiceAccountMutation();
  const [tenantId, setTenantId] = React.useState('');
  const [displayName, setDisplayName] = React.useState('');
  const [description, setDescription] = React.useState('');

  if (tenantsQuery.isPending) {
    return <LoadingState title="Loading tenants" />;
  }

  if (tenantsQuery.error || !tenantsQuery.data) {
    return <ErrorState description={getErrorMessage(tenantsQuery.error)} />;
  }

  const tenantOptions = tenantsQuery.data.items.map((tenant) => ({
    value: tenant.tenantId,
    content: `${tenant.name} (${tenant.tenantId})`,
  }));

  const canSubmit = tenantId.trim().length > 0 && displayName.trim().length > 1;

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Machine identity"
        title="Create Service Account"
        description="Provision a machine identity through the IAM identity facade."
        actions={
          <Flex gap="2" wrap>
            <Button view="flat" onClick={() => navigate({to: '/service-accounts'})}>
              Cancel
            </Button>
            <Button
              view="action"
              loading={mutation.isPending}
              disabled={!canSubmit}
              onClick={() => {
                mutation.mutate(
                  {
                    tenantId,
                    displayName: displayName.trim(),
                    description: description.trim(),
                  },
                  {
                    onSuccess: (account) => {
                      appToaster.add({
                        name: `sa-created-page-${account.serviceAccountId}`,
                        title: 'Service account created',
                        content: `${account.name} for ${account.tenantName}`,
                        theme: 'success',
                      });
                      navigate({
                        to: '/service-accounts/$serviceAccountId',
                        params: {serviceAccountId: account.serviceAccountId},
                      });
                    },
                  },
                );
              }}
            >
              Create Service Account
            </Button>
          </Flex>
        }
      />
      <SectionCard
        title="Service account metadata"
        description="Choose the tenant scope and machine identity attributes."
      >
        <div className="form-grid">
          <Select
            label="Tenant"
            value={tenantId ? [tenantId] : []}
            options={tenantOptions}
            placeholder="Choose tenant"
            onUpdate={(value) => setTenantId(value[0] ?? '')}
          />
          <TextInput
            label="Display name"
            placeholder="billing-worker"
            value={displayName}
            onUpdate={setDisplayName}
          />
          <div className="form-grid__full">
            <Text variant="body-2">Description</Text>
            <TextArea
              rows={4}
              value={description}
              onUpdate={setDescription}
              placeholder="Worker for invoice sync and exports"
            />
          </div>
          <Text variant="body-1" color="secondary">
            The service account is created in the selected tenant and becomes available immediately in the IAM inventory.
          </Text>
        </div>
      </SectionCard>
    </div>
  );
}

export function ServiceAccountsPage() {
  const accountsQuery = useServiceAccountsQuery();
  const navigate = useNavigate();

  if (accountsQuery.isPending) {
    return <LoadingState title="Loading service accounts" />;
  }

  if (accountsQuery.error || !accountsQuery.data) {
    return <ErrorState description={getErrorMessage(accountsQuery.error)} />;
  }

  const columns: ColumnDef<(typeof accountsQuery.data.items)[number]>[] = [
    {accessorKey: 'serviceAccountId', header: 'ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'tenantName', header: 'Tenant'},
    {accessorKey: 'ownerTeam', header: 'Owner team'},
    {accessorKey: 'keysCount', header: 'Keys'},
    {accessorKey: 'apiKeysCount', header: 'API Keys'},
    {accessorKey: 'updatedAt', header: 'Updated', cell: ({row}) => formatDateTime(row.original.updatedAt)},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button
          size="s"
          view="flat"
          onClick={() =>
            navigate({
              to: '/service-accounts/$serviceAccountId',
              params: {serviceAccountId: row.original.serviceAccountId},
            })
          }
        >
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader
        title="Service Accounts"
        description="Machine identities, keys and token footprints."
        actions={
          <Button view="action" onClick={() => navigate({to: '/service-accounts/create'})}>
            Create Service Account
          </Button>
        }
      />
      <DataTableCard
        title="Service accounts"
        description="All machine identities visible in the current IAM context."
        data={accountsQuery.data.items}
        columns={columns}
        emptyTitle="No service accounts"
      />
    </div>
  );
}

export function ServiceAccountDetailPage({tab}: {tab: 'overview' | 'keys'}) {
  const params = useParams({strict: false}) as {serviceAccountId: string};
  const accountQuery = useServiceAccountQuery(params.serviceAccountId);
  const effectiveAccessQuery = useEffectiveAccessQuery();
  const navigate = useNavigate();
  const [innerTab, setInnerTab] = React.useState<'overview' | 'keys' | 'tokens' | 'access'>(tab);

  React.useEffect(() => {
    setInnerTab(tab);
  }, [tab]);

  if (accountQuery.isPending) {
    return <LoadingState title="Loading service account" />;
  }

  if (accountQuery.error || !accountQuery.data) {
    return <ErrorState description={getErrorMessage(accountQuery.error)} />;
  }

  const account = accountQuery.data;
  const effectiveAccess = (effectiveAccessQuery.data ?? []).filter((row) => row.subjectId === account.serviceAccountId);

  const keyColumns: ColumnDef<(typeof account.asymmetricKeys)[number]>[] = [
    {accessorKey: 'id', header: 'Key ID'},
    {accessorKey: 'algorithm', header: 'Algorithm'},
    {accessorKey: 'createdAt', header: 'Created', cell: ({row}) => formatDateTime(row.original.createdAt)},
    {accessorKey: 'lastUsedAt', header: 'Last used', cell: ({row}) => formatDateTime(row.original.lastUsedAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];
  const apiKeyColumns: ColumnDef<(typeof account.apiKeys)[number]>[] = [
    {accessorKey: 'prefix', header: 'Prefix'},
    {accessorKey: 'createdAt', header: 'Created', cell: ({row}) => formatDateTime(row.original.createdAt)},
    {accessorKey: 'lastUsedAt', header: 'Last used', cell: ({row}) => formatDateTime(row.original.lastUsedAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];
  const tokenColumns: ColumnDef<(typeof account.tokens)[number]>[] = [
    {accessorKey: 'client', header: 'Client'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'protectionLevel', header: 'Protection'},
    {accessorKey: 'expiresAt', header: 'Expires', cell: ({row}) => formatDateTime(row.original.expiresAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow={account.tenantName}
        title={account.name}
        description={account.description}
        actions={
          <Button view="outlined" onClick={() => showBulkToast('Service account rotation flow pending', 1)}>
            Rotate Keys
          </Button>
        }
      />
      <SectionCard title="Identity summary" description="Service account metadata and auth posture.">
        <KeyValueGrid
          items={[
            {label: 'Service account ID', value: account.serviceAccountId},
            {label: 'Tenant', value: account.tenantName},
            {label: 'Owner team', value: account.ownerTeam},
            {label: 'Status', value: <StatusBadge status={account.status} />},
            {label: 'Created', value: formatShortDate(account.createdAt)},
            {label: 'Last auth', value: formatDateTime(account.lastAuthAt)},
          ]}
        />
      </SectionCard>
      <DetailTabs
        items={[
          {id: 'overview', title: 'Overview'},
          {id: 'keys', title: 'Keys'},
          {id: 'tokens', title: 'Tokens'},
          {id: 'access', title: 'Access'},
        ]}
        activeTab={innerTab}
        onSelectTab={(nextTab) => {
          setInnerTab(nextTab as typeof innerTab);
          if (nextTab === 'keys') {
            navigate({
              to: '/service-accounts/$serviceAccountId/keys',
              params: {serviceAccountId: account.serviceAccountId},
            });
            return;
          }
          navigate({
            to: '/service-accounts/$serviceAccountId',
            params: {serviceAccountId: account.serviceAccountId},
          });
        }}
      />
      {innerTab === 'overview' ? (
        <SectionCard title="Description" description="Primary ownership context.">
          <Text variant="body-2">{account.description}</Text>
        </SectionCard>
      ) : null}
      {innerTab === 'keys' ? (
        <div className="two-column-grid">
          <DataTableCard title="Asymmetric keys" description="Keypair inventory" data={account.asymmetricKeys} columns={keyColumns} emptyTitle="No keys" />
          <DataTableCard title="API keys" description="Issued API keys" data={account.apiKeys} columns={apiKeyColumns} emptyTitle="No API keys" />
        </div>
      ) : null}
      {innerTab === 'tokens' ? (
        <DataTableCard title="Tokens" description="Current issued tokens" data={account.tokens} columns={tokenColumns} emptyTitle="No tokens" />
      ) : null}
      {innerTab === 'access' ? (
        <EffectiveAccessTable rows={effectiveAccess} title="Effective access" description="Runtime permissions granted to this service account." />
      ) : null}
    </div>
  );
}

export function ServiceAccountsOverviewRoutePage() {
  return <ServiceAccountDetailPage tab="overview" />;
}

export function ServiceAccountsKeysRoutePage() {
  return <ServiceAccountDetailPage tab="keys" />;
}

export function OAuthClientsPage() {
  const clientsQuery = useOAuthClientsQuery();
  const navigate = useNavigate();

  if (clientsQuery.isPending) {
    return <LoadingState title="Loading OAuth clients" />;
  }

  if (clientsQuery.error || !clientsQuery.data) {
    return <ErrorState description={getErrorMessage(clientsQuery.error)} />;
  }

  const columns: ColumnDef<(typeof clientsQuery.data.items)[number]>[] = [
    {accessorKey: 'clientId', header: 'Client ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'tenantName', header: 'Tenant'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'scopesCount', header: 'Scopes'},
    {accessorKey: 'redirectUrisCount', header: 'Redirect URIs'},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/oauth-clients/$clientId', params: {clientId: row.original.clientId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="OAuth Clients" description="OIDC client registry and secret operations." />
      <DataTableCard title="Clients" description="Current interactive and machine-facing OAuth clients." data={clientsQuery.data.items} columns={columns} emptyTitle="No clients" />
    </div>
  );
}

export function OAuthClientDetailPage() {
  const params = useParams({strict: false}) as {clientId: string};
  const clientQuery = useOAuthClientQuery(params.clientId);
  const [rotateOpen, setRotateOpen] = React.useState(false);

  if (clientQuery.isPending) {
    return <LoadingState title="Loading OAuth client" />;
  }

  if (clientQuery.error || !clientQuery.data) {
    return <ErrorState description={getErrorMessage(clientQuery.error)} />;
  }

  const client = clientQuery.data;
  const secretColumns: ColumnDef<(typeof client.secrets)[number]>[] = [
    {accessorKey: 'name', header: 'Secret'},
    {accessorKey: 'createdAt', header: 'Created', cell: ({row}) => formatDateTime(row.original.createdAt)},
    {accessorKey: 'expiresAt', header: 'Expires', cell: ({row}) => formatDateTime(row.original.expiresAt)},
    {accessorKey: 'note', header: 'Note'},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];
  const tokenColumns: ColumnDef<(typeof client.tokens)[number]>[] = [
    {accessorKey: 'client', header: 'Client'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'expiresAt', header: 'Expires', cell: ({row}) => formatDateTime(row.original.expiresAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow={client.tenantName}
        title={client.name}
        description={`${client.type} OAuth client`}
        actions={<Button view="action" onClick={() => setRotateOpen(true)}>Rotate Secret</Button>}
      />
      <SectionCard title="Client summary" description="Primary client metadata and redirect scope.">
        <KeyValueGrid
          items={[
            {label: 'Client ID', value: client.clientId},
            {label: 'Tenant', value: client.tenantName},
            {label: 'Type', value: titleFromId(client.type)},
            {label: 'Status', value: <StatusBadge status={client.status} />},
            {label: 'Created', value: formatShortDate(client.createdAt)},
            {label: 'Updated', value: formatDateTime(client.updatedAt)},
          ]}
        />
      </SectionCard>
      <div className="two-column-grid">
        <SectionCard title="Redirect URIs" description="Configured redirect endpoints.">
          <ResourceList items={client.redirectUris} />
        </SectionCard>
        <SectionCard title="Scopes" description="Configured client scopes.">
          <PillList items={client.scopes} />
        </SectionCard>
      </div>
      <div className="two-column-grid">
        <DataTableCard title="Secrets" description="Secret rotation history and expiration metadata." data={client.secrets} columns={secretColumns} emptyTitle="No secrets" />
        <DataTableCard title="Tokens" description="Issued tokens for this client." data={client.tokens} columns={tokenColumns} emptyTitle="No tokens" />
      </div>
      <RotateSecretDialog open={rotateOpen} clientId={client.clientId} clientName={client.name} onClose={() => setRotateOpen(false)} />
    </div>
  );
}

export function RolesPage() {
  const rolesQuery = useRolesQuery();
  const navigate = useNavigate();

  if (rolesQuery.isPending) {
    return <LoadingState title="Loading roles" />;
  }

  if (rolesQuery.error || !rolesQuery.data) {
    return <ErrorState description={getErrorMessage(rolesQuery.error)} />;
  }

  const columns: ColumnDef<(typeof rolesQuery.data.items)[number]>[] = [
    {accessorKey: 'roleId', header: 'Role ID'},
    {accessorKey: 'namespace', header: 'Namespace'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'permissionsCount', header: 'Permissions'},
    {accessorKey: 'system', header: 'System', cell: ({row}) => (row.original.system ? 'Yes' : 'No')},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/roles/$roleId', params: {roleId: row.original.roleId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="Roles Catalog" description="Canonical role model and permission expansion." />
      <DataTableCard title="Roles" description="System and custom roles resolved by the authorization facade." data={rolesQuery.data.items} columns={columns} emptyTitle="No roles" />
    </div>
  );
}

export function RoleDetailPage() {
  const params = useParams({strict: false}) as {roleId: string};
  const roleQuery = useRoleQuery(params.roleId);

  if (roleQuery.isPending) {
    return <LoadingState title="Loading role" />;
  }

  if (roleQuery.error || !roleQuery.data) {
    return <ErrorState description={getErrorMessage(roleQuery.error)} />;
  }

  const role = roleQuery.data as RoleDetail;
  const permissionColumns: ColumnDef<(typeof role.permissions)[number]>[] = [
    {accessorKey: 'id', header: 'Permission'},
    {accessorKey: 'displayName', header: 'Display name'},
    {accessorKey: 'description', header: 'Description'},
  ];

  return (
    <div className="page-stack">
      <PageHeader eyebrow={role.namespace} title={role.name} description={role.description} />
      <SectionCard title="Role summary" description="Permission set and reuse footprint.">
        <KeyValueGrid
          items={[
            {label: 'Role ID', value: role.roleId},
            {label: 'Namespace', value: role.namespace},
            {label: 'System role', value: role.system ? 'Yes' : 'No'},
            {label: 'Permissions', value: formatCount(role.permissionsCount)},
          ]}
        />
      </SectionCard>
      <div className="two-column-grid">
        <DataTableCard title="Permissions" description="Expanded runtime permissions." data={role.permissions} columns={permissionColumns} emptyTitle="No permissions" />
        <SectionCard title="Used by" description="Known bindings and inherited assignments.">
          <ResourceList items={role.usedBy} />
        </SectionCard>
      </div>
    </div>
  );
}

export function ResourceAccessPage() {
  const params = useParams({strict: false}) as {resourceType: string; resourceId: string};
  const {context} = useAppUI();
  const bindingsQuery = useAccessBindingsQuery(params.resourceType, params.resourceId);
  const [drawerOpen, setDrawerOpen] = React.useState(false);

  if (bindingsQuery.isPending) {
    return <LoadingState title="Loading resource access" />;
  }

  if (bindingsQuery.error) {
    return <ErrorState description={getErrorMessage(bindingsQuery.error)} />;
  }

  const bindings = bindingsQuery.data ?? [];
  const directCount = bindings.filter((item) => item.source === 'direct').length;
  const groupCount = bindings.filter((item) => item.source === 'group').length;
  const policyCount = bindings.filter((item) => item.source === 'policy').length;

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="Resource access"
        title={`${params.resourceType}/${params.resourceId}`}
        description="Direct and inherited bindings resolved for a single protected resource."
        actions={<Button view="action" onClick={() => setDrawerOpen(true)}>Grant Access</Button>}
      />
      <MetricGrid
        items={[
          {id: 'all', title: 'Bindings', value: formatCount(bindings.length), tone: 'info'},
          {id: 'direct', title: 'Direct', value: formatCount(directCount), tone: 'success'},
          {id: 'group', title: 'Group', value: formatCount(groupCount), tone: 'warning'},
          {id: 'policy', title: 'Policy', value: formatCount(policyCount), tone: 'info'},
        ]}
      />
      <AccessBindingsTable bindings={bindings} />
      <GrantAccessDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        tenantId={context.tenantId}
        resourceType={params.resourceType as ResourceType}
        resourceId={params.resourceId}
      />
    </div>
  );
}

export function AccessExplorerPage() {
  const effectiveAccessQuery = useEffectiveAccessQuery();
  const [query, setQuery] = React.useState('');
  const deferredQuery = React.useDeferredValue(query);

  if (effectiveAccessQuery.isPending) {
    return <LoadingState title="Loading access explorer" />;
  }

  if (effectiveAccessQuery.error || !effectiveAccessQuery.data) {
    return <ErrorState description={getErrorMessage(effectiveAccessQuery.error)} />;
  }

  const rows = effectiveAccessQuery.data.filter((row) =>
    matchesQuery(deferredQuery, [row.subjectName, row.subjectId, row.resourceId, row.permission, row.roleId]),
  );

  return (
    <div className="page-stack">
      <PageHeader title="Access Explorer" description="Flattened effective permission rows for subjects and resources." />
      <SectionCard title="Search" description="Filter by subject, resource or permission.">
        <div className="filter-bar">
          <TextInput label="Search" value={query} onUpdate={setQuery} placeholder="subject / permission / resource" />
        </div>
      </SectionCard>
      <EffectiveAccessTable rows={rows} title="Effective access" />
    </div>
  );
}

export function AccessExplainPage() {
  const usersQuery = useUsersQuery();
  const [subjectId, setSubjectId] = React.useState('user-demo-admin');
  const [resourceId, setResourceId] = React.useState('project-demo-infra');
  const [permission, setPermission] = React.useState('project.write');
  const [requested, setRequested] = React.useState(true);
  const explainQuery = useExplainAccessQuery(subjectId, resourceId, permission, requested);

  const subjectOptions = (usersQuery.data?.items ?? []).map((user) => ({
    value: user.userId,
    content: `${user.name} (${user.userId})`,
  }));

  return (
    <div className="page-stack">
      <PageHeader title="Explain Access" description="Trace runtime authorization decisions step by step." />
      <SectionCard title="Explain request" description="Evaluate one subject, resource and permission combination.">
        <div className="filter-bar">
          <Select label="Subject" value={[subjectId]} options={subjectOptions} onUpdate={(value) => setSubjectId(value[0] ?? '')} />
          <TextInput label="Resource ID" value={resourceId} onUpdate={setResourceId} placeholder="project-demo-infra" />
          <TextInput label="Permission" value={permission} onUpdate={setPermission} placeholder="project.write" />
          <Button
            view="action"
            onClick={() => setRequested(true)}
          >
            Evaluate
          </Button>
        </div>
      </SectionCard>
      {explainQuery.isPending ? <LoadingState title="Evaluating authorization path" /> : null}
      {explainQuery.error ? <ErrorState description={getErrorMessage(explainQuery.error)} /> : null}
      {explainQuery.data ? (
        <div className="two-column-grid">
          <SectionCard title="Decision" description="Final runtime decision and path identifiers.">
            <KeyValueGrid
              items={[
                {label: 'Decision', value: <ToneBadge tone={explainQuery.data.decision === 'allow' ? 'success' : 'danger'}>{titleFromId(explainQuery.data.decision)}</ToneBadge>},
                {label: 'Permission', value: explainQuery.data.permission},
                {label: 'Evaluated at', value: formatDateTime(explainQuery.data.evaluatedAt)},
              ]}
            />
            <Text variant="body-2">{explainQuery.data.summary}</Text>
            <PillList items={explainQuery.data.pathIds} />
          </SectionCard>
          <SectionCard title="Evaluation steps" description="Collapsed explain tree prepared for operators.">
            <div className="timeline-list">
              {explainQuery.data.steps.map((step) => (
                <div key={step.id} className="timeline-list__item">
                  <ToneBadge tone="info">{step.id}</ToneBadge>
                  <div>
                    <Text variant="subheader-1">{step.title}</Text>
                    <ResourceList items={step.details} />
                  </div>
                </div>
              ))}
            </div>
          </SectionCard>
        </div>
      ) : null}
    </div>
  );
}

export function AccessSimulatePage() {
  const usersQuery = useUsersQuery();
  const rolesQuery = useRolesQuery();
  const [subjectId, setSubjectId] = React.useState('user-demo-analyst');
  const [resourceId, setResourceId] = React.useState('project-demo-analytics');
  const [roleId, setRoleId] = React.useState('project-viewer');
  const [requested, setRequested] = React.useState(true);
  const impactQuery = useImpactSimulationQuery(resourceId, subjectId, roleId, requested);

  const subjectOptions = (usersQuery.data?.items ?? []).map((user) => ({
    value: user.userId,
    content: `${user.name} (${user.userId})`,
  }));
  const roleOptions = (rolesQuery.data?.items ?? []).map((role) => ({
    value: role.roleId,
    content: `${role.name} (${role.roleId})`,
  }));

  const columns: ColumnDef<ImpactRow>[] = [
    {accessorKey: 'subjectName', header: 'Subject'},
    {accessorKey: 'before', header: 'Before'},
    {accessorKey: 'after', header: 'After'},
    {accessorKey: 'status', header: 'Impact', cell: ({row}) => <ToneBadge tone={row.original.status}>{titleFromId(row.original.status)}</ToneBadge>},
    {accessorKey: 'affectedPermissions', header: 'Affected permissions', cell: ({row}) => row.original.affectedPermissions.join(', ')},
  ];

  return (
    <div className="page-stack">
      <PageHeader title="Simulate Access Change" description="Preview permission deltas before applying authz mutations." />
      <SectionCard title="Simulation input" description="The current API exposes impact simulation over graph mutations.">
        <div className="filter-bar">
          <Select label="Subject" value={[subjectId]} options={subjectOptions} onUpdate={(value) => setSubjectId(value[0] ?? '')} />
          <TextInput label="Resource ID" value={resourceId} onUpdate={setResourceId} />
          <Select label="Role" value={[roleId]} options={roleOptions} onUpdate={(value) => setRoleId(value[0] ?? '')} />
          <Button view="action" onClick={() => setRequested(true)}>Simulate</Button>
        </div>
      </SectionCard>
      {impactQuery.isPending ? <LoadingState title="Running impact analysis" /> : null}
      {impactQuery.error ? <ErrorState description={getErrorMessage(impactQuery.error)} /> : null}
      {impactQuery.data ? <DataTableCard title="Impact analysis" description="Subjects and permissions affected by the delta." data={impactQuery.data} columns={columns} emptyTitle="No impact" /> : null}
    </div>
  );
}

export function PoliciesPage() {
  const policiesQuery = usePolicyTemplatesQuery();
  const navigate = useNavigate();

  if (policiesQuery.isPending) {
    return <LoadingState title="Loading policy templates" />;
  }

  if (policiesQuery.error || !policiesQuery.data) {
    return <ErrorState description={getErrorMessage(policiesQuery.error)} />;
  }

  const columns: ColumnDef<(typeof policiesQuery.data.items)[number]>[] = [
    {accessorKey: 'templateId', header: 'Template ID'},
    {accessorKey: 'name', header: 'Name'},
    {accessorKey: 'scope', header: 'Scope'},
    {accessorKey: 'parameters', header: 'Parameters', cell: ({row}) => row.original.parameters.join(', ')},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/policies/$templateId', params: {templateId: row.original.templateId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="Policy Templates" description="Reusable binding blueprints for common access shapes." />
      <DataTableCard title="Templates" description="Catalog of reusable IAM templates." data={policiesQuery.data.items} columns={columns} emptyTitle="No policy templates" />
    </div>
  );
}

export function PolicyDetailPage() {
  const params = useParams({strict: false}) as {templateId: string};
  const policyQuery = usePolicyTemplateQuery(params.templateId);

  if (policyQuery.isPending) {
    return <LoadingState title="Loading policy template" />;
  }

  if (policyQuery.error || !policyQuery.data) {
    return <ErrorState description={getErrorMessage(policyQuery.error)} />;
  }

  const policy = policyQuery.data;

  return (
    <div className="page-stack">
      <PageHeader eyebrow={policy.scope} title={policy.name} description={policy.description} />
      <div className="two-column-grid">
        <SectionCard title="Parameters" description="Template inputs required by the workflow.">
          <PillList items={policy.parameters} />
        </SectionCard>
        <SectionCard title="Generated bindings" description="Bindings emitted by the template expansion.">
          <PillList items={policy.generatedBindings} />
        </SectionCard>
      </div>
    </div>
  );
}

export function SupportAccessPage() {
  const supportQuery = useSupportGrantsQuery();
  const [wizardOpen, setWizardOpen] = React.useState(false);

  if (supportQuery.isPending) {
    return <LoadingState title="Loading support access" />;
  }

  if (supportQuery.error || !supportQuery.data) {
    return <ErrorState description={getErrorMessage(supportQuery.error)} />;
  }

  const columns: ColumnDef<SupportGrant>[] = [
    {accessorKey: 'tenantName', header: 'Tenant'},
    {accessorKey: 'subjectName', header: 'Subject'},
    {accessorKey: 'roleId', header: 'Role'},
    {accessorKey: 'incidentId', header: 'Incident'},
    {accessorKey: 'expiresAt', header: 'Expires', cell: ({row}) => formatDateTime(row.original.expiresAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  return (
    <div className="page-stack">
      <PageHeader
        title="Support Access"
        description="Temporary customer support sessions with explicit expiry."
        actions={<Button view="action" onClick={() => setWizardOpen(true)}>Grant Temporary Access</Button>}
      />
      <HighlightAlert
        title="Support workflows"
        message="This screen drives the same support-grant API used by the backend seed and Postman collection."
      />
      <DataTableCard title="Support grants" description="Temporary support sessions and approval outcomes." data={supportQuery.data.items} columns={columns} emptyTitle="No support grants" />
      <SupportGrantWizard open={wizardOpen} onClose={() => setWizardOpen(false)} />
    </div>
  );
}

export function SessionsPage() {
  const sessionsQuery = useSessionsQuery();
  const [statusOnlyActive, setStatusOnlyActive] = React.useState(false);

  if (sessionsQuery.isPending) {
    return <LoadingState title="Loading sessions and tokens" />;
  }

  if (sessionsQuery.error || !sessionsQuery.data) {
    return <ErrorState description={getErrorMessage(sessionsQuery.error)} />;
  }

  const items = statusOnlyActive
    ? sessionsQuery.data.items.filter((token) => token.status === 'active')
    : sessionsQuery.data.items;

  const columns: ColumnDef<(typeof items)[number]>[] = [
    {accessorKey: 'client', header: 'Client'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'protectionLevel', header: 'Protection'},
    {accessorKey: 'lastUsedAt', header: 'Last used', cell: ({row}) => formatDateTime(row.original.lastUsedAt)},
    {accessorKey: 'expiresAt', header: 'Expires', cell: ({row}) => formatDateTime(row.original.expiresAt)},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
  ];

  return (
    <div className="page-stack">
      <PageHeader title="Sessions / Tokens" description="Issued token inventory and current session posture." />
      <SectionCard title="Filters" description="Operational token view filters.">
        <Flex gap="4" alignItems="center" wrap>
          <Switch checked={statusOnlyActive} onUpdate={setStatusOnlyActive} content="Only active tokens" />
        </Flex>
      </SectionCard>
      <DataTableCard title="Tokens" description="Cross-subject token inventory." data={items} columns={columns} emptyTitle="No tokens" />
    </div>
  );
}

export function AuditPage() {
  const auditQuery = useAuditEventsQuery();
  const navigate = useNavigate();
  const [query, setQuery] = React.useState('');
  const deferredQuery = React.useDeferredValue(query);

  if (auditQuery.isPending) {
    return <LoadingState title="Loading audit log" />;
  }

  if (auditQuery.error || !auditQuery.data) {
    return <ErrorState description={getErrorMessage(auditQuery.error)} />;
  }

  const items = auditQuery.data.items.filter((event) =>
    matchesQuery(deferredQuery, [event.eventType, event.actor, event.resource, event.tenantId]),
  );

  const columns: ColumnDef<AuditEvent>[] = [
    {accessorKey: 'eventType', header: 'Event'},
    {accessorKey: 'actor', header: 'Actor'},
    {accessorKey: 'resource', header: 'Resource'},
    {accessorKey: 'tenantId', header: 'Tenant'},
    {accessorKey: 'occurredAt', header: 'Occurred', cell: ({row}) => formatDateTime(row.original.occurredAt)},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/audit/$eventId', params: {eventId: row.original.eventId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="Audit Log" description="Searchable record of authn, authz and lifecycle changes." />
      <SectionCard title="Search" description="Filter by event, actor, resource or tenant.">
        <div className="filter-bar">
          <TextInput label="Search" value={query} onUpdate={setQuery} placeholder="binding.added / actor / tenant" />
        </div>
      </SectionCard>
      <DataTableCard title="Audit events" description="Append-only event stream projected for operators." data={items} columns={columns} emptyTitle="No audit events" />
    </div>
  );
}

export function AuditEventPage() {
  const params = useParams({strict: false}) as {eventId: string};
  const eventQuery = useAuditEventQuery(params.eventId);

  if (eventQuery.isPending) {
    return <LoadingState title="Loading audit event" />;
  }

  if (eventQuery.error || !eventQuery.data) {
    return <ErrorState description={getErrorMessage(eventQuery.error)} />;
  }

  const event = eventQuery.data;

  return (
    <div className="page-stack">
      <PageHeader eyebrow={event.tenantId} title={event.eventType} description={event.eventId} />
      <SectionCard title="Event summary" description="Audit envelope and business context.">
        <KeyValueGrid
          items={[
            {label: 'Actor', value: event.actor},
            {label: 'Resource', value: event.resource},
            {label: 'Result', value: event.result},
            {label: 'Occurred', value: formatDateTime(event.occurredAt)},
            {label: 'Reason', value: event.reason ?? '—'},
          ]}
        />
      </SectionCard>
      <SectionCard title="Payload" description="Raw event payload returned by the current API.">
        <JsonCodeBlock value={event.payload} />
      </SectionCard>
    </div>
  );
}

export function OperationsPage() {
  const operationsQuery = useOperationsQuery();
  const navigate = useNavigate();

  if (operationsQuery.isPending) {
    return <LoadingState title="Loading operations" />;
  }

  if (operationsQuery.error || !operationsQuery.data) {
    return <ErrorState description={getErrorMessage(operationsQuery.error)} />;
  }

  const columns: ColumnDef<Operation>[] = [
    {accessorKey: 'operationId', header: 'Operation ID'},
    {accessorKey: 'type', header: 'Type'},
    {accessorKey: 'resource', header: 'Resource'},
    {accessorKey: 'actor', header: 'Actor'},
    {accessorKey: 'status', header: 'Status', cell: ({row}) => <StatusBadge status={row.original.status} />},
    {accessorKey: 'startedAt', header: 'Started', cell: ({row}) => formatDateTime(row.original.startedAt)},
    {
      id: 'open',
      header: '',
      cell: ({row}) => (
        <Button size="s" view="flat" onClick={() => navigate({to: '/operations/$operationId', params: {operationId: row.original.operationId}})}>
          Open
        </Button>
      ),
    },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="Operations" description="Long-running IAM workflows and orchestration state." />
      <DataTableCard title="Operations" description="Temporal-backed asynchronous workflows and progress state." data={operationsQuery.data.items} columns={columns} emptyTitle="No operations" />
    </div>
  );
}

export function OperationDetailPage() {
  const params = useParams({strict: false}) as {operationId: string};
  const operationQuery = useOperationQuery(params.operationId);

  if (operationQuery.isPending) {
    return <LoadingState title="Loading operation" />;
  }

  if (operationQuery.error || !operationQuery.data) {
    return <ErrorState description={getErrorMessage(operationQuery.error)} />;
  }

  const operation = operationQuery.data;

  return (
    <div className="page-stack">
      <PageHeader eyebrow={operation.tenantId} title={operation.type} description={operation.operationId} />
      <SectionCard title="Operation summary" description="Workflow metadata and completion state.">
        <KeyValueGrid
          items={[
            {label: 'Resource', value: operation.resource},
            {label: 'Actor', value: operation.actor},
            {label: 'Status', value: <StatusBadge status={operation.status} />},
            {label: 'Started', value: formatDateTime(operation.startedAt)},
            {label: 'Updated', value: formatDateTime(operation.updatedAt)},
            {label: 'Completed', value: formatDateTime(operation.completedAt)},
          ]}
        />
      </SectionCard>
      <div className="two-column-grid">
        <SectionCard title="Workflow steps" description="Execution progress across orchestration stages.">
          <OperationTimeline steps={operation.steps} />
        </SectionCard>
        <SectionCard title="Logs" description="Latest operation log lines.">
          <JsonCodeBlock value={operation.logs} />
        </SectionCard>
      </div>
    </div>
  );
}

export function SettingsPage() {
  const settingsQuery = useSettingsQuery();
  const saveMutation = useSaveSettingsMutation();
  const [draft, setDraft] = React.useState<SettingsSection[]>([]);

  React.useEffect(() => {
    if (settingsQuery.data) {
      setDraft(settingsQuery.data);
    }
  }, [settingsQuery.data]);

  if (settingsQuery.isPending && draft.length === 0) {
    return <LoadingState title="Loading settings" />;
  }

  if (settingsQuery.error && draft.length === 0) {
    return <ErrorState description={getErrorMessage(settingsQuery.error)} />;
  }

  return (
    <div className="page-stack">
      <PageHeader
        title="Settings / Integrations"
        description="Global defaults that shape token policies, audit retention and runtime behavior."
        actions={
          <Button
            view="action"
            loading={saveMutation.isPending}
            onClick={() =>
              saveMutation.mutate(draft, {
                onSuccess: () => {
                  appToaster.add({
                    name: 'settings-saved',
                    title: 'Settings saved',
                    content: 'Global IAM configuration has been updated.',
                    theme: 'success',
                  });
                },
              })
            }
          >
            Save Changes
          </Button>
        }
      />
      <div className="settings-stack">
        {draft.map((section) => (
          <SectionCard key={section.id} title={section.title} description={section.description}>
            <div className="settings-fields">
              {section.fields.map((field) => (
                <div key={field.id} className="settings-field">
                  {field.type === 'switch' ? (
                    <Switch
                      checked={Boolean(field.value)}
                      content={field.label}
                      onUpdate={(checked) =>
                        setDraft((current) =>
                          current.map((item) =>
                            item.id === section.id
                              ? {
                                  ...item,
                                  fields: item.fields.map((entry) =>
                                    entry.id === field.id ? {...entry, value: checked} : entry,
                                  ),
                                }
                              : item,
                          ),
                        )
                      }
                    />
                  ) : field.type === 'select' ? (
                    <Select
                      label={field.label}
                      value={[String(field.value)]}
                      options={(field.options ?? []).map((option) => ({value: option, content: option}))}
                      onUpdate={(value) =>
                        setDraft((current) =>
                          current.map((item) =>
                            item.id === section.id
                              ? {
                                  ...item,
                                  fields: item.fields.map((entry) =>
                                    entry.id === field.id ? {...entry, value: value[0] ?? ''} : entry,
                                  ),
                                }
                              : item,
                          ),
                        )
                      }
                    />
                  ) : (
                    <div className="settings-field-textarea">
                      <Text variant="body-2">{field.label}</Text>
                      <TextArea
                        value={String(field.value)}
                        rows={3}
                        onUpdate={(value) =>
                          setDraft((current) =>
                            current.map((item) =>
                              item.id === section.id
                                ? {
                                    ...item,
                                    fields: item.fields.map((entry) =>
                                      entry.id === field.id ? {...entry, value} : entry,
                                    ),
                                  }
                                : item,
                            ),
                          )
                        }
                      />
                    </div>
                  )}
                </div>
              ))}
            </div>
          </SectionCard>
        ))}
      </div>
    </div>
  );
}

export function SearchPage() {
  const {globalSearch, setGlobalSearch} = useAppUI();
  const [query, setQuery] = React.useState(globalSearch);
  const deferredQuery = React.useDeferredValue(query);
  const searchQuery = useSearchResultsQuery(deferredQuery);
  const navigate = useNavigate();

  React.useEffect(() => {
    setGlobalSearch(query);
  }, [query, setGlobalSearch]);

  return (
    <div className="page-stack">
      <PageHeader title="Global Search" description="Cross-entity search results over the seeded domain catalog." />
      <SectionCard title="Search query" description="Search roles, groups, audit events and more.">
        <div className="filter-bar">
          <TextInput
            label="Search"
            value={query}
            placeholder="project-editor / grp-demo-ops / evt-..."
            onUpdate={setQuery}
          />
        </div>
      </SectionCard>
      {deferredQuery.trim().length === 0 ? (
        <EmptyState title="Start typing" description="Search catalog results appear here as soon as a query is entered." />
      ) : searchQuery.isPending ? (
        <LoadingState title="Searching catalog" />
      ) : searchQuery.error ? (
        <ErrorState description={getErrorMessage(searchQuery.error)} />
      ) : (
        <SectionCard title="Results" description={`${formatCount(searchQuery.data?.items.length ?? 0)} matches`}>
          <div className="search-results">
            {(searchQuery.data?.items ?? []).map((result: SearchResult) => (
              <button
                key={result.id}
                type="button"
                className="search-result-card"
                onClick={() => navigate({to: result.href})}
              >
                <div>
                  <Flex gap="2" alignItems="center">
                    <Label theme="utility" size="s">{result.type}</Label>
                    <Text variant="subheader-1">{result.title}</Text>
                  </Flex>
                  <Text variant="body-1" color="secondary">
                    {result.context}
                  </Text>
                  <Text variant="body-2">{result.description}</Text>
                </div>
                <Icon data={ArrowUpRightFromSquare} />
              </button>
            ))}
          </div>
        </SectionCard>
      )}
    </div>
  );
}
