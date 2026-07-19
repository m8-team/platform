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
  signal?: AbortSignal
}

export async function fetchOrganizations({signal}: FetchOrganizationsOptions = {}): Promise<ListOrganizationsResponse> {
  const apiBaseUrl = (import.meta.env.VITE_RESOURCE_MANAGER_API_URL ?? '').replace(/\/$/, '')
  const response = await loggedFetch(
    'resource-manager',
    `${apiBaseUrl}/resource-manager/v1/organizations?pageSize=100&orderBy=name&showDeleted=false`,
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
