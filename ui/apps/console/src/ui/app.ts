import {createNextApp} from '@json-render/next/server';

import {appSpec} from '@/ui/specs/app';

export const {
  getPageData,
  generateMetadata,
  generateStaticParams,
} = createNextApp({
  spec: appSpec,

  loaders: {},
});
