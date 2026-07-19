import {keepPreviousData, queryOptions, useQuery} from '@tanstack/react-query'

import {fetchOrganizations} from '../api/organizations'

export const organizationQueryKeys = {
  all: ['resource-manager', 'organizations'] as const,
  list: (parameters: OrganizationsQueryParameters) => [...organizationQueryKeys.all, 'list', parameters] as const,
}

export interface OrganizationsQueryParameters {
  pageSize: number
  pageToken?: string
}

export function organizationsQueryOptions(parameters: OrganizationsQueryParameters) {
  return queryOptions({
    queryKey: organizationQueryKeys.list(parameters),
    queryFn: ({signal}) => fetchOrganizations({...parameters, signal}),
    placeholderData: keepPreviousData,
    staleTime: 30_000,
  })
}

export function useOrganizationsQuery(parameters: OrganizationsQueryParameters) {
  return useQuery(organizationsQueryOptions(parameters))
}
