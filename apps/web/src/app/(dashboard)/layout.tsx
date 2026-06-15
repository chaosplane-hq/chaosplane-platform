import { Content } from '@carbon/react';
import { AppHeader } from '@/components/layout/header';
import { AuthGuard } from '@/components/auth-guard';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthGuard>
      <AppHeader />
      <Content>{children}</Content>
    </AuthGuard>
  );
}
