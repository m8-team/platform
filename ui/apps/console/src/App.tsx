import {useCallback, useEffect, useMemo, useState} from 'react'
import {
  Alert,
  Button,
  Card,
  configure,
  DefinitionList,
  Icon,
  Label,
  Progress,
  Select,
  Text,
  TextInput,
  ThemeProvider,
} from '@gravity-ui/uikit'
import {ActionBar, AsideHeader} from '@gravity-ui/navigation'
import type {AsideHeaderItem, MenuGroup} from '@gravity-ui/navigation'
import {
  Check,
  Clock,
  Cloud,
  Database,
  Gear,
  ListUl,
  Magnifier,
  Person,
  Plus,
  Rocket,
  Shield,
  TriangleExclamation,
  ArrowRotateRight,
} from '@gravity-ui/icons'

import './App.css'

type ProjectStatus = 'Active' | 'Suspended' | 'Failed' | 'Provisioning' | 'Deleting'

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
  version: string
  etag: string
  lastOperation: string
  conditions: Array<{
    label: string
    value: number
    text: string
    tone: 'success' | 'default' | 'warning' | 'danger'
  }>
  auditSummary: string
}

const initialLanguage = 'en'
const navigationCompactStorageKey = 'm8.console.navigation.compact'

const workspaceOptions = [
  {value: 'ws_prod-eu1', content: 'ws_prod-eu1'},
  {value: 'ws_shared-eu1', content: 'ws_shared-eu1'},
  {value: 'ws_legacy-eu1', content: 'ws_legacy-eu1'},
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
  {value: 'usr_7ac391e2_ops', content: 'usr_7ac391e2_ops'},
  {value: 'usr_19bd4027_sre', content: 'usr_19bd4027_sre'},
  {value: 'usr_2f0c81aa_sec', content: 'usr_2f0c81aa_sec'},
]

const projects: Project[] = [
  {
    name: 'payments-ledger',
    projectId: 'prj_8f3a91c2e7b04d6a',
    workspace: 'ws_prod-eu1',
    organization: 'org_m8_finance_6b21d0',
    status: 'Active',
    desiredState: 'Running',
    actualState: 'Running',
    updated: '2026-06-23 09:42',
    owner: 'usr_7ac391e2_ops',
    version: '42',
    etag: 'etag_prj_8f3a91c2e7b04d6a_v42',
    lastOperation: 'op_0c91b6f33e2a4d1b',
    conditions: [
      {label: 'Ready', value: 100, text: 'True', tone: 'success'},
      {label: 'Policy bound', value: 100, text: 'True', tone: 'success'},
      {label: 'Quota healthy', value: 100, text: 'True', tone: 'success'},
      {label: 'Drift detected', value: 0, text: 'False', tone: 'default'},
    ],
    auditSummary:
      '18 events in 24h. Last actor usr_7ac391e2_ops updated runtime policy through op_0c91b6f33e2a4d1b.',
  },
  {
    name: 'risk-scoring',
    projectId: 'prj_2e41d7a9c0bf4e55',
    workspace: 'ws_prod-eu1',
    organization: 'org_m8_finance_6b21d0',
    status: 'Provisioning',
    desiredState: 'Running',
    actualState: 'Provisioning',
    updated: '2026-06-23 09:31',
    owner: 'usr_19bd4027_sre',
    version: '17',
    etag: 'etag_prj_2e41d7a9c0bf4e55_v17',
    lastOperation: 'op_9fe2304db1a44e88',
    conditions: [
      {label: 'Ready', value: 48, text: 'False', tone: 'warning'},
      {label: 'Policy bound', value: 100, text: 'True', tone: 'success'},
      {label: 'Quota healthy', value: 100, text: 'True', tone: 'success'},
      {label: 'Drift detected', value: 0, text: 'False', tone: 'default'},
    ],
    auditSummary:
      '11 events in 24h. Provisioning operation op_9fe2304db1a44e88 is still applying runtime descriptors.',
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
    version: '29',
    etag: 'etag_prj_6d90aa31f48c4b8e_v29',
    lastOperation: 'op_63ab7e02d4104ba1',
    conditions: [
      {label: 'Ready', value: 0, text: 'False', tone: 'warning'},
      {label: 'Policy bound', value: 100, text: 'True', tone: 'success'},
      {label: 'Quota healthy', value: 100, text: 'True', tone: 'success'},
      {label: 'Drift detected', value: 0, text: 'False', tone: 'default'},
    ],
    auditSummary: '7 events in 24h. Suspension was requested by usr_2f0c81aa_sec.',
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
    version: '8',
    etag: 'etag_prj_41c2de83b7764a09_v8',
    lastOperation: 'op_3a7c8a2148ff47a2',
    conditions: [
      {label: 'Ready', value: 0, text: 'False', tone: 'danger'},
      {label: 'Policy bound', value: 100, text: 'True', tone: 'success'},
      {label: 'Quota healthy', value: 32, text: 'Degraded', tone: 'warning'},
      {label: 'Drift detected', value: 100, text: 'True', tone: 'danger'},
    ],
    auditSummary: '23 events in 24h. Last reconciliation failed after provider timeout.',
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
    version: '61',
    etag: 'etag_prj_9aa4c11d0e744f3a_v61',
    lastOperation: 'op_bdc4221d8b714c9d',
    conditions: [
      {label: 'Ready', value: 0, text: 'False', tone: 'warning'},
      {label: 'Policy bound', value: 100, text: 'True', tone: 'success'},
      {label: 'Quota healthy', value: 100, text: 'True', tone: 'success'},
      {label: 'Finalizers cleared', value: 66, text: 'Partial', tone: 'warning'},
    ],
    auditSummary: '14 events in 24h. Delete operation is waiting for finalizers.',
  },
]

const menuGroups: MenuGroup[] = [
  {id: 'core', title: 'Core modules', icon: Database},
  {id: 'operations', title: 'Platform operations', icon: Cloud},
  {id: 'governance', title: 'Governance', icon: Shield},
]

const menuItems: AsideHeaderItem[] = [
  {id: 'resource-manager', title: 'Resource Manager', icon: Database, current: true, groupId: 'core'},
  {id: 'identity', title: 'Identity', icon: Person, groupId: 'core'},
  {id: 'authentication', title: 'Authentication', icon: Shield, groupId: 'core'},
  {id: 'access', title: 'Access', icon: Gear, groupId: 'core'},
  {id: 'provisioning', title: 'Provisioning', icon: Plus, groupId: 'operations'},
  {id: 'runtime', title: 'Runtime', icon: Cloud, groupId: 'operations'},
  {id: 'delivery', title: 'Delivery', icon: Rocket, groupId: 'operations'},
  {id: 'operations', title: 'Operations', icon: Clock, groupId: 'operations'},
  {id: 'audit', title: 'Audit', icon: ListUl, groupId: 'governance'},
  {id: 'compliance', title: 'Compliance', icon: Check, groupId: 'governance'},
  {id: 'settings', title: 'Settings', icon: Gear, groupId: 'governance'},
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
  const [workspace, setWorkspace] = useState('ws_prod-eu1')
  const [projectId, setProjectId] = useState('prj_8f3a91c2e7b04d6a')
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

  const selectedProject = useMemo(
    () => projects.find((project) => project.projectId === projectId) ?? projects[0],
    [projectId],
  )
  const projectOptions = useMemo(
    () =>
      projects
        .filter((project) => project.workspace === workspace || workspace === 'ws_prod-eu1')
        .map((project) => ({
          value: project.projectId,
          content: `${project.name} / ${project.projectId}`,
        })),
    [workspace],
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
        project.workspace === workspace &&
        (status === 'all' || project.status === status) &&
        (owner === 'all' || project.owner === owner)
      )
    })
  }, [owner, search, status, workspace])

  return (
    <ThemeProvider theme="light" lang={initialLanguage} fallbackLang="en">
      <AsideHeader
        compact={compact}
        logo={{text: 'M8 Platform', icon: Shield, href: '/'}}
        menuItems={menuItems}
        menuGroups={menuGroups}
        menuOverflow="scroll"
        onChangeCompact={handleNavigationCompactChange}
        renderContent={() => (
          <div className="m8-page">
            <ActionBar aria-label="M8 Platform action bar" className="m8-actionbar">
              <ActionBar.Section>
                <ActionBar.Group>
                  <ActionBar.Item>
                    <Switcher
                      label="Workspace"
                      value={[workspace]}
                      options={workspaceOptions}
                      onUpdate={(next) => {
                        const nextWorkspace = next[0] ?? workspace
                        setWorkspace(nextWorkspace)
                        const nextProject = projects.find((project) => project.workspace === nextWorkspace)
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
                  <ActionBar.Item>
                    <Label theme="normal">Region: eu-west-1</Label>
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
                      <span>Resource Manager</span>
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
                      onUpdate={(next) => setWorkspace(next[0] ?? workspace)}
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
                          Selected row: {selectedProject.name}
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
                      selectedProjectId={selectedProject.projectId}
                      onSelectProject={setProjectId}
                    />
                  </Card>

                  <ProjectDetails project={selectedProject} />
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

interface SwitcherProps {
  label: string
  value: string[]
  options: Array<{value: string; content: string}>
  onUpdate: (value: string[]) => void
}

function Switcher({label, value, options, onUpdate}: SwitcherProps) {
  return (
    <label className="m8-field m8-switcher">
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Select value={value} options={options} width="max" onUpdate={onUpdate} />
    </label>
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

function ProjectDetails({project}: {project: Project}) {
  return (
    <aside className="m8-details">
      <Card view="outlined" type="container" className="m8-details__card">
        <div className="m8-details__header">
          <div>
            <Text as="h2" variant="header-1">
              {project.name}
            </Text>
            <Text variant="caption-2" color="secondary" className="m8-mono">
              {project.projectId}
            </Text>
          </div>
          <StatusLabel status={project.status} />
        </div>

        <div className="m8-tabs" aria-label="Project details tabs">
          <button className="m8-tabs__item m8-tabs__item_active">Overview</button>
          <button className="m8-tabs__item">Conditions</button>
          <button className="m8-tabs__item">Operations</button>
          <button className="m8-tabs__item">Audit</button>
        </div>

        <DefinitionList>
          <DefinitionList.Item name="Lifecycle">{project.status}</DefinitionList.Item>
          <DefinitionList.Item name="Desired state">{project.desiredState}</DefinitionList.Item>
          <DefinitionList.Item name="Actual state">{project.actualState}</DefinitionList.Item>
          <DefinitionList.Item name="Workspace">
            <span className="m8-mono">{project.workspace}</span>
          </DefinitionList.Item>
          <DefinitionList.Item name="Organization">
            <span className="m8-mono">{project.organization}</span>
          </DefinitionList.Item>
          <DefinitionList.Item name="Owner">
            <span className="m8-mono">{project.owner}</span>
          </DefinitionList.Item>
          <DefinitionList.Item name="Version">{project.version}</DefinitionList.Item>
          <DefinitionList.Item name="ETag">
            <span className="m8-mono">{project.etag}</span>
          </DefinitionList.Item>
          <DefinitionList.Item name="Last operation">
            <span className="m8-mono">{project.lastOperation}</span>
          </DefinitionList.Item>
        </DefinitionList>

        <div className="m8-conditions">
          {project.conditions.map((condition) => (
            <div key={condition.label} className="m8-condition">
              <div className="m8-condition__header">
                <Text variant="body-2">{condition.label}</Text>
                <Text variant="caption-2" color="secondary">
                  {condition.text}
                </Text>
              </div>
              <Progress value={condition.value} theme={condition.tone} />
            </div>
          ))}
        </div>

        <Alert
          theme="info"
          view="outlined"
          title="Audit summary"
          message={project.auditSummary}
          icon={<Icon data={ListUl} />}
        />

        <div className="m8-details__actions">
          <Button view="outlined-warning">
            <Icon data={TriangleExclamation} size={14} />
            Suspend
          </Button>
          <Button view="outlined-success" disabled={project.status !== 'Suspended'}>
            <Icon data={Check} size={14} />
            Resume
          </Button>
          <Button view="outlined-danger">
            <Icon data={TriangleExclamation} size={14} />
            Delete
          </Button>
          <Button view="outlined">
            <Icon data={ListUl} size={14} />
            View audit
          </Button>
          <Button view="normal">
            <Icon data={Clock} size={14} />
            Open operation
          </Button>
        </div>
      </Card>
    </aside>
  )
}

export default App
