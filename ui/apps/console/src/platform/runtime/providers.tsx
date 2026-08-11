'use client';

import {type ReactNode, useMemo, useState} from 'react';

import {ToasterComponent, ToasterProvider} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';
import {M8OperationRuntime, OperationRegistry} from '@m8/operation';
import {M8QueryRuntime, QueryRegistry} from '@m8/query';
import {QueryClient} from '@m8/query/react';
import {M8RuntimeProvider} from '@m8/runtime';
import {
  createProjectOperation,
  deleteProjectOperation,
  getOrganizationQuery,
  getProjectQuery,
  listOrganizationsQuery,
  listProjectsQuery,
} from '@m8/resource-manager';

import {moduleRegistry} from '@/platform/modules/registry';
import {registry} from '@/platform/registry/registry';
import {ThemeContext, type Theme} from '@/platform/runtime/theme-context';

const runtimeContext = {
  permissions: ['resource-manager.read', 'resource-manager.projects.delete'],
  features: ['resource-manager'],
  edition: 'community',
} as const;

const queryRuntime = new M8QueryRuntime(new QueryRegistry([
  listOrganizationsQuery,
  getOrganizationQuery,
  listProjectsQuery,
  getProjectQuery,
]));

const operationRegistry = new OperationRegistry([
  createProjectOperation,
  deleteProjectOperation,
]);

export function Providers({children}: {children: ReactNode}) {
  const [theme, setTheme] = useState<Theme>('light');
  const [queryClient] = useState(() => new QueryClient());
  const operationRuntime = useMemo(
    () => new M8OperationRuntime(operationRegistry, {
      authorization: {
        check: async ({permission, context}) =>
          context.permissions?.includes(permission) ?? false,
      },
      confirmation: {
        confirm: async ({confirmation}) => globalThis.confirm(confirmation.title),
      },
      queryInvalidation: {
        invalidate: queryId => queryClient.invalidateQueries({queryKey: [queryId]}).then(() => undefined),
      },
    }),
    [queryClient],
  );

  return (
    <ThemeContext.Provider value={{theme, setTheme}}>
      <M8RuntimeProvider
        modules={moduleRegistry}
        registry={registry}
        queryClient={queryClient}
        queries={queryRuntime}
        operations={operationRuntime}
        context={runtimeContext}
        theme={theme}
      >
        <ToasterProvider toaster={toaster}>
            {children}
          <ToasterComponent />
        </ToasterProvider>
      </M8RuntimeProvider>
    </ThemeContext.Provider>
  );
}
