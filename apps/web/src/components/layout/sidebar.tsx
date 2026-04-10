'use client';

import { useState } from 'react';
import {
  SideNav,
  SideNavItems,
  SideNavLink,
  SideNavMenu,
  SideNavMenuItem,
} from '@carbon/react';
import {
  Dashboard,
  Chemistry,
  Cloud,
  Settings,
} from '@carbon/icons-react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';

interface AppSidebarProps {
  isExpanded: boolean;
  onToggle: () => void;
}

const navItems = [
  { href: '/', label: 'Dashboard', icon: Dashboard },
  { href: '/experiments', label: 'Experiments', icon: Chemistry },
  { href: '/environments', label: 'Environments', icon: Cloud },
  { href: '/settings', label: 'Settings', icon: Settings },
];

export function AppSidebar({ isExpanded, onToggle }: AppSidebarProps) {
  const pathname = usePathname();

  return (
    <SideNav
      aria-label="Side navigation"
      expanded={isExpanded}
      isPersistent={false}
      onSideNavBlur={onToggle}
    >
      <SideNavItems>
        {navItems.map(({ href, label, icon: Icon }) => (
          <SideNavLink
            key={href}
            renderIcon={Icon}
            href={href}
            as={Link}
            isActive={pathname === href}
            large
          >
            {label}
          </SideNavLink>
        ))}
      </SideNavItems>
    </SideNav>
  );
}
