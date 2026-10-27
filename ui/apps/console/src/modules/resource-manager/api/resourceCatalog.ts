import {loggedFetch} from '../../../platform/http/loggedFetch'

export interface CatalogResource {
  id: string
  kind: string
  name: string
  description: string
  state: string
  revision?: number
  module?: string
  owningOrganizationId?: string
  owner?: {kind: string; id: string}
  labels: Record<string, string>
  createdAt?: string
  updatedAt?: string
}

export interface ResourceScope {
  owningOrganizationId: string
  ancestry: Array<{kind: string; id: string}>
}

export interface CatalogPage {
  resources: CatalogResource[]
  nextPageToken?: string
}

const baseUrl = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
const resourceKinds: Record<string, string> = {
  tenant: 'tenant', group: 'group', organization: 'organization',
  workspace: 'workspace', project: 'project', environment: 'environment',
  workloadService: 'workload_service',
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function stringValue(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function stringMap(value: unknown): Record<string, string> {
  if (!isRecord(value)) return {}
  return Object.fromEntries(Object.entries(value).filter((item): item is [string, string] =>
    typeof item[1] === 'string'))
}

function parseOwner(value: unknown): CatalogResource['owner'] {
  if (!isRecord(value)) return undefined
  for (const kind of ['organization', 'workspace', 'project']) {
    const id = stringValue(value[kind + 'Id'])
    if (id) return {kind, id}
  }
  return undefined
}

export function parseCatalogResource(raw: unknown): CatalogResource {
  if (!isRecord(raw)) throw new Error('Invalid Resource Manager record')
  const branch = Object.keys(raw).find((key) => key === 'descriptor' || key in resourceKinds)
  if (!branch || !isRecord(raw[branch])) throw new Error('Unknown Resource Manager record')
  const value = raw[branch] as Record<string, unknown>
  if (!isRecord(value.metadata)) throw new Error('Missing resource metadata')
  const meta = value.metadata
  const descriptorRef = branch === 'descriptor' && isRecord(value.resource) ? value.resource : undefined
  const kind = descriptorRef ? stringValue(descriptorRef.kind) : resourceKinds[branch]
  const id = stringValue(meta.id)
  if (!kind || !id) throw new Error('Resource kind or ID is missing')
  if (descriptorRef && stringValue(descriptorRef.id) !== id) {
    throw new Error('Resource descriptor identity mismatch')
  }
  const fallbackOwnerId = stringValue(value.organizationId)
  const owner = parseOwner(value.owner) ??
    (fallbackOwnerId && branch !== 'organization' ? {kind: 'organization', id: fallbackOwnerId} : undefined)
  return {
    id, kind,
    name: stringValue(meta.name) || id,
    description: stringValue(meta.description),
    state: stringValue(meta.state).replace(/^RESOURCE_STATE_/, '') || 'UNSPECIFIED',
    revision: typeof meta.revision === 'number' ? meta.revision : undefined,
    module: stringValue(value.module) || undefined,
    owningOrganizationId: stringValue(value.owningOrganizationId) ||
      (branch === 'organization' ? id : undefined),
    owner, labels: stringMap(meta.labels),
    createdAt: stringValue(meta.createTime) || undefined,
    updatedAt: stringValue(meta.updateTime) || undefined,
  }
}

async function getJson(path: string, signal?: AbortSignal): Promise<unknown> {
  const response = await loggedFetch('resource-manager', baseUrl + path, {
    credentials: 'same-origin', signal,
    headers: {Accept: 'application/json'},
  })
  if (!response.ok) throw new Error('Resource Manager: HTTP ' + response.status)
  return response.json()
}

export async function listCatalogResources(
  organizationId: string,
  pageToken: string,
  signal?: AbortSignal,
): Promise<CatalogPage> {
  if (!organizationId) throw new Error('Organization scope is required')
  const query = new URLSearchParams({organizationId, pageSize: '100'})
  if (pageToken) query.set('pageToken', pageToken)
  const raw = await getJson('/resource-manager/v1/resources?' + query, signal)
  if (!isRecord(raw) || !Array.isArray(raw.resources)) {
    throw new Error('Invalid ListResources response')
  }
  const resources = raw.resources.map(parseCatalogResource)
  const seen = new Set<string>()
  for (const resource of resources) {
    if (resource.owningOrganizationId && resource.owningOrganizationId !== organizationId) {
      throw new Error('Resource returned outside selected Organization')
    }
    const key = resource.kind + ':' + resource.id
    if (seen.has(key)) throw new Error('Duplicate resource in catalog page')
    seen.add(key)
  }
  return {resources, nextPageToken: stringValue(raw.nextPageToken) || undefined}
}

export async function getCatalogResource(kind: string, id: string, signal?: AbortSignal) {
  const raw = await getJson('/resource-manager/v1/resources/' +
    encodeURIComponent(kind) + '/' + encodeURIComponent(id), signal)
  const resource = parseCatalogResource(raw)
  if (resource.id !== id || resource.kind !== kind) {
    throw new Error('Resource Manager returned a different resource')
  }
  return resource
}

export async function resolveCatalogResourceScope(kind: string, id: string, signal?: AbortSignal): Promise<ResourceScope> {
  const raw = await getJson('/resource-manager/v1/resources/' +
    encodeURIComponent(kind) + '/' + encodeURIComponent(id) + ':resolveScope', signal)
  if (!isRecord(raw) || !Array.isArray(raw.ancestry)) throw new Error('Invalid ResolveResourceScope response')
  return {
    owningOrganizationId: stringValue(raw.owningOrganizationId),
    ancestry: raw.ancestry.map((entry) => {
      if (!isRecord(entry) || !stringValue(entry.kind) || !stringValue(entry.id)) {
        throw new Error('Invalid resource ancestry')
      }
      return {kind: stringValue(entry.kind), id: stringValue(entry.id)}
    }),
  }
}
