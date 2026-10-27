import {useMemo, useState} from 'react'
import {ArrowRotateRight, ChevronDown, ChevronRight, Magnifier} from '@gravity-ui/icons'
import {Button, Card, ClipboardButton, Icon, Label, Text, TextInput} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import {useParams} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {AppLanguage, Translate} from '../../../i18n'
import {
  fetchOrganizationDetail,
  fetchOrganizationResources,
  fetchOrganizationStructure,
} from '../api/organizationDetail'
import type {OrganizationDetail, OrganizationResource, OrganizationStructure, OrgUnit} from '../api/organizationDetail'
import {OrganizationSettings} from '../components/OrganizationSettings'
import {OrgUnitChangeSetEditor} from '../components/OrgUnitChangeSetEditor'
import {ResourceOwnershipCards} from '../components/ResourceOwnershipCards'
import './organization-details.css'
import './organization-editor.css'

type Section = 'overview' | 'units' | 'changes' | 'resources' | 'ownership' | 'settings' | 'sync' | 'history'
interface Props {language: AppLanguage; t: Translate}
type DictionaryKey =
  | 'overview' | 'units' | 'changes' | 'resources' | 'ownership' | 'settings' | 'sync' | 'history'
  | 'refresh' | 'loading' | 'retry' | 'notAvailable' | 'empty' | 'unavailable'
  | 'structureRevision' | 'tenant' | 'organization' | 'resourceCount' | 'orgUnitCount'
  | 'searchUnits' | 'unitDetails' | 'state' | 'type' | 'source' | 'parent'
  | 'created' | 'updated' | 'management' | 'root' | 'unlinked'
  | 'collapseAll' | 'expandAll' | 'noUnits' | 'noMatches' | 'selectedUnit'
  | 'name' | 'id' | 'apiMissing' | 'ownershipMessage' | 'syncMessage'
  | 'historyMessage' | 'overviewHint' | 'resourceScopeHint' | 'accessDenied'
  | 'code' | 'noSource' | 'all' | 'initialRoot' | 'loadError'

const dictionary: Record<AppLanguage, Record<DictionaryKey, string>> = {
  ru: {
    overview: 'Обзор', units: 'Подразделения', changes: 'Изменения структуры', settings: 'Настройки', resources: 'Ресурсы',
    ownership: 'Ответственность', sync: 'Источники данных', history: 'История',
    refresh: 'Обновить', loading: 'Загрузка…', retry: 'Повторить',
    notAvailable: 'Недоступно', empty: 'Нет данных', unavailable: 'Информация пока недоступна',
    structureRevision: 'Ревизия структуры', tenant: 'Tenant', organization: 'Организация',
    resourceCount: 'Ресурсов в организации', orgUnitCount: 'Подразделений',
    searchUnits: 'Поиск подразделений', unitDetails: 'Сведения о подразделении',
    state: 'Состояние', type: 'Тип', source: 'Источник', parent: 'Родитель',
    created: 'Создана', updated: 'Обновлена', management: 'Управление',
    root: 'Корневой уровень', unlinked: 'Видимые подразделения без доступного родителя',
    collapseAll: 'Свернуть всё', expandAll: 'Развернуть всё',
    noUnits: 'В организации пока нет подразделений', noMatches: 'Ничего не найдено',
    selectedUnit: 'Выберите подразделение для просмотра его свойств',
    name: 'Название', id: 'Идентификатор', apiMissing: 'API ещё не подключён',
    ownershipMessage: 'Назначения владельцев ресурсов и ответственных ещё не доступны через API Resource Manager. Права не выводятся из положения подразделения в дереве.',
    syncMessage: 'Directory Sync и подключение 1С:ЗУП будут добавлены после реализации реестра источников и журнала синхронизаций. Этот экран не показывает вымышленные подключения.',
    historyMessage: 'Исторический API OrgUnitVersion пока возвращает UNIMPLEMENTED. Текущая структура доступна во вкладке «Подразделения».',
    overviewHint: 'Фактическое состояние Resource Manager. Сотрудники и членства хранятся отдельно в M8 Identity/Directory.',
    resourceScopeHint: 'Список ресурсов текущей организации. Владельцы и разрешения M8 Access здесь не рассчитываются.',
    accessDenied: 'Если API вернул HTTP 403, проверьте авторизацию M8 Access.',
    code: 'Код', noSource: 'Не указан', all: 'Все подразделения',
    initialRoot: 'У организации нет обязательного искусственного корневого подразделения.',
    loadError: 'Ошибка загрузки данных',
  },
  en: {
    overview: 'Overview', units: 'Org units', changes: 'ChangeSets', settings: 'Settings', resources: 'Resources',
    ownership: 'Ownership', sync: 'Data sources', history: 'History',
    refresh: 'Refresh', loading: 'Loading…', retry: 'Retry',
    notAvailable: 'Unavailable', empty: 'No data', unavailable: 'Data is not available yet',
    structureRevision: 'Structure revision', tenant: 'Tenant', organization: 'Organization',
    resourceCount: 'Organization resources', orgUnitCount: 'Org units',
    searchUnits: 'Search org units', unitDetails: 'Org unit details',
    state: 'Lifecycle', type: 'Type', source: 'Source', parent: 'Parent',
    created: 'Created', updated: 'Updated', management: 'Managed by',
    root: 'Root level', unlinked: 'Visible units with an inaccessible parent',
    collapseAll: 'Collapse all', expandAll: 'Expand all',
    noUnits: 'No org units yet', noMatches: 'No matching org units',
    selectedUnit: 'Select an org unit to view its properties',
    name: 'Name', id: 'Identifier', apiMissing: 'API not connected',
    ownershipMessage: 'Resource ownership and organizational accountability are not yet served by the Resource Manager API. Permissions are not derived from a position in the organizational tree.',
    syncMessage: 'Directory Sync and 1C:ZUP connections require the source registry and synchronization journal. No simulated connections are shown.',
    historyMessage: 'The OrgUnitVersion history endpoint currently returns UNIMPLEMENTED. The active hierarchy is available under Org units.',
    overviewHint: 'Live Resource Manager state. Employees and memberships are managed separately in M8 Identity/Directory.',
    resourceScopeHint: 'Organization-scoped resource list. M8 Access grants and owners are not inferred.',
    accessDenied: 'For HTTP 403 responses, check M8 Access authorization.',
    code: 'Code', noSource: 'Not specified', all: 'All org units',
    initialRoot: 'No artificial root OrgUnit is required for an Organization.',
    loadError: 'Could not load data',
  },
}

function formatDate(value: string | undefined, language: AppLanguage) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat(language === 'ru' ? 'ru-RU' : 'en-US', {
    dateStyle: 'medium', timeStyle: 'short',
  }).format(date)
}

function Lifecycle({value}: {value: string}) {
  const theme = value === 'ACTIVE' ? 'success' : value === 'CLOSED' || value === 'SUSPENDED'
    ? 'warning' : value === 'FAILED' || value === 'DELETED' ? 'danger' : 'normal'
  return <Label theme={theme}>{value}</Label>
}

function RequestState({error, onRetry, label}: {error: unknown; onRetry: () => void; label: string}) {
  return (
    <div className="m8-org-detail__error" role="alert">
      <Text variant="body-2">{label}</Text>
      <Text variant="caption-2" color="secondary">
        {error instanceof Error ? error.message : String(error)}
      </Text>
      <Button view="outlined" onClick={onRetry}>Retry / Повторить</Button>
    </div>
  )
}

interface Branch {unit: OrgUnit; children: Branch[]}
function makeTree(units: OrgUnit[]): Branch[] {
  const nodes = new Map(units.map((unit) => [unit.id, {unit, children: [] as Branch[]}]))
  const roots: Branch[] = []
  for (const unit of units) {
    const node = nodes.get(unit.id)!
    const parent = nodes.get(unit.parentId)
    // Inaccessible parents are not assumed to exist or be authorized.
    if (!parent || parent === node) roots.push(node)
    else parent.children.push(node)
  }
  const byName = (a: Branch, b: Branch) => a.unit.name.localeCompare(b.unit.name)
  for (const node of nodes.values()) node.children.sort(byName)
  return roots.sort(byName)
}

function TreeRow({
  node, depth, selectedId, expanded, onSelect, onToggle, searchMatched, ancestors,
}: {
  node: Branch; depth: number; selectedId: string; expanded: Set<string>;
  onSelect: (id: string) => void; onToggle: (id: string) => void;
  searchMatched?: Set<string>; ancestors: Set<string>;
}) {
  if (ancestors.has(node.unit.id)) return null
  const nextAncestors = new Set(ancestors)
  nextAncestors.add(node.unit.id)
  const visible = !searchMatched || searchMatched.has(node.unit.id)
  if (!visible) return null
  const isExpanded = expanded.has(node.unit.id) || Boolean(searchMatched)
  const children = node.children.filter((child) => !searchMatched || searchMatched.has(child.unit.id))
  return (
    <>
      <div className={`m8-org-tree__row ${selectedId === node.unit.id ? 'm8-org-tree__row_selected' : ''}`}
        style={{paddingInlineStart: `${depth * 19 + 6}px`}}>
        {children.length ? (
          <button type="button" className="m8-org-tree__expander" aria-label={isExpanded ? 'Collapse' : 'Expand'}
            aria-expanded={isExpanded} onClick={() => onToggle(node.unit.id)}>
            <Icon data={isExpanded ? ChevronDown : ChevronRight} size={14} />
          </button>
        ) : <span className="m8-org-tree__expander" aria-hidden="true" />}
        <button className="m8-org-tree__item" type="button"
          aria-current={selectedId === node.unit.id ? 'true' : undefined}
          onClick={() => onSelect(node.unit.id)}>
          <span className="m8-org-tree__name">{node.unit.name}</span>
          <span className="m8-org-tree__status">{node.unit.state}</span>
        </button>
      </div>
      {isExpanded && children.map((child) => (
        <TreeRow key={child.unit.id} node={child} depth={depth + 1}
          selectedId={selectedId} expanded={expanded}
          onSelect={onSelect} onToggle={onToggle}
          searchMatched={searchMatched} ancestors={nextAncestors} />
      ))}
    </>
  )
}

function OrgUnitExplorer({data, language}: {data: OrganizationStructure; language: AppLanguage}) {
  const d = dictionary[language]
  const [search, setSearch] = useState('')
  const [selectedId, setSelectedId] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set())
  const selected = data.units.find((unit) => unit.id === selectedId)
  const tree = useMemo(() => makeTree(data.units), [data.units])
  const filter = search.trim().toLocaleLowerCase()
  const matched = useMemo(() => {
    if (!filter) return undefined
    const map = new Map(data.units.map((u) => [u.id, u]))
    const visible = new Set<string>()
    for (const unit of data.units) {
      if (![unit.name, unit.code ?? '', unit.id].some((s) => s.toLocaleLowerCase().includes(filter))) continue
      let current: OrgUnit | undefined = unit
      const visited = new Set<string>()
      while (current && !visited.has(current.id)) {
        visited.add(current.id)
        visible.add(current.id)
        current = map.get(current.parentId)
      }
    }
    return visible
  }, [data.units, filter])
  const toggle = (id: string) => setExpanded((value) => {
    const next = new Set(value)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    return next
  })
  return (
    <div className="m8-org-explorer">
      <section className="m8-org-explorer__tree" aria-label={d.units}>
        <div className="m8-org-explorer__toolbar">
          <TextInput size="m" placeholder={d.searchUnits} aria-label={d.searchUnits} value={search}
            startContent={<Icon data={Magnifier} size={15}/>} onUpdate={setSearch}/>
          <div className="m8-org-explorer__controls">
            <Button view="flat" size="s" onClick={() => setExpanded(new Set(data.units.map((u) => u.id)))}>
              {d.expandAll}
            </Button>
            <Button view="flat" size="s" onClick={() => setExpanded(new Set())}>{d.collapseAll}</Button>
          </div>
        </div>
        <div className="m8-org-explorer__nodes">
          {data.units.length === 0 ? <Text color="secondary">{d.noUnits}</Text> : null}
          {matched && matched.size === 0 ? <Text color="secondary">{d.noMatches}</Text> : null}
          {tree.map((node) => (
            <TreeRow key={node.unit.id} node={node} depth={0} selectedId={selectedId}
              expanded={expanded} onSelect={setSelectedId} onToggle={toggle}
              searchMatched={matched} ancestors={new Set()} />
          ))}
        </div>
      </section>
      <section className="m8-org-explorer__inspector" aria-label={d.unitDetails}>
        <Text as="h3" variant="header-1">{d.unitDetails}</Text>
        {selected ? (
          <>
            <div className="m8-org-detail__unit-heading">
              <Text as="h4" variant="header-2">{selected.name}</Text>
              <Lifecycle value={selected.state}/>
            </div>
            <dl className="m8-org-detail__definitions">
              <dt>{d.id}</dt><dd className="m8-org-detail__value-copy">
                <span className="m8-mono">{selected.id}</span><ClipboardButton text={selected.id} size="s"/>
              </dd>
              <dt>{d.code}</dt><dd>{selected.code ?? '—'}</dd>
              <dt>{d.parent}</dt><dd className="m8-mono">{selected.parentId || d.root}</dd>
              <dt>{d.type}</dt><dd>{selected.kind ?? '—'}</dd>
              <dt>{d.management}</dt><dd>{selected.managementMode}</dd>
              <dt>{d.source}</dt><dd>{selected.authoritativeSourceId ?? d.noSource}</dd>
              <dt>{d.structureRevision}</dt><dd>{selected.structureRevision}</dd>
            </dl>
          </>
        ) : <Text color="secondary">{d.selectedUnit}</Text>}
        <Text variant="caption-2" color="secondary">{d.initialRoot}</Text>
      </section>
    </div>
  )
}

function ResourceInventory({resources, language}: {resources: OrganizationResource[]; language: AppLanguage}) {
  const d = dictionary[language]
  return (
    <>
      <Text variant="body-2" color="secondary">{d.resourceScopeHint}</Text>
      {resources.length === 0 ? <div className="m8-org-detail__empty">{d.empty}</div> : (
        <div className="m8-org-detail__table-wrap">
          <table className="m8-org-detail__table">
            <thead><tr><th>{d.name}</th><th>{d.type}</th><th>{d.state}</th><th>{d.id}</th></tr></thead>
            <tbody>{resources.map((r) => (
              <tr key={`${r.kind}:${r.id}`}>
                <td>{r.name}</td><td>{r.kind}</td><td><Lifecycle value={r.state}/></td>
                <td className="m8-mono">{r.id}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </>
  )
}

function UnsupportedFeature({message, language}: {message: string; language: AppLanguage}) {
  const d = dictionary[language]
  return <div className="m8-org-detail__unsupported">
    <Label theme="warning">{d.apiMissing}</Label>
    <Text variant="body-2">{message}</Text>
  </div>
}

function OrganizationDetailContent({organizationId, language, t}: Props & {organizationId: string}) {
  const d = dictionary[language]
  const [section, setSection] = useState<Section>('overview')
  const organization = useQuery({
    queryKey: ['resource-manager', 'organization', organizationId],
    queryFn: ({signal}) => fetchOrganizationDetail(organizationId, signal),
    staleTime: 30_000,
  })
  const structure = useQuery({
    queryKey: ['resource-manager', 'organization', organizationId, 'structure'],
    queryFn: ({signal}) => fetchOrganizationStructure(organizationId, signal),
    staleTime: 30_000,
    enabled: section === 'units' || section === 'changes' || section === 'overview',
  })
  const resources = useQuery({
    queryKey: ['resource-manager', 'organization', organizationId, 'resources'],
    queryFn: ({signal}) => fetchOrganizationResources(organizationId, signal),
    staleTime: 30_000,
    enabled: section === 'resources' || section === 'ownership' || section === 'overview',
  })
  const tabs: Section[] = ['overview', 'units', 'changes', 'resources', 'ownership', 'settings', 'sync', 'history']
  const current: OrganizationDetail | undefined = organization.data
  const refresh = () => {
    void organization.refetch()
    if (structure.isFetched) void structure.refetch()
    if (resources.isFetched) void resources.refetch()
  }

  return (
    <main className="m8-page__body">
      <section className="m8-page__content m8-org-detail">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs items={[
              {text: t('breadcrumb.resourceManager'), href: '/resource-manager'},
              {text: t('menu.resources.organizations'), href: '/resource-manager/organizations'},
              {text: current?.name ?? organizationId},
            ]}/>
            <Text as="h1" variant="display-1">{current?.name ?? d.organization}</Text>
            <div className="m8-org-detail__heading-info">
              {current ? <Lifecycle value={current.state}/> : null}
              <span className="m8-mono">{organizationId}</span>
              <ClipboardButton text={organizationId} size="s"/>
            </div>
          </div>
          <Button view="outlined" loading={organization.isFetching || structure.isFetching || resources.isFetching}
            onClick={refresh}>
            <Icon data={ArrowRotateRight} size={16}/>{d.refresh}
          </Button>
        </div>

        {organization.isPending ? <Card view="outlined" className="m8-org-detail__card">
          <Text color="secondary">{d.loading}</Text></Card> : null}
        {organization.isError ? <Card view="outlined" className="m8-org-detail__card">
          <RequestState label={d.loadError} error={organization.error}
            onRetry={() => void organization.refetch()}/></Card> : null}

        {current ? (
          <>
            <div className="m8-org-detail__summary">
              <Card view="outlined" className="m8-org-detail__metric">
                <Text variant="caption-2" color="secondary">{d.orgUnitCount}</Text>
                <Text variant="header-2">{structure.data ? structure.data.units.length : '—'}</Text>
              </Card>
              <Card view="outlined" className="m8-org-detail__metric">
                <Text variant="caption-2" color="secondary">{d.resourceCount}</Text>
                <Text variant="header-2">{resources.data ? resources.data.length : '—'}</Text>
              </Card>
              <Card view="outlined" className="m8-org-detail__metric">
                <Text variant="caption-2" color="secondary">{d.structureRevision}</Text>
                <Text variant="header-2">{structure.data?.revision ?? '—'}</Text>
              </Card>
            </div>
            <div className="m8-org-detail__tabs" role="tablist" aria-label={d.organization}>
              {tabs.map((tab) => (
                <button type="button" key={tab} id={`org-tab-${tab}`} role="tab"
                  aria-selected={section === tab} aria-controls="org-tab-content"
                  className={`m8-org-detail__tab ${section === tab ? 'm8-org-detail__tab_active' : ''}`}
                  onClick={() => setSection(tab)}>{d[tab]}</button>
              ))}
            </div>
            <Card view="outlined" type="container" className="m8-org-detail__card">
              <div id="org-tab-content" role="tabpanel" aria-labelledby={`org-tab-${section}`}>
                {section === 'overview' ? (
                  <div className="m8-org-detail__overview">
                    <div>
                      <Text as="h2" variant="header-1">{d.overview}</Text>
                      <Text variant="body-2" color="secondary">{d.overviewHint}</Text>
                    </div>
                    <dl className="m8-org-detail__definitions">
                      <dt>{d.organization}</dt><dd>{current.name}</dd>
                      <dt>{d.tenant}</dt><dd className="m8-mono">{current.tenantId}</dd>
                      <dt>{d.state}</dt><dd><Lifecycle value={current.state}/></dd>
                      <dt>{d.created}</dt><dd>{formatDate(current.createdAt, language)}</dd>
                      <dt>{d.updated}</dt><dd>{formatDate(current.updatedAt, language)}</dd>
                    </dl>
                    {(structure.isError || resources.isError) ? (
                      <Text color="secondary">{d.accessDenied}</Text>
                    ) : null}
                  </div>
                ) : null}
                {section === 'units' ? (
                  structure.isPending ? <Text color="secondary">{d.loading}</Text> :
                    structure.isError ? <RequestState label={d.loadError} error={structure.error}
                      onRetry={() => void structure.refetch()}/> :
                      structure.data ? <OrgUnitExplorer data={structure.data} language={language}/> : null
                ) : null}
                {section === 'resources' ? (
                  resources.isPending ? <Text color="secondary">{d.loading}</Text> :
                    resources.isError ? <RequestState label={d.loadError} error={resources.error}
                      onRetry={() => void resources.refetch()}/> :
                      resources.data ? <ResourceInventory resources={resources.data} language={language}/> : null
                ) : null}
                {section === 'changes' ? (
                  structure.isPending ? <Text color="secondary">{d.loading}</Text> :
                    structure.isError ? <RequestState label={d.loadError} error={structure.error}
                      onRetry={() => void structure.refetch()}/> :
                      structure.data ? <OrgUnitChangeSetEditor key={structure.data.revision}
                        structure={structure.data} organizationId={organizationId} language={language}/> : null
                ) : null}
                {section === 'ownership' ? (
                  resources.isPending ? <Text color="secondary">{d.loading}</Text> :
                    resources.isError ? <RequestState label={d.loadError} error={resources.error}
                      onRetry={() => void resources.refetch()}/> :
                      resources.data ? <ResourceOwnershipCards resources={resources.data}
                        organizationId={organizationId} language={language}/> : null
                ) : null}
                {section === 'settings' ? <OrganizationSettings key={`${current.id}:${current.revision ?? 0}`} organization={current} language={language}/> : null}
                {section === 'sync' ? <UnsupportedFeature language={language} message={d.syncMessage}/> : null}
                {section === 'history' ? <UnsupportedFeature language={language} message={d.historyMessage}/> : null}
              </div>
            </Card>
          </>
        ) : null}
      </section>
    </main>
  )
}

export function OrganizationDetailsPage({language, t}: Props) {
  const {organizationId} = useParams({from: '/resource-manager/organizations/$organizationId'})
  return <OrganizationDetailContent key={organizationId} organizationId={organizationId}
    language={language} t={t}/>
}
