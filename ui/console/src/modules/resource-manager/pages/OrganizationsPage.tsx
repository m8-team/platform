import {useCallback, useMemo, useState} from 'react'
import {ArrowRotateRight} from '@gravity-ui/icons'
import type {SortingState} from '@gravity-ui/table/tanstack'
import {Button, Card, Icon, Text, TextInput} from '@gravity-ui/uikit'
import {useRouter} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {AppLanguage, Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'
import {OrganizationsTable} from '../components/OrganizationsTable'
import {OrganizationCreatePanel} from '../components/OrganizationCreatePanel'
import {useOrganizationsQuery} from '../queries/organizations'

const defaultSorting: SortingState = [{id: 'name', desc: false}]
const tenantIDPattern = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
const tenantScopeKey = 'm8.resource-manager.organizations.tenant-id'
function defaultTenantID(): string {
  // Preserve the explicitly selected Tenant across reloads. The env value is
  // only a fallback for development, never an override of the user's scope.
  let saved = ''
  try {
    saved = (sessionStorage.getItem(tenantScopeKey) ?? '').trim()
  } catch {
    // Session storage can be unavailable in locked-down browsers.
  }
  const configured = (import.meta.env.VITE_RESOURCE_MANAGER_TENANT_ID ?? '').trim()
  if (tenantIDPattern.test(saved)) return saved
  return tenantIDPattern.test(configured) ? configured : ''
}

export interface OrganizationsPageProps {
  language: AppLanguage
  t: Translate
}

export function OrganizationsPage({language, t}: OrganizationsPageProps) {
  const router = useRouter()
  const [creating, setCreating] = useState(false)
  const [tenantId, setTenantId] = useState(defaultTenantID)
  const [tenantInput, setTenantInput] = useState(defaultTenantID)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [pageTokens, setPageTokens] = useState<Record<number, string>>({1: ''})
  const [nameFilter, setNameFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<Organization['state'] | ''>('')
  const [sorting, setSorting] = useState<SortingState>(defaultSorting)
  const filter = useMemo(() => buildNameFilter(nameFilter), [nameFilter])
  const orderBy = useMemo(() => buildOrderBy(sorting), [sorting])
  const organizationsQuery = useOrganizationsQuery({
    tenantId,
    pageSize,
    pageToken: pageTokens[page],
    filter,
    stateFilter: statusFilter || undefined,
    orderBy,
  })

  const organizations = organizationsQuery.data?.organizations ?? []
  const resetPagination = useCallback(() => {
    setPage(1)
    setPageTokens({1: ''})
  }, [])
  const handleFilterUpdate = useCallback(
    (value: string) => {
      setNameFilter(value)
      resetPagination()
    },
    [resetPagination],
  )
  const handleStatusUpdate = useCallback(
    (value: Organization['state'] | '') => {
      setStatusFilter(value)
      resetPagination()
    },
    [resetPagination],
  )
  const clearFilters = useCallback(() => {
    setNameFilter('')
    setStatusFilter('')
    resetPagination()
  }, [resetPagination])
  const handleSortingUpdate = useCallback(
    (value: SortingState) => {
      setSorting(value.length > 0 ? [value[0]] : defaultSorting)
      resetPagination()
    },
    [resetPagination],
  )
  const handlePaginationUpdate = useCallback(
    (nextPage: number, nextPageSize: number) => {
      if (organizationsQuery.isFetching) return
      if (nextPageSize !== pageSize) {
        setPageSize(nextPageSize)
        resetPagination()
        return
      }

      if (nextPage === page + 1) {
        const nextPageToken = organizationsQuery.data?.nextPageToken
        if (!nextPageToken) return
        setPageTokens((current) => ({...current, [nextPage]: nextPageToken}))
      } else if (nextPage > 1 && pageTokens[nextPage] === undefined) {
        return
      }
      setPage(nextPage)
    },
    [organizationsQuery.data?.nextPageToken, organizationsQuery.isFetching, page, pageSize, pageTokens, resetPagination],
  )

  const selectTenant = () => {
    const selected = tenantInput.trim()
    if (!tenantIDPattern.test(selected)) return
    try {
      sessionStorage.setItem(tenantScopeKey, selected)
    } catch {
      // The selected Tenant remains usable even without session storage.
    }
    setTenantId(selected)
    resetPagination()
  }
  const validTenant = tenantIDPattern.test(tenantInput.trim())

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: '/resource-manager'},
                {text: t('menu.resources.organizations')},
              ]}
            />
            <Text as="h1" variant="display-1">
              {t('page.organizations.title')}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {t('page.organizations.description')}
            </Text>
          </div>
          <div className="m8-org-editor__actions">
            <Button view="action" onClick={() => setCreating((value) => !value)}>
              {t('organizations.create.title')}
            </Button>
            <Button
              view="outlined"
              loading={organizationsQuery.isFetching}
              onClick={() => void organizationsQuery.refetch()}
            >
              <Icon data={ArrowRotateRight} size={16} />
              {t('organizations.refresh')}
            </Button>
          </div>
        </div>

        {creating ? (
          <OrganizationCreatePanel
            key={tenantId || 'unselected-tenant'}
            t={t}
            initialTenantId={tenantId}
            onCancel={() => setCreating(false)}
          />
        ) : null}

        <Card view="outlined" className="m8-org-editor">
          <form className="m8-org-editor__form" onSubmit={(e) => {e.preventDefault();selectTenant()}}>
            <Text variant="body-2">
              {t('organizations.tenant.selectTitle')}
            </Text>
            <Text variant="caption-2" color="secondary">
              {t('organizations.tenant.selectDescription')}
            </Text>
            <div className="m8-org-editor__actions">
              <div style={{flex: 1, minWidth: '220px'}}>
                <TextInput value={tenantInput} onUpdate={setTenantInput} size="m"
                  placeholder={t('organizations.tenant.inputPlaceholder')} aria-label={t('organizations.tenant.inputLabel')}/>
              </div>
              <Button type="submit" view="action" disabled={!validTenant}>
                {t('organizations.tenant.load')}
              </Button>
            </div>
          </form>
        </Card>

        <Card view="outlined" type="container" className="m8-table-card">
          <div className="m8-card-header">
            <div>
              <Text as="h2" variant="header-1">
                {t('organizations.inventory')}
              </Text>
              <Text variant="caption-2" color="secondary">
                {t('organizations.total')}: {organizationsQuery.data?.totalSize ?? organizations.length}
              </Text>
            </div>
          </div>

          <div className="m8-organizations-toolbar">
            <label className="m8-organizations-toolbar__status">
              <Text variant="caption-2">{t('organizations.filters.status')}</Text>
              <select
                className="m8-org-editor__select"
                value={statusFilter}
                onChange={(event) => handleStatusUpdate(event.target.value as Organization['state'] | '')}
                aria-label={t('organizations.filters.statusAriaLabel')}
              >
                <option value="">{t('organizations.filters.allStatuses')}</option>
                <option value="ACTIVE">{t('organizations.state.ACTIVE')}</option>
                <option value="SUSPENDED">{t('organizations.state.SUSPENDED')}</option>
                <option value="ARCHIVED">{t('organizations.state.ARCHIVED')}</option>
              </select>
            </label>
            {statusFilter || nameFilter ? (
              <Button view="flat-secondary" onClick={clearFilters}>
                {t('organizations.filters.clear')}
              </Button>
            ) : null}
          </div>

          {!tenantId ? (
            <div className="m8-organizations-message">
              <Text variant="body-2" color="secondary">{t('organizations.tenant.required')}</Text>
            </div>
          ) : organizationsQuery.isError ? (
            <div className="m8-organizations-message" role="alert">
              <Text variant="body-2">{t('organizations.error')}</Text>
              <Text variant="caption-2" color="secondary">
                {organizationsQuery.error instanceof Error ? organizationsQuery.error.message : null}
              </Text>
            </div>
          ) : (
            <OrganizationsTable
              key={`${tenantId}:${page}:${pageSize}:${filter}:${statusFilter}:${orderBy}`}
              organizations={organizations}
              language={language}
              loading={organizationsQuery.isFetching}
              page={page}
              pageSize={pageSize}
              total={organizationsQuery.data?.totalSize ?? organizations.length}
              paginationDisabled={organizationsQuery.isFetching}
              nameFilter={nameFilter}
              sorting={sorting}
              onFilterUpdate={handleFilterUpdate}
              onSortingUpdate={handleSortingUpdate}
              onPaginationUpdate={handlePaginationUpdate}
              t={t}
              onOrganizationActivate={(organization) =>
                void router.navigate({
                  to: '/resource-manager/organizations/$organizationId',
                  params: {organizationId: organization.id},
                })
              }
            />
          )}
        </Card>
      </section>
    </main>
  )
}

function buildNameFilter(value: string) {
  const name = value.trim()
  return name || undefined
}

function buildOrderBy(sorting: SortingState) {
  const selected = sorting[0] ?? defaultSorting[0]
  const fieldByColumn: Record<string, string> = {
    id: 'id',
    name: 'name',
    createTime: 'create_time',
    updateTime: 'update_time',
  }
  const field = fieldByColumn[selected.id] ?? 'name'
  return `${field} ${selected.desc ? 'desc' : 'asc'}`
}
