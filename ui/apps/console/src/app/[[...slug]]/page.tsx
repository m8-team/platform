import {PageRenderer} from '@json-render/next';
import {notFound} from 'next/navigation';

import {
  generateMetadata,
  generateStaticParams,
  getPageData,
} from '@/platform/app';

export {generateMetadata, generateStaticParams};

export default async function Page(props: {
  params: Promise<{slug?: string[]}>;
}) {
  const data = await getPageData(props);
  if (!data) notFound();
  return <PageRenderer {...data} />;
}
