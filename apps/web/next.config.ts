import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  sassOptions: {
    includePaths: ['./node_modules'],
  },
};

export default nextConfig;
