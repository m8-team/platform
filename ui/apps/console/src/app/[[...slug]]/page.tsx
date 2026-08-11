import {PageRenderer} from '@json-render/next';
import {matchRoute, slugToPath} from '@json-render/next/server';
import {notFound} from 'next/navigation';

import {generateMetadata, generateStaticParams, getPageData,} from '@/platform/app';
import {appSpec} from '@/platform/specs/app';

export {generateMetadata, generateStaticParams};

export default async function Page({
                                     params,
                                   }: {
  params: Promise<{ slug?: string[] }>;
}) {
  const resolvedParams = await params;
  const data = await getPageData({params: Promise.resolve(resolvedParams)});

  if (!data) {
    notFound();
  }

  const matched = matchRoute(appSpec, slugToPath(resolvedParams.slug));
  const runtime = data.initialState?.__runtime;
  const runtimeState = runtime && typeof runtime === 'object' ? runtime : {};

  return <PageRenderer {...data} initialState={{
    ...data.initialState,
    __runtime: {...runtimeState, params: matched?.params ?? {}},
  }} />;
}
