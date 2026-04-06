import type {IconData} from '@gravity-ui/uikit';
import {
  ArrowsRotateRight,
  Books,
  Boxes3,
  Briefcase,
  Gear,
  House,
  Key,
  Lock,
  Magnifier,
  Persons,
  PersonsLock,
  Shield,
} from '@gravity-ui/icons';

import {titleFromId} from '@/shared/lib/format';
import type {BreadcrumbItem} from '@/shared/types/iam';

export type NavigationItemConfig = {
  id: string;
  title: string;
  to: string;
  icon: IconData;
  matchPrefix?: string;
};

export const navigationItems: NavigationItemConfig[] = [
  {id: 'dashboard', title: 'Обзор', to: '/dashboard', icon: House, matchPrefix: '/dashboard'},
  {id: 'tenants', title: 'Тенанты', to: '/tenants', icon: Boxes3, matchPrefix: '/tenants'},
  {id: 'users', title: 'Пользователи', to: '/users', icon: Persons, matchPrefix: '/users'},
  {id: 'groups', title: 'Группы', to: '/groups', icon: PersonsLock, matchPrefix: '/groups'},
  {
    id: 'service-accounts',
    title: 'Service Accounts',
    to: '/service-accounts',
    icon: Briefcase,
    matchPrefix: '/service-accounts',
  },
  {
    id: 'oauth-clients',
    title: 'OAuth Clients',
    to: '/oauth-clients',
    icon: Key,
    matchPrefix: '/oauth-clients',
  },
  {id: 'roles', title: 'Роли', to: '/roles', icon: Shield, matchPrefix: '/roles'},
  {id: 'access', title: 'Доступ', to: '/access/explorer', icon: Lock, matchPrefix: '/access'},
  {id: 'audit', title: 'Аудит', to: '/audit', icon: Books, matchPrefix: '/audit'},
  {
    id: 'operations',
    title: 'Операции',
    to: '/operations',
    icon: ArrowsRotateRight,
    matchPrefix: '/operations',
  },
  {id: 'settings', title: 'Настройки', to: '/settings', icon: Gear, matchPrefix: '/settings'},
];

export const utilityItems: NavigationItemConfig[] = [
  {id: 'search', title: 'Глобальный поиск', to: '/search', icon: Magnifier, matchPrefix: '/search'},
];

const segmentLabels: Record<string, string> = {
  dashboard: 'Dashboard',
  tenants: 'Tenants',
  users: 'Users',
  groups: 'Groups',
  'service-accounts': 'Service Accounts',
  'oauth-clients': 'OAuth Clients',
  roles: 'Roles',
  access: 'Access',
  resources: 'Resource Access',
  explorer: 'Access Explorer',
  explain: 'Explain Access',
  simulate: 'Impact Analysis',
  policies: 'Policy Templates',
  'support-access': 'Support Access',
  sessions: 'Sessions',
  audit: 'Audit',
  operations: 'Operations',
  settings: 'Settings',
  members: 'Members',
  keys: 'Keys',
};

export function buildBreadcrumbs(pathname: string): BreadcrumbItem[] {
  const parts = pathname.split('/').filter(Boolean);
  const breadcrumbs: BreadcrumbItem[] = [{label: 'IAM', href: '/dashboard'}];
  let currentPath = '';

  for (const part of parts) {
    currentPath += `/${part}`;

    breadcrumbs.push({
      label: segmentLabels[part] ?? titleFromId(part),
      href: currentPath,
    });
  }

  return breadcrumbs;
}

export function getActiveNavigation(pathname: string): string {
  if (pathname === '/' || pathname.startsWith('/dashboard')) {
    return 'dashboard';
  }

  const matched = navigationItems.find((item) => pathname.startsWith(item.matchPrefix ?? item.to));
  return matched?.id ?? 'dashboard';
}
