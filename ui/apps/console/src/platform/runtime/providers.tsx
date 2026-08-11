'use client';

import {type ReactNode, useState} from 'react';

import {NextAppProvider} from '@json-render/next';
import {ThemeProvider, ToasterComponent, ToasterProvider,} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';

import {registry} from '@/platform/registry/registry';
import {ThemeContext, type Theme} from '@/platform/runtime/theme-context';

export function Providers({children}: {children: ReactNode}) {
  const [theme, setTheme] = useState<Theme>('light');

  return (
    <ThemeContext.Provider value={{theme, setTheme}}>
      <ThemeProvider theme={theme}>
        <ToasterProvider toaster={toaster}>
          <NextAppProvider registry={registry}>
            {children}
          </NextAppProvider>
          <ToasterComponent />
        </ToasterProvider>
      </ThemeProvider>
    </ThemeContext.Provider>
  );
}
