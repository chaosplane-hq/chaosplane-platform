'use client';

import { useState } from 'react';
import {
  Header,
  HeaderGlobalAction,
  HeaderGlobalBar,
  HeaderMenuButton,
  HeaderName,
  HeaderPanel,
  SkipToContent,
} from '@carbon/react';
import { Notification, UserAvatar } from '@carbon/icons-react';
import Link from 'next/link';
import { AppSidebar } from './sidebar';

export function AppHeader() {
  const [isSideNavExpanded, setIsSideNavExpanded] = useState(false);
  const [isUserPanelOpen, setIsUserPanelOpen] = useState(false);

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
            <p style={{ margin: 0, fontWeight: 600 }}>User Name</p>
            <p style={{ margin: '4px 0 0', fontSize: '0.875rem', color: 'var(--cds-text-secondary)' }}>
              user@company.com
            </p>
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
