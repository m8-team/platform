import {useMemo, useState} from 'react'
import {ArrowRotateRight, Magnifier} from '@gravity-ui/icons'
import {Button, Card, ClipboardButton, Icon, Label, Select, Text, TextInput} from '@gravity-ui/uikit'
import {useQuery} from '@tanstack/react-query'
import {useParams, useRouter} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {Translate} from '../../../i18n'
import {
  getCatalogResource,
  listCatalogResources,
  resolveCatalogResourceScope,
} from '../api/resourceCatalog'
import type {CatalogResource} from '../api/resourceCatalog'
import './resource-catalog.css'

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
const savedOrganizationKey = 'm8.resource-manager.catalog.organization-id'
type InspectorTab = 'overview' | 'ownership' | 'dependencies'

function initialOrganization() {
  if (typeof window === 'undefined') return ''
  const fromLink = new URLSearchParams(window.location.search).get('organizationId') || ''
  if (uuidPattern.test(fromLink)) return fromLink
  try {
    const saved = window.sessionStorage.getItem(savedOrganizationKey) || ''
    if (uuidPattern.test(saved)) return saved
  } catch {
    // Do not assume storage access.
  }
  return ''
}

function stateTheme(state: string): 'success' | 'warning' | 'danger' | 'info' | 'normal' {
  if (state === 'ACTIVE') return 'success'
  if (state === 'SUSPENDED' || state === 'ARCHIVED') return 'warning'
  if (state === 'DELETED' || state === 'FAILED') return 'danger'
  if (state === 'CREATING') return 'info'
  return 'normal'
}

function ResourceState({state}: {state: string}) {
  return <Label theme={stateTheme(state)}>{state}</Label>
}

function ErrorMessage({error, t, retry}: {error: Error | null; t: Translate; retry: () => void}) {
  return (
    <Card view="outlined" className="m8-catalog__message">
      <Text color="danger" variant="body-2">{t('catalog.error')}</Text>
      <Text variant="body-2" color="secondary">{error?.message ?? t('catalog.unknownError')}</Text>
      <Button view="outlined" onClick={retry}>{t('catalog.retry')}</Button>
    </Card>
  )
}

function InfoField({label, value}: {label: string; value: string | number | undefined}) {
  return (
    <div className="m8-catalog__field">
      <Text variant="caption-2" color="secondary">{label}</Text>
      <Text variant="body-2" className="m8-catalog__break">{value ?? '—'}</Text>
    </div>
  )
}

function formatTime(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function ResourceInspector({resource, t}: {resource: CatalogResource; t: Translate}) {
  const [tab, setTab] = useState<InspectorTab>('overview')
  const scope = useQuery({
    queryKey: ['resource-manager', 'resource-catalog', 'scope', resource.kind, resource.id],
    queryFn: ({signal}) => resolveCatalogResourceScope(resource.kind, resource.id, signal),
    retry: false,
    enabled: tab === 'ownership' || tab === 'dependencies',
  })
  const tabs: InspectorTab[] = ['overview', 'ownership', 'dependencies']
  return (
    <section className="m8-catalog__inspector" aria-label={t('catalog.card')}>
      <div className="m8-catalog__title">
        <div>
          <Text variant="caption-2" color="secondary">{resource.kind}{resource.module ? ' · ' + resource.module : ''}</Text>
          <Text as="h2" variant="header-2">{resource.name}</Text>
        </div>
        <ResourceState state={resource.state} />
      </div>
      <div className="m8-catalog__identity">
        <Text variant="caption-2" color="secondary">ID</Text>
        <Text variant="body-2" className="m8-catalog__break">{resource.id}</Text>
        <ClipboardButton text={resource.id} size="s"/>
      </div>
      <div className="m8-catalog__tabs" role="tablist">
        {tabs.map((current) => (
          <Button key={current} view={tab === current ? 'action' : 'flat'}
            onClick={() => setTab(current)}>{t(('catalog.tab.' + current) as 'catalog.tab.overview' | 'catalog.tab.ownership' | 'catalog.tab.dependencies')}</Button>
        ))}
      </div>
      {tab === 'overview' && (
        <div className="m8-catalog__section">
          <Text variant="body-2">{resource.description || t('catalog.noDescription')}</Text>
          <div className="m8-catalog__properties">
            <InfoField label={t('catalog.kind')} value={resource.kind}/>
            <InfoField label={t('catalog.module')} value={resource.module}/>
            <InfoField label={t('catalog.revision')} value={resource.revision}/>
            <InfoField label={t('catalog.created')} value={formatTime(resource.createdAt)}/>
            <InfoField label={t('catalog.updated')} value={formatTime(resource.updatedAt)}/>
          </div>
          <Text as="h3" variant="subheader-2">{t('catalog.labels')}</Text>
          {Object.keys(resource.labels).length ? (
            <div className="m8-catalog__labels">
              {Object.entries(resource.labels).map(([key, value]) => (
                <Label key={key} theme="normal">{key + ': ' + value}</Label>
              ))}
            </div>
          ) : <Text color="secondary" variant="body-2">{t('catalog.noLabels')}</Text>}
        </div>
      )}
      {tab === 'ownership' && (
        <div className="m8-catalog__section">
          <Text as="h3" variant="subheader-2">{t('catalog.structuralOwnership')}</Text>
          <div className="m8-catalog__properties">
            <InfoField label={t('catalog.organization')} value={scope.data?.owningOrganizationId || resource.owningOrganizationId}/>
            <InfoField label={t('catalog.parent')} value={resource.owner ? resource.owner.kind + ': ' + resource.owner.id : undefined}/>
          </div>
          {scope.isPending && scope.fetchStatus === 'fetching' && <Text color="secondary">{t('catalog.loading')}</Text>}
          {scope.isError && <Text color="secondary">{t('catalog.scopeUnavailable')}</Text>}
          <Text as="h3" variant="subheader-2">{t('catalog.accountableOwners')}</Text>
          <div className="m8-catalog__properties">
            <InfoField label={t('catalog.businessOwner')} value={t('catalog.notAvailable')}/>
            <InfoField label={t('catalog.technicalOwner')} value={t('catalog.notAvailable')}/>
            <InfoField label={t('catalog.operationsTeam')} value={t('catalog.notAvailable')}/>
            <InfoField label={t('catalog.accountableUnit')} value={t('catalog.notAvailable')}/>
          </div>
          <Text color="secondary" variant="body-2">{t('catalog.ownerApiNote')}</Text>
          <Button disabled>{t('catalog.assignOwner')}</Button>
        </div>
      )}
      {tab === 'dependencies' && (
        <div className="m8-catalog__section">
          <Text as="h3" variant="subheader-2">{t('catalog.containment')}</Text>
          {scope.isPending && scope.fetchStatus === 'fetching' ? <Text color="secondary">{t('catalog.loading')}</Text> : null}
          {scope.isError ? <Text color="secondary">{t('catalog.scopeUnavailable')}</Text> : null}
          {scope.data && scope.data.ancestry.length ? (
            <ol className="m8-catalog__ancestry">
              {scope.data.ancestry.map((node, i) => (
                <li key={node.kind + ':' + node.id + ':' + i}>
                  <Text variant="body-2">{node.kind}</Text>
                  <Text color="secondary" variant="caption-2" className="m8-catalog__break">{node.id}</Text>
                </li>
              ))}
            </ol>
          ) : scope.isSuccess ? <Text color="secondary">{t('catalog.noAncestry')}</Text> : null}
          <Text color="secondary" variant="body-2">{t('catalog.notDependency')}</Text>
          <Text as="h3" variant="subheader-2">{t('catalog.functionalDependencies')}</Text>
          <Text color="secondary" variant="body-2">{t('catalog.dependencyApiNote')}</Text>
          <Button disabled>{t('catalog.addDependency')}</Button>
        </div>
      )}
    </section>
  )
}

export function ResourceCatalogPage({t}: {t: Translate}) {
  const router = useRouter()
  const params = useParams({strict: false}) as {kind?: string; resourceId?: string}
  const [organizationId, setOrganizationId] = useState(initialOrganization)
  const [organizationInput, setOrganizationInput] = useState(initialOrganization)
  const [pageTokens, setPageTokens] = useState<string[]>([''])
  const [page, setPage] = useState(0)
  const [search, setSearch] = useState('')
  const [kindFilter, setKindFilter] = useState('')
  const [stateFilter, setStateFilter] = useState('')
  const currentPageToken = pageTokens[page] ?? ''
  const catalog = useQuery({
    queryKey: ['resource-manager', 'resource-catalog', organizationId, currentPageToken],
    queryFn: ({signal}) => listCatalogResources(organizationId, currentPageToken, signal),
    enabled: Boolean(organizationId),
    retry: false,
    staleTime: 30_000,
  })
  const selected = useQuery({
    queryKey: ['resource-manager', 'resource-catalog', 'detail', params.kind, params.resourceId],
    queryFn: ({signal}) => getCatalogResource(params.kind!, params.resourceId!, signal),
    enabled: Boolean(params.kind && params.resourceId),
    retry: false,
  })
  const rows = useMemo(() => catalog.data?.resources ?? [], [catalog.data?.resources])
  const kinds = useMemo(() => [...new Set(rows.map((r) => r.kind))].sort(), [rows])
  const states = useMemo(() => [...new Set(rows.map((r) => r.state))].sort(), [rows])
  const visible = useMemo(() => rows.filter((r) => {
    const q = search.trim().toLocaleLowerCase()
    return (!kindFilter || r.kind === kindFilter) &&
      (!stateFilter || r.state === stateFilter) &&
      (!q || [r.name, r.id, r.kind, r.module ?? ''].some((v) => v.toLocaleLowerCase().includes(q)))
  }), [rows, search, kindFilter, stateFilter])

  const selectOrganization = () => {
    const id = organizationInput.trim()
    if (!uuidPattern.test(id)) return
    setOrganizationId(id)
    setPageTokens([''])
    setPage(0)
    try {sessionStorage.setItem(savedOrganizationKey, id)} catch { /* optional */ }
  }
  const openResource = (resource: CatalogResource) => {
    void router.navigate({
      to: '/resource-manager/catalog/$kind/$resourceId',
      params: {kind: resource.kind, resourceId: resource.id},
    })
  }
  const showNext = () => {
    const next = catalog.data?.nextPageToken
    if (!next || next === currentPageToken || pageTokens.includes(next)) return
    setPageTokens((existing) => [...existing, next])
    setPage((current) => current + 1)
  }

  return (
    <main className="m8-page__body">
      <section className="m8-page__content">
        <div className="m8-page__heading">
          <div>
            <ConsoleBreadcrumbs items={[
              {text: t('breadcrumb.resourceManager'), href: '/resource-manager'},
              {text: t('menu.resources.catalog')},
            ]}/>
            <Text as="h1" variant="display-1">{t('catalog.title')}</Text>
            <Text as="p" variant="body-2" color="secondary">{t('catalog.description')}</Text>
          </div>
          <Button view="outlined" disabled={!organizationId} onClick={() => void catalog.refetch()}>
            <Icon data={ArrowRotateRight}/> {t('catalog.refresh')}
          </Button>
        </div>
        <Card view="outlined" className="m8-catalog__scope">
          <div className="m8-catalog__scope-form">
            <div className="m8-catalog__scope-input">
              <Text variant="subheader-2">{t('catalog.organization')}</Text>
              <TextInput value={organizationInput} onUpdate={setOrganizationInput}
                placeholder={t('catalog.organizationPlaceholder')} />
            </div>
            <Button view="action" disabled={!uuidPattern.test(organizationInput.trim())}
              onClick={selectOrganization}>{t('catalog.applyScope')}</Button>
          </div>
          <Text color="secondary" variant="caption-2">{t('catalog.scopeNote')}</Text>
        </Card>
        {organizationId ? (
          <>
            <div className="m8-catalog__stats">
              <Card view="outlined" className="m8-catalog__stat">
                <Text color="secondary" variant="caption-2">{t('catalog.loaded')}</Text>
                <Text variant="header-2">{rows.length}</Text>
              </Card>
              <Card view="outlined" className="m8-catalog__stat">
                <Text color="secondary" variant="caption-2">{t('catalog.types')}</Text>
                <Text variant="header-2">{kinds.length}</Text>
              </Card>
              <Card view="outlined" className="m8-catalog__stat">
                <Text color="secondary" variant="caption-2">{t('catalog.active')}</Text>
                <Text variant="header-2">{rows.filter((r) => r.state === 'ACTIVE').length}</Text>
              </Card>
            </div>
            <div className="m8-catalog__grid">
              <Card view="outlined" className="m8-catalog__inventory">
                <div className="m8-catalog__toolbar">
                  <Text as="h2" variant="header-2">{t('catalog.inventory')}</Text>
                  <TextInput value={search} onUpdate={setSearch}
                    startContent={<Icon data={Magnifier}/>} placeholder={t('catalog.search')}/>
                  <Select width="max" value={[kindFilter]}
                    options={[{value: '', content: t('catalog.allKinds')}, ...kinds.map((k) => ({value: k, content: k}))]}
                    onUpdate={(v) => setKindFilter(v[0] ?? '')}/>
                  <Select width="max" value={[stateFilter]}
                    options={[{value: '', content: t('catalog.allStates')}, ...states.map((s) => ({value: s, content: s}))]}
                    onUpdate={(v) => setStateFilter(v[0] ?? '')}/>
                </div>
                {catalog.isPending && <Text color="secondary">{t('catalog.loading')}</Text>}
                {catalog.isError && <ErrorMessage error={catalog.error} t={t}
                  retry={() => void catalog.refetch()}/>}
                {catalog.isSuccess && (
                  <>
                    <div className="m8-catalog__table-wrap">
                      <table className="m8-catalog__table">
                        <thead><tr>
                          <th>{t('catalog.name')}</th><th>{t('catalog.kind')}</th>
                          <th>{t('catalog.state')}</th><th>{t('catalog.module')}</th>
                        </tr></thead>
                        <tbody>
                          {visible.map((r) => (
                            <tr key={r.kind + ':' + r.id}>
                              <td><Button view="flat" className="m8-catalog__resource-name"
                                onClick={() => openResource(r)}>{r.name}</Button>
                                <div className="m8-catalog__sub-id">{r.id}</div>
                              </td>
                              <td>{r.kind}</td><td><ResourceState state={r.state}/></td>
                              <td>{r.module ?? '—'}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    {!visible.length && <Text color="secondary">{t('catalog.empty')}</Text>}
                    <div className="m8-catalog__pagination">
                      <Text color="secondary" variant="caption-2">{t('catalog.page')} {page + 1} · {t('catalog.pageOnly')}</Text>
                      <div className="m8-catalog__pagination-actions">
                        <Button disabled={page === 0} onClick={() => setPage((p) => p - 1)}>{t('catalog.previous')}</Button>
                        <Button disabled={!catalog.data.nextPageToken ||
                          pageTokens.includes(catalog.data.nextPageToken)}
                          onClick={showNext}>{t('catalog.next')}</Button>
                      </div>
                    </div>
                  </>
                )}
              </Card>
              <Card view="outlined" className="m8-catalog__detail">
                {params.kind && params.resourceId ? (
                  <>
                    <Button view="flat" onClick={() => void router.navigate({to: '/resource-manager/catalog'})}>
                      {t('catalog.backToCatalog')}
                    </Button>
                    {selected.isPending && <Text color="secondary">{t('catalog.loading')}</Text>}
                    {selected.isError && <ErrorMessage error={selected.error} t={t}
                      retry={() => void selected.refetch()}/>}
                    {selected.data && <ResourceInspector key={selected.data.kind + ':' + selected.data.id}
                      resource={selected.data} t={t}/>}
                  </>
                ) : (
                  <div className="m8-catalog__empty-detail">
                    <Text variant="header-1">{t('catalog.card')}</Text>
                    <Text color="secondary">{t('catalog.selectResource')}</Text>
                  </div>
                )}
              </Card>
            </div>
          </>
        ) : (
          <Card view="outlined" className="m8-catalog__message">
            <Text color="secondary">{t('catalog.noScope')}</Text>
          </Card>
        )}
      </section>
    </main>
  )
}
