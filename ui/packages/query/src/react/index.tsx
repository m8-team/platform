'use client';

import {createContext, useContext, type ReactNode} from 'react';
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';

import {QueryRuntime} from '../registry';
import type {RuntimeContext} from '@m8/core';

const RuntimeContext = createContext<QueryRuntime | null>(null);

export function QueryProvider({
  runtime,
  queryClient,
  children,
}: {
  runtime: QueryRuntime;
  queryClient: QueryClient;
  children: ReactNode;
}) {
  return (
    <QueryClientProvider client={queryClient}>
      <RuntimeContext.Provider value={runtime}>{children}</RuntimeContext.Provider>
    </QueryClientProvider>
  );
}

export function useRegisteredQuery(queryId: string, input: unknown, enabled = true, context: RuntimeContext = {}) {
  const runtime = useContext(RuntimeContext);
  if (!runtime) throw new Error('useRegisteredQuery must be used within QueryProvider.');
  return useQuery({
    queryKey: runtime.queryKey(queryId, input),
    queryFn: ({signal}) => runtime.execute(queryId, input, signal, context),
    enabled,
  });
}

export function useInvalidateQuery() {
  const runtime = useContext(RuntimeContext);
  const queryClient = useQueryClient();
  if (!runtime) throw new Error('useInvalidateQuery must be used within QueryProvider.');
  return (queryId: string) => {
    const definition = runtime.registry.require(queryId);
    return queryClient.invalidateQueries({queryKey: [definition.id]});
  };
}

export function invalidateQuery(
  queryClient: QueryClient,
  runtime: QueryRuntime,
  queryId: string,
) {
  runtime.registry.require(queryId);
  return queryClient.invalidateQueries({queryKey: [queryId]});
}

export {QueryClient};
