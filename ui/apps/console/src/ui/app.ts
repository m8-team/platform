import {createNextApp} from '@json-render/next/server';

import {appSpec} from '@/ui/specs/app';

export const {
    Page,
    generateMetadata,
    generateStaticParams,
} = createNextApp({
    spec: appSpec,

    loaders: {},
});
