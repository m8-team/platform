import {loggedFetch} from '../../../platform/http/loggedFetch'

const workspaceStates = new Set<Workspace['state']>([
  'CREATING',
  'ACTIVE',
  'SUSPENDED',
  'DELETING',
  'DELETED',
  'FAILED',
  'STATE_UNSPECIFIED',
])

export interface Workspace {
  id: string
  organizationId: string
  organizationName?: string
  state: 'CREATING' | 'ACTIVE' | 'SUSPENDED' | 'DELETING' | 'DELETED' | 'FAILED' | 'STATE_UNSPECIFIED'
  name?: string
  description?: string
  createTime?: string
  updateTime?: string
  version?: string | number
  labels?: Record<string, string>
}

export async function fetchAllWorkspaces(
  organizationIds: string[],
  signal?: AbortSignal,
): Promise<ListWorkspacesResponse> {
  const pages = await Promise.all(organizationIds.map((organizationId) => fetchOrganizationWorkspaces(organizationId, signal)))
  const workspaces = pages.flat()
  return {workspaces, totalSize: workspaces.length}
}

async function fetchOrganizationWorkspaces(organizationId: string, signal?: AbortSignal) {
  const result: Workspace[] = []
  let pageToken: string | undefined
  do {
    const page = await fetchWorkspaces({organizationId, pageSize: 1000, pageToken, orderBy: 'name asc', signal})
    result.push(...page.workspaces)
    pageToken = page.nextPageToken
  } while (pageToken)
  return result
}

export interface ListWorkspacesResponse {
  workspaces: Workspace[]
  nextPageToken?: string
  totalSize: number
}

export interface FetchWorkspacesOptions {
  organizationId: string
  pageSize: number
  pageToken?: string
  filter?: string
  orderBy?: string
  signal?: AbortSignal
}

export async function fetchWorkspaces({
  organizationId,
  pageSize,
  pageToken,
  filter,
  orderBy = 'name asc',
  signal,
}: FetchWorkspacesOptions): Promise<ListWorkspacesResponse> {
  const apiBaseUrl = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
  const parameters = new URLSearchParams({
    organizationId,
    pageSize: String(pageSize),
    orderBy,
    showDeleted: 'false',
  })
  if (pageToken) parameters.set('pageToken', pageToken)
  if (filter) parameters.set('filter', filter)

  const response = await loggedFetch(
    'resource-manager',
    `${apiBaseUrl}/resource-manager/v1/workspaces?${parameters}`,
    {headers: {Accept: 'application/json'}, credentials: 'same-origin', signal},
  )
  if (!response.ok) throw new Error(`Resource Manager returned HTTP ${response.status}`)
  return parseListWorkspacesResponse(await response.json())
}

function parseListWorkspacesResponse(value: unknown): ListWorkspacesResponse {
  if (!isRecord(value)) throw new Error('Resource Manager returned an invalid workspaces response')
  const rawWorkspaces = value.workspaces ?? []
  if (!Array.isArray(rawWorkspaces)) throw new Error('Resource Manager returned an invalid workspaces collection')
  const workspaces = rawWorkspaces.map(parseWorkspace)
  const totalSize =
    typeof value.totalSize === 'number' && Number.isInteger(value.totalSize) && value.totalSize >= 0
      ? value.totalSize
      : workspaces.length
  return {
    workspaces,
    totalSize,
    nextPageToken: typeof value.nextPageToken === 'string' && value.nextPageToken ? value.nextPageToken : undefined,
  }
}

function parseWorkspace(value: unknown, index: number): Workspace {
  if (!isRecord(value) || typeof value.id !== 'string' || !value.id) {
    throw new Error(`Resource Manager returned an invalid workspace at index ${index}`)
  }
  const state = workspaceStates.has(value.state as Workspace['state'])
    ? (value.state as Workspace['state'])
    : 'STATE_UNSPECIFIED'
  return {
    id: value.id,
    organizationId: typeof value.organizationId === 'string' ? value.organizationId : '',
    state,
    name: optionalString(value.name),
    description: optionalString(value.description),
    createTime: optionalString(value.createTime),
    updateTime: optionalString(value.updateTime),
    version: typeof value.version === 'string' || typeof value.version === 'number' ? value.version : undefined,
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
