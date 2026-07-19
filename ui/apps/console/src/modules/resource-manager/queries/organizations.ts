import {queryOptions, useQuery} from '@tanstack/react-query'

import {fetchOrganizations} from '../api/organizations'

export const organizationQueryKeys = {
  all: ['resource-manager', 'organizations'] as const,
  list: () => [...organizationQueryKeys.all, 'list'] as const,
}

export function organizationsQueryOptions() {
  return queryOptions({
    queryKey: organizationQueryKeys.list(),
    queryFn: ({signal}) => fetchOrganizations({signal}),
    staleTime: 30_000,
  })
}

export function useOrganizationsQuery() {
  return useQuery(organizationsQueryOptions())
}
