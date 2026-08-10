'use client';

import type {ReactNode} from 'react';

import {NextAppProvider} from '@json-render/next';
import {
    ThemeProvider,
    ToasterComponent,
    ToasterProvider,
} from '@gravity-ui/uikit';
import {toaster} from '@gravity-ui/uikit/toaster-singleton';

import {registry} from '@/ui/registry/registry';
import {handlers} from '@/ui/registry/handlers';

export function Providers({
                              children,
                          }: {
    children: ReactNode;
}) {
    return (
        <ThemeProvider theme="light">
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
    );
}
