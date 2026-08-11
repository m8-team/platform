import {resolve} from 'node:path';
import type {NextConfig} from 'next';

const nextConfig: NextConfig = {
  transpilePackages: [
    '@m8/module-sdk',
    '@m8/resource-manager',
  ],
  turbopack: {
    root: resolve(__dirname, '../..'),
  },
};

export default nextConfig;
