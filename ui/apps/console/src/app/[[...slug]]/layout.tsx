import type {ReactNode} from 'react';

import '@gravity-ui/uikit/styles/fonts.css';
import '@gravity-ui/uikit/styles/styles.css';

import {Providers} from '@/ui/runtime/providers';

export default function Layout({
                                   children,
                               }: {
    children: ReactNode;
}) {
    return (
        <html lang="ru">
        <body>
        <Providers>
            {children}
        </Providers>
        </body>
        </html>
    );
}
