'use client';

import type {ReactNode} from 'react';

import {NextAppProvider} from '@json-render/next';
import {ThemeProvider} from '@gravity-ui/uikit';

import {registry} from '@/ui/registry/registry';
import {handlers} from '@/ui/registry/handlers';

export function Providers({
                              children,
                          }: {
    children: ReactNode;
}) {
    return (
        <ThemeProvider theme="light">
            <NextAppProvider
                registry={registry}
                handlers={handlers}
            >
                {children}
            </NextAppProvider>
        </ThemeProvider>
    );
}
