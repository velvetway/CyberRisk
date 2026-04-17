// CyberRisk — Sidebar.
// Ported from /tmp/design/cyberrisk/project/src/shell.jsx (lines 16-175).
// Uses React Router for navigation + longest-prefix active matching.
import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Icon } from './Icon';
import { IconBtn } from './Primitives';
import { NAV_ITEMS, NavItem, findActiveNavItem } from './NavConfig';

export interface SidebarProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ collapsed, onToggleCollapsed }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const activeItem = findActiveNavItem(location.pathname);
  const activeId = activeItem?.id;

  const main = NAV_ITEMS.filter((i) => i.section === 'main');
  const refs = NAV_ITEMS.filter((i) => i.section === 'refs');
  const sys = NAV_ITEMS.filter((i) => i.section === 'system');

  const renderItem = (item: NavItem) => {
    const isActive = item.id === activeId;
    return (
      <button
        key={item.id}
        onClick={() => navigate(item.path)}
        title={collapsed ? item.label : undefined}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          width: '100%',
          height: 32,
          padding: collapsed ? '0 0' : '0 10px',
          justifyContent: collapsed ? 'center' : 'flex-start',
          background: isActive ? 'var(--accent-ghost)' : 'transparent',
          color: isActive ? 'var(--accent)' : 'var(--fg-muted)',
          border: 'none',
          borderRadius: 'var(--r-sm)',
          fontSize: 'var(--text-sm)',
          fontWeight: isActive ? 500 : 400,
          cursor: 'pointer',
          transition: 'var(--transition)',
          position: 'relative',
          textAlign: 'left',
        }}
        onMouseEnter={(e: React.MouseEvent<HTMLButtonElement>) => {
          if (!isActive) {
            e.currentTarget.style.background = 'var(--bg-hover)';
            e.currentTarget.style.color = 'var(--fg)';
          }
        }}
        onMouseLeave={(e: React.MouseEvent<HTMLButtonElement>) => {
          if (!isActive) {
            e.currentTarget.style.background = 'transparent';
            e.currentTarget.style.color = 'var(--fg-muted)';
          }
        }}
      >
        {isActive && !collapsed && (
          <span
            style={{
              position: 'absolute',
              left: -12,
              top: 8,
              bottom: 8,
              width: 2,
              background: 'var(--accent)',
              borderRadius: 2,
            }}
          />
        )}
        <Icon name={item.icon} size={16} />
        {!collapsed && (
          <>
            <span style={{ flex: 1 }}>{item.label}</span>
            {item.badge && (
              <span
                style={{
                  fontSize: 'var(--text-2xs)',
                  fontFamily: 'var(--font-mono)',
                  padding: '1px 5px',
                  borderRadius: 'var(--r-xs)',
                  background: item.badge === 'NEW' ? 'var(--accent-ghost)' : 'var(--bg-elev-3)',
                  color: item.badge === 'NEW' ? 'var(--accent)' : 'var(--fg-dim)',
                  letterSpacing: '0.04em',
                }}
              >
                {item.badge}
              </span>
            )}
          </>
        )}
      </button>
    );
  };

  return (
    <aside
      style={{
        width: collapsed ? 'var(--sidebar-w-collapsed)' : 'var(--sidebar-w)',
        flexShrink: 0,
        background: 'var(--bg-elev-1)',
        borderRight: '1px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        transition: 'width var(--transition)',
        height: '100vh',
        position: 'sticky',
        top: 0,
      }}
    >
      {/* Logo block */}
      <div
        style={{
          height: 'var(--topbar-h)',
          padding: collapsed ? 0 : '0 14px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          gap: 10,
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            width: 26,
            height: 26,
            borderRadius: 6,
            background: 'var(--accent)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--accent-fg)',
            flexShrink: 0,
            boxShadow: '0 4px 10px oklch(0 0 0 / 0.3)',
          }}
        >
          <Icon name="shield" size={15} />
        </div>
        {!collapsed && (
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 'var(--text-sm)', fontWeight: 600, letterSpacing: '-0.01em' }}>
              CyberRisk
            </div>
            <div
              style={{
                fontSize: 10,
                color: 'var(--fg-faint)',
                fontFamily: 'var(--font-mono)',
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                marginTop: -1,
              }}
            >
              Enterprise Platform
            </div>
          </div>
        )}
      </div>

      {/* Workspace selector */}
      {!collapsed && (
        <div style={{ padding: '10px 12px', borderBottom: '1px solid var(--border)' }}>
          <button
            style={{
              width: '100%',
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              padding: '6px 10px',
              background: 'var(--bg-elev-2)',
              border: '1px solid var(--border)',
              borderRadius: 'var(--r-sm)',
              cursor: 'pointer',
              color: 'var(--fg)',
              textAlign: 'left',
            }}
          >
            <div
              style={{
                width: 18,
                height: 18,
                borderRadius: 4,
                background: 'linear-gradient(135deg, oklch(0.62 0.17 265), oklch(0.65 0.15 320))',
              }}
            />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', lineHeight: 1.2 }}>
                Организация
              </div>
              <div
                style={{
                  fontSize: 'var(--text-sm)',
                  lineHeight: 1.3,
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                АО «Энерго-Центр»
              </div>
            </div>
            <Icon name="chevronD" size={14} color="var(--fg-dim)" />
          </button>
        </div>
      )}

      {/* Nav */}
      <nav style={{ flex: 1, overflow: 'auto', padding: collapsed ? '8px 8px' : '10px 12px' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>{main.map(renderItem)}</div>
        {!collapsed && (
          <div
            style={{
              fontSize: 'var(--text-2xs)',
              color: 'var(--fg-faint)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              fontWeight: 500,
              padding: '14px 10px 6px',
            }}
          >
            Справочники
          </div>
        )}
        {collapsed && <div style={{ height: 1, background: 'var(--border)', margin: '10px 4px' }} />}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>{refs.map(renderItem)}</div>

        {!collapsed && (
          <div
            style={{
              fontSize: 'var(--text-2xs)',
              color: 'var(--fg-faint)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              fontWeight: 500,
              padding: '14px 10px 6px',
            }}
          >
            Система
          </div>
        )}
        {collapsed && <div style={{ height: 1, background: 'var(--border)', margin: '10px 4px' }} />}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>{sys.map(renderItem)}</div>
      </nav>

      {/* Status footer */}
      <div
        style={{
          padding: collapsed ? 8 : 12,
          borderTop: '1px solid var(--border)',
          display: 'flex',
          alignItems: 'center',
          gap: 10,
        }}
      >
        {!collapsed ? (
          <>
            <div
              style={{
                width: 8,
                height: 8,
                borderRadius: 999,
                background: 'var(--risk-low)',
                boxShadow: '0 0 8px var(--risk-low)',
              }}
            />
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 'var(--text-xs)' }}>Система активна</div>
              <div style={{ fontSize: 10, color: 'var(--fg-faint)', fontFamily: 'var(--font-mono)' }}>
                api v1.4.2 · uptime 41d 12h
              </div>
            </div>
            <IconBtn size={24} onClick={onToggleCollapsed} title="Свернуть">
              <Icon name="chevronR" size={14} />
            </IconBtn>
          </>
        ) : (
          <IconBtn size={28} onClick={onToggleCollapsed} title="Развернуть">
            <Icon name="menu" size={16} />
          </IconBtn>
        )}
      </div>
    </aside>
  );
};
