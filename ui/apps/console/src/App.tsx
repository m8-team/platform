import {createContext, useCallback, useContext, useEffect, useMemo, useState} from 'react'
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
  ToasterComponent,
  ToasterProvider,
} from '@gravity-ui/uikit'
import {toaster} from '@gravity-ui/uikit/toaster-singleton'
import {AsideHeader, FooterItem} from '@gravity-ui/navigation'
import type {AsideHeaderItem, MenuGroup, PanelItemProps} from '@gravity-ui/navigation'
import {Outlet, useRouter, useRouterState} from '@tanstack/react-router'
import {
  ArrowShapeRightFromLine,
  BellDot,
  BarsPlay,
  BranchesDown,
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
  GearPlay,
  Layers,
  ListUl,
  Magnifier,
  NodesRight,
  OctagonXmark,
  Person,
  Persons,
  Rocket,
  Shield,
  ShieldCheck,
  Signal,
  Speedometer,
  TriangleExclamation,
  ArrowRotateRight,
  EnvelopeOpenXmark,
} from '@gravity-ui/icons'

import {ConsoleActionBar} from './components/ConsoleActionBar'
import {ConsoleBreadcrumbs} from './components/ConsoleBreadcrumbs'
import {Metric} from './components/Metric'
import {
  createTranslator,
  fallbackLanguage,
  isAppLanguage,
  languageOptions as languageOptionConfigs,
} from './i18n'
import type {AppLanguage, Translate, TranslationKey} from './i18n'
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

interface ConsoleSelection {
  organization: string
  workspace: string
  projectId: string
  projectOptions: Array<{value: string; content: string}>
  setWorkspace: (value: string) => void
  setProjectId: (value: string) => void
}

interface ConsoleI18n {
  language: AppLanguage
  t: Translate
}

const ConsoleSelectionContext = createContext<ConsoleSelection | null>(null)
const ConsoleI18nContext = createContext<ConsoleI18n | null>(null)

const languageStorageKey = 'm8.console.language'
const navigationCompactStorageKey = 'm8.console.navigation.compact'
const menuGroupCollapsedStorageKey = 'm8.console.menu-groups.collapsed'

const resourceManagerRoutes = {
  overview: '/resource-manager',
  organizations: {
    list: '/resource-manager/organizations',
    detail: '/resource-manager/organizations/:organizationId',
  },
  workspaces: {
    list: '/resource-manager/workspaces',
    detail: '/resource-manager/workspaces/:workspaceId',
  },
  projects: {
    list: '/resource-manager/projects',
    detail: '/resource-manager/projects/:projectId',
  },
} as const

const organizationOptions = [
  {value: 'org_m8_finance_6b21d0', content: 'Acme'},
  {value: 'org_m8_billing_91f2c5', content: 'Billing'},
]

const workspaceOptionConfigs = [
  {value: 'ws_prod-eu1', titleKey: 'workspace.platform'},
  {value: 'ws_shared-eu1', titleKey: 'workspace.sharedServices'},
  {value: 'ws_legacy-eu1', titleKey: 'workspace.legacy'},
] satisfies Array<{value: string; titleKey: TranslationKey}>

const statusOptionConfigs = [
  {value: 'all', titleKey: 'status.all'},
  {value: 'Active', titleKey: 'status.Active'},
  {value: 'Suspended', titleKey: 'status.Suspended'},
  {value: 'Failed', titleKey: 'status.Failed'},
  {value: 'Provisioning', titleKey: 'status.Provisioning'},
  {value: 'Deleting', titleKey: 'status.Deleting'},
] satisfies Array<{value: ProjectStatus | 'all'; titleKey: TranslationKey}>

const ownerOptionConfigs = [
  {value: 'all', titleKey: 'owner.all'},
  {value: 'usr_19bd4027_sre', content: 'usr_19bd4027_sre'},
  {value: 'usr_2f0c81aa_sec', content: 'usr_2f0c81aa_sec'},
] satisfies Array<{value: string; content?: string; titleKey?: TranslationKey}>

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

const resourceOverviewMix = [
  {labelKey: 'overview.resource.database', value: 42, tone: 'brand'},
  {labelKey: 'overview.resource.kafka', value: 31, tone: 'info'},
  {labelKey: 'overview.resource.cache', value: 18, tone: 'positive'},
  {labelKey: 'overview.resource.storage', value: 9, tone: 'warning'},
] satisfies Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>

const resourceOverviewLifecycle = [
  {labelKey: 'status.Active', value: 78, tone: 'positive'},
  {labelKey: 'status.Provisioning', value: 12, tone: 'info'},
  {labelKey: 'status.Suspended', value: 6, tone: 'warning'},
  {labelKey: 'status.Failed', value: 4, tone: 'danger'},
] satisfies Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>

const resourceOverviewOperations = [
  {
    titleKey: 'overview.operation.projectCreate',
    descriptionKey: 'overview.operation.projectCreateDescription',
    theme: 'info',
  },
  {
    titleKey: 'overview.operation.workspaceSuspend',
    descriptionKey: 'overview.operation.workspaceSuspendDescription',
    theme: 'warning',
  },
  {
    titleKey: 'overview.operation.resourceReconcile',
    descriptionKey: 'overview.operation.resourceReconcileDescription',
    theme: 'success',
  },
] satisfies Array<{
  titleKey: TranslationKey
  descriptionKey: TranslationKey
  theme: 'success' | 'warning' | 'danger' | 'info' | 'normal'
}>

const resourceOverviewSignals = [
  {
    titleKey: 'overview.signal.sourceOfTruth',
    descriptionKey: 'overview.signal.sourceOfTruthDescription',
  },
  {
    titleKey: 'overview.signal.safeMutations',
    descriptionKey: 'overview.signal.safeMutationsDescription',
  },
  {
    titleKey: 'overview.signal.auditReady',
    descriptionKey: 'overview.signal.auditReadyDescription',
  },
] satisfies Array<{titleKey: TranslationKey; descriptionKey: TranslationKey}>

type OverviewTone = 'brand' | 'info' | 'positive' | 'warning' | 'danger'

type MenuGroupConfig = Omit<MenuGroup, 'title'> & {titleKey: TranslationKey}
type MenuItemConfig = Omit<AsideHeaderItem, 'title'> & {titleKey: TranslationKey}

const menuGroupConfigs: MenuGroupConfig[] = [
  {id: 'resources', titleKey: 'menu.resources', icon: BranchesDown},
  {id: 'platform-operations', titleKey: 'menu.platformOperations', icon: GearPlay},
  {id: 'identity-access', titleKey: 'menu.identityAccess', icon: Shield},
  {id: 'gateway', titleKey: 'menu.gateway', icon: Cloud},
  {id: 'security', titleKey: 'menu.security', icon: Shield},
  {id: 'observability', titleKey: 'menu.observability', icon: Clock},
  {id: 'audit', titleKey: 'menu.audit', icon: ListUl},
  {id: 'settings', titleKey: 'menu.settings', icon: Gear},
]

const menuItemConfigs: MenuItemConfig[] = [
  {
    id: 'resources-overview',
    titleKey: 'menu.resources.overview',
    icon: Rocket,
    href: resourceManagerRoutes.overview,
    groupId: 'resources',
  },
  {
    id: 'resources-organizations',
    titleKey: 'menu.resources.organizations',
    icon: Briefcase,
    href: resourceManagerRoutes.organizations.list,
    groupId: 'resources',
  },
  {
    id: 'resources-workspaces',
    titleKey: 'menu.resources.workspaces',
    icon: Folders,
    href: resourceManagerRoutes.workspaces.list,
    groupId: 'resources',
  },
  {
    id: 'resources-project',
    titleKey: 'menu.resources.projects',
    icon: Database,
    href: resourceManagerRoutes.projects.list,
    groupId: 'resources',
  },
  {
    id: 'platform-operations-long-running',
    titleKey: 'menu.operations.longRunning',
    icon: Clock,
    groupId: 'platform-operations',
  },
  {
    id: 'platform-operations-quotas-limits',
    titleKey: 'menu.operations.quotasLimits',
    icon: Speedometer,
    groupId: 'platform-operations',
  },
  {id: 'platform-operations-jobs', titleKey: 'menu.operations.jobs', icon: BarsPlay, groupId: 'platform-operations'},
  {id: 'platform-operations-queues', titleKey: 'menu.operations.queues', icon: Layers, groupId: 'platform-operations'},
  {id: 'platform-operations-outbox', titleKey: 'menu.operations.outbox', icon: EnvelopeOpenXmark, groupId: 'platform-operations'},
  {
    id: 'platform-operations-failed-events',
    titleKey: 'menu.operations.failedEvents',
    icon: TriangleExclamation,
    groupId: 'platform-operations',
  },
  {id: 'platform-operations-retries', titleKey: 'menu.operations.retries', icon: ArrowRotateRight, groupId: 'platform-operations'},
  {
    id: 'platform-operations-dead-letter-queue',
    titleKey: 'menu.operations.deadLetterQueue',
    icon: OctagonXmark,
    groupId: 'platform-operations',
  },
  {id: 'identity-access-identity', titleKey: 'menu.identity.identity', icon: Person, groupId: 'identity-access'},
  {id: 'identity-access-authentication', titleKey: 'menu.identity.authentication', icon: Shield, groupId: 'identity-access'},
  {id: 'identity-access-control', titleKey: 'menu.identity.accessControl', icon: ShieldCheck, groupId: 'identity-access'},
  {id: 'gateway-api-services', titleKey: 'menu.gateway.apiServices', icon: Cloud, groupId: 'gateway'},
  {id: 'gateway-routes', titleKey: 'menu.gateway.routes', icon: ArrowShapeRightFromLine, groupId: 'gateway'},
  {id: 'gateway-consumers', titleKey: 'menu.gateway.consumers', icon: Persons, groupId: 'gateway'},
  {id: 'gateway-policies', titleKey: 'menu.gateway.policies', icon: Check, groupId: 'gateway'},
  {id: 'gateway-rate-limits', titleKey: 'menu.gateway.rateLimits', icon: Speedometer, groupId: 'gateway'},
  {id: 'gateway-yaml', titleKey: 'menu.gateway.yaml', icon: Code, groupId: 'gateway'},
  {id: 'security-dashboard', titleKey: 'menu.security.dashboard', icon: Rocket, groupId: 'security'},
  {id: 'security-risk-rules', titleKey: 'menu.security.riskRules', icon: Shield, groupId: 'security'},
  {id: 'security-device-fingerprints', titleKey: 'menu.security.deviceFingerprints', icon: Fingerprint, groupId: 'security'},
  {id: 'security-velocity-rules', titleKey: 'menu.security.velocityRules', icon: Speedometer, groupId: 'security'},
  {id: 'security-signals', titleKey: 'menu.security.signals', icon: Signal, groupId: 'security'},
  {id: 'security-decisions', titleKey: 'menu.security.decisions', icon: Check, groupId: 'security'},
  {id: 'security-challenges', titleKey: 'menu.security.challenges', icon: TriangleExclamation, groupId: 'security'},
  {id: 'security-fraud-cases', titleKey: 'menu.security.fraudCases', icon: Briefcase, groupId: 'security'},
  {id: 'security-events', titleKey: 'menu.security.securityEvents', icon: ListUl, groupId: 'security'},
  {id: 'security-access-reviews', titleKey: 'menu.security.accessReviews', icon: Persons, groupId: 'security'},
  {id: 'security-policy-violations', titleKey: 'menu.security.policyViolations', icon: TriangleExclamation, groupId: 'security'},
  {id: 'observability-metrics', titleKey: 'menu.observability.metrics', icon: Speedometer, groupId: 'observability'},
  {id: 'observability-logs', titleKey: 'menu.observability.logs', icon: ListUl, groupId: 'observability'},
  {id: 'observability-traces', titleKey: 'menu.observability.traces', icon: NodesRight, groupId: 'observability'},
  {id: 'observability-alerts', titleKey: 'menu.observability.alerts', icon: TriangleExclamation, groupId: 'observability'},
  {id: 'observability-slo', titleKey: 'menu.observability.slo', icon: Check, groupId: 'observability'},
  {id: 'audit-events', titleKey: 'menu.audit.events', icon: ListUl, groupId: 'audit'},
  {id: 'audit-exports', titleKey: 'menu.audit.exports', icon: ArrowRotateRight, groupId: 'audit'},
  {id: 'settings-project', titleKey: 'menu.settings.project', icon: Gear, groupId: 'settings'},
  {id: 'settings-modules', titleKey: 'menu.settings.modules', icon: Database, groupId: 'settings'},
  {id: 'settings-integrations', titleKey: 'menu.settings.integrations', icon: Cloud, groupId: 'settings'},
  {id: 'settings-webhooks', titleKey: 'menu.settings.webhooks', icon: ArrowShapeRightFromLine, groupId: 'settings'},
  {id: 'settings-api-tokens', titleKey: 'menu.settings.apiTokens', icon: Shield, groupId: 'settings'},
]

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

function readInitialLanguage() {
  if (typeof window === 'undefined') {
    return fallbackLanguage
  }

  try {
    const storedLanguage = window.localStorage.getItem(languageStorageKey)
    return isAppLanguage(storedLanguage) ? storedLanguage : fallbackLanguage
  } catch {
    return fallbackLanguage
  }
}

function readCurrentPathname() {
  if (typeof window === 'undefined') {
    return resourceManagerRoutes.projects.list
  }

  return window.location.pathname
}

function getCurrentMenuItemId(pathname: string) {
  if (pathname === resourceManagerRoutes.overview) {
    return 'resources-overview'
  }

  if (pathname.startsWith(`${resourceManagerRoutes.organizations.list}/`)) {
    return 'resources-organizations'
  }

  if (pathname === resourceManagerRoutes.organizations.list) {
    return 'resources-organizations'
  }

  if (pathname.startsWith(`${resourceManagerRoutes.workspaces.list}/`)) {
    return 'resources-workspaces'
  }

  if (pathname === resourceManagerRoutes.workspaces.list) {
    return 'resources-workspaces'
  }

  if (pathname.startsWith(`${resourceManagerRoutes.projects.list}/`)) {
    return 'resources-project'
  }

  if (pathname === resourceManagerRoutes.projects.list) {
    return 'resources-project'
  }

  return 'resources-project'
}

function getCurrentMenuGroupId(pathname = readCurrentPathname()) {
  const currentMenuItemId = getCurrentMenuItemId(pathname)
  return menuItemConfigs.find((item) => item.id === currentMenuItemId)?.groupId
}

function createDefaultCollapsedMenuGroups(pathname = readCurrentPathname()) {
  const currentGroupId = getCurrentMenuGroupId(pathname)

  return menuGroupConfigs.reduce<Record<string, boolean>>((collapsedGroups, group) => {
    collapsedGroups[group.id] = group.id !== currentGroupId
    return collapsedGroups
  }, {})
}

function normalizeCollapsedMenuGroups(storedGroups?: Record<string, unknown>, pathname = readCurrentPathname()) {
  const currentGroupId = getCurrentMenuGroupId(pathname)
  const collapsedGroups = createDefaultCollapsedMenuGroups(pathname)

  if (storedGroups) {
    for (const group of menuGroupConfigs) {
      const storedValue = storedGroups[group.id]
      if (typeof storedValue === 'boolean') {
        collapsedGroups[group.id] = storedValue
      }
    }
  }

  if (currentGroupId) {
    collapsedGroups[currentGroupId] = false
  }

  return collapsedGroups
}

function readInitialCollapsedMenuGroups() {
  if (typeof window === 'undefined') {
    return createDefaultCollapsedMenuGroups()
  }

  try {
    const storedValue = window.localStorage.getItem(menuGroupCollapsedStorageKey)
    if (!storedValue) {
      return createDefaultCollapsedMenuGroups()
    }

    const parsedValue: unknown = JSON.parse(storedValue)
    if (!parsedValue || typeof parsedValue !== 'object' || Array.isArray(parsedValue)) {
      return createDefaultCollapsedMenuGroups()
    }

    return normalizeCollapsedMenuGroups(parsedValue as Record<string, unknown>)
  } catch {
    return createDefaultCollapsedMenuGroups()
  }
}

function translateOptions<T extends string>(
  options: Array<{value: T; content?: string; titleKey?: TranslationKey}>,
  t: Translate,
) {
  return options.map((option) => ({
    value: option.value,
    content: option.titleKey ? t(option.titleKey) : option.content ?? option.value,
  }))
}

function App() {
  const [language, setLanguage] = useState<AppLanguage>(readInitialLanguage)
  const [compact, setCompact] = useState(readInitialNavigationCompact)
  const [collapsedMenuGroupIds, setCollapsedMenuGroupIds] = useState(readInitialCollapsedMenuGroups)
  const [activeFooterPanel, setActiveFooterPanel] = useState<FooterPanel | null>(null)
  const [organization, setOrganization] = useState('org_m8_finance_6b21d0')
  const [workspace, setWorkspace] = useState('ws_prod-eu1')
  const [projectId, setProjectId] = useState('prj_2e41d7a9c0bf4e55')
  const router = useRouter()
  const pathname = useRouterState({select: (state) => state.location.pathname})
  const t = useMemo(() => createTranslator(language), [language])

  const handleLanguageUpdate = useCallback(
    (next: string[]) => {
      const nextLanguage = next[0]
      setLanguage(isAppLanguage(nextLanguage) ? nextLanguage : language)
    },
    [language],
  )

  const handleNavigationCompactChange = useCallback((nextCompact: boolean) => {
    setCompact(nextCompact)

    try {
      window.localStorage.setItem(navigationCompactStorageKey, String(nextCompact))
    } catch {
      // Storage can be unavailable in private or restricted browser contexts.
    }
  }, [])

  const handleToggleMenuGroupCollapsed = useCallback((groupId: string) => {
    setCollapsedMenuGroupIds((currentCollapsedGroups) => {
      const nextCollapsedGroups = normalizeCollapsedMenuGroups(
        {
          ...currentCollapsedGroups,
          [groupId]: !currentCollapsedGroups[groupId],
        },
        pathname,
      )

      try {
        window.localStorage.setItem(menuGroupCollapsedStorageKey, JSON.stringify(nextCollapsedGroups))
      } catch {
        // Storage can be unavailable in private or restricted browser contexts.
      }

      return nextCollapsedGroups
    })
  }, [pathname])

  useEffect(() => {
    configure({
      lang: language,
      fallbackLang: fallbackLanguage,
    })
    document.documentElement.lang = language

    try {
      window.localStorage.setItem(languageStorageKey, language)
    } catch {
      // Storage can be unavailable in private or restricted browser contexts.
    }
  }, [language])

  const currentMenuItemId = getCurrentMenuItemId(pathname)
  const effectiveCollapsedMenuGroupIds = useMemo(
    () => normalizeCollapsedMenuGroups(collapsedMenuGroupIds, pathname),
    [collapsedMenuGroupIds, pathname],
  )
  const menuGroups = useMemo<MenuGroup[]>(
    () => menuGroupConfigs.map(({titleKey, ...group}) => ({...group, title: t(titleKey)})),
    [t],
  )
  const navigationMenuItems = useMemo(
    () =>
      menuItemConfigs.map(({titleKey, ...item}) => {
        const title = t(titleKey)

        if (!item.href) {
          return {
            ...item,
            title,
            current: item.id === currentMenuItemId,
          }
        }

        const href = item.href
        const onItemClick: NonNullable<AsideHeaderItem['onItemClick']> = (_item, _collapsed, event) => {
          event.preventDefault()
          void router.navigate({to: href})
        }

        return {
          ...item,
          title,
          current: item.id === currentMenuItemId,
          onItemClick,
        }
      }),
    [currentMenuItemId, router, t],
  )

  const subheaderItems = useMemo<AsideHeaderItem[]>(
    () => [
      {
        id: 'subheader-dashboard',
        title: t('menu.security.dashboard'),
        icon: Rocket,
      },
    ],
    [t],
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
            title={t('footer.notifications')}
            description={t('panel.notifications.description')}
            items={[
              t('panel.notifications.item.quota'),
              t('panel.notifications.item.gateway'),
              t('panel.notifications.item.audit'),
            ]}
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
            title={t('footer.support')}
            description={t('panel.support.description')}
            items={[t('panel.support.item.create'), t('panel.support.item.docs'), t('panel.support.item.status')]}
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
            title={t('footer.account')}
            description={t('panel.account.description')}
            items={[t('panel.account.item.profile'), t('panel.account.item.security'), t('panel.account.item.sessions')]}
          />
        ),
      },
    ],
    [activeFooterPanel, t],
  )

  const workspaceOptions = useMemo(() => translateOptions(workspaceOptionConfigs, t), [t])
  const languageOptions = useMemo(
    () =>
      languageOptionConfigs.map((option) => ({
        value: option.value,
        content: t(option.labelKey),
      })),
    [t],
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

  const handleOrganizationUpdate = useCallback(
    (next: string[]) => {
      const nextOrganization = next[0] ?? organization
      const nextProject =
        projects.find(
          (project) => project.organization === nextOrganization && project.workspace === workspace,
        ) ?? projects.find((project) => project.organization === nextOrganization)

      setOrganization(nextOrganization)

      if (nextProject) {
        setWorkspace(nextProject.workspace)
        setProjectId(nextProject.projectId)
      }
    },
    [organization, workspace],
  )

  const handleWorkspaceUpdate = useCallback(
    (next: string[]) => {
      const nextWorkspace = next[0] ?? workspace
      setWorkspace(nextWorkspace)

      const nextProject = projects.find(
        (project) => project.organization === organization && project.workspace === nextWorkspace,
      )
      if (nextProject) {
        setProjectId(nextProject.projectId)
      }
    },
    [organization, workspace],
  )

  const handleProjectUpdate = useCallback(
    (next: string[]) => {
      setProjectId(next[0] ?? projectId)
    },
    [projectId],
  )

  const selection = useMemo<ConsoleSelection>(
    () => ({
      organization,
      workspace,
      projectId,
      projectOptions,
      setWorkspace,
      setProjectId,
    }),
    [organization, projectId, projectOptions, workspace],
  )
  const i18nValue = useMemo<ConsoleI18n>(() => ({language, t}), [language, t])

  if (pathname.startsWith('/commerce-intelligence')) {
    return (
      <ThemeProvider theme="light" lang="ru" fallbackLang={fallbackLanguage}>
        <ToasterProvider toaster={toaster}>
          <ToasterComponent />
          <Outlet />
        </ToasterProvider>
      </ThemeProvider>
    )
  }

  return (
    <ThemeProvider theme="light" lang={language} fallbackLang={fallbackLanguage}>
      <ToasterProvider toaster={toaster}>
        <ToasterComponent />
        <AsideHeader
          compact={compact}
          logo={{text: 'M8 Platform', icon: Shield, href: '/'}}
          topAlert={{
            title: t('topAlert.title'),
            message: t('topAlert.message'),
            theme: 'info',
            view: 'filled',
            dense: true,
            closable: true,
            preloadHeight: true,
          }}
          panelItems={panelItems}
          subheaderItems={subheaderItems}
          menuItems={navigationMenuItems}
          menuGroups={menuGroups}
          menuOverflow="scroll"
          collapsedMenuGroupIds={effectiveCollapsedMenuGroupIds}
          onClosePanel={() => setActiveFooterPanel(null)}
          onChangeCompact={handleNavigationCompactChange}
          onToggleMenuGroupCollapsed={handleToggleMenuGroupCollapsed}
          renderFooter={({compact: footerCompact}) => (
            <>
              <FooterItem
                id="notifications"
                icon={BellDot}
                title={t('footer.notifications')}
                tooltipText={t('footer.notifications')}
                current={activeFooterPanel === 'notifications'}
                onItemClick={() => {
                  setActiveFooterPanel(activeFooterPanel === 'notifications' ? null : 'notifications')
                }}
                compact={footerCompact}
              />
              <FooterItem
                id="support"
                icon={CircleQuestion}
                title={t('footer.support')}
                tooltipText={t('footer.support')}
                current={activeFooterPanel === 'support'}
                onItemClick={() => {
                  setActiveFooterPanel(activeFooterPanel === 'support' ? null : 'support')
                }}
                compact={footerCompact}
              />
              <FooterItem
                id="account"
                icon={Person}
                title={t('footer.account')}
                tooltipText={t('footer.account')}
                current={activeFooterPanel === 'account'}
                itemWrapper={(params, makeItem) =>
                  makeItem({
                    ...params,
                    icon: <Avatar text="СC" size="xs" theme="brand" />,
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
            <ConsoleI18nContext.Provider value={i18nValue}>
              <ConsoleSelectionContext.Provider value={selection}>
                <div className="m8-page">
                  <ConsoleActionBar
                    language={language}
                    organization={organization}
                    workspace={workspace}
                    projectId={projectId}
                    languageOptions={languageOptions}
                    organizationOptions={organizationOptions}
                    workspaceOptions={workspaceOptions}
                    projectOptions={projectOptions}
                    labels={{
                      organization: t('action.org'),
                      workspace: t('action.workspace'),
                      project: t('action.project'),
                      language: t('action.language'),
                      refresh: t('action.refresh'),
                      openOperation: t('action.openOperation'),
                      newProject: t('action.newProject'),
                    }}
                    onLanguageUpdate={handleLanguageUpdate}
                    onOrganizationUpdate={handleOrganizationUpdate}
                    onWorkspaceUpdate={handleWorkspaceUpdate}
                    onProjectUpdate={handleProjectUpdate}
                  />
                  <Outlet />
                </div>
              </ConsoleSelectionContext.Provider>
            </ConsoleI18nContext.Provider>
          )}
        />
      </ToasterProvider>
    </ThemeProvider>
  )
}

function useConsoleSelection() {
  const selection = useContext(ConsoleSelectionContext)

  if (!selection) {
    throw new Error('Console selection context is not available')
  }

  return selection
}

function useConsoleI18n() {
  const i18n = useContext(ConsoleI18nContext)

  if (!i18n) {
    throw new Error('Console i18n context is not available')
  }

  return i18n
}

export function ResourceManagerOverviewPage() {
  const {t} = useConsoleI18n()

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: resourceManagerRoutes.overview},
                {text: t('menu.resources.overview')},
              ]}
            />
            <Text as="h1" variant="display-1">
              {t('page.resourceManager.title')}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {t('page.resourceManager.description')}
            </Text>
          </div>

          <div className="m8-summary m8-summary_overview">
            <Metric
              label={t('overview.metric.organizations')}
              value="2"
              description={t('overview.metric.organizationsDescription')}
            />
            <Metric
              label={t('overview.metric.workspaces')}
              value="3"
              description={t('overview.metric.workspacesDescription')}
            />
            <Metric
              label={t('overview.metric.projects')}
              value="147"
              description={t('overview.metric.projectsDescription')}
            />
            <Metric
              label={t('overview.metric.operations')}
              value="9"
              description={t('overview.metric.operationsDescription')}
              tone="warning"
            />
          </div>
        </div>

        <div className="m8-overview-grid">
          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.hierarchy.title')} description={t('overview.hierarchy.description')} />
            <div className="m8-overview-hierarchy">
              <OverviewHierarchyNode label={t('overview.hierarchy.organizations')} value="2" />
              <OverviewHierarchyNode label={t('overview.hierarchy.workspaces')} value="3" />
              <OverviewHierarchyNode label={t('overview.hierarchy.projects')} value="147" />
              <OverviewHierarchyNode label={t('overview.hierarchy.resources')} value="612" />
            </div>
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.resourceMix.title')} description={t('overview.resourceMix.description')} />
            <OverviewBarChart items={resourceOverviewMix} t={t} />
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.lifecycle.title')} description={t('overview.lifecycle.description')} />
            <OverviewStackChart items={resourceOverviewLifecycle} t={t} />
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card">
            <OverviewCardHeader title={t('overview.operations.title')} description={t('overview.operations.description')} />
            <div className="m8-overview-list">
              {resourceOverviewOperations.map((operation) => (
                <div className="m8-overview-list__item" key={operation.titleKey}>
                  <div>
                    <Text variant="body-2">{t(operation.titleKey)}</Text>
                    <Text variant="caption-2" color="secondary">
                      {t(operation.descriptionKey)}
                    </Text>
                  </div>
                  <Label theme={operation.theme}>{t('overview.operation.running')}</Label>
                </div>
              ))}
            </div>
          </Card>

          <Card view="outlined" type="container" className="m8-overview-card m8-overview-card_full">
            <OverviewCardHeader title={t('overview.info.title')} description={t('overview.info.description')} />
            <div className="m8-overview-signals">
              {resourceOverviewSignals.map((signal) => (
                <div className="m8-overview-signal" key={signal.titleKey}>
                  <Icon data={Check} size={16} />
                  <div>
                    <Text variant="body-2">{t(signal.titleKey)}</Text>
                    <Text variant="caption-2" color="secondary">
                      {t(signal.descriptionKey)}
                    </Text>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </section>
    </main>
  )
}

function OverviewCardHeader({title, description}: {title: string; description: string}) {
  return (
    <div>
      <Text as="h2" variant="header-1">
        {title}
      </Text>
      <Text variant="caption-2" color="secondary">
        {description}
      </Text>
    </div>
  )
}

function OverviewHierarchyNode({label, value}: {label: string; value: string}) {
  return (
    <div className="m8-overview-hierarchy__node">
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Text variant="header-2">{value}</Text>
    </div>
  )
}

function OverviewBarChart({
  items,
  t,
}: {
  items: Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>
  t: Translate
}) {
  return (
    <div className="m8-bar-chart">
      {items.map((item) => (
        <div className="m8-bar-chart__row" key={item.labelKey}>
          <Text variant="caption-2" color="secondary">
            {t(item.labelKey)}
          </Text>
          <div className="m8-bar-chart__track" aria-hidden="true">
            <div className={`m8-bar-chart__bar m8-overview-tone_${item.tone}`} style={{width: `${item.value}%`}} />
          </div>
          <Text variant="caption-2">{item.value}%</Text>
        </div>
      ))}
    </div>
  )
}

function OverviewStackChart({
  items,
  t,
}: {
  items: Array<{labelKey: TranslationKey; value: number; tone: OverviewTone}>
  t: Translate
}) {
  return (
    <div className="m8-stack-chart-shell">
      <div className="m8-stack-chart" aria-hidden="true">
        {items.map((item) => (
          <span
            className={`m8-stack-chart__segment m8-overview-tone_${item.tone}`}
            key={item.labelKey}
            style={{width: `${item.value}%`}}
          />
        ))}
      </div>
      <div className="m8-stack-chart__legend">
        {items.map((item) => (
          <div className="m8-stack-chart__legend-item" key={item.labelKey}>
            <span className={`m8-stack-chart__dot m8-overview-tone_${item.tone}`} />
            <Text variant="caption-2" color="secondary">
              {t(item.labelKey)}
            </Text>
            <Text variant="caption-2">{item.value}%</Text>
          </div>
        ))}
      </div>
    </div>
  )
}

export function ResourceOrganizationsPage() {
  const {t} = useConsoleI18n()

  return (
    <ResourcePlaceholderPage
      current={t('menu.resources.organizations')}
      title={t('page.organizations.title')}
      description={t('page.organizations.description')}
    />
  )
}

export function ResourceOrganizationDetailsPage() {
  const {t} = useConsoleI18n()

  return (
    <ResourcePlaceholderPage
      current={t('menu.resources.organizations')}
      title={t('page.organizationDetails.title')}
      description={t('page.organizationDetails.description')}
    />
  )
}

export function ResourceWorkspacesPage() {
  const {t} = useConsoleI18n()

  return (
    <ResourcePlaceholderPage
      current={t('menu.resources.workspaces')}
      title={t('page.workspaces.title')}
      description={t('page.workspaces.description')}
    />
  )
}

export function ResourceWorkspaceDetailsPage() {
  const {t} = useConsoleI18n()

  return (
    <ResourcePlaceholderPage
      current={t('menu.resources.workspaces')}
      title={t('page.workspaceDetails.title')}
      description={t('page.workspaceDetails.description')}
    />
  )
}

export function ResourceProjectDetailsPage() {
  return <ResourceProjectsPage />
}

export function ResourceProjectsPage() {
  const {organization, workspace, projectId, setWorkspace, setProjectId} = useConsoleSelection()
  const {t} = useConsoleI18n()
  const [status, setStatus] = useState('all')
  const [owner, setOwner] = useState('all')
  const [search, setSearch] = useState('')
  const workspaceOptions = useMemo(() => translateOptions(workspaceOptionConfigs, t), [t])
  const statusOptions = useMemo(() => translateOptions(statusOptionConfigs, t), [t])
  const ownerOptions = useMemo(() => translateOptions(ownerOptionConfigs, t), [t])

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
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: resourceManagerRoutes.overview},
                {text: t('projects.title')},
              ]}
            />
            <Text as="h1" variant="display-1">
              {t('projects.title')}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {t('projects.description')}
            </Text>
          </div>

          <div className="m8-summary">
            <Metric label={t('projects.metric.projects')} value="147" description={t('projects.metric.projectsDescription')} />
            <Metric label={t('projects.metric.failed')} value="2" description={t('projects.metric.failedDescription')} tone="danger" />
            <Metric label={t('projects.metric.deleting')} value="4" description={t('projects.metric.deletingDescription')} tone="warning" />
          </div>
        </div>

        <Card view="outlined" type="container" className="m8-filter-card">
          <div className="m8-filters">
            <label className="m8-field">
              <Text variant="caption-2" color="secondary">
                {t('projects.search')}
              </Text>
              <TextInput
                value={search}
                placeholder={t('projects.searchPlaceholder')}
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
                  (project) => project.organization === organization && project.workspace === nextWorkspace,
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
                  {t('projects.inventory')}
                </Text>
                <Text variant="caption-2" color="secondary">
                  {t('projects.inventoryDescription')}
                </Text>
              </div>
                <div className="m8-labels">
                  {statusOptions.slice(1).map((option) => (
                  <StatusLabel key={option.value} status={option.value as ProjectStatus} t={t} />
                  ))}
                </div>
            </div>

              <ProjectTable
                projects={visibleProjects}
                selectedProjectId={projectId}
                onSelectProject={setProjectId}
                t={t}
              />
          </Card>
        </div>
      </section>
    </main>
  )
}

function ResourcePlaceholderPage({
  current,
  title,
  description,
}: {
  current: string
  title: string
  description: string
}) {
  const {t} = useConsoleI18n()

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: resourceManagerRoutes.overview},
                {text: current},
              ]}
            />
            <Text as="h1" variant="display-1">
              {title}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {description}
            </Text>
          </div>
        </div>

        <Card view="outlined" type="container" className="m8-placeholder-card">
          <Text as="h2" variant="header-1">
            {t('page.placeholder.title')}
          </Text>
          <Text variant="body-2" color="secondary">
            {t('page.placeholder.description')}
          </Text>
        </Card>
      </section>
    </main>
  )
}

function ProjectTable({
  projects,
  selectedProjectId,
  onSelectProject,
  t,
}: {
  projects: Project[]
  selectedProjectId: string
  onSelectProject: (projectId: string) => void
  t: Translate
}) {
  if (projects.length === 0) {
    return (
      <div className="m8-empty-table">
        <Text variant="body-2">{t('projects.empty')}</Text>
        <Text variant="caption-2" color="secondary">
          {t('projects.emptyDescription')}
        </Text>
      </div>
    )
  }

  return (
    <div className="m8-table-shell">
      <table className="m8-project-table">
        <thead>
          <tr>
            <th>{t('projects.column.project')}</th>
            <th>{t('projects.column.projectId')}</th>
            <th>{t('projects.column.workspace')}</th>
            <th>{t('projects.column.organization')}</th>
            <th>{t('projects.column.status')}</th>
            <th>{t('projects.column.desiredState')}</th>
            <th>{t('projects.column.actualState')}</th>
            <th>{t('projects.column.updated')}</th>
            <th>{t('projects.column.owner')}</th>
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
                <StatusLabel status={project.status} t={t} />
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

function StatusLabel({status, t}: {status: ProjectStatus; t: Translate}) {
  const themeByStatus: Record<ProjectStatus, 'success' | 'warning' | 'danger' | 'info' | 'normal'> = {
    Active: 'success',
    Suspended: 'warning',
    Failed: 'danger',
    Provisioning: 'info',
    Deleting: 'warning',
  }
  const statusTitleKey: Record<ProjectStatus, TranslationKey> = {
    Active: 'status.Active',
    Suspended: 'status.Suspended',
    Failed: 'status.Failed',
    Provisioning: 'status.Provisioning',
    Deleting: 'status.Deleting',
  }

  return <Label theme={themeByStatus[status]}>{t(statusTitleKey[status])}</Label>
}

export default App
