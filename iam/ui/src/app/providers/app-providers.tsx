import * as React from 'react';

import {QueryClientProvider} from '@tanstack/react-query';
import {Lang as DynamicFormsLang, configure as configureDynamicForms} from '@gravity-ui/dynamic-forms';
import {settings as dateSettings} from '@gravity-ui/date-utils';
import {
  Lang as UIKitLang,
  ThemeProvider, ToasterComponent,
  ToasterProvider,
  configure as configureUIKit,
} from '@gravity-ui/uikit';

import {appQueryClient} from '@/app/providers/app-query-client';
import {appToaster} from '@/app/providers/app-toaster';
import {AppUIContext, defaultAppContext} from '@/app/providers/app-ui-context';

configureUIKit({lang: UIKitLang.Ru});
configureDynamicForms({lang: DynamicFormsLang.Ru});
void dateSettings.loadLocale('ru');

export function AppProviders({children}: React.PropsWithChildren) {
  const [context, setContext] = React.useState(defaultAppContext);
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
