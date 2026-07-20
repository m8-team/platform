import {queryOptions, useQuery} from '@tanstack/react-query'

import {fetchAllWorkspaces} from '../api/workspaces'

export const workspaceQueryKeys = {
  all: ['resource-manager', 'workspaces'] as const,
  list: () => [...workspaceQueryKeys.all, 'list'] as const,
}

export function workspacesQueryOptions() {
  return queryOptions({
    queryKey: workspaceQueryKeys.list(),
    queryFn: ({signal}) => fetchAllWorkspaces(signal),
    staleTime: 30_000,
  })
}

export function useWorkspacesQuery() {
  return useQuery(workspacesQueryOptions())
}
