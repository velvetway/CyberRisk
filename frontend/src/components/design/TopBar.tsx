// CyberRisk — TopBar.
// Ported from /tmp/design/cyberrisk/project/src/shell.jsx (lines 177-249).
// Integrates with AuthContext to show real user + role.
import React from 'react';
import { Icon } from './Icon';
import { IconBtn, KBD } from './Primitives';
import { useAuth } from '../../context/AuthContext';

export interface TopBarProps {
  breadcrumbs: string[];
  onCmdK: () => void;
  onThemeToggle: () => void;
  theme: 'dark' | 'light';
}

const roleLabel: Record<string, string> = {
  admin: 'ADMIN',
  auditor: 'AUDITOR',
  viewer: 'VIEWER',
};

export const TopBar: React.FC<TopBarProps> = ({ breadcrumbs, onCmdK, onThemeToggle, theme }) => {
  // AuthContext may not be mounted in isolated previews — guard it.
  let username = 'И. Петров';
  let role = 'CISO · ADMIN';
  let initials = 'ИП';
  try {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    const { user } = useAuth();
    if (user) {
      username = user.username;
      initials = user.username.slice(0, 2).toUpperCase();
      const badge = roleLabel[user.role] || user.role.toUpperCase();
      role = `CISO · ${badge}`;
    }
  } catch {
    /* outside AuthProvider — use defaults */
  }

  return (
    <div
      style={{
        height: 'var(--topbar-h)',
        borderBottom: '1px solid var(--border)',
        background: 'var(--bg-glass)',
        backdropFilter: 'blur(12px)',
        display: 'flex',
        alignItems: 'center',
        padding: '0 16px',
        gap: 12,
        position: 'sticky',
        top: 0,
        zIndex: 10,
      }}
    >
      {/* Breadcrumbs */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-dim)',
        }}
      >
        {breadcrumbs.map((b, i) => (
          <React.Fragment key={i}>
            {i > 0 && <Icon name="chevronR" size={12} color="var(--fg-faint)" />}
            <span
              style={{
                color: i === breadcrumbs.length - 1 ? 'var(--fg)' : 'var(--fg-dim)',
                fontWeight: i === breadcrumbs.length - 1 ? 500 : 400,
              }}
            >
              {b}
            </span>
          </React.Fragment>
        ))}
      </div>

      <div style={{ flex: 1 }} />

      {/* Command bar */}
      <button
        onClick={onCmdK}
        style={{
          width: 260,
          height: 28,
          padding: '0 10px',
          background: 'var(--bg-elev-2)',
          border: '1px solid var(--border)',
          borderRadius: 'var(--r-sm)',
          color: 'var(--fg-dim)',
          fontSize: 'var(--text-xs)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          cursor: 'pointer',
          transition: 'var(--transition)',
        }}
        onMouseEnter={(e: React.MouseEvent<HTMLButtonElement>) => {
          e.currentTarget.style.background = 'var(--bg-elev-3)';
        }}
        onMouseLeave={(e: React.MouseEvent<HTMLButtonElement>) => {
          e.currentTarget.style.background = 'var(--bg-elev-2)';
        }}
      >
        <Icon name="search" size={13} />
        <span style={{ flex: 1, textAlign: 'left' }}>Поиск активов, угроз, ПО…</span>
        <KBD>⌘K</KBD>
      </button>

      <IconBtn onClick={onThemeToggle} title={theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'}>
        <Icon name={theme === 'dark' ? 'sun' : 'moon'} size={15} />
      </IconBtn>
      <IconBtn title="Уведомления">
        <div style={{ position: 'relative' }}>
          <Icon name="bell" size={15} />
          <div
            style={{
              position: 'absolute',
              top: -2,
              right: -2,
              width: 6,
              height: 6,
              borderRadius: 999,
              background: 'var(--risk-critical)',
            }}
          />
        </div>
      </IconBtn>
      <IconBtn title="Помощь">
        <Icon name="help" size={15} />
      </IconBtn>

      <div style={{ width: 1, height: 20, background: 'var(--border)', margin: '0 4px' }} />

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '4px 8px',
          borderRadius: 'var(--r-sm)',
          cursor: 'pointer',
        }}
      >
        <div
          style={{
            width: 24,
            height: 24,
            borderRadius: 999,
            background: 'linear-gradient(135deg, oklch(0.55 0.14 320), oklch(0.55 0.18 265))',
            color: 'white',
            fontSize: 11,
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {initials}
        </div>
        <div style={{ lineHeight: 1.2 }}>
          <div style={{ fontSize: 'var(--text-xs)' }}>{username}</div>
          <div style={{ fontSize: 10, color: 'var(--fg-faint)', fontFamily: 'var(--font-mono)' }}>
            {role}
          </div>
        </div>
      </div>
    </div>
  );
};
