'use client';

import {type ReactNode, useMemo, useState} from 'react';

import {ToasterComponent, ToasterProvider} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';
import {QueryClient} from '@m8/query/react';
import {createRuntime, RuntimeProvider} from '@m8/runtime';

import {moduleRegistry} from '@/platform/modules/registry';
import {registry} from '@/platform/registry/registry';
import {ThemeContext, type Theme} from '@/platform/runtime/theme-context';

const runtimeContext = {
  permissions: ['resource-manager.read', 'resource-manager.projects.delete'],
  features: ['resource-manager'],
  edition: 'community',
} as const;

export function Providers({children}: {children: ReactNode}) {
  const [theme, setTheme] = useState<Theme>('light');
  const [queryClient] = useState(() => new QueryClient());
  const runtime = useMemo(
    () => createRuntime({modules: moduleRegistry, catalog: registry, adapters: {
      authorization: {
        can: async ({permission, context}) =>
          context.permissions?.includes(permission) ?? false,
      },
      confirmation: {
        confirm: async ({confirmation}) => globalThis.confirm(confirmation.title),
      },
      queryInvalidation: {
        invalidate: queryId => queryClient.invalidateQueries({queryKey: [queryId]}).then(() => undefined),
      },
      longRunningOperations: {
        wait: async (operationId, options) => {
          options?.signal?.throwIfAborted();
          return {id: operationId, status: 'SUCCEEDED'};
        },
      },
      errorReporter: {report: error => console.error('M8 runtime secondary effect failed', error)},
    }}),
    [queryClient],
  );

  return (
    <ThemeContext.Provider value={{theme, setTheme}}>
      <RuntimeProvider
        modules={moduleRegistry}
        registry={registry}
        queryClient={queryClient}
        queries={runtime.queries}
        operations={runtime.operations}
        authorization={runtime.authorization}
        context={runtimeContext}
        theme={theme}
      >
        <ToasterProvider toaster={toaster}>
            {children}
          <ToasterComponent />
        </ToasterProvider>
      </RuntimeProvider>
    </ThemeContext.Provider>
  );
}
