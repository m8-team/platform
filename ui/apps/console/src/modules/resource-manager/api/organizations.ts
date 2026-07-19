import {loggedFetch} from '../../../platform/http/loggedFetch'

const organizationStates = new Set<Organization['state']>([
  'CREATING',
  'ACTIVE',
  'SUSPENDED',
  'DELETING',
  'DELETED',
  'FAILED',
  'STATE_UNSPECIFIED',
])

export interface Organization {
  id: string
  state: 'CREATING' | 'ACTIVE' | 'SUSPENDED' | 'DELETING' | 'DELETED' | 'FAILED' | 'STATE_UNSPECIFIED'
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
  pageSize: number
  pageToken?: string
  filter?: string
  orderBy?: string
  signal?: AbortSignal
}

export async function fetchOrganizations({
  pageSize,
  pageToken,
  filter,
  orderBy = 'name asc',
  signal,
}: FetchOrganizationsOptions): Promise<ListOrganizationsResponse> {
  const apiBaseUrl = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
  const parameters = new URLSearchParams({
    pageSize: String(pageSize),
    orderBy,
    showDeleted: 'false',
  })
  if (pageToken) parameters.set('pageToken', pageToken)
  if (filter) parameters.set('filter', filter)
  const response = await loggedFetch(
    'resource-manager',
    `${apiBaseUrl}/resource-manager/v1/organizations?${parameters}`,
    {
      headers: {Accept: 'application/json'},
      credentials: 'same-origin',
      signal,
    },
  )
  if (!response.ok) {
    throw new Error(`Resource Manager returned HTTP ${response.status}`)
  }
  return parseListOrganizationsResponse(await response.json())
}

function parseListOrganizationsResponse(value: unknown): ListOrganizationsResponse {
  if (!isRecord(value)) throw new Error('Resource Manager returned an invalid organizations response')

  const rawOrganizations = value.organizations ?? []
  if (!Array.isArray(rawOrganizations)) {
    throw new Error('Resource Manager returned an invalid organizations collection')
  }

  const organizations = rawOrganizations.map((organization, index) => parseOrganization(organization, index))
  const totalSize =
    typeof value.totalSize === 'number' && Number.isInteger(value.totalSize) && value.totalSize >= 0
      ? value.totalSize
      : organizations.length

  return {
    organizations,
    totalSize,
    nextPageToken: typeof value.nextPageToken === 'string' && value.nextPageToken ? value.nextPageToken : undefined,
  }
}

function parseOrganization(value: unknown, index: number): Organization {
  if (!isRecord(value) || typeof value.id !== 'string' || value.id.length === 0) {
    throw new Error(`Resource Manager returned an invalid organization at index ${index}`)
  }

  const state = organizationStates.has(value.state as Organization['state'])
    ? (value.state as Organization['state'])
    : 'STATE_UNSPECIFIED'

  return {
    id: value.id,
    state,
    name: optionalString(value.name),
    description: optionalString(value.description),
    createTime: optionalString(value.createTime),
    updateTime: optionalString(value.updateTime),
    version:
      typeof value.version === 'string' || typeof value.version === 'number' ? value.version : undefined,
    labels: isStringRecord(value.labels) ? value.labels : undefined,
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
