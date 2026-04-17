// CyberRisk — CommandPalette.
// Ported from /tmp/design/cyberrisk/project/src/shell.jsx (lines 251-308).
// Replaces ASSETS preview items with nav + static actions (no mock dataset in real app).
import React, { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon, IconName } from './Icon';
import { Chip, KBD } from './Primitives';
import { NAV_ITEMS } from './NavConfig';

export interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
}

interface PaletteItem {
  type: 'Переход' | 'Действие';
  label: string;
  path: string;
  icon: IconName;
}

export const CommandPalette: React.FC<CommandPaletteProps> = ({ open, onClose }) => {
  const navigate = useNavigate();
  const [q, setQ] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      const t = setTimeout(() => inputRef.current?.focus(), 30);
      return () => clearTimeout(t);
    }
  }, [open]);

  // Close on Escape.
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  const all: PaletteItem[] = [
    ...NAV_ITEMS.map<PaletteItem>((i) => ({
      type: 'Переход',
      label: i.label,
      path: i.path,
      icon: i.icon,
    })),
    { type: 'Действие', label: 'Рассчитать риск (что-если)', path: '/risk/preview', icon: 'zap' },
    { type: 'Действие', label: 'Экспортировать PDF-отчёт',   path: '/reports',      icon: 'download' },
    { type: 'Действие', label: 'Добавить актив',             path: '/assets/new',   icon: 'plus' },
  ];

  const items = q
    ? all.filter((it) => it.label.toLowerCase().includes(q.toLowerCase()))
    : all;

  if (!open) return null;
  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        background: 'oklch(0 0 0 / 0.5)',
        zIndex: 100,
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'center',
        paddingTop: '12vh',
        backdropFilter: 'blur(4px)',
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: 560,
          maxWidth: '90vw',
          background: 'var(--bg-elev-2)',
          border: '1px solid var(--border-strong)',
          borderRadius: 'var(--r-lg)',
          boxShadow: 'var(--sh-lg)',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '12px 14px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <Icon name="search" size={16} color="var(--fg-dim)" />
          <input
            ref={inputRef}
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Введите команду или найдите актив, угрозу…"
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              fontSize: 'var(--text-md)',
              color: 'var(--fg)',
            }}
          />
          <KBD>ESC</KBD>
        </div>
        <div style={{ maxHeight: 420, overflow: 'auto', padding: 6 }}>
          {items.length === 0 && (
            <div
              style={{
                padding: 20,
                textAlign: 'center',
                color: 'var(--fg-dim)',
                fontSize: 'var(--text-sm)',
              }}
            >
              Ничего не найдено
            </div>
          )}
          {items.map((it, i) => (
            <div
              key={i}
              onClick={() => {
                navigate(it.path);
                onClose();
                setQ('');
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: '8px 10px',
                borderRadius: 'var(--r-sm)',
                cursor: 'pointer',
              }}
              onMouseEnter={(e: React.MouseEvent<HTMLDivElement>) => {
                e.currentTarget.style.background = 'var(--bg-hover)';
              }}
              onMouseLeave={(e: React.MouseEvent<HTMLDivElement>) => {
                e.currentTarget.style.background = 'transparent';
              }}
            >
              <Icon name={it.icon} size={14} color="var(--fg-muted)" />
              <span style={{ flex: 1, fontSize: 'var(--text-sm)' }}>{it.label}</span>
              <Chip tone="ghost">{it.type}</Chip>
            </div>
          ))}
        </div>
        <div
          style={{
            padding: '8px 14px',
            borderTop: '1px solid var(--border)',
            fontSize: 'var(--text-xs)',
            color: 'var(--fg-dim)',
            display: 'flex',
            gap: 14,
          }}
        >
          <span>
            <KBD>↑</KBD>
            <KBD>↓</KBD> навигация
          </span>
          <span>
            <KBD>↵</KBD> открыть
          </span>
          <span>
            <KBD>⌘K</KBD> закрыть
          </span>
        </div>
      </div>
    </div>
  );
};
