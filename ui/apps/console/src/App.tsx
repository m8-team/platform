import {useCallback, useEffect, useMemo, useState} from 'react'
import {
  Avatar,
  Button,
  Card,
  configure,
  Icon,
  Label,
  Select,
  Text,
  TextInput,
  ThemeProvider,
} from '@gravity-ui/uikit'
import {ActionBar, AsideHeader, FooterItem} from '@gravity-ui/navigation'
import type {AsideHeaderItem, MenuGroup, PanelItemProps} from '@gravity-ui/navigation'
import {
  ArrowShapeRightFromLine,
  BellDot,
  Briefcase,
  Check,
  CircleQuestion,
  Clock,
  Cloud,
  Code,
  Database,
  Fingerprint,
  Folders,
  Gear,
  ListUl,
  Magnifier,
  NodesRight,
  Person,
  Persons,
  Plus,
  Rocket,
  Shield,
  ShieldCheck,
  Signal,
  Speedometer,
  TriangleExclamation,
  ArrowRotateRight,
} from '@gravity-ui/icons'

import './App.css'

type ProjectStatus = 'Active' | 'Suspended' | 'Failed' | 'Provisioning' | 'Deleting'
type FooterPanel = 'notifications' | 'support' | 'account'

interface Project {
  name: string
  projectId: string
  workspace: string
  organization: string
  status: ProjectStatus
  desiredState: string
  actualState: string
  updated: string
  owner: string
  lastOperation: string
}

const initialLanguage = 'en'
const navigationCompactStorageKey = 'm8.console.navigation.compact'

const organizationOptions = [
  {value: 'org_m8_finance_6b21d0', content: 'Acme'},
  {value: 'org_m8_billing_91f2c5', content: 'Billing'},
]

const workspaceOptions = [
  {value: 'ws_prod-eu1', content: 'Platform'},
  {value: 'ws_shared-eu1', content: 'Shared Services'},
  {value: 'ws_legacy-eu1', content: 'Legacy'},
]

const statusOptions = [
  {value: 'all', content: 'All statuses'},
  {value: 'Active', content: 'Active'},
  {value: 'Suspended', content: 'Suspended'},
  {value: 'Failed', content: 'Failed'},
  {value: 'Provisioning', content: 'Provisioning'},
  {value: 'Deleting', content: 'Deleting'},
]

const ownerOptions = [
  {value: 'all', content: 'All owners'},
  {value: 'usr_19bd4027_sre', content: 'usr_19bd4027_sre'},
  {value: 'usr_2f0c81aa_sec', content: 'usr_2f0c81aa_sec'},
]

const projects: Project[] = [
  {
    name: 'IAM',
    projectId: 'prj_2e41d7a9c0bf4e55',
    workspace: 'ws_prod-eu1',
    organization: 'org_m8_finance_6b21d0',
    status: 'Provisioning',
    desiredState: 'Running',
    actualState: 'Provisioning',
    updated: '2026-06-23 09:31',
    owner: 'usr_19bd4027_sre',
    lastOperation: 'op_9fe2304db1a44e88',
  },
  {
    name: 'partner-settlement',
    projectId: 'prj_6d90aa31f48c4b8e',
    workspace: 'ws_prod-eu1',
    organization: 'org_m8_finance_6b21d0',
    status: 'Suspended',
    desiredState: 'Suspended',
    actualState: 'Suspended',
    updated: '2026-06-22 18:07',
    owner: 'usr_2f0c81aa_sec',
    lastOperation: 'op_63ab7e02d4104ba1',
  },
  {
    name: 'invoice-export',
    projectId: 'prj_41c2de83b7764a09',
    workspace: 'ws_shared-eu1',
    organization: 'org_m8_billing_91f2c5',
    status: 'Failed',
    desiredState: 'Running',
    actualState: 'Failed',
    updated: '2026-06-23 08:55',
    owner: 'usr_8b17d6f0_ops',
    lastOperation: 'op_3a7c8a2148ff47a2',
  },
  {
    name: 'legacy-reports',
    projectId: 'prj_9aa4c11d0e744f3a',
    workspace: 'ws_legacy-eu1',
    organization: 'org_m8_finance_6b21d0',
    status: 'Deleting',
    desiredState: 'Deleted',
    actualState: 'Deleting',
    updated: '2026-06-23 07:14',
    owner: 'usr_64ea18c2_admin',
    lastOperation: 'op_bdc4221d8b714c9d',
  },
]

const menuGroups: MenuGroup[] = [
  {id: 'resources', title: 'Resources', icon: Database},
  {id: 'identity-access', title: 'Identity & Access', icon: Shield},
  {id: 'gateway', title: 'Gateway', icon: Cloud},
  {id: 'security', title: 'Security & Risk', icon: Shield},
  {id: 'observability', title: 'Observability', icon: Clock},
  {id: 'audit', title: 'Audit', icon: ListUl},
  {id: 'settings', title: 'Settings', icon: Gear},
]

const menuItems: AsideHeaderItem[] = [
  {id: 'resources-organizations', title: 'Organizations', icon: Briefcase, groupId: 'resources'},
  {id: 'resources-workspaces', title: 'Workspaces', icon: Folders, groupId: 'resources'},
  {id: 'resources-project', title: 'Projects', icon: Database, current: true, groupId: 'resources'},
  {id: 'resources-quotas-limits', title: 'Quotas & Limits', icon: Speedometer, groupId: 'resources'},
  {id: 'identity-access-identity', title: 'Identity', icon: Person, groupId: 'identity-access'},
  {id: 'identity-access-authentication', title: 'Authentication', icon: Shield, groupId: 'identity-access'},
  {id: 'identity-access-control', title: 'Access Control', icon: ShieldCheck, groupId: 'identity-access'},
  {id: 'gateway-api-services', title: 'API Services', icon: Cloud, groupId: 'gateway'},
  {id: 'gateway-routes', title: 'Routes', icon: ArrowShapeRightFromLine, groupId: 'gateway'},
  {id: 'gateway-consumers', title: 'Consumers', icon: Persons, groupId: 'gateway'},
  {id: 'gateway-policies', title: 'Policies', icon: Check, groupId: 'gateway'},
  {id: 'gateway-rate-limits', title: 'Rate Limits', icon: Speedometer, groupId: 'gateway'},
  {id: 'gateway-yaml', title: 'Gateway YAML', icon: Code, groupId: 'gateway'},
  {id: 'security-dashboard', title: 'Dashboard', icon: Rocket, groupId: 'security'},
  {id: 'security-risk-rules', title: 'Risk Rules', icon: Shield, groupId: 'security'},
  {id: 'security-device-fingerprints', title: 'Device Fingerprints', icon: Fingerprint, groupId: 'security'},
  {id: 'security-velocity-rules', title: 'Velocity Rules', icon: Speedometer, groupId: 'security'},
  {id: 'security-signals', title: 'Signals', icon: Signal, groupId: 'security'},
  {id: 'security-decisions', title: 'Decisions', icon: Check, groupId: 'security'},
  {id: 'security-challenges', title: 'Challenges', icon: TriangleExclamation, groupId: 'security'},
  {id: 'security-fraud-cases', title: 'Fraud Cases', icon: Briefcase, groupId: 'security'},
  {id: 'security-events', title: 'Security Events', icon: ListUl, groupId: 'security'},
  {id: 'security-access-reviews', title: 'Access Reviews', icon: Persons, groupId: 'security'},
  {id: 'security-policy-violations', title: 'Policy Violations', icon: TriangleExclamation, groupId: 'security'},
  {id: 'observability-metrics', title: 'Metrics', icon: Speedometer, groupId: 'observability'},
  {id: 'observability-logs', title: 'Logs', icon: ListUl, groupId: 'observability'},
  {id: 'observability-traces', title: 'Traces', icon: NodesRight, groupId: 'observability'},
  {id: 'observability-alerts', title: 'Alerts', icon: TriangleExclamation, groupId: 'observability'},
  {id: 'observability-slo', title: 'SLO', icon: Check, groupId: 'observability'},
  {id: 'audit-events', title: 'Audit Events', icon: ListUl, groupId: 'audit'},
  {id: 'audit-exports', title: 'Exports', icon: ArrowRotateRight, groupId: 'audit'},
  {id: 'settings-project', title: 'Project Settings', icon: Gear, groupId: 'settings'},
  {id: 'settings-modules', title: 'Modules', icon: Database, groupId: 'settings'},
  {id: 'settings-integrations', title: 'Integrations', icon: Cloud, groupId: 'settings'},
  {id: 'settings-webhooks', title: 'Webhooks', icon: ArrowShapeRightFromLine, groupId: 'settings'},
  {id: 'settings-api-tokens', title: 'API Tokens', icon: Shield, groupId: 'settings'},
]

const defaultMenuItems: AsideHeaderItem[] = [
  ...menuItems,
  {
    id: 'observability-incidents',
    title: 'Incidents',
    icon: TriangleExclamation,
    groupId: 'observability',
    hidden: true,
  },
]

configure({
  lang: initialLanguage,
  fallbackLang: 'en',
})

function readInitialNavigationCompact() {
  if (typeof window === 'undefined') {
    return false
  }

  try {
    return window.localStorage.getItem(navigationCompactStorageKey) === 'true'
  } catch {
    return false
  }
}

function App() {
  const [compact, setCompact] = useState(readInitialNavigationCompact)
  const [navigationItems, setNavigationItems] = useState(menuItems)
  const [activeFooterPanel, setActiveFooterPanel] = useState<FooterPanel | null>(null)
  const [organization, setOrganization] = useState('org_m8_finance_6b21d0')
  const [workspace, setWorkspace] = useState('ws_prod-eu1')
  const [projectId, setProjectId] = useState('prj_2e41d7a9c0bf4e55')
  const [status, setStatus] = useState('all')
  const [owner, setOwner] = useState('all')
  const [search, setSearch] = useState('')

  const handleNavigationCompactChange = useCallback((nextCompact: boolean) => {
    setCompact(nextCompact)

    try {
      window.localStorage.setItem(navigationCompactStorageKey, String(nextCompact))
    } catch {
      // Storage can be unavailable in private or restricted browser contexts.
    }
  }, [])

  useEffect(() => {
    document.documentElement.lang = initialLanguage
  }, [])

  const subheaderItems = useMemo<AsideHeaderItem[]>(
    () => [
      {
        id: 'subheader-dashboard',
        title: 'Dashboard',
        icon: Rocket,
      },
    ],
    [],
  )

  const panelItems = useMemo<PanelItemProps[]>(
    () => [
      {
        id: 'notifications',
        open: activeFooterPanel === 'notifications',
        size: 360,
        hideVeil: true,
        children: (
          <AsidePanel
            title="Notifications"
            description="Recent platform events that need operator attention."
            items={['Quota warning in Platform workspace', 'Gateway route policy updated', 'Audit export completed']}
          />
        ),
      },
      {
        id: 'support',
        open: activeFooterPanel === 'support',
        size: 360,
        hideVeil: true,
        children: (
          <AsidePanel
            title="Support Center"
            description="Help, support requests, and quick access to platform documentation."
            items={['Create support request', 'Documentation', 'Platform status']}
          />
        ),
      },
      {
        id: 'account',
        open: activeFooterPanel === 'account',
        size: 360,
        hideVeil: true,
        children: (
          <AsidePanel
            title="Account"
            description="Profile and access settings for the current user."
            items={['Profile', 'Security', 'Active sessions']}
          />
        ),
      },
    ],
    [activeFooterPanel],
  )

  const projectOptions = useMemo(
    () =>
      projects
        .filter((project) => project.organization === organization && project.workspace === workspace)
        .map((project) => ({
          value: project.projectId,
          content: project.name,
        })),
    [organization, workspace],
  )

  const visibleProjects = useMemo(() => {
    const searchValue = search.trim().toLowerCase()

    return projects.filter((project) => {
      const matchesSearch =
        searchValue.length === 0 ||
        [project.name, project.projectId, project.owner, project.lastOperation].some((value) =>
          value.toLowerCase().includes(searchValue),
        )

      return (
        matchesSearch &&
        project.organization === organization &&
        project.workspace === workspace &&
        (status === 'all' || project.status === status) &&
        (owner === 'all' || project.owner === owner)
      )
    })
  }, [organization, owner, search, status, workspace])

  return (
    <ThemeProvider theme="light" lang={initialLanguage} fallbackLang="en">
      <AsideHeader
        compact={compact}
        logo={{text: 'M8 Platform', icon: Shield, href: '/'}}
        topAlert={{
          title: 'Demo environment',
          message: 'Project data is mocked for the M8 Platform console prototype.',
          theme: 'info',
          view: 'filled',
          dense: true,
          closable: true,
          preloadHeight: true,
        }}
        panelItems={panelItems}
        subheaderItems={subheaderItems}
        menuItems={navigationItems}
        menuGroups={menuGroups}
        defaultMenuItems={defaultMenuItems}
        menuOverflow="scroll"
        onClosePanel={() => setActiveFooterPanel(null)}
        onChangeCompact={handleNavigationCompactChange}
        onMenuItemsChanged={setNavigationItems}
        renderFooter={({compact: footerCompact}) => (
          <>
            <FooterItem
              id="notifications"
              icon={BellDot}
              title="Notifications"
              tooltipText="Notifications"
              current={activeFooterPanel === 'notifications'}
              onItemClick={() => {
                setActiveFooterPanel(activeFooterPanel === 'notifications' ? null : 'notifications')
              }}
              compact={footerCompact}
            />
            <FooterItem
              id="support"
              icon={CircleQuestion}
              title="Support Center"
              tooltipText="Support Center"
              current={activeFooterPanel === 'support'}
              onItemClick={() => {
                setActiveFooterPanel(activeFooterPanel === 'support' ? null : 'support')
              }}
              compact={footerCompact}
            />
            <FooterItem
              id="account"
              icon={Person}
              title="Account"
              tooltipText="Account"
              current={activeFooterPanel === 'account'}
              itemWrapper={(params, makeItem) =>
                makeItem({
                  ...params,
                  icon: <Avatar className="m8-account-avatar" text="С" size="xs" theme="brand" />,
                })
              }
              onItemClick={() => {
                setActiveFooterPanel(activeFooterPanel === 'account' ? null : 'account')
              }}
              compact={footerCompact}
            />
          </>
        )}
        renderContent={() => (
          <div className="m8-page">
            <ActionBar aria-label="M8 Platform action bar" className="m8-actionbar">
              <ActionBar.Section>
                <ActionBar.Group>
                  <ActionBar.Item>
                    <Switcher
                      label="Org"
                      value={[organization]}
                      options={organizationOptions}
                      onUpdate={(next) => {
                        const nextOrganization = next[0] ?? organization
                        const nextProject =
                          projects.find(
                            (project) =>
                              project.organization === nextOrganization && project.workspace === workspace,
                          ) ?? projects.find((project) => project.organization === nextOrganization)

                        setOrganization(nextOrganization)

                        if (nextProject) {
                          setWorkspace(nextProject.workspace)
                          setProjectId(nextProject.projectId)
                        }
                      }}
                    />
                  </ActionBar.Item>
                  <ActionBar.Item>
                    <Switcher
                      label="Workspace"
                      value={[workspace]}
                      options={workspaceOptions}
                      onUpdate={(next) => {
                        const nextWorkspace = next[0] ?? workspace
                        setWorkspace(nextWorkspace)
                        const nextProject = projects.find(
                          (project) =>
                            project.organization === organization && project.workspace === nextWorkspace,
                        )
                        if (nextProject) {
                          setProjectId(nextProject.projectId)
                        }
                      }}
                    />
                  </ActionBar.Item>
                  <ActionBar.Item>
                    <Switcher
                      label="Project"
                      value={[projectId]}
                      options={projectOptions}
                      onUpdate={(next) => setProjectId(next[0] ?? projectId)}
                    />
                  </ActionBar.Item>
                </ActionBar.Group>
                <ActionBar.Group pull="right">
                  <ActionBar.Item>
                    <Button view="normal">
                      <Icon data={ArrowRotateRight} size={14} />
                      Refresh
                    </Button>
                  </ActionBar.Item>
                  <ActionBar.Item>
                    <Button view="outlined">
                      <Icon data={Clock} size={14} />
                      Open operation
                    </Button>
                  </ActionBar.Item>
                  <ActionBar.Item>
                    <Button view="action">
                      <Icon data={Plus} size={14} />
                      New project
                    </Button>
                  </ActionBar.Item>
                </ActionBar.Group>
              </ActionBar.Section>
            </ActionBar>

            <main className="m8-page__body">
              <section className="m8-page__content">
                <div className="m8-page__heading">
                  <div>
                    <div className="m8-breadcrumbs">
                      <span>M8</span>
                      <span>/</span>
                      <span>Resources</span>
                      <span>/</span>
                      <strong>Projects</strong>
                    </div>
                    <Text as="h1" variant="display-1">
                      Projects
                    </Text>
                    <Text as="p" variant="body-2" color="secondary">
                      Manage project lifecycle, desired state, operations, and auditability across
                      M8 workspaces.
                    </Text>
                  </div>

                  <div className="m8-summary">
                    <Metric label="Projects" value="147" description="3 provisioning" />
                    <Metric label="Failed" value="2" description="requires review" tone="danger" />
                    <Metric label="Deleting" value="4" description="pending finalizers" tone="warning" />
                  </div>
                </div>

                <Card view="outlined" type="container" className="m8-filter-card">
                  <div className="m8-filters">
                    <label className="m8-field">
                      <Text variant="caption-2" color="secondary">
                        Search
                      </Text>
                      <TextInput
                        value={search}
                        placeholder="Project, opaque ID, owner, operation"
                        startContent={<Icon data={Magnifier} size={14} />}
                        onUpdate={setSearch}
                      />
                    </label>
                    <Switcher
                      label="Workspace"
                      value={[workspace]}
                      options={workspaceOptions}
                      onUpdate={(next) => {
                        const nextWorkspace = next[0] ?? workspace
                        setWorkspace(nextWorkspace)
                        const nextProject = projects.find(
                          (project) =>
                            project.organization === organization && project.workspace === nextWorkspace,
                        )
                        if (nextProject) {
                          setProjectId(nextProject.projectId)
                        }
                      }}
                    />
                    <Switcher
                      label="Status"
                      value={[status]}
                      options={statusOptions}
                      onUpdate={(next) => setStatus(next[0] ?? status)}
                    />
                    <Switcher
                      label="Owner"
                      value={[owner]}
                      options={ownerOptions}
                      onUpdate={(next) => setOwner(next[0] ?? owner)}
                    />
                  </div>
                </Card>

                <div className="m8-workspace">
                  <Card view="outlined" type="container" className="m8-table-card">
                    <div className="m8-card-header">
                      <div>
                        <Text as="h2" variant="header-1">
                          Project inventory
                        </Text>
                        <Text variant="caption-2" color="secondary">
                          Filtered list of projects in the selected workspace.
                        </Text>
                      </div>
                      <div className="m8-labels">
                        {statusOptions.slice(1).map((option) => (
                          <StatusLabel key={option.value} status={option.value as ProjectStatus} />
                        ))}
                      </div>
                    </div>

                    <ProjectTable
                      projects={visibleProjects}
                      selectedProjectId={projectId}
                      onSelectProject={setProjectId}
                    />
                  </Card>
                </div>
              </section>
            </main>
          </div>
        )}
      />
    </ThemeProvider>
  )
}

function ProjectTable({
  projects,
  selectedProjectId,
  onSelectProject,
}: {
  projects: Project[]
  selectedProjectId: string
  onSelectProject: (projectId: string) => void
}) {
  if (projects.length === 0) {
    return (
      <div className="m8-empty-table">
        <Text variant="body-2">No projects match the current filters.</Text>
        <Text variant="caption-2" color="secondary">
          Adjust workspace, status, owner, or search criteria.
        </Text>
      </div>
    )
  }

  return (
    <div className="m8-table-shell">
      <table className="m8-project-table">
        <thead>
          <tr>
            <th>Project</th>
            <th>Project ID</th>
            <th>Workspace</th>
            <th>Organization</th>
            <th>Status</th>
            <th>Desired State</th>
            <th>Actual State</th>
            <th>Updated</th>
            <th>Owner</th>
          </tr>
        </thead>
        <tbody>
          {projects.map((project) => (
            <tr
              key={project.projectId}
              className={project.projectId === selectedProjectId ? 'm8-project-table__row_selected' : undefined}
              onClick={() => onSelectProject(project.projectId)}
            >
              <td>
                <div className="m8-project-cell">
                  <span className={`m8-status-dot m8-status-dot_${project.status.toLowerCase()}`} />
                  <div>
                    <Text variant="body-2">{project.name}</Text>
                    <Text variant="caption-2" color="secondary">
                      {project.lastOperation}
                    </Text>
                  </div>
                </div>
              </td>
              <td className="m8-mono">{project.projectId}</td>
              <td className="m8-mono">{project.workspace}</td>
              <td className="m8-mono">{project.organization}</td>
              <td>
                <StatusLabel status={project.status} />
              </td>
              <td>{project.desiredState}</td>
              <td>{project.actualState}</td>
              <td>{project.updated}</td>
              <td className="m8-mono">{project.owner}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function AsidePanel({
  title,
  description,
  items,
}: {
  title: string
  description: string
  items: string[]
}) {
  return (
    <div className="m8-aside-panel">
      <div>
        <Text as="h2" variant="header-1">
          {title}
        </Text>
        <Text variant="body-2" color="secondary">
          {description}
        </Text>
      </div>

      <div className="m8-aside-panel__items">
        {items.map((item) => (
          <Button key={item} view="outlined" width="max">
            {item}
          </Button>
        ))}
      </div>
    </div>
  )
}

interface SwitcherProps {
  label: string
  value: string[]
  options: Array<{value: string; content: string}>
  onUpdate: (value: string[]) => void
}

function Switcher({label, value, options, onUpdate}: SwitcherProps) {
  return (
    <div className="m8-field m8-switcher">
      <Select aria-label={label} value={value} options={options} width="max" onUpdate={onUpdate} />
    </div>
  )
}

function Metric({
  label,
  value,
  description,
  tone = 'normal',
}: {
  label: string
  value: string
  description: string
  tone?: 'normal' | 'warning' | 'danger'
}) {
  return (
    <Card view="outlined" type="container" className={`m8-metric m8-metric_${tone}`}>
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Text variant="header-2">{value}</Text>
      <Text variant="caption-2" color="secondary">
        {description}
      </Text>
    </Card>
  )
}

function StatusLabel({status}: {status: ProjectStatus}) {
  const themeByStatus: Record<ProjectStatus, 'success' | 'warning' | 'danger' | 'info' | 'normal'> = {
    Active: 'success',
    Suspended: 'warning',
    Failed: 'danger',
    Provisioning: 'info',
    Deleting: 'warning',
  }

  return <Label theme={themeByStatus[status]}>{status}</Label>
}

export default App
