import {useMemo, useState} from 'react'
import {Magnifier} from '@gravity-ui/icons'
import {Card, Icon, Select, Text, TextInput} from '@gravity-ui/uikit'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import {Metric} from '../../../components/Metric'
import {useConsoleI18n, useConsoleSelection} from '../../../console/ConsoleContext'
import {ProjectTable, StatusLabel} from '../components/ProjectTable'
import {ownerOptionConfigs, statusOptionConfigs, workspaceOptionConfigs} from '../config/projectFilters'
import {translateOptions} from '../lib/translateOptions'
import type {ProjectStatus} from '../model/project'
import {projects} from '../model/projectFixtures'
import {resourceManagerRoutes} from '../routes'

export function ResourceProjectDetailsPage() {
  return <ResourceProjectsPage />
}

export function ResourceProjectsPage() {
  const {organization, workspace, projectId, setWorkspace, setProjectId} = useConsoleSelection()
  const {t} = useConsoleI18n()
  const [status, setStatus] = useState('all')
  const [owner, setOwner] = useState('all')
  const [search, setSearch] = useState('')
  const workspaceOptions = useMemo(() => translateOptions(workspaceOptionConfigs, t), [t])
  const statusOptions = useMemo(() => translateOptions(statusOptionConfigs, t), [t])
  const ownerOptions = useMemo(() => translateOptions(ownerOptionConfigs, t), [t])

  const visibleProjects = useMemo(() => {
    const searchValue = search.trim().toLowerCase()

    return projects.filter((project) => {
      const matchesSearch =
        searchValue.length === 0 ||
        [project.name, project.projectId, project.owner, project.lastOperation].some((value) =>
          value.toLowerCase().includes(searchValue),
        )

      return (
        matchesSearch &&
        project.organization === organization &&
        project.workspace === workspace &&
        (status === 'all' || project.status === status) &&
        (owner === 'all' || project.owner === owner)
      )
    })
  }, [organization, owner, search, status, workspace])

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs
              items={[
                {text: t('breadcrumb.resourceManager'), href: resourceManagerRoutes.overview},
                {text: t('projects.title')},
              ]}
            />
            <Text as="h1" variant="display-1">
              {t('projects.title')}
            </Text>
            <Text as="p" variant="body-2" color="secondary">
              {t('projects.description')}
            </Text>
          </div>

          <div className="m8-summary">
            <Metric label={t('projects.metric.projects')} value="147" description={t('projects.metric.projectsDescription')} />
            <Metric label={t('projects.metric.failed')} value="2" description={t('projects.metric.failedDescription')} tone="danger" />
            <Metric label={t('projects.metric.deleting')} value="4" description={t('projects.metric.deletingDescription')} tone="warning" />
          </div>
        </div>

        <Card view="outlined" type="container" className="m8-filter-card">
          <div className="m8-filters">
            <label className="m8-field">
              <Text variant="caption-2" color="secondary">
                {t('projects.search')}
              </Text>
              <TextInput
                value={search}
                placeholder={t('projects.searchPlaceholder')}
                startContent={<Icon data={Magnifier} size={14} />}
                onUpdate={setSearch}
              />
            </label>
            <Switcher
              label={t('action.workspace')}
              value={[workspace]}
              options={workspaceOptions}
              onUpdate={(next) => {
                const nextWorkspace = next[0] ?? workspace
                setWorkspace(nextWorkspace)
                const nextProject = projects.find(
                  (project) => project.organization === organization && project.workspace === nextWorkspace,
                )
                if (nextProject) {
                  setProjectId(nextProject.projectId)
                }
              }}
            />
            <Switcher
              label={t('projects.column.status')}
              value={[status]}
              options={statusOptions}
              onUpdate={(next) => setStatus(next[0] ?? status)}
            />
            <Switcher
              label={t('projects.column.owner')}
              value={[owner]}
              options={ownerOptions}
              onUpdate={(next) => setOwner(next[0] ?? owner)}
            />
          </div>
        </Card>

        <div className="m8-workspace">
          <Card view="outlined" type="container" className="m8-table-card">
            <div className="m8-card-header">
              <div>
                <Text as="h2" variant="header-1">
                  {t('projects.inventory')}
                </Text>
                <Text variant="caption-2" color="secondary">
                  {t('projects.inventoryDescription')}
                </Text>
              </div>
              <div className="m8-labels">
                {statusOptions.slice(1).map((option) => (
                  <StatusLabel key={option.value} status={option.value as ProjectStatus} t={t} />
                ))}
              </div>
            </div>

            <ProjectTable
              projects={visibleProjects}
              selectedProjectId={projectId}
              onSelectProject={setProjectId}
              t={t}
            />
          </Card>
        </div>
      </section>
    </main>
  )
}

interface SwitcherProps {
  label: string
  value: string[]
  options: Array<{value: string; content: string}>
  onUpdate: (value: string[]) => void
}

function Switcher({label, value, options, onUpdate}: SwitcherProps) {
  return (
    <div className="m8-field m8-switcher">
      <Text variant="caption-2" color="secondary">
        {label}
      </Text>
      <Select aria-label={label} value={value} options={options} width="max" onUpdate={onUpdate} />
    </div>
  )
}
