import {useMemo} from 'react'
import {ClipboardButton, Text} from '@gravity-ui/uikit'

import type {ResourceTableColumn} from '../../../components/ResourceTable'
import type {AppLanguage, Translate} from '../../../i18n'
import type {Organization} from '../api/organizations'
import {CopyableOrganizationID, OrganizationStateLabel} from './OrganizationTableCells'

export function useOrganizationColumns(language: AppLanguage, t: Translate) {
  return useMemo<ResourceTableColumn<Organization>[]>(
    () => [
      {
        id: 'name',
        name: t('organizations.column.name'),
        width: 260,
        template: (organization) => (
          <div className="m8-organization-name">
            <div className="m8-copyable-cell">
              <Text variant="body-2" ellipsis>{organization.name || t('organizations.unnamed')}</Text>
              {organization.name ? (
                <ClipboardButton
                  text={organization.name}
                  view="flat-secondary"
                  size="s"
                  tooltipInitialText={t('resource.copy')}
                  tooltipSuccessText={t('resource.copied')}
                />
              ) : null}
            </div>
            {organization.description ? (
              <Text variant="caption-2" color="secondary" ellipsis>{organization.description}</Text>
            ) : null}
          </div>
        ),
      },
      {
        id: 'state',
        name: t('organizations.column.state'),
        width: 150,
        meta: {sortable: false},
        template: (organization) => <OrganizationStateLabel state={organization.state} />,
      },
      {
        id: 'id',
        name: t('organizations.column.id'),
        width: 300,
        template: (organization) => <CopyableOrganizationID id={organization.id} t={t} />,
      },
      {
        id: 'version',
        name: t('organizations.column.version'),
        width: 100,
        meta: {sortable: false},
        template: (organization) => organization.version ?? '—',
      },
      {
        id: 'createTime',
        name: t('organizations.column.created'),
        width: 190,
        template: (organization) => formatDate(organization.createTime, language),
      },
      {
        id: 'updateTime',
        name: t('organizations.column.updated'),
        width: 190,
        template: (organization) => formatDate(organization.updateTime, language),
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
