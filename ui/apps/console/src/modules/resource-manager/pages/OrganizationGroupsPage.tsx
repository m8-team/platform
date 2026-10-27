import {useMemo, useRef, useState} from 'react'
import {ArrowRotateRight, Magnifier} from '@gravity-ui/icons'
import {Button, Card, Icon, Label, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'
import {useParams, useRouter} from '@tanstack/react-router'

import {ConsoleBreadcrumbs} from '../../../components/ConsoleBreadcrumbs'
import type {Translate} from '../../../i18n'
import {verifyTenant} from '../api/organizationMutations'
import {fetchOrganizations} from '../api/organizations'
import {
  addGroupOrganization, createGroup, getGroup, listGroupLinks,
  listGroups, removeGroupOrganization, verifyOrganizationInTenant,
} from '../api/organizationGroups'
import './organization-groups.css'

const uuid = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
const tenantStorageKey = 'm8.resource-manager.organizations.tenant-id'

function initialTenantId() {
  let saved = ''
  try {saved = sessionStorage.getItem(tenantStorageKey) ?? ''} catch { /* optional */ }
  const configured = (import.meta.env.VITE_RESOURCE_MANAGER_TENANT_ID ?? '').trim()
  return uuid.test(saved) ? saved : uuid.test(configured) ? configured : ''
}

function Status({state, t}: {state: string; t: Translate}) {
  const theme = state === 'ACTIVE' ? 'success' : state === 'SUSPENDED' ? 'warning' : 'normal'
  const labels: Record<string, 'organizations.state.ACTIVE' | 'organizations.state.SUSPENDED' |
    'organizations.state.ARCHIVED' | 'organizations.state.DELETED' | 'organizations.state.FAILED'> = {
    ACTIVE: 'organizations.state.ACTIVE', SUSPENDED: 'organizations.state.SUSPENDED',
    ARCHIVED: 'organizations.state.ARCHIVED', DELETED: 'organizations.state.DELETED',
    FAILED: 'organizations.state.FAILED',
  }
  return <Label theme={theme}>{labels[state] ? t(labels[state]) : state}</Label>
}

function ErrorPanel({t, error, retry}: {t: Translate; error: Error | null; retry?: () => void}) {
  return <div className="m8-groups__error" role="alert">
    <Text color="danger" variant="body-2">{t('groups.error')}</Text>
    <Text color="secondary" variant="caption-2">{error?.message}</Text>
    {retry && <Button onClick={retry}>{t('groups.retry')}</Button>}
  </div>
}

export function OrganizationGroupsPage({t}: {t: Translate}) {
  const router = useRouter()
  const params = useParams({strict: false}) as {groupId?: string}
  const queryClient = useQueryClient()
  const [tenantId, setTenantId] = useState(initialTenantId)
  const [tenantInput, setTenantInput] = useState(initialTenantId)
  const [search, setSearch] = useState('')
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [organizationId, setOrganizationId] = useState('')
  const [pendingRemoval, setPendingRemoval] = useState<string | null>(null)
  const createKey = useRef(crypto.randomUUID())
  const linkKey = useRef(crypto.randomUUID())
  const removalKeys = useRef<Record<string, string>>({})

  const groups = useQuery({
    queryKey: ['resource-manager', 'groups', tenantId],
    queryFn: ({signal}) => listGroups(tenantId, signal),
    enabled: uuid.test(tenantId),
    retry: false,
  })
  const selected = useQuery({
    queryKey: ['resource-manager', 'groups', 'detail', tenantId, params.groupId],
    queryFn: ({signal}) => getGroup(params.groupId!, tenantId, signal),
    enabled: Boolean(params.groupId) && uuid.test(tenantId),
    retry: false,
  })
  const links = useQuery({
    queryKey: ['resource-manager', 'groups', 'links', params.groupId],
    queryFn: ({signal}) => listGroupLinks(params.groupId!, signal),
    enabled: Boolean(params.groupId) && selected.isSuccess,
    retry: false,
  })
  const organizations = useQuery({
    queryKey: ['resource-manager', 'groups', 'organizations', tenantId],
    queryFn: ({signal}) => fetchOrganizations({tenantId, pageSize: 10000, signal}),
    enabled: uuid.test(tenantId) && Boolean(params.groupId) && selected.isSuccess,
    retry: false,
  })
  const orgNames = new Map((organizations.data?.organizations ?? []).map((org) => [org.id, org.name ?? org.id]))
  const filtered = useMemo(() => (groups.data ?? []).filter((g) =>
    [g.name, g.description, g.id].some((part) => part.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase())),
  ), [groups.data, search])

  const create = useMutation({
    mutationFn: async () => {
      await verifyTenant(tenantId)
      return createGroup(tenantId, name, description, createKey.current)
    },
    onSuccess: async (group) => {
      await queryClient.invalidateQueries({queryKey: ['resource-manager', 'groups', tenantId]})
      setCreating(false)
      void router.navigate({to: '/resource-manager/groups/$groupId', params: {groupId: group.id}})
    },
  })
  const add = useMutation({
    mutationFn: async () => {
      const id = organizationId.trim()
      await verifyOrganizationInTenant(id, tenantId)
      await addGroupOrganization(params.groupId!, id, linkKey.current)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({queryKey: ['resource-manager', 'groups', 'links', params.groupId]})
      setOrganizationId('')
      linkKey.current = crypto.randomUUID()
    },
  })
  const remove = useMutation({
    mutationFn: async (id: string) => {
      if (!params.groupId) throw new Error('Group must be selected')
      const key = removalKeys.current[id] ?? (removalKeys.current[id] = crypto.randomUUID())
      await removeGroupOrganization(params.groupId, id, key)
      return id
    },
    onSuccess: async (id) => {
      delete removalKeys.current[id]
      setPendingRemoval(null)
      await queryClient.invalidateQueries({queryKey: ['resource-manager', 'groups', 'links', params.groupId]})
    },
  })
  const openGroup = (groupId: string) => {
    setPendingRemoval(null)
    setOrganizationId('')
    linkKey.current = crypto.randomUUID()
    void router.navigate({to: '/resource-manager/groups/$groupId', params: {groupId}})
  }
  const selectTenant = () => {
    const next = tenantInput.trim()
    if (!uuid.test(next)) return
    try {sessionStorage.setItem(tenantStorageKey, next)} catch { /* optional */ }
    setTenantId(next)
    setOrganizationId('')
    setPendingRemoval(null)
    void router.navigate({to: '/resource-manager/groups'})
  }
  const resetCreateKey = () => {createKey.current = crypto.randomUUID();create.reset()}
  const validGroup = name.trim().length > 0 && name.trim().length <= 256 && description.length <= 1024
  const validOrgId = uuid.test(organizationId.trim())
  const activeGroup = selected.data?.state === 'ACTIVE'

  return <main className="m8-page__body">
    <section className="m8-page__content">
      <div className="m8-page__heading">
        <div>
          <ConsoleBreadcrumbs items={[
            {text: t('breadcrumb.resourceManager'), href: '/resource-manager'},
            {text: t('groups.title'), href: params.groupId ? '/resource-manager/groups' : undefined},
          ]}/>
          <Text as="h1" variant="display-1">{t('groups.title')}</Text>
          <Text as="p" variant="body-2" color="secondary">{t('groups.description')}</Text>
        </div>
        <div className="m8-groups__actions">
          <Button view="action" disabled={!uuid.test(tenantId)} onClick={() => setCreating(!creating)}>
            {t('groups.create')}
          </Button>
          <Button view="outlined" disabled={!uuid.test(tenantId)} loading={groups.isFetching}
            onClick={() => void groups.refetch()}><Icon data={ArrowRotateRight}/>{t('groups.refresh')}</Button>
        </div>
      </div>

      <Card view="outlined" className="m8-groups__scope">
        <Text variant="subheader-2">{t('groups.scope')}</Text>
        <form className="m8-groups__row" onSubmit={(e) => {e.preventDefault();selectTenant()}}>
          <TextInput className="m8-groups__input" value={tenantInput}
            placeholder={t('groups.tenantLabel')} aria-label={t('groups.tenantLabel')}
            onUpdate={setTenantInput}/>
          <Button type="submit" view="action" disabled={!uuid.test(tenantInput.trim())}>
            {t('groups.tenantSelect')}
          </Button>
        </form>
        <Text color="secondary" variant="caption-2">{t('groups.scopeHelp')}</Text>
      </Card>

      {creating && <Card view="outlined" className="m8-groups__create">
        <Text as="h2" variant="header-2">{t('groups.create')}</Text>
        <Text color="secondary" variant="body-2">{t('groups.createHelp')}</Text>
        <form className="m8-groups__fields" onSubmit={(e) => {e.preventDefault();if(validGroup && !create.isPending) create.mutate()}}>
          <label className="m8-groups__fields"><Text variant="caption-2">{t('groups.name')}</Text>
            <TextInput value={name} onUpdate={(value) => {setName(value);resetCreateKey()}}
              placeholder={t('groups.namePlaceholder')} disabled={create.isPending}/>
          </label>
          <label className="m8-groups__fields"><Text variant="caption-2">{t('groups.descriptionLabel')}</Text>
            <textarea className="m8-groups__textarea" value={description} maxLength={1024}
              onChange={(e) => {setDescription(e.target.value);resetCreateKey()}}
              disabled={create.isPending}/>
          </label>
          {create.isError && <ErrorPanel t={t} error={create.error}/>}
          <div className="m8-groups__actions">
            <Button onClick={() => setCreating(false)} disabled={create.isPending}>{t('groups.cancel')}</Button>
            <Button type="submit" view="action" disabled={!validGroup || create.isPending}
              loading={create.isPending}>{t('groups.save')}</Button>
          </div>
        </form>
      </Card>}

      {!tenantId ? <Card view="outlined" className="m8-groups__message">
        <Text color="secondary">{t('groups.tenantMissing')}</Text>
      </Card> : <div className="m8-groups__layout">
        <Card view="outlined" className="m8-groups__inventory">
          <Text as="h2" variant="header-2">{t('groups.list')}</Text>
          <TextInput value={search} onUpdate={setSearch}
            startContent={<Icon data={Magnifier}/>} placeholder={t('groups.search')}/>
          {groups.isPending && <Text color="secondary">{t('groups.loading')}</Text>}
          {groups.isError && <ErrorPanel t={t} error={groups.error} retry={() => void groups.refetch()}/>}
          {groups.isSuccess && !groups.data.length && <Text color="secondary">{t('groups.empty')}</Text>}
          {groups.isSuccess && groups.data.length > 0 && !filtered.length &&
            <Text color="secondary">{t('groups.noMatches')}</Text>}
          {filtered.map((group) => <button key={group.id} type="button"
            className={'m8-groups__item' + (params.groupId === group.id ? ' m8-groups__item_selected' : '')}
            onClick={() => openGroup(group.id)}>
            <span className="m8-groups__item-info">
              <Text variant="subheader-2">{group.name}</Text>
              <Text variant="caption-2" color="secondary" className="m8-groups__uuid">{group.id}</Text>
            </span>
            <Status state={group.state} t={t}/>
          </button>)}
        </Card>

        <Card view="outlined" className="m8-groups__detail">
          {!params.groupId ? <Text color="secondary">{t('groups.select')}</Text> : <>
            {selected.isPending && <Text color="secondary">{t('groups.loading')}</Text>}
            {selected.isError && <ErrorPanel t={t} error={selected.error} retry={() => void selected.refetch()}/>}
            {selected.data && <>
              <div className="m8-groups__heading">
                <div><Text color="secondary" variant="caption-2">{t('groups.details')}</Text>
                  <Text as="h2" variant="header-2">{selected.data.name}</Text></div>
                <Status state={selected.data.state} t={t}/>
              </div>
              <Text variant="body-2">{selected.data.description}</Text>
              <dl className="m8-groups__meta">
                <dt>ID</dt><dd>{selected.data.id}</dd>
                <dt>{t('groups.tenant')}</dt><dd>{selected.data.tenantId}</dd>
                <dt>{t('groups.revision')}</dt><dd>{selected.data.revision}</dd>
              </dl>
              <hr className="m8-groups__divider"/>
              <Text as="h3" variant="header-1">{t('groups.members')} ({links.data?.length ?? '—'})</Text>
              <Text variant="body-2" color="secondary">{t('groups.membersHelp')}</Text>
              {links.isPending && <Text color="secondary">{t('groups.loading')}</Text>}
              {links.isError && <ErrorPanel t={t} error={links.error} retry={() => void links.refetch()}/>}
              {links.isSuccess && !links.data.length && <Text color="secondary">{t('groups.noMembers')}</Text>}
              {links.data?.map((link) => <div className="m8-groups__member" key={link.organizationId}>
                <div className="m8-groups__item-info">
                  <Text variant="subheader-2">{orgNames.get(link.organizationId) ?? link.organizationId}</Text>
                  <Text variant="caption-2" color="secondary" className="m8-groups__uuid">{link.organizationId}</Text>
                </div>
                <div className="m8-groups__actions">
                  <Button view="flat" onClick={() => void router.navigate({
                    to: '/resource-manager/organizations/$organizationId',
                    params: {organizationId: link.organizationId},
                  })}>{t('groups.viewOrganization')}</Button>
                  {pendingRemoval === link.organizationId ? <>
                    <Text variant="caption-2" color="secondary">{t('groups.removeNote')}</Text>
                    <Button view="outlined" onClick={() => setPendingRemoval(null)}>{t('groups.removeCancel')}</Button>
                    <Button view="outlined-danger" loading={remove.isPending}
                      disabled={remove.isPending}
                      onClick={() => remove.mutate(link.organizationId)}>{t('groups.removeConfirm')}</Button>
                  </> : <Button view="flat-danger" disabled={!activeGroup || remove.isPending}
                    onClick={() => setPendingRemoval(link.organizationId)}>{t('groups.remove')}</Button>}
                </div>
              </div>)}
              {remove.isError && <ErrorPanel t={t} error={remove.error}/>}
              {activeGroup && <form className="m8-groups__link" onSubmit={(e) => {
                e.preventDefault()
                if (validOrgId && !add.isPending) add.mutate()
              }}>
                <Text as="h3" variant="subheader-2">{t('groups.link')}</Text>
                <div className="m8-groups__row">
                  <TextInput value={organizationId} className="m8-groups__input"
                    placeholder={t('groups.linkId')} aria-label={t('groups.linkId')}
                    onUpdate={(value) => {setOrganizationId(value);linkKey.current = crypto.randomUUID();add.reset()}}
                    disabled={add.isPending}/>
                  <Button type="submit" view="action" disabled={!validOrgId || add.isPending}
                    loading={add.isPending}>{t('groups.link')}</Button>
                </div>
                <Text variant="caption-2" color="secondary">{t('groups.linkHelp')}</Text>
                {add.isError && <ErrorPanel t={t} error={add.error}/>}
              </form>}
              <Card view="outlined" className="m8-groups__notice">
                <Text variant="body-2">{t('groups.noAccessInheritance')}</Text>
              </Card>
              <div className="m8-groups__relations">
                <div className="m8-groups__heading">
                  <Text as="h3" variant="header-1">{t('groups.relationships')}</Text>
                  <Label theme="normal">{t('groups.unsupported')}</Label>
                </div>
                <Text variant="body-2" color="secondary">{t('groups.relationshipsUnavailable')}</Text>
              </div>
            </>}
          </>}
        </Card>
      </div>}
      <Text variant="caption-2" color="secondary">{t('groups.pageNote')}</Text>
    </section>
  </main>
}
