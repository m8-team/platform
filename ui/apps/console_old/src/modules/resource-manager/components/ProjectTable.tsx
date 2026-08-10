import {useMemo} from 'react'
import {Label, Table, Text} from '@gravity-ui/uikit'
import type {TableColumnConfig} from '@gravity-ui/uikit'

import type {Translate, TranslationKey} from '../../../i18n'
import type {Project, ProjectStatus} from '../model/project'

const themeByStatus: Record<ProjectStatus, 'success' | 'warning' | 'danger' | 'info' | 'normal'> = {
  Active: 'success',
  Suspended: 'warning',
  Failed: 'danger',
  Provisioning: 'info',
  Deleting: 'warning',
}

const statusTitleKey: Record<ProjectStatus, TranslationKey> = {
  Active: 'status.Active',
  Suspended: 'status.Suspended',
  Failed: 'status.Failed',
  Provisioning: 'status.Provisioning',
  Deleting: 'status.Deleting',
}

interface ProjectTableProps {
  projects: Project[]
  selectedProjectId: string
  onSelectProject: (projectId: string) => void
  t: Translate
}

export function ProjectTable({projects, selectedProjectId, onSelectProject, t}: ProjectTableProps) {
  const columns = useMemo<TableColumnConfig<Project>[]>(
    () => [
      {
        id: 'name',
        name: t('projects.column.project'),
        width: 250,
        template: (project) => (
          <div className="m8-project-cell">
            <span
              aria-hidden="true"
              className={`m8-status-dot m8-status-dot_${project.status.toLowerCase()}`}
            />
            <div>
              <Text variant="body-2">{project.name}</Text>
              <Text variant="caption-2" color="secondary">
                {project.lastOperation}
              </Text>
            </div>
          </div>
        ),
      },
      {id: 'projectId', name: t('projects.column.projectId'), width: 180, className: 'm8-mono'},
      {id: 'workspace', name: t('projects.column.workspace'), width: 150, className: 'm8-mono'},
      {id: 'organization', name: t('projects.column.organization'), width: 150, className: 'm8-mono'},
      {
        id: 'status',
        name: t('projects.column.status'),
        width: 130,
        template: (project) => <StatusLabel status={project.status} t={t} />,
      },
      {id: 'desiredState', name: t('projects.column.desiredState'), width: 140},
      {id: 'actualState', name: t('projects.column.actualState'), width: 140},
      {id: 'updated', name: t('projects.column.updated'), width: 150},
      {id: 'owner', name: t('projects.column.owner'), width: 180, className: 'm8-mono'},
    ],
    [t],
  )

  if (projects.length === 0) {
    return (
      <div className="m8-empty-table" role="status">
        <Text variant="body-2">{t('projects.empty')}</Text>
        <Text variant="caption-2" color="secondary">
          {t('projects.emptyDescription')}
        </Text>
      </div>
    )
  }

  return (
    <div className="m8-table-shell">
      <Table
        data={projects}
        columns={columns}
        width="max"
        className="m8-project-table"
        getRowDescriptor={(project) => ({
          id: project.projectId,
          interactive: true,
          classNames:
            project.projectId === selectedProjectId
              ? ['m8-project-table__row_selected']
              : [],
        })}
        onRowClick={(project) => onSelectProject(project.projectId)}
      />
    </div>
  )
}

export function StatusLabel({status, t}: {status: ProjectStatus; t: Translate}) {
  return <Label theme={themeByStatus[status]}>{t(statusTitleKey[status])}</Label>
}
