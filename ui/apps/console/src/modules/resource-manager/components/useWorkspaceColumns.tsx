import {useMemo} from 'react'
import {ClipboardButton, Label, Text} from '@gravity-ui/uikit'

import type {ResourceTableColumn} from '../../../components/ResourceTable'
import type {AppLanguage, Translate} from '../../../i18n'
import type {Workspace} from '../api/workspaces'

export function useWorkspaceColumns(language: AppLanguage, t: Translate) {
  return useMemo<ResourceTableColumn<Workspace>[]>(
    () => [
      {
        id: 'name',
        name: t('workspaces.column.name'),
        width: 260,
        template: (workspace) => (
          <div className="m8-organization-name">
            <div className="m8-copyable-cell">
              <Text variant="body-2" ellipsis>{workspace.name || t('workspaces.unnamed')}</Text>
              {workspace.name ? <ClipboardButton text={workspace.name} view="flat-secondary" size="s" /> : null}
            </div>
            {workspace.description ? <Text variant="caption-2" color="secondary" ellipsis>{workspace.description}</Text> : null}
          </div>
        ),
      },
      {
        id: 'organizationName',
        name: t('workspaces.column.organization'),
        width: 240,
        template: (workspace) => (
          <div className="m8-organization-name">
            <Text variant="body-2" ellipsis>{workspace.organizationName || workspace.organizationId || '—'}</Text>
            {workspace.organizationName && workspace.organizationId ? (
              <Text variant="caption-2" color="secondary" ellipsis>{workspace.organizationId}</Text>
            ) : null}
          </div>
        ),
      },
      {
        id: 'state',
        name: t('workspaces.column.state'),
        width: 150,
        meta: {sortable: false},
        template: ({state}) => <Label theme={stateTheme(state)}>{state.replace('STATE_', '')}</Label>,
      },
      {
        id: 'id',
        name: t('workspaces.column.id'),
        width: 300,
        template: ({id}) => (
          <div className="m8-copyable-cell">
            <span className="m8-mono">{id}</span>
            <ClipboardButton text={id} view="flat-secondary" size="s" />
          </div>
        ),
      },
      {id: 'version', name: t('workspaces.column.version'), width: 100, meta: {sortable: false}, template: ({version}) => version ?? '—'},
      {id: 'createTime', name: t('workspaces.column.created'), width: 190, template: ({createTime}) => formatDate(createTime, language)},
      {id: 'updateTime', name: t('workspaces.column.updated'), width: 190, template: ({updateTime}) => formatDate(updateTime, language)},
    ],
    [language, t],
  )
}

function stateTheme(state: Workspace['state']) {
  if (state === 'ACTIVE') return 'success'
  if (state === 'FAILED') return 'danger'
  if (state === 'SUSPENDED' || state === 'DELETING') return 'warning'
  return 'normal'
}

function formatDate(value: string | undefined, language: AppLanguage) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(language === 'ru' ? 'ru-RU' : 'en-US', {dateStyle: 'medium', timeStyle: 'short'}).format(date)
}
