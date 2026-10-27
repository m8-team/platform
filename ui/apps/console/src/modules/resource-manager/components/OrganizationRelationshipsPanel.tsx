import {useMemo, useRef, useState} from 'react'
import {Button, Card, Label, Text, TextInput} from '@gravity-ui/uikit'
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'
import type {Translate} from '../../../i18n'
import {fetchOrganizations} from '../api/organizations'
import {
  closeRelationship, createRelationship, fetchRelationships, fetchRelationshipHistory, updateRelationship,
} from '../api/organizationRelationships'
import type {OrganizationRelationship, RelationshipDraft, RelationshipType} from '../api/organizationRelationships'
import './organization-relationships.css'

const types: RelationshipType[] = ['PARENT_SUBSIDIARY', 'AFFILIATE', 'PARTNER', 'SERVICE_PROVIDER']
const uuid = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
const localNow = () => {
  const now = new Date()
  return new Date(now.getTime() - now.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
}
const localTime = (time: string) => {
  const date = new Date(time)
  return Number.isNaN(date.getTime()) ? '' :
    new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
}
const toISO = (value: string) => value ? new Date(value).toISOString() : undefined
const pct = (value: number) => (value / 100).toLocaleString(undefined, {maximumFractionDigits: 2})

interface Props {t: Translate; tenantId: string; memberIds?: string[]}

export function OrganizationRelationshipsPanel({t, tenantId, memberIds}: Props) {
  const queryClient = useQueryClient()
  const [creating, setCreating] = useState(false)
  const [fromId, setFromId] = useState('')
  const [toId, setToId] = useState('')
  const [type, setType] = useState<RelationshipType>('PARENT_SUBSIDIARY')
  const [stake, setStake] = useState('0')
  const [validFrom, setValidFrom] = useState(localNow)
  const [validTo, setValidTo] = useState('')
  const [evidence, setEvidence] = useState('')
  const [editing, setEditing] = useState<string | null>(null)
  const [confirmClose, setConfirmClose] = useState<string | null>(null)
  const [historyId, setHistoryId] = useState<string | null>(null)
  const [editStake, setEditStake] = useState('0')
  const [editTo, setEditTo] = useState('')
  const [editEvidence, setEditEvidence] = useState('')
  const createKey = useRef(crypto.randomUUID())
  const updateKey = useRef(crypto.randomUUID())
  const closeKeys = useRef<Record<string, {key: string; timestamp: string}>>({})
  const queryKey = ['resource-manager', 'organization-relationships', tenantId]
  const all = useQuery({
    queryKey, queryFn: ({signal}) => fetchRelationships(tenantId, signal),
    enabled: uuid.test(tenantId), retry: false,
  })
  const organizations = useQuery({
    queryKey: ['resource-manager', 'organization-relationships', 'organizations', tenantId],
    queryFn: ({signal}) => fetchOrganizations({tenantId, pageSize: 10000, signal}),
    enabled: uuid.test(tenantId), retry: false,
  })
  const history = useQuery({
    queryKey: ['resource-manager', 'organization-relationships', 'history', tenantId, historyId],
    queryFn: ({signal}) => fetchRelationshipHistory(tenantId, historyId!, signal),
    enabled: uuid.test(tenantId) && Boolean(historyId), retry: false,
  })
  const orgs = useMemo(() => organizations.data?.organizations ?? [], [organizations.data])
  const orgNames = useMemo(() => new Map(orgs.map((o) => [o.id, o.name || o.id])), [orgs])
  const members = useMemo(() => new Set(memberIds), [memberIds])
  const visible = useMemo(() => (all.data ?? []).filter((rel) => !memberIds ||
    members.has(rel.fromOrganizationId) || members.has(rel.toOrganizationId)),
  [all.data, memberIds, members])
  const selected = all.data?.find((rel) => rel.id === editing)
  const canCreate = fromId !== toId && uuid.test(fromId) && uuid.test(toId) &&
    Number.isFinite(Date.parse(toISO(validFrom) || '')) &&
    (!validTo || Date.parse(toISO(validTo) || '') > Date.parse(toISO(validFrom) || '')) &&
    Number(stake) >= 0 && Number(stake) <= 100 &&
    (type === 'PARENT_SUBSIDIARY' || Number(stake) === 0) && evidence.length <= 1024
  const create = useMutation({
    mutationFn: async () => {
      const draft: RelationshipDraft = {
        fromOrganizationId: fromId, toOrganizationId: toId, type,
        ownershipBasisPoints: type === 'PARENT_SUBSIDIARY' ? Math.round(Number(stake) * 100) : 0,
        validFrom: toISO(validFrom)!, validTo: toISO(validTo), evidenceRef: evidence.trim(),
      }
      return createRelationship(tenantId, draft, createKey.current)
    },
    onSuccess: async () => {
      setCreating(false)
      createKey.current = crypto.randomUUID()
      await queryClient.invalidateQueries({queryKey})
    },
  })
  const update = useMutation({
    mutationFn: async (rel: OrganizationRelationship) =>
      updateRelationship(tenantId, rel.id, rel.revision,
        rel.type === 'PARENT_SUBSIDIARY' ? Math.round(Number(editStake) * 100) : 0,
        toISO(editTo), editEvidence.trim(), updateKey.current),
    onSuccess: async () => {
      setEditing(null)
      updateKey.current = crypto.randomUUID()
      await queryClient.invalidateQueries({queryKey})
    },
  })
  const close = useMutation({
    mutationFn: async (rel: OrganizationRelationship) => {
      const command = closeKeys.current[rel.id] ||
        (closeKeys.current[rel.id] = {key: crypto.randomUUID(), timestamp: new Date().toISOString()})
      return closeRelationship(tenantId, rel.id, rel.revision, command.key, command.timestamp)
    },
    onSuccess: async (rel) => {
      delete closeKeys.current[rel.id]
      setConfirmClose(null)
      await queryClient.invalidateQueries({queryKey})
      await queryClient.invalidateQueries({queryKey: ['resource-manager', 'organization-relationships', 'history', tenantId]})
    },
  })
  const startEditing = (rel: OrganizationRelationship) => {
    setEditing(rel.id)
    setEditStake(String(rel.ownershipBasisPoints / 100))
    setEditTo(rel.validTo ? localTime(rel.validTo) : '')
    setEditEvidence(rel.evidenceRef)
    updateKey.current = crypto.randomUUID()
    update.reset()
  }
  const label = (kind: RelationshipType) =>
    t(('relations.type.' + kind) as 'relations.type.PARENT_SUBSIDIARY' |
      'relations.type.AFFILIATE' | 'relations.type.PARTNER' | 'relations.type.SERVICE_PROVIDER')
  const name = (id: string) => orgNames.get(id) || id
  const error = create.error || update.error || close.error || all.error || organizations.error
  const disabledEdit = !selected || Number(editStake) < 0 || Number(editStake) > 100 ||
    (editTo && Date.parse(toISO(editTo) || '') <= Date.parse(selected.validFrom))

  return <section className="m8-relations">
    <div className="m8-relations__heading">
      <div>
        <Text as="h3" variant="header-1">{t('relations.title')}</Text>
        <Text variant="body-2" color="secondary">{t('relations.intro')}</Text>
      </div>
      <div className="m8-relations__buttons">
        <Button onClick={() => void all.refetch()} disabled={!tenantId || all.isFetching}>{t('relations.refresh')}</Button>
        <Button view="action" disabled={!tenantId || orgs.length < 2} onClick={() => setCreating(!creating)}>
          {t('relations.create')}
        </Button>
      </div>
    </div>
    <Text variant="caption-2" color="secondary">
      {memberIds ? t('relations.groupFiltered') : t('relations.allTenant')}
    </Text>
    <Card view="outlined" className="m8-relations__notice">
      <Text variant="body-2">{t('relations.noInheritance')}</Text>
    </Card>
    {all.isPending || organizations.isPending ? <Text color="secondary">{t('relations.loading')}</Text> : null}
    {error ? <Card view="outlined" className="m8-relations__error">
      <Text color="danger" variant="body-2">{t('relations.error')}</Text>
      <Text color="secondary">{error instanceof Error ? error.message : String(error)}</Text>
    </Card> : null}
    {creating && <Card view="outlined" className="m8-relations__form">
      <Text as="h4" variant="header-2">{t('relations.create')}</Text>
      <div className="m8-relations__form-grid">
        <label><Text variant="caption-2">{t('relations.from')}</Text>
          <select value={fromId} onChange={(e) => {setFromId(e.target.value);createKey.current = crypto.randomUUID();create.reset()}}>
            <option value="">—</option>
            {orgs.map((o) => <option value={o.id} key={o.id}>{o.name || o.id}</option>)}
          </select>
        </label>
        <label><Text variant="caption-2">{t('relations.to')}</Text>
          <select value={toId} onChange={(e) => {setToId(e.target.value);createKey.current = crypto.randomUUID();create.reset()}}>
            <option value="">—</option>
            {orgs.map((o) => <option value={o.id} key={o.id}>{o.name || o.id}</option>)}
          </select>
        </label>
        <label><Text variant="caption-2">{t('relations.type')}</Text>
          <select value={type} onChange={(e) => {setType(e.target.value as RelationshipType);setStake('0');createKey.current=crypto.randomUUID()}}>
            {types.map((kind) => <option value={kind} key={kind}>{label(kind)}</option>)}
          </select>
        </label>
        {type === 'PARENT_SUBSIDIARY' && <label><Text variant="caption-2">{t('relations.stake')}</Text>
          <input type="number" value={stake} min={0} max={100} step={0.01}
            onChange={(e) => {setStake(e.target.value);createKey.current=crypto.randomUUID()}}/>
          <Text variant="caption-2" color="secondary">{t('relations.stakeHelp')}</Text>
        </label>}
        <label><Text variant="caption-2">{t('relations.validFrom')}</Text>
          <input type="datetime-local" value={validFrom} onChange={(e) => {setValidFrom(e.target.value);createKey.current=crypto.randomUUID()}}/>
        </label>
        <label><Text variant="caption-2">{t('relations.validTo')}</Text>
          <input type="datetime-local" value={validTo} onChange={(e) => {setValidTo(e.target.value);createKey.current=crypto.randomUUID()}}/>
        </label>
        <label className="m8-relations__wide"><Text variant="caption-2">{t('relations.evidence')}</Text>
          <TextInput value={evidence} onUpdate={(v) => {setEvidence(v);createKey.current=crypto.randomUUID()}}/>
          <Text variant="caption-2" color="secondary">{t('relations.evidenceHint')}</Text>
        </label>
      </div>
      <Text variant="caption-2" color="secondary">{t('relations.duplicateHelp')}</Text>
      <div className="m8-relations__buttons">
        <Button disabled={create.isPending} onClick={() => setCreating(false)}>{t('relations.cancel')}</Button>
        <Button view="action" disabled={!canCreate || create.isPending} loading={create.isPending}
          onClick={() => create.mutate()}>{t('relations.create')}</Button>
      </div>
    </Card>}
    {all.isSuccess && !visible.length && <Text color="secondary">{t('relations.empty')}</Text>}
    {visible.map((rel) => <Card key={rel.id} view="outlined" className="m8-relations__item">
      <div className="m8-relations__heading">
        <div className="m8-relations__relation">
          <Text variant="subheader-2">{name(rel.fromOrganizationId)} → {name(rel.toOrganizationId)}</Text>
          <Text variant="body-2" color="secondary">{label(rel.type)}</Text>
          <Text variant="caption-2" color="secondary">{rel.id}</Text>
        </div>
        <Label theme={rel.state === 'ACTIVE' ? 'success' : 'normal'}>
          {t(rel.state === 'ACTIVE' ? 'relations.state.ACTIVE' : 'relations.state.CLOSED')}
        </Label>
      </div>
      <dl className="m8-relations__info">
        <dt>{t('relations.validFrom')}</dt><dd>{new Date(rel.validFrom).toLocaleString()}</dd>
        <dt>{t('relations.validTo')}</dt><dd>{rel.validTo ? new Date(rel.validTo).toLocaleString() : '—'}</dd>
        {rel.type === 'PARENT_SUBSIDIARY' && <>
          <dt>{t('relations.stake')}</dt><dd>{rel.ownershipBasisPoints ? pct(rel.ownershipBasisPoints) + '%' : '—'}</dd>
        </>}
        <dt>{t('relations.revision')}</dt><dd>{rel.revision}</dd>
        <dt>{t('relations.evidence')}</dt><dd>{rel.evidenceRef || '—'}</dd>
      </dl>
      {rel.state === 'ACTIVE' && <div className="m8-relations__buttons">
        <Button onClick={() => startEditing(rel)}>{t('relations.edit')}</Button>
        <Button view="outlined-danger" onClick={() => setConfirmClose(rel.id)}>{t('relations.close')}</Button>
      </div>}
      <div className="m8-relations__buttons">
        <Button view="flat" onClick={() => setHistoryId(historyId === rel.id ? null : rel.id)}>
          {t('relations.history')}
        </Button>
      </div>
      {historyId === rel.id && <div className="m8-relations__form">
        <Text as="h4" variant="subheader-2">{t('relations.history')}</Text>
        {history.isPending && <Text color="secondary">{t('relations.loading')}</Text>}
        {history.isError && <Text color="danger">{history.error.message}</Text>}
        {history.data?.map((version) => <div className="m8-relations__revision" key={version.revision}>
          <Text variant="subheader-2">{t('relations.revision')} {version.revision}</Text>
          <Text color="secondary" variant="caption-2">{version.updateTime || '—'}</Text>
          <Text variant="body-2">{t(version.state === 'CLOSED' ? 'relations.state.CLOSED' : 'relations.state.ACTIVE')}</Text>
          <Text variant="body-2">{version.type === 'PARENT_SUBSIDIARY' ?
            pct(version.ownershipBasisPoints) + '%' : label(version.type)}</Text>
          <Text variant="caption-2" color="secondary">{version.evidenceRef || '—'}</Text>
        </div>)}
      </div>}
      {editing === rel.id && <div className="m8-relations__form">
        {rel.type === 'PARENT_SUBSIDIARY' && <label>
          <Text variant="caption-2">{t('relations.stake')}</Text>
          <input type="number" min={0} max={100} step={0.01} value={editStake} onChange={(e) => {
            setEditStake(e.target.value);updateKey.current=crypto.randomUUID();update.reset()
          }}/>
        </label>}
        <label><Text variant="caption-2">{t('relations.validTo')}</Text>
          <input type="datetime-local" value={editTo} onChange={(e) => {
            setEditTo(e.target.value);updateKey.current=crypto.randomUUID();update.reset()
          }}/>
        </label>
        <label><Text variant="caption-2">{t('relations.evidence')}</Text>
          <TextInput value={editEvidence} onUpdate={(v) => {
            setEditEvidence(v);updateKey.current=crypto.randomUUID();update.reset()
          }}/>
        </label>
        <div className="m8-relations__buttons">
          <Button onClick={() => setEditing(null)}>{t('relations.cancel')}</Button>
          <Button view="action" disabled={Boolean(disabledEdit) || update.isPending}
            loading={update.isPending} onClick={() => update.mutate(rel)}>{t('relations.save')}</Button>
        </div>
      </div>}
      {confirmClose === rel.id && <div className="m8-relations__close">
        <Text variant="body-2">{t('relations.closeNotice')}</Text>
        {Date.parse(rel.validFrom) >= Date.now() && <Text color="danger">{t('relations.cannotClose')}</Text>}
        <div className="m8-relations__buttons">
          <Button onClick={() => setConfirmClose(null)}>{t('relations.cancel')}</Button>
          <Button view="outlined-danger" disabled={close.isPending || Date.parse(rel.validFrom) >= Date.now()}
            loading={close.isPending} onClick={() => close.mutate(rel)}>{t('relations.confirmClose')}</Button>
        </div>
      </div>}
    </Card>)}
  </section>
}
