import {ResourceTable} from '../../../components/ResourceTable'
import type {AppLanguage, Translate} from '../../../i18n'
import type {Workspace} from '../api/workspaces'
import {useWorkspaceColumns} from './useWorkspaceColumns'

interface WorkspacesTableProps {
  workspaces: Workspace[]
  language: AppLanguage
  loading: boolean
  onWorkspaceActivate: (workspace: Workspace) => void
  t: Translate
}

export function WorkspacesTable(props: WorkspacesTableProps) {
  const columns = useWorkspaceColumns(props.language, props.t)
  return (
    <ResourceTable
      data={props.workspaces}
      columns={columns}
      getRowId={(workspace) => workspace.id}
      loading={props.loading}
      loadingContent={props.t('workspaces.loading')}
      emptyContent={props.t('workspaces.empty')}
      sortable
      filtering={{
        mode: 'client',
        searchPlaceholder: props.t('workspaces.filters.searchPlaceholder'),
        ariaLabel: props.t('workspaces.filters.searchPlaceholder'),
      }}
      pagination={{
        mode: 'client',
        defaultPageSize: 20,
        pageSizeOptions: [10, 20, 50, 100],
      }}
      settings={{storageKey: 'm8.resource-manager.workspaces.table-settings'}}
      onRowActivate={props.onWorkspaceActivate}
    />
  )
}
