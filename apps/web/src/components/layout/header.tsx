'use client';

import { useState, useEffect } from 'react';
import {
  Header,
  HeaderGlobalAction,
  HeaderGlobalBar,
  HeaderMenuButton,
  HeaderName,
  HeaderPanel,
  SkipToContent,
} from '@carbon/react';
import { Notification, UserAvatar, Logout } from '@carbon/icons-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { AppSidebar } from './sidebar';
import { getCurrentUser, logout, type CurrentUserResponse } from '@/lib/auth';

export function AppHeader() {
  const router = useRouter();
  const [isSideNavExpanded, setIsSideNavExpanded] = useState(false);
  const [isUserPanelOpen, setIsUserPanelOpen] = useState(false);
  const [user, setUser] = useState<CurrentUserResponse | null>(null);
  const [userLoading, setUserLoading] = useState(true);

  useEffect(() => {
    getCurrentUser()
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setUserLoading(false));
  }, []);

  async function handleLogout() {
    await logout();
    router.push('/login');
  }

  return (
    <>
      <Header aria-label="ChaosPlane">
        <SkipToContent />
        <HeaderMenuButton
          aria-label={isSideNavExpanded ? 'Close menu' : 'Open menu'}
          onClick={() => setIsSideNavExpanded((prev) => !prev)}
          isCollapsible
          isActive={isSideNavExpanded}
        />
        <HeaderName as={Link} href="/" prefix="">
          ChaosPlane
        </HeaderName>
        <HeaderGlobalBar>
          <HeaderGlobalAction aria-label="Notifications">
            <Notification size={20} />
          </HeaderGlobalAction>
          <HeaderGlobalAction
            aria-label="User profile"
            isActive={isUserPanelOpen}
            onClick={() => setIsUserPanelOpen((prev) => !prev)}
          >
            <UserAvatar size={20} />
          </HeaderGlobalAction>
        </HeaderGlobalBar>
        <HeaderPanel expanded={isUserPanelOpen}>
          <div style={{ padding: 'var(--cds-spacing-05)', color: 'var(--cds-text-primary)' }}>
            <p style={{ margin: 0, fontWeight: 600 }}>{userLoading ? '' : (user?.user.name ?? 'Not signed in')}</p>
            <p style={{ margin: '4px 0 0', fontSize: '0.875rem', color: 'var(--cds-text-secondary)' }}>
              {userLoading ? '' : (user?.user.email ?? '')}
            </p>
            {user?.tenant && (
              <p style={{ margin: '4px 0 0', fontSize: '0.75rem', color: 'var(--cds-text-helper)' }}>
                {user.tenant.name}
              </p>
            )}
            <div style={{ marginTop: 'var(--cds-spacing-05)', borderTop: '1px solid var(--cds-border-subtle-01)', paddingTop: 'var(--cds-spacing-04)' }}>
              <button
                type="button"
                onClick={handleLogout}
                style={{
                  background: 'none', border: 'none', color: 'var(--cds-text-error)',
                  cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem',
                  fontSize: '0.875rem', padding: 0,
                }}
              >
                <Logout size={16} /> Sign out
              </button>
            </div>
          </div>
        </HeaderPanel>
      </Header>
      <AppSidebar
        isExpanded={isSideNavExpanded}
        onToggle={() => setIsSideNavExpanded(false)}
      />
    </>
  );
}
