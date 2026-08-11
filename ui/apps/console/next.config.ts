import {resolve} from 'node:path';
import type {NextConfig} from 'next';

const nextConfig: NextConfig = {
  transpilePackages: [
    '@m8/core',
    '@m8/query',
    '@m8/operation',
    '@m8/runtime',
    '@m8/resource-manager',
  ],
  turbopack: {
    root: resolve(__dirname, '../..'),
  },
};

export default nextConfig;
