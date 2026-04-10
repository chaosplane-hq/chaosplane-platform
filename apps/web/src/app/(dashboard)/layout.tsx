import { Content } from '@carbon/react';
import { AppHeader } from '@/components/layout/header';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <AppHeader />
      <Content>{children}</Content>
    </>
  );
}
