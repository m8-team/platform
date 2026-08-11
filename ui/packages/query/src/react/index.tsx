'use client';

import {createContext, useContext, type ReactNode} from 'react';
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
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
    queryKey: runtime.queryKey(queryId, input, context),
    queryFn: ({signal}) => runtime.execute(queryId, input, signal, context),
    enabled,
  });
}

export {QueryClient};
