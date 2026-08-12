'use client';

import {type ReactNode, useMemo, useState} from 'react';

import {
  ThemeProvider,
  ToasterComponent,
  ToasterProvider,
} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';
import {QueryClient} from '@m8/query/react';
import {createRuntime, RuntimeProvider} from '@m8/runtime';

import {moduleRegistry} from '@/platform/modules/registry';
import {registry} from '@/platform/registry/registry';
import {ThemeContext, type Theme} from '@/platform/runtime/theme-context';

const runtimeContext = {
  permissions: [
    'resource-manager.overview.read',
    'resource-manager.organizations.read',
    'resource-manager.projects.read',
    'resource-manager.projects.create',
    'resource-manager.projects.delete',
  ],
  features: ['resource-manager'],
  edition: 'community',
} as const;

export function Providers({children}: {children: ReactNode}) {
  const [theme, setTheme] = useState<Theme>('light');
  const [queryClient] = useState(() => new QueryClient());
  const runtime = useMemo(
    () => createRuntime({modules: moduleRegistry, adapters: {
      authorization: {
        can: async ({permission, context}) =>
          context.permissions?.includes(permission) ?? false,
      },
      queryInvalidation: {
        invalidate: queryId => queryClient.invalidateQueries({queryKey: ['query', queryId]}).then(() => undefined),
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
      <ThemeProvider theme={theme}>
        <RuntimeProvider
          runtime={runtime}
          registry={registry}
          queryClient={queryClient}
          context={runtimeContext}
        >
          <ToasterProvider toaster={toaster}>
            {children}
            <ToasterComponent />
          </ToasterProvider>
        </RuntimeProvider>
      </ThemeProvider>
    </ThemeContext.Provider>
  );
}
