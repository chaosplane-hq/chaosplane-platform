import type { Metadata } from 'next';
import '@/styles/globals.scss';
import { Providers } from './providers';

export const metadata: Metadata = {
  title: 'ChaosPlane',
  description: 'Chaos engineering platform',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
