import {useMemo, useState} from 'react'
import {ArrowRotateRight} from '@gravity-ui/icons'
import {Button, Card, Icon, Select, Text} from '@gravity-ui/uikit'
import {useRouter} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {AppLanguage, Translate} from '../../../i18n'
import {WorkspacesTable} from '../components/WorkspacesTable'
import {useOrganizationsByIdsQuery} from '../queries/organizations'
import {useWorkspacesQuery} from '../queries/workspaces'

const allOrganizationsValue = '__all__'

export function WorkspacesPage({language, t}: {language: AppLanguage; t: Translate}) {
  const router = useRouter()
  const [organizationFilter, setOrganizationFilter] = useState(allOrganizationsValue)
  const workspacesQuery = useWorkspacesQuery()
  const organizationIds = useMemo(
    () => [...new Set((workspacesQuery.data?.workspaces ?? []).map(({organizationId}) => organizationId).filter(Boolean))].sort(),
    [workspacesQuery.data?.workspaces],
  )
  const organizationsQuery = useOrganizationsByIdsQuery(organizationIds)
  const organizations = useMemo(() => organizationsQuery.data?.organizations ?? [], [organizationsQuery.data?.organizations])
  const organizationNames = useMemo(
    () => new Map(organizations.map(({id, name}) => [id, name || id])),
    [organizations],
  )
  const allWorkspaces = useMemo(
    () => (workspacesQuery.data?.workspaces ?? []).map((workspace) => ({
      ...workspace,
      organizationName: organizationNames.get(workspace.organizationId),
    })),
    [organizationNames, workspacesQuery.data?.workspaces],
  )
  const workspaces = useMemo(
    () => organizationFilter === allOrganizationsValue
      ? allWorkspaces
      : allWorkspaces.filter(({organizationId}) => organizationId === organizationFilter),
    [allWorkspaces, organizationFilter],
  )
  const organizationOptions = useMemo(
    () => [
      {value: allOrganizationsValue, content: t('workspaces.organizations.all')},
      ...organizations.map(({id, name}) => ({value: id, content: name || id})),
    ],
    [organizations, t],
  )
  const loading = organizationsQuery.isLoading || workspacesQuery.isFetching
  const error = organizationsQuery.error ?? workspacesQuery.error

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs items={[{text: t('breadcrumb.resourceManager'), href: '/resource-manager'}, {text: t('menu.resources.workspaces')}]} />
            <Text as="h1" variant="display-1">{t('page.workspaces.title')}</Text>
            <Text as="p" variant="body-2" color="secondary">{t('page.workspaces.description')}</Text>
          </div>
          <Button view="outlined" loading={loading} onClick={() => void Promise.all([organizationsQuery.refetch(), workspacesQuery.refetch()])}>
            <Icon data={ArrowRotateRight} size={16} />{t('workspaces.refresh')}
          </Button>
        </div>
        <Card view="outlined" type="container" className="m8-table-card">
          <div className="m8-card-header">
            <div>
              <Text as="h2" variant="header-1">{t('workspaces.inventory')}</Text>
              <Text variant="caption-2" color="secondary">{t('workspaces.total')}: {workspaces.length}</Text>
            </div>
            <Select
              aria-label={t('workspaces.organization')}
              value={[organizationFilter]}
              options={organizationOptions}
              width="max"
              loading={organizationsQuery.isLoading}
              onUpdate={(value) => setOrganizationFilter(value[0] ?? allOrganizationsValue)}
            />
          </div>
          {error ? (
            <div className="m8-organizations-message" role="alert">
              <Text variant="body-2">{t('workspaces.error')}</Text>
              <Text variant="caption-2" color="secondary">{error instanceof Error ? error.message : null}</Text>
            </div>
          ) : (
            <WorkspacesTable
              key={organizationFilter}
              workspaces={workspaces}
              language={language}
              loading={loading}
              onWorkspaceActivate={(workspace) => void router.navigate({to: '/resource-manager/workspaces/$workspaceId', params: {workspaceId: workspace.id}})}
              t={t}
            />
          )}
        </Card>
      </section>
    </main>
  )
}
