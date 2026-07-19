import {useMemo} from 'react'
import {ActionsPanel, ClipboardButton, Icon, Label, Text} from '@gravity-ui/uikit'
import {ArrowRight, Copy} from '@gravity-ui/icons'
import type {ColumnDef} from '@gravity-ui/table/tanstack'
import {toaster} from '@gravity-ui/uikit/toaster-singleton'

import {ResourceTable} from '../../../components/ResourceTable'
import type {AppLanguage, Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'

export interface OrganizationsTableProps {
  organizations: Organization[]
  language: AppLanguage
  loading: boolean
  page: number
  pageSize: number
  total: number
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
  onPaginationUpdate,
  onOrganizationActivate,
  t,
}: OrganizationsTableProps) {
  const columns = useMemo<ColumnDef<Organization>[]>(
    () => [
      {
        accessorKey: 'name',
        header: t('organizations.column.name'),
        size: 260,
        cell: ({row}) => (
          <div className="m8-organization-name">
            <div className="m8-copyable-cell">
              <Text variant="body-2" ellipsis>{row.original.name || t('organizations.unnamed')}</Text>
              {row.original.name ? (
                <ClipboardButton
                  text={row.original.name}
                  view="flat-secondary"
                  size="s"
                  tooltipInitialText={t('resource.copy')}
                  tooltipSuccessText={t('resource.copied')}
                />
              ) : null}
            </div>
            {row.original.description ? (
              <Text variant="caption-2" color="secondary" ellipsis>
                {row.original.description}
              </Text>
            ) : null}
          </div>
        ),
      },
      {
        accessorKey: 'state',
        header: t('organizations.column.state'),
        size: 150,
        cell: ({getValue}) => <OrganizationState state={getValue<Organization['state']>()} />,
      },
      {
        accessorKey: 'id',
        header: t('organizations.column.id'),
        size: 300,
        cell: ({getValue}) => {
          const id = getValue<string>()
          return (
            <div className="m8-copyable-cell">
              <span className="m8-mono">{id}</span>
              <ClipboardButton
                text={id}
                view="flat-secondary"
                size="s"
                tooltipInitialText={t('resource.copy')}
                tooltipSuccessText={t('resource.copied')}
              />
            </div>
          )
        },
      },
      {
        accessorKey: 'version',
        header: t('organizations.column.version'),
        size: 100,
        cell: ({getValue}) => getValue<string | number>() ?? '—',
      },
      {
        accessorKey: 'createTime',
        header: t('organizations.column.created'),
        size: 190,
        cell: ({getValue}) => formatDate(getValue<string>(), language),
      },
      {
        accessorKey: 'updateTime',
        header: t('organizations.column.updated'),
        size: 190,
        cell: ({getValue}) => formatDate(getValue<string>(), language),
      },
    ],
    [language, t],
  )

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
      filtering={{searchPlaceholder: t('organizations.filters.searchPlaceholder')}}
      pagination={{
        page,
        pageSize,
        total,
        pageSizeOptions: [10, 20, 50, 100],
        onUpdate: onPaginationUpdate,
      }}
      settings={{
        storageKey: 'm8.resource-manager.organizations.table-settings',
        enableSearch: true,
        searchPlaceholder: t('organizations.settings.searchPlaceholder'),
      }}
      onRowActivate={onOrganizationActivate}
      renderSelectionActions={({selectedItems, clearSelection}) => {
        const openSelected = () => {
          if (selectedItems.length === 1) onOrganizationActivate(selectedItems[0])
        }
        const copySelectedIds = async () => {
          try {
            await navigator.clipboard.writeText(selectedItems.map((organization) => organization.id).join('\n'))
            toaster.add({
              name: 'organization-ids-copied',
              title: t('organizations.actions.idsCopied'),
              theme: 'success',
              autoHiding: 3000,
            })
          } catch {
            toaster.add({
              name: 'organization-ids-copy-failed',
              title: t('organizations.actions.copyFailed'),
              theme: 'danger',
              autoHiding: 5000,
            })
          }
        }

        return (
          <ActionsPanel
            className="m8-selection-actions-panel"
            renderNote={() => `${t('organizations.actions.selected')}: ${selectedItems.length}`}
            onClose={clearSelection}
            actions={[
              {
                id: 'open',
                button: {
                  props: {
                    children: [
                      <Icon key="icon" data={ArrowRight} size={16} />,
                      t('organizations.actions.open'),
                    ],
                    disabled: selectedItems.length !== 1,
                    onClick: openSelected,
                  },
                },
                dropdown: {
                  item: {
                    text: t('organizations.actions.open'),
                    disabled: selectedItems.length !== 1,
                    action: openSelected,
                    iconStart: <Icon data={ArrowRight} size={16} />,
                  },
                },
              },
              {
                id: 'copy-ids',
                button: {
                  props: {
                    children: [<Icon key="icon" data={Copy} size={16} />, t('organizations.actions.copyIds')],
                    onClick: () => void copySelectedIds(),
                  },
                },
                dropdown: {
                  item: {
                    text: t('organizations.actions.copyIds'),
                    action: () => void copySelectedIds(),
                    iconStart: <Icon data={Copy} size={16} />,
                  },
                },
              },
            ]}
          />
        )
      }}
    />
  )
}

function OrganizationState({state}: {state: Organization['state']}) {
  const theme =
    state === 'ACTIVE'
      ? 'success'
      : state === 'FAILED'
        ? 'danger'
        : state === 'SUSPENDED' || state === 'DELETING'
          ? 'warning'
          : 'normal'
  return <Label theme={theme}>{state.replace('STATE_', '')}</Label>
}

function formatDate(value: string | undefined, language: AppLanguage) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(language === 'ru' ? 'ru-RU' : 'en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}
