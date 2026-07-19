import {useMemo} from 'react'
import {ClipboardButton, Text} from '@gravity-ui/uikit'
import type {ColumnDef} from '@gravity-ui/table/tanstack'

import type {AppLanguage, Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'
import {CopyableOrganizationID, OrganizationStateLabel} from './OrganizationTableCells'

export function useOrganizationColumns(language: AppLanguage, t: Translate) {
  return useMemo<ColumnDef<Organization>[]>(
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
              <Text variant="caption-2" color="secondary" ellipsis>{row.original.description}</Text>
            ) : null}
          </div>
        ),
      },
      {
        accessorKey: 'state',
        header: t('organizations.column.state'),
        size: 150,
        enableSorting: false,
        cell: ({getValue}) => <OrganizationStateLabel state={getValue<Organization['state']>()} />,
      },
      {
        accessorKey: 'id',
        header: t('organizations.column.id'),
        size: 300,
        cell: ({getValue}) => <CopyableOrganizationID id={getValue<string>()} t={t} />,
      },
      {
        accessorKey: 'version',
        header: t('organizations.column.version'),
        size: 100,
        enableSorting: false,
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
