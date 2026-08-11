'use client';

import {createContext, useContext, type ReactNode} from 'react';
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';

import {M8QueryRuntime} from '../registry';

const RuntimeContext = createContext<M8QueryRuntime | null>(null);

export function QueryProvider({
  runtime,
  queryClient,
  children,
}: {
  runtime: M8QueryRuntime;
  queryClient: QueryClient;
  children: ReactNode;
}) {
  return (
    <QueryClientProvider client={queryClient}>
      <RuntimeContext.Provider value={runtime}>{children}</RuntimeContext.Provider>
    </QueryClientProvider>
  );
}

export function useM8Query(queryId: string, input: unknown, enabled = true) {
  const runtime = useContext(RuntimeContext);
  if (!runtime) throw new Error('useM8Query must be used within QueryProvider.');
  return useQuery({
    queryKey: runtime.queryKey(queryId, input),
    queryFn: ({signal}) => runtime.execute(queryId, input, signal),
    enabled,
  });
}

export function useInvalidateM8Query() {
  const runtime = useContext(RuntimeContext);
  const queryClient = useQueryClient();
  if (!runtime) throw new Error('useInvalidateM8Query must be used within QueryProvider.');
  return (queryId: string) => {
    const definition = runtime.registry.require(queryId);
    return queryClient.invalidateQueries({queryKey: [definition.id]});
  };
}

export function invalidateQuery(
  queryClient: QueryClient,
  runtime: M8QueryRuntime,
  queryId: string,
) {
  runtime.registry.require(queryId);
  return queryClient.invalidateQueries({queryKey: [queryId]});
}

export {QueryClient};
