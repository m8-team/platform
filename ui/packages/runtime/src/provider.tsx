'use client';

import {
  createContext,
  useContext,
  useMemo,
  useRef,
  type ReactNode,
} from 'react';
import {ThemeProvider} from '@gravity-ui/uikit';
import {NextAppProvider} from '@json-render/next';
import type {ComponentRegistry} from '@json-render/react';
import type {ModuleRegistry, M8ModuleDefinition, M8RuntimeContext} from '@m8/core';
import type {M8OperationRuntime} from '@m8/operation';
import {M8QueryRuntime} from '@m8/query';
import {QueryProvider, type QueryClient} from '@m8/query/react';

import {createM8ActionHandlers} from './actions';
import type {RuntimeAuthorizationAdapter} from './create-runtime';
import {RouteRuntimeBoundary} from './route-runtime';

interface M8RuntimeValue {
  readonly modules: ModuleRegistry<readonly M8ModuleDefinition[]>;
  readonly context: M8RuntimeContext;
  readonly authorization?: RuntimeAuthorizationAdapter;
}

const RuntimeContext = createContext<M8RuntimeValue | null>(null);

export interface M8RuntimeProviderProps {
  readonly modules: ModuleRegistry<readonly M8ModuleDefinition[]>;
  readonly registry: ComponentRegistry;
  readonly queryClient: QueryClient;
  readonly queries: M8QueryRuntime;
  readonly operations: M8OperationRuntime;
  readonly context: M8RuntimeContext;
  readonly authorization?: RuntimeAuthorizationAdapter;
  readonly navigate?: (href: string) => void;
  readonly theme?: 'light' | 'dark' | 'light-hc' | 'dark-hc';
  readonly children: ReactNode;
}

export function M8RuntimeProvider({
  modules,
  registry,
  queryClient,
  queries,
  operations,
  context,
  authorization,
  navigate,
  theme = 'light',
  children,
}: M8RuntimeProviderProps) {
  const contextRef = useRef(context);
  contextRef.current = context;
  const handlers = useMemo(
    () => createM8ActionHandlers({
      operations,
      queryClient,
      getContext: () => contextRef.current,
      navigate,
    }),
    [navigate, operations, queryClient],
  );

  const runtimeRegistry = useMemo(() => ({
    ...registry,
    __M8RouteRuntime: ({element, children}: {element: {props: Parameters<typeof RouteRuntimeBoundary>[0]}; children?: ReactNode}) =>
      <RouteRuntimeBoundary {...element.props}>{children}</RouteRuntimeBoundary>,
  }), [registry]);
  return (
    <RuntimeContext.Provider value={{modules, context, authorization}}>
      <ThemeProvider theme={theme}>
        <QueryProvider runtime={queries} queryClient={queryClient}>
          <NextAppProvider registry={runtimeRegistry} handlers={handlers}>
            {children}
          </NextAppProvider>
        </QueryProvider>
      </ThemeProvider>
    </RuntimeContext.Provider>
  );
}

export function useM8Runtime(): M8RuntimeValue {
  const value = useContext(RuntimeContext);
  if (!value) throw new Error('useM8Runtime must be used within M8RuntimeProvider.');
  return value;
}
