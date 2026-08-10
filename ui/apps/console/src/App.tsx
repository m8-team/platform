import {useCallback, useEffect, useMemo, useState, useSyncExternalStore} from 'react'
import {
  Avatar,
  Button,
  configure,
  Label,
  Text,
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
import {ServiceRequestConsole} from './components/ServiceRequestConsole'
import {ConsoleI18nContext, ConsoleSelectionContext} from './console/ConsoleContext'
import type {ConsoleI18n, ConsoleSelection} from './console/ConsoleContext'
import {workspaceOptionConfigs} from './modules/resource-manager/config/projectFilters'
import {translateOptions} from './modules/resource-manager/lib/translateOptions'
import {projects} from './modules/resource-manager/model/projectFixtures'
import {resourceManagerRoutes} from './modules/resource-manager/routes'
import {isServiceRequestLoggingEnabled} from './platform/http/loggedFetch'
import {serviceRequestLog} from './platform/http/serviceRequestLog'
import {
  createTranslator,
  fallbackLanguage,
  isAppLanguage,
  languageOptions as languageOptionConfigs,
} from './i18n'
import type {AppLanguage, TranslationKey} from './i18n'
import './App.css'

type FooterPanel = 'notifications' | 'support' | 'request-console' | 'account'

const languageStorageKey = 'm8.console.language'
const navigationCompactStorageKey = 'm8.console.navigation.compact'
const menuGroupCollapsedStorageKey = 'm8.console.menu-groups.collapsed'

const organizationOptions = [
  {value: 'org_m8_finance_6b21d0', content: 'Acme'},
  {value: 'org_m8_billing_91f2c5', content: 'Billing'},
]

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

function App() {
  const [language, setLanguage] = useState<AppLanguage>(readInitialLanguage)
  const [compact, setCompact] = useState(readInitialNavigationCompact)
  const [collapsedMenuGroupIds, setCollapsedMenuGroupIds] = useState(readInitialCollapsedMenuGroups)
  const [activeFooterPanel, setActiveFooterPanel] = useState<FooterPanel | null>(null)
  const serviceRequestRecords = useSyncExternalStore(
    serviceRequestLog.subscribe,
    serviceRequestLog.getSnapshot,
    serviceRequestLog.getSnapshot,
  )
  const serviceRequestCounts = useMemo(
    () => serviceRequestRecords.reduce(
      (counts, record) => {
        if (record.pending) return counts
        if (record.status !== undefined && record.status < 400 && !record.error) counts.success += 1
        else counts.failure += 1
        return counts
      },
      {success: 0, failure: 0},
    ),
    [serviceRequestRecords],
  )
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
      ...(isServiceRequestLoggingEnabled
        ? [
            {
              id: 'request-console',
              open: activeFooterPanel === 'request-console',
              size: 560,
              className: 'm8-api-debug-panel',
              contentOverflow: 'auto' as const,
              hideVeil: false,
              children: <ServiceRequestConsole t={t} onClose={() => setActiveFooterPanel(null)} />,
            },
          ]
        : []),
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
              {isServiceRequestLoggingEnabled ? (
                <FooterItem
                  id="request-console"
                  icon={Code}
                  title={t('footer.requestConsole')}
                  tooltipText={t('footer.requestConsole')}
                  rightAdornment={(
                    <span className="m8-request-console__menu-counts">
                      <Label theme="success">{serviceRequestCounts.success}</Label>
                      <Label theme="danger">{serviceRequestCounts.failure}</Label>
                    </span>
                  )}
                  current={activeFooterPanel === 'request-console'}
                  onItemClick={() => {
                    setActiveFooterPanel(activeFooterPanel === 'request-console' ? null : 'request-console')
                  }}
                  compact={footerCompact}
                />
              ) : null}
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

export default App
