import {loggedFetch} from '../../../platform/http/loggedFetch'

const organizationStates = new Set<Organization['state']>([
  'CREATING', 'ACTIVE', 'SUSPENDED', 'ARCHIVED', 'DELETING',
  'DELETED', 'FAILED', 'STATE_UNSPECIFIED',
])

export interface Organization {
  id: string
  state: 'CREATING' | 'ACTIVE' | 'SUSPENDED' | 'ARCHIVED' | 'DELETING' | 'DELETED' | 'FAILED' | 'STATE_UNSPECIFIED'
  name?: string
  description?: string
  createTime?: string
  updateTime?: string
  version?: string | number
  labels?: Record<string, string>
}

export interface ListOrganizationsResponse {
  organizations: Organization[]
  nextPageToken?: string
  totalSize: number
}

export interface FetchOrganizationsOptions {
  tenantId: string
  pageSize: number
  pageToken?: string
  filter?: string
  stateFilter?: Organization['state']
  orderBy?: string
  signal?: AbortSignal
}

/**
 * ListOrganizations is no longer registered at /resource-manager/v1/organizations.
 * The canonical v1 ResourceManagerService exposes a Tenant-scoped ListResources.
 * Pagination and display sorting below are client-side until durable indexed
 * server pagination is implemented; never query across Tenant boundaries.
 */
export async function fetchOrganizations({
  tenantId,
  pageSize,
  pageToken,
  filter,
  stateFilter,
  orderBy = 'name asc',
  signal,
}: FetchOrganizationsOptions): Promise<ListOrganizationsResponse> {
  if (!tenantId) throw new Error('Tenant ID is required for Organization listing')
  const baseUrl = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
  const parameters = new URLSearchParams({tenantId, kind: 'organization'})
  const response = await loggedFetch(
    'resource-manager',
    `${baseUrl}/resource-manager/v1/resources?${parameters}`,
    {credentials: 'same-origin', headers: {Accept: 'application/json'}, signal},
  )
  if (!response.ok) {
    if (response.status === 404) {
      throw new Error('Parent Tenant was not found in the current Resource Manager')
    }
    if (response.status === 401 || response.status === 403) {
      throw new Error(`Organization list access denied (HTTP ${response.status})`)
    }
    throw new Error(`Resource Manager returned HTTP ${response.status}`)
  }
  const value: unknown = await response.json()
  if (!isRecord(value) || !Array.isArray(value.resources)) {
    throw new Error('Resource Manager returned an invalid resource list')
  }
  const organizations = value.resources.map((item, index) => parseOrganizationResource(item, index, tenantId))
  const substring = (filter ?? '').trim().toLocaleLowerCase()
  const filtered = organizations.filter((org) => {
    const matchesSearch = !substring || [org.name, org.id, org.description]
      .some((value) => value?.toLocaleLowerCase().includes(substring))
    return matchesSearch && (!stateFilter || org.state === stateFilter)
  })

  // The existing table accepts name/id/create_time/update_time ordering.
  const [field, direction] = orderBy.split(' ')
  const property = field === 'id' ? 'id'
    : field === 'create_time' ? 'createTime'
      : field === 'update_time' ? 'updateTime' : 'name'
  filtered.sort((a, b) => {
    const result = (a[property] ?? '').localeCompare(b[property] ?? '', undefined, {numeric: true})
    return direction === 'desc' ? -result : result
  })
  const offset = pageToken === undefined ? 0 : Number(pageToken)
  if (!Number.isSafeInteger(offset) || offset < 0 || pageSize < 1) {
    throw new Error('Invalid Organization page offset')
  }
  const next = offset + pageSize
  return {
    organizations: filtered.slice(offset, next),
    totalSize: filtered.length,
    nextPageToken: next < filtered.length ? String(next) : undefined,
  }
}

function parseOrganizationResource(value: unknown, index: number, requestedTenantId: string): Organization {
  if (!isRecord(value) || !isRecord(value.organization)) {
    throw new Error(`Expected Organization resource at index ${index}`)
  }
  const org = value.organization
  if (org.tenantId !== requestedTenantId || !isRecord(org.metadata)) {
    throw new Error('Resource Manager returned an Organization outside the requested Tenant')
  }
  const metadata = org.metadata
  if (typeof metadata.id !== 'string' || metadata.id.length === 0) {
    throw new Error(`Missing Organization ID at index ${index}`)
  }
  const stateValue = typeof metadata.state === 'string'
    ? metadata.state.replace(/^RESOURCE_STATE_/, '') : 'STATE_UNSPECIFIED'
  const state = organizationStates.has(stateValue as Organization['state'])
    ? stateValue as Organization['state'] : 'STATE_UNSPECIFIED'
  return {
    id: metadata.id,
    state,
    name: optionalString(metadata.name),
    description: optionalString(metadata.description),
    createTime: optionalString(metadata.createTime),
    updateTime: optionalString(metadata.updateTime),
    version: typeof metadata.revision === 'number' || typeof metadata.revision === 'string'
      ? metadata.revision : undefined,
    labels: isStringRecord(metadata.labels) ? metadata.labels : undefined,
  }
}

function optionalString(value: unknown) {
  return typeof value === 'string' ? value : undefined
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function isStringRecord(value: unknown): value is Record<string, string> {
  return isRecord(value) && Object.values(value).every((item) => typeof item === 'string')
}
