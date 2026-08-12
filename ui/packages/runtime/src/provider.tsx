'use client';

import {
  createContext,
  useContext,
  useMemo,
  useRef,
  type ReactNode,
} from 'react';
import {NextAppProvider} from '@json-render/next';
import type {ComponentRegistry} from '@json-render/react';
import type {RuntimeContext} from '@m8/core';
import {QueryProvider, type QueryClient} from '@m8/query/react';

import {createActionHandlers} from './actions';
import type {Runtime} from './create-runtime';
import {RouteRuntimeBoundary} from './route-runtime';

export interface RuntimeValue {
  readonly runtime: Runtime;
  readonly context: RuntimeContext;
}

const RuntimeContextValue = createContext<RuntimeValue | null>(null);

export interface RuntimeProviderProps {
  readonly runtime: Runtime;
  readonly registry: ComponentRegistry;
  readonly queryClient: QueryClient;
  readonly context: RuntimeContext;
  readonly children: ReactNode;
}

export function RuntimeProvider({
  runtime,
  registry,
  queryClient,
  context,
  children,
}: RuntimeProviderProps) {
  const contextRef = useRef(context);
  contextRef.current = context;
  const handlers = useMemo(
    () => createActionHandlers({
      operations: runtime.operations,
      getContext: () => contextRef.current,
    }),
    [runtime.operations],
  );

  // __RouteRuntime is an internal integration renderer. The application-owned
  // json-render registry remains the component registry and source of truth.
  const runtimeRegistry = useMemo(() => ({
    ...registry,
    __RouteRuntime: ({element, children}: {
      element: {props: Parameters<typeof RouteRuntimeBoundary>[0]};
      children?: ReactNode;
    }) => <RouteRuntimeBoundary {...element.props}>{children}</RouteRuntimeBoundary>,
  }), [registry]);

  return (
    <RuntimeContextValue.Provider value={{runtime, context}}>
      <QueryProvider runtime={runtime.queries} queryClient={queryClient}>
        <NextAppProvider registry={runtimeRegistry} handlers={handlers}>
          {children}
        </NextAppProvider>
      </QueryProvider>
    </RuntimeContextValue.Provider>
  );
}

export function useRuntime(): RuntimeValue {
  const value = useContext(RuntimeContextValue);
  if (!value) throw new Error('useRuntime must be used within RuntimeProvider.');
  return value;
}
