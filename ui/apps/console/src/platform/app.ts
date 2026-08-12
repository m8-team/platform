import {createNextApp} from '@json-render/next/server';
import {runtimeNextLoaders} from '@m8/runtime';

import {appSpec} from '@/platform/specs/app';

export const {generateMetadata, generateStaticParams, getPageData} = createNextApp({
  spec: appSpec,
  loaders: runtimeNextLoaders,
});
