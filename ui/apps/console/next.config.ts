import {resolve} from 'node:path';
import type {NextConfig} from 'next';

const nextConfig: NextConfig = {
  transpilePackages: [
    '@m8/json-render-module-sdk',
    '@m8/resource-manager-module',
  ],
  turbopack: {
    root: resolve(__dirname, '../..'),
  },
};

export default nextConfig;
