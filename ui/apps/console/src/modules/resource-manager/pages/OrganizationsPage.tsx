import {useCallback, useMemo, useState} from 'react'
import {ArrowRotateRight} from '@gravity-ui/icons'
import type {SortingState} from '@gravity-ui/table/tanstack'
import {Button, Card, Icon, Text} from '@gravity-ui/uikit'
import {useRouter} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {AppLanguage, Translate} from '../../../i18n'
import {OrganizationsTable} from '../components/OrganizationsTable'
import {useOrganizationsQuery} from '../queries/organizations'

const defaultSorting: SortingState = [{id: 'name', desc: false}]

export interface OrganizationsPageProps {
  language: AppLanguage
  t: Translate
}

export function OrganizationsPage({language, t}: OrganizationsPageProps) {
  const router = useRouter()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [pageTokens, setPageTokens] = useState<Record<number, string>>({1: ''})
  const [nameFilter, setNameFilter] = useState('')
  const [sorting, setSorting] = useState<SortingState>(defaultSorting)
  const filter = useMemo(() => buildNameFilter(nameFilter), [nameFilter])
  const orderBy = useMemo(() => buildOrderBy(sorting), [sorting])
  const organizationsQuery = useOrganizationsQuery({
    pageSize,
    pageToken: pageTokens[page],
    filter,
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
          <Button
            view="outlined"
            loading={organizationsQuery.isFetching}
            onClick={() => void organizationsQuery.refetch()}
          >
            <Icon data={ArrowRotateRight} size={16} />
            {t('organizations.refresh')}
          </Button>
        </div>

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

          {organizationsQuery.isError ? (
            <div className="m8-organizations-message" role="alert">
              <Text variant="body-2">{t('organizations.error')}</Text>
              <Text variant="caption-2" color="secondary">
                {organizationsQuery.error instanceof Error ? organizationsQuery.error.message : null}
              </Text>
            </div>
          ) : (
            <OrganizationsTable
              key={`${page}:${pageSize}:${filter}:${orderBy}`}
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
  return name ? `name = ${JSON.stringify(name)}` : undefined
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
