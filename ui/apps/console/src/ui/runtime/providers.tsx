'use client';

import {useState, type ReactNode} from 'react';

import {NextAppProvider} from '@json-render/next';
import {
    ThemeProvider,
    ToasterComponent,
    ToasterProvider,
} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';

import {registry} from '@/ui/registry/registry';
import {handlers} from '@/ui/registry/handlers';
import {
    AppThemeContext,
    type AppTheme,
} from '@/ui/runtime/theme-context';

export function Providers({
                              children,
                          }: {
    children: ReactNode;
}) {
    const [theme, setTheme] = useState<AppTheme>('light');

    return (
        <AppThemeContext.Provider value={{theme, setTheme}}>
            <ThemeProvider theme={theme}>
                <ToasterProvider toaster={toaster}>
                    <NextAppProvider
                        registry={registry}
                        handlers={handlers}
                    >
                        {children}
                    </NextAppProvider>
                    <ToasterComponent />
                </ToasterProvider>
            </ThemeProvider>
        </AppThemeContext.Provider>
    );
}
