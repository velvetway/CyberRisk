// CyberRisk — navigation config.
// Shared between Sidebar and CommandPalette.
import type { IconName } from './Icon';

export type NavSection = 'main' | 'refs' | 'system';

export interface NavItem {
  id: string;
  path: string; // React Router path
  label: string;
  icon: IconName;
  badge?: string | null;
  section: NavSection;
}

export const NAV_ITEMS: NavItem[] = [
  { id: 'dashboard', path: '/',                label: 'Обзор',            icon: 'grid',     badge: null,  section: 'main' },
  { id: 'riskmap',   path: '/risk/map',        label: 'Карта рисков',     icon: 'target',   badge: '5×5', section: 'main' },
  { id: 'graph',     path: '/risk/graph',      label: 'Граф атаки',       icon: 'flow',     badge: null,  section: 'main' },
  { id: 'assets',    path: '/assets',          label: 'Реестр активов',   icon: 'layers',   badge: null,  section: 'main' },
  { id: 'software',  path: '/software',        label: 'Справочник ПО',    icon: 'package',  badge: null,  section: 'main' },
  { id: 'simulator', path: '/risk/preview',    label: 'Симулятор риска',  icon: 'zap',      badge: 'NEW', section: 'main' },
  { id: 'threats',   path: '/threats',         label: 'Каталог угроз',    icon: 'alert',    badge: 'БДУ', section: 'refs' },
  { id: 'vulns',     path: '/vulnerabilities', label: 'Уязвимости',       icon: 'activity', badge: null,  section: 'refs' },
  { id: 'reports',   path: '/reports',         label: 'Отчёты',           icon: 'file',     badge: null,  section: 'refs' },
  { id: 'settings',  path: '/settings',        label: 'Настройки',        icon: 'settings', badge: null,  section: 'system' },
];

/**
 * Pick the nav item whose `path` best matches the given URL.
 * Uses longest-prefix-match so `/assets/42/risks` resolves to `assets`.
 * Root `/` only matches the dashboard entry to avoid swallowing everything.
 */
export function findActiveNavItem(pathname: string): NavItem | undefined {
  let best: NavItem | undefined;
  let bestLen = -1;
  for (const item of NAV_ITEMS) {
    if (item.path === '/') {
      if (pathname === '/' && bestLen < 1) {
        best = item;
        bestLen = 1;
      }
      continue;
    }
    if (pathname === item.path || pathname.startsWith(item.path + '/')) {
      if (item.path.length > bestLen) {
        best = item;
        bestLen = item.path.length;
      }
    }
  }
  return best;
}
