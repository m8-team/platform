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
  organizations?: Organization[]
  nextPageToken?: string
  totalSize?: number
}

export interface FetchOrganizationsOptions {
  pageSize: number
  pageToken?: string
  signal?: AbortSignal
}

export async function fetchOrganizations({pageSize, pageToken, signal}: FetchOrganizationsOptions): Promise<ListOrganizationsResponse> {
  const apiBaseUrl = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
  const parameters = new URLSearchParams({
    pageSize: String(pageSize),
    orderBy: 'name',
    showDeleted: 'false',
  })
  if (pageToken) parameters.set('pageToken', pageToken)
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
  return (await response.json()) as ListOrganizationsResponse
}
import {loggedFetch} from '../../../platform/http/serviceRequestLog'
