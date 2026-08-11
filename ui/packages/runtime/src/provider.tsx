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
import type {ModuleRegistry, ModuleDefinition, RuntimeContext} from '@m8/core';
import type {OperationRuntime} from '@m8/operation';
import {QueryRuntime} from '@m8/query';
import {QueryProvider, type QueryClient} from '@m8/query/react';

import {createActionHandlers} from './actions';
import type {RuntimeAuthorizationAdapter} from './create-runtime';
import {RouteRuntimeBoundary} from './route-runtime';

interface RuntimeValue {
  readonly modules: ModuleRegistry<readonly ModuleDefinition[]>;
  readonly context: RuntimeContext;
  readonly authorization?: RuntimeAuthorizationAdapter;
}

const RuntimeContext = createContext<RuntimeValue | null>(null);

export interface RuntimeProviderProps {
  readonly modules: ModuleRegistry<readonly ModuleDefinition[]>;
  readonly registry: ComponentRegistry;
  readonly queryClient: QueryClient;
  readonly queries: QueryRuntime;
  readonly operations: OperationRuntime;
  readonly context: RuntimeContext;
  readonly authorization?: RuntimeAuthorizationAdapter;
  readonly navigate?: (href: string) => void;
  readonly theme?: 'light' | 'dark' | 'light-hc' | 'dark-hc';
  readonly children: ReactNode;
}

export function RuntimeProvider({
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
}: RuntimeProviderProps) {
  const contextRef = useRef(context);
  contextRef.current = context;
  const handlers = useMemo(
    () => createActionHandlers({
      operations,
      queryClient,
      getContext: () => contextRef.current,
      navigate,
    }),
    [navigate, operations, queryClient],
  );

  const runtimeRegistry = useMemo(() => ({
    ...registry,
    __RouteRuntime: ({element, children}: {element: {props: Parameters<typeof RouteRuntimeBoundary>[0]}; children?: ReactNode}) =>
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

export function useRuntime(): RuntimeValue {
  const value = useContext(RuntimeContext);
  if (!value) throw new Error('useRuntime must be used within RuntimeProvider.');
  return value;
}
