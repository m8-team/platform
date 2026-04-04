import * as React from 'react';

import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {Lang as DynamicFormsLang, configure as configureDynamicForms} from '@gravity-ui/dynamic-forms';
import {settings as dateSettings} from '@gravity-ui/date-utils';
import {
  Lang as UIKitLang,
  ThemeProvider,
  Toaster,
  ToasterComponent,
  ToasterProvider,
  configure as configureUIKit,
} from '@gravity-ui/uikit';

import {env} from '@/shared/config/env';
import type {AppContextSelection} from '@/shared/types/iam';

configureUIKit({lang: UIKitLang.Ru});
configureDynamicForms({lang: DynamicFormsLang.Ru});
void dateSettings.loadLocale('ru');

export const appQueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
  },
});

export const appToaster = new Toaster();

type AppUIContextValue = {
  context: AppContextSelection;
  globalSearch: string;
  navCompact: boolean;
  setContext: React.Dispatch<React.SetStateAction<AppContextSelection>>;
  setGlobalSearch: React.Dispatch<React.SetStateAction<string>>;
  setNavCompact: React.Dispatch<React.SetStateAction<boolean>>;
};

const AppUIContext = React.createContext<AppUIContextValue | null>(null);

const defaultContext: AppContextSelection = {
  tenantId: env.defaultTenantId,
  organizationId: 'org-1',
  environment: 'prod',
  region: 'eu-central',
};

export function AppProviders({children}: React.PropsWithChildren) {
  const [context, setContext] = React.useState<AppContextSelection>(defaultContext);
  const [globalSearch, setGlobalSearch] = React.useState('');
  const [navCompact, setNavCompact] = React.useState(false);

  return (
    <QueryClientProvider client={appQueryClient}>
      <ToasterProvider toaster={appToaster}>
        <ThemeProvider theme="light" lang="ru">
          <AppUIContext.Provider
            value={{
              context,
              globalSearch,
              navCompact,
              setContext,
              setGlobalSearch,
              setNavCompact,
            }}
          >
            {children}
            <ToasterComponent />
          </AppUIContext.Provider>
        </ThemeProvider>
      </ToasterProvider>
    </QueryClientProvider>
  );
}

export function useAppUI() {
  const context = React.useContext(AppUIContext);

  if (!context) {
    throw new Error('useAppUI must be used within AppProviders');
  }

  return context;
}
