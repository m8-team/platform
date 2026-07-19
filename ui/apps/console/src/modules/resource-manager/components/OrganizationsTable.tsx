import {ResourceTable} from '../../../components/ResourceTable'
import type {SortingState} from '@gravity-ui/table/tanstack'

import type {AppLanguage, Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'
import {OrganizationActionsPanel} from './OrganizationActionsPanel'
import {useOrganizationColumns} from './useOrganizationColumns'

export interface OrganizationsTableProps {
  organizations: Organization[]
  language: AppLanguage
  loading: boolean
  page: number
  pageSize: number
  total: number
  paginationDisabled: boolean
  nameFilter: string
  sorting: SortingState
  onFilterUpdate: (value: string) => void
  onSortingUpdate: (value: SortingState) => void
  onPaginationUpdate: (page: number, pageSize: number) => void
  onOrganizationActivate: (organization: Organization) => void
  t: Translate
}

export function OrganizationsTable({
  organizations,
  language,
  loading,
  page,
  pageSize,
  total,
  paginationDisabled,
  nameFilter,
  sorting,
  onFilterUpdate,
  onSortingUpdate,
  onPaginationUpdate,
  onOrganizationActivate,
  t,
}: OrganizationsTableProps) {
  const columns = useOrganizationColumns(language, t)

  return (
    <ResourceTable
      data={organizations}
      columns={columns}
      getRowId={(organization) => organization.id}
      loading={loading}
      loadingContent={t('organizations.loading')}
      emptyContent={t('organizations.empty')}
      className="m8-organizations-table"
      selectable
      sortable
      sorting={{mode: 'server', value: sorting, onUpdate: onSortingUpdate}}
      filtering={{
        mode: 'server',
        value: nameFilter,
        onUpdate: onFilterUpdate,
        searchPlaceholder: t('organizations.filters.searchPlaceholder'),
        ariaLabel: t('organizations.filters.searchPlaceholder'),
      }}
      pagination={{
        mode: 'server',
        page,
        pageSize,
        total,
        disabled: paginationDisabled,
        pageSizeOptions: [10, 20, 50, 100],
        onUpdate: onPaginationUpdate,
      }}
      settings={{
        storageKey: 'm8.resource-manager.organizations.table-settings',
        enableSearch: true,
        searchPlaceholder: t('organizations.settings.searchPlaceholder'),
      }}
      onRowActivate={onOrganizationActivate}
      getRowAriaLabel={(organization) => organization.name || organization.id}
      renderSelectionActions={({selectedItems, clearSelection}) => (
        <OrganizationActionsPanel
          organizations={selectedItems}
          onClear={clearSelection}
          onOpen={onOrganizationActivate}
          t={t}
        />
      )}
    />
  )
}
