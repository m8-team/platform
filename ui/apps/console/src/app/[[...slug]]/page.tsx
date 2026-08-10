import {PageRenderer} from '@json-render/next';
import {notFound} from 'next/navigation';

import {
    generateMetadata,
    generateStaticParams,
    getPageData,
} from '@/ui/app';

export {generateMetadata, generateStaticParams};

export default async function Page({
    params,
}: {
    params: Promise<{slug?: string[]}>;
}) {
    const data = await getPageData({params});

    if (!data) {
        notFound();
    }

    return <PageRenderer {...data} />;
}
