import {loggedFetch} from '../../../platform/http/loggedFetch'

const base = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '') +
  '/resource-manager/v1/organization-relationships'

export type RelationshipType = 'PARENT_SUBSIDIARY' | 'AFFILIATE' | 'PARTNER' | 'SERVICE_PROVIDER'
export interface OrganizationRelationship {
  id: string
  tenantId: string
  fromOrganizationId: string
  toOrganizationId: string
  type: RelationshipType
  ownershipBasisPoints: number
  validFrom: string
  validTo?: string
  state: 'ACTIVE' | 'CLOSED'
  revision: number
  createTime?: string
  updateTime?: string
  evidenceRef: string
}
interface ListRelationships {relationships: OrganizationRelationship[];nextPageToken?: string}

async function send(path: string, options: RequestInit = {}): Promise<unknown> {
  const response = await loggedFetch('resource-manager', base + path, {
    credentials: 'same-origin',
    headers: {Accept: 'application/json', ...(options.body ? {'Content-Type': 'application/json'} : {})},
    ...options,
  })
  if (!response.ok) {
    const message = response.status === 403 ? 'Permission denied' :
      response.status === 409 || response.status === 412 ? 'Relationship revision conflict' :
      response.status === 400 ? 'Invalid relationship data' :
      response.status === 404 ? 'Relationship or Tenant not found' :
      response.status === 501 ? 'Relationship API is not implemented' :
      'Relationship API request failed'
    throw new Error(message + ' (HTTP ' + response.status + ')')
  }
  return response.json() as Promise<unknown>
}
const isRecord = (v: unknown): v is Record<string, unknown> =>
  v !== null && typeof v === 'object' && !Array.isArray(v)
const string = (v: unknown) => typeof v === 'string' ? v : ''
const uuid = /^[a-f\d]{8}-[a-f\d]{4}-[1-8][a-f\d]{3}-[89ab][a-f\d]{3}-[a-f\d]{12}$/i
const types = new Set(['PARENT_SUBSIDIARY', 'AFFILIATE', 'PARTNER', 'SERVICE_PROVIDER'])

function parse(raw: unknown, tenantId: string): OrganizationRelationship {
  if (!isRecord(raw) || !uuid.test(string(raw.id)) || raw.tenantId !== tenantId ||
      !uuid.test(string(raw.fromOrganizationId)) || !uuid.test(string(raw.toOrganizationId))) {
    throw new Error('Invalid or cross-Tenant relationship response')
  }
  const type = string(raw.type).replace(/^ORGANIZATION_RELATIONSHIP_TYPE_/, '')
  const state = string(raw.state).replace(/^ORGANIZATION_RELATIONSHIP_STATE_/, '')
  if (!types.has(type) || (state !== 'ACTIVE' && state !== 'CLOSED')) throw new Error('Unknown relationship state/type')
  return {
    id: string(raw.id), tenantId, fromOrganizationId: string(raw.fromOrganizationId),
    toOrganizationId: string(raw.toOrganizationId), type: type as RelationshipType,
    ownershipBasisPoints: Number(raw.ownershipBasisPoints) || 0,
    validFrom: string(raw.validFrom), validTo: string(raw.validTo) || undefined,
    state, revision: Number(raw.revision) || 0,
    createTime: string(raw.createTime), updateTime: string(raw.updateTime),
    evidenceRef: string(raw.evidenceRef),
  }
}

export async function fetchRelationships(tenantId: string, signal?: AbortSignal): Promise<OrganizationRelationship[]> {
  if (!uuid.test(tenantId)) throw new Error('Select a valid Tenant UUID')
  const items: OrganizationRelationship[] = []
  const tokens = new Set<string>()
  let token = ''
  for (let index = 0; index < 100; index++) {
    const params = new URLSearchParams({tenantId, includeClosed: 'true', pageSize: '200'})
    if (token) params.set('pageToken', token)
    const raw = await send('?' + params.toString(), {signal})
    if (!isRecord(raw) || !Array.isArray(raw.relationships)) throw new Error('Invalid relationship list response')
    const page = raw as unknown as ListRelationships
    for (const entry of page.relationships) items.push(parse(entry, tenantId))
    const next = string(page.nextPageToken)
    if (!next) {
      const unique = new Set(items.map((r) => r.id))
      if (unique.size !== items.length) throw new Error('Duplicate corporate relationships')
      return items
    }
    if (tokens.has(next)) throw new Error('Repeated relationship page token')
    tokens.add(next)
    token = next
  }
  throw new Error('Relationship pagination limit exceeded')
}

export interface RelationshipDraft {
  fromOrganizationId: string
  toOrganizationId: string
  type: RelationshipType
  ownershipBasisPoints: number
  validFrom: string
  validTo?: string
  evidenceRef: string
}
export async function createRelationship(tenantId: string, draft: RelationshipDraft, idempotencyKey: string) {
  const raw = await send('', {method: 'POST', body: JSON.stringify({
    relationship: {...draft, tenantId, type: 'ORGANIZATION_RELATIONSHIP_TYPE_' + draft.type},
    idempotencyKey,
  })})
  return parse(raw, tenantId)
}
export async function updateRelationship(tenantId: string, id: string, revision: number,
  ownershipBasisPoints: number, validTo: string | undefined, evidenceRef: string, idempotencyKey: string) {
  const raw = await send('/' + encodeURIComponent(id), {method: 'PATCH', body: JSON.stringify({
    id, tenantId, expectedRevision: revision, idempotencyKey,
    ownershipBasisPoints, ...(validTo ? {validTo} : {}), evidenceRef,
  })})
  return parse(raw, tenantId)
}
export async function fetchRelationshipHistory(tenantId: string, id: string, signal?: AbortSignal) {
  const versions: OrganizationRelationship[] = []
  const tokens = new Set<string>()
  let token = ''
  for (let i = 0; i < 100; i++) {
    const params = new URLSearchParams({tenantId, pageSize: '100'})
    if (token) params.set('pageToken', token)
    const raw = await send('/' + encodeURIComponent(id) + '/versions?' + params, {signal})
    if (!isRecord(raw) || !Array.isArray(raw.versions)) throw new Error('Invalid relationship revision history')
    for (const entry of raw.versions) {
      const version = parse(entry, tenantId)
      if (version.id !== id) throw new Error('Relationship history identity mismatch')
      versions.push(version)
    }
    const next = string(raw.nextPageToken)
    if (!next) return versions
    if (tokens.has(next)) throw new Error('Repeated relationship history token')
    tokens.add(next)
    token = next
  }
  throw new Error('Relationship version pagination limit exceeded')
}

export async function closeRelationship(tenantId: string, id: string, revision: number, idempotencyKey: string, closedAt: string) {
  const raw = await send('/' + encodeURIComponent(id) + ':close', {method: 'POST', body: JSON.stringify({
    id, tenantId, expectedRevision: revision, idempotencyKey,
    closedAt,
  })})
  return parse(raw, tenantId)
}
