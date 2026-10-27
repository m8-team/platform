import {queryOptions, useQuery} from '@tanstack/react-query'

import {fetchOrganizations} from '../api/organizations'

export const organizationQueryKeys = {
  all: ['resource-manager', 'organizations'] as const,
  list: (parameters: OrganizationsQueryParameters) => [...organizationQueryKeys.all, 'list', parameters] as const,
}

export interface OrganizationsQueryParameters {
  tenantId: string
  pageSize: number
  pageToken?: string
  filter?: string
  stateFilter?: import('../api/organizations').Organization['state']
  orderBy: string
}

export function organizationsQueryOptions(parameters: OrganizationsQueryParameters) {
  return queryOptions({
    queryKey: organizationQueryKeys.list(parameters),
    queryFn: ({signal}) => fetchOrganizations({...parameters, signal}),
    staleTime: 30_000,
    enabled: parameters.tenantId.length > 0,
  })
}

export function useOrganizationsQuery(parameters: OrganizationsQueryParameters) {
  return useQuery(organizationsQueryOptions(parameters))
}
