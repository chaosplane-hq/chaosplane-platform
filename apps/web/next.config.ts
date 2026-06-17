import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  sassOptions: {
    includePaths: ['./node_modules'],
  },
  async redirects() {
    return [
      { source: '/get-started', destination: '/onboarding', permanent: true },
      { source: '/dashboard', destination: '/', permanent: true },
    ];
  },
};

export default nextConfig;
