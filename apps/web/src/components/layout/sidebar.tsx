'use client';

import {
  SideNav,
  SideNavItems,
  SideNavLink,
} from '@carbon/react';
import {
  Dashboard,
  Chemistry,
  Cloud,
  Settings,
  Rocket,
  Network_3,
  Security,
  Rule,
  Idea,
  ChartLine,
  Chat,
  GameConsole,
  Certificate,
  Flow,
  Store,
  Partnership,
  Analytics,
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
  { href: '/policies', label: 'Policies', icon: Rule },
  { href: '/environments', label: 'Environments', icon: Cloud },
  { href: '/topology', label: 'Topology', icon: Network_3 },
  { href: '/vulnerabilities', label: 'Vulnerabilities', icon: Security },
  { href: '/suggestions', label: 'Suggestions', icon: Idea },
  { href: '/analysis', label: 'Analysis', icon: ChartLine },
  { href: '/ai-chat', label: 'AI Chat', icon: Chat },
  { href: '/gamedays', label: 'GameDays', icon: GameConsole },
  { href: '/resilience', label: 'Resilience', icon: Certificate },
  { href: '/workflows', label: 'Workflows', icon: Flow },
  { href: '/marketplace', label: 'Marketplace', icon: Store },
  { href: '/federation', label: 'Federation', icon: Partnership },
  { href: '/predictions', label: 'Predictions', icon: Analytics },
  { href: '/onboarding', label: 'Get Started', icon: Rocket },
  { href: '/settings', label: 'Settings', icon: Settings },
];

export function AppSidebar({ isExpanded, onToggle }: AppSidebarProps) {
  const pathname = usePathname();

  return (
    <SideNav
      aria-label="Side navigation"
      expanded={isExpanded}
      onSideNavBlur={onToggle}
    >
      <SideNavItems>
        {navItems.map(({ href, label, icon: Icon }) => (
          <SideNavLink
            key={href}
            renderIcon={Icon}
            href={href}
            as={Link}
            isActive={pathname === href || (href !== '/' && pathname.startsWith(href))}
            large
          >
            {label}
          </SideNavLink>
        ))}
      </SideNavItems>
    </SideNav>
  );
}
