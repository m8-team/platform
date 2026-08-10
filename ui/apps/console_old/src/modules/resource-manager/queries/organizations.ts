import {queryOptions, useQuery} from '@tanstack/react-query'

import {fetchOrganizations, fetchOrganizationsByIds} from '../api/organizations'

export const organizationQueryKeys = {
  all: ['resource-manager', 'organizations'] as const,
  list: (parameters: OrganizationsQueryParameters) => [...organizationQueryKeys.all, 'list', parameters] as const,
  byIds: (ids: string[]) => [...organizationQueryKeys.all, 'by-ids', ids] as const,
}

export function useOrganizationsByIdsQuery(ids: string[]) {
  return useQuery({
    queryKey: organizationQueryKeys.byIds(ids),
    queryFn: ({signal}) => fetchOrganizationsByIds(ids, signal),
    enabled: ids.length > 0,
    staleTime: 30_000,
  })
}

export interface OrganizationsQueryParameters {
  pageSize: number
  pageToken?: string
  filter?: string
  orderBy: string
}

export function organizationsQueryOptions(parameters: OrganizationsQueryParameters) {
  return queryOptions({
    queryKey: organizationQueryKeys.list(parameters),
    queryFn: ({signal}) => fetchOrganizations({...parameters, signal}),
    staleTime: 30_000,
  })
}

export function useOrganizationsQuery(parameters: OrganizationsQueryParameters) {
  return useQuery(organizationsQueryOptions(parameters))
}
