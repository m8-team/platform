import {queryOptions, useQuery} from '@tanstack/react-query'

import {fetchAllWorkspaces} from '../api/workspaces'

export interface WorkspacesQueryParameters {
  organizationIds: string[]
}

export const workspaceQueryKeys = {
  all: ['resource-manager', 'workspaces'] as const,
  list: (parameters: WorkspacesQueryParameters) => [...workspaceQueryKeys.all, 'list', parameters] as const,
}

export function workspacesQueryOptions(parameters: WorkspacesQueryParameters) {
  return queryOptions({
    queryKey: workspaceQueryKeys.list(parameters),
    queryFn: ({signal}) => fetchAllWorkspaces(parameters.organizationIds, signal),
    enabled: parameters.organizationIds.length > 0,
    staleTime: 30_000,
  })
}

export function useWorkspacesQuery(parameters: WorkspacesQueryParameters) {
  return useQuery(workspacesQueryOptions(parameters))
}
