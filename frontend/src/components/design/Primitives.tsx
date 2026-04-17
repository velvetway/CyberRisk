// CyberRisk — Primitives: RiskBadge, Chip, Btn, IconBtn, KBD, Card, StatCard, Sparkline.
// Ported from /tmp/design/cyberrisk/project/src/primitives.jsx to TypeScript.
import React, { useState, CSSProperties, ReactNode } from 'react';

export type RiskLevel = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface RiskBadgeProps {
  level: RiskLevel;
  score?: number;
  compact?: boolean;
}

export const RiskBadge: React.FC<RiskBadgeProps> = ({ level, score, compact }) => {
  const map = {
    critical: { label: 'Критический', short: 'CRIT', color: 'var(--risk-critical)', bg: 'var(--risk-critical-bg)', br: 'var(--risk-critical-br)' },
    high:     { label: 'Высокий',     short: 'HIGH', color: 'var(--risk-high)',     bg: 'var(--risk-high-bg)',     br: 'var(--risk-high-br)' },
    medium:   { label: 'Средний',     short: 'MED',  color: 'var(--risk-medium)',   bg: 'var(--risk-medium-bg)',   br: 'var(--risk-medium-br)' },
    low:      { label: 'Низкий',      short: 'LOW',  color: 'var(--risk-low)',      bg: 'var(--risk-low-bg)',      br: 'var(--risk-low-br)' },
    info:     { label: 'Инфо',        short: 'INFO', color: 'var(--risk-info)',     bg: 'var(--risk-info-bg)',     br: 'var(--risk-info-br)' },
  } as const;
  const c = map[level] || map.info;
  const displayScore: string | null =
    score !== undefined && score !== null
      ? (typeof score === 'number' ? score.toFixed(1) : String(score))
      : null;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      padding: compact ? '2px 6px' : '3px 8px',
      borderRadius: 'var(--r-sm)',
      background: c.bg, color: c.color,
      border: `1px solid ${c.br}`,
      fontSize: 'var(--text-2xs)',
      fontWeight: 600, letterSpacing: '0.04em', textTransform: 'uppercase',
      fontFamily: 'var(--font-mono)',
      whiteSpace: 'nowrap',
    }}>
      <span style={{ width: 6, height: 6, borderRadius: 999, background: c.color, boxShadow: `0 0 8px ${c.color}` }} />
      {compact ? c.short : c.label}
      {displayScore !== null && <span style={{ opacity: 0.75 }}>{displayScore}</span>}
    </span>
  );
};

export type ChipTone = 'neutral' | 'accent' | 'success' | 'warn' | 'danger' | 'ghost';

export interface ChipProps {
  children?: ReactNode;
  tone?: ChipTone;
  icon?: ReactNode;
  mono?: boolean;
  onClick?: (e: React.MouseEvent<HTMLSpanElement>) => void;
}

export const Chip: React.FC<ChipProps> = ({ children, tone = 'neutral', icon, mono, onClick }) => {
  const toneStyles: Record<ChipTone, { bg: string; fg: string; br: string }> = {
    neutral: { bg: 'var(--bg-elev-3)',         fg: 'var(--fg-muted)',    br: 'var(--border)' },
    accent:  { bg: 'var(--accent-ghost)',      fg: 'var(--accent)',      br: 'transparent' },
    success: { bg: 'var(--risk-low-bg)',       fg: 'var(--risk-low)',    br: 'var(--risk-low-br)' },
    warn:    { bg: 'var(--risk-medium-bg)',    fg: 'var(--risk-medium)', br: 'var(--risk-medium-br)' },
    danger:  { bg: 'var(--risk-critical-bg)',  fg: 'var(--risk-critical)', br: 'var(--risk-critical-br)' },
    ghost:   { bg: 'transparent',              fg: 'var(--fg-dim)',      br: 'var(--border)' },
  };
  const t = toneStyles[tone];
  return (
    <span onClick={onClick} style={{
      display: 'inline-flex', alignItems: 'center', gap: 5,
      padding: '2px 7px', borderRadius: 'var(--r-sm)',
      background: t.bg, color: t.fg,
      border: `1px solid ${t.br}`,
      fontSize: 'var(--text-xs)', fontWeight: 500,
      fontFamily: mono ? 'var(--font-mono)' : 'inherit',
      cursor: onClick ? 'pointer' : 'default',
      transition: 'var(--transition)',
    }}>
      {icon}{children}
    </span>
  );
};

export interface IconBtnProps {
  children?: ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  active?: boolean;
  title?: string;
  size?: number;
}

export const IconBtn: React.FC<IconBtnProps> = ({ children, onClick, active, title, size = 30 }) => (
  <button
    title={title}
    onClick={onClick}
    style={{
      width: size, height: size,
      display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
      background: active ? 'var(--bg-active)' : 'transparent',
      color: active ? 'var(--fg)' : 'var(--fg-muted)',
      border: '1px solid transparent',
      borderColor: active ? 'var(--border)' : 'transparent',
      borderRadius: 'var(--r-sm)',
      cursor: 'pointer',
      transition: 'var(--transition)',
    }}
    onMouseEnter={(e) => {
      if (!active) {
        e.currentTarget.style.background = 'var(--bg-hover)';
        e.currentTarget.style.color = 'var(--fg)';
      }
    }}
    onMouseLeave={(e) => {
      if (!active) {
        e.currentTarget.style.background = 'transparent';
        e.currentTarget.style.color = 'var(--fg-muted)';
      }
    }}
  >
    {children}
  </button>
);

export type BtnVariant = 'primary' | 'secondary' | 'ghost' | 'outline' | 'danger';
export type BtnSize = 'sm' | 'md' | 'lg';

export interface BtnProps {
  children?: ReactNode;
  variant?: BtnVariant;
  size?: BtnSize;
  icon?: ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  disabled?: boolean;
  style?: CSSProperties;
  fullWidth?: boolean;
  type?: 'button' | 'submit' | 'reset';
}

export const Btn: React.FC<BtnProps> = ({
  children, variant = 'primary', size = 'md', icon, onClick, disabled, style, fullWidth, type = 'button',
}) => {
  const variants: Record<BtnVariant, { bg: string; fg: string; br: string; bgH: string }> = {
    primary:   { bg: 'var(--accent)',         fg: 'var(--accent-fg)', br: 'var(--accent)',         bgH: 'var(--accent-hover)' },
    secondary: { bg: 'var(--bg-elev-3)',      fg: 'var(--fg)',        br: 'var(--border-strong)',  bgH: 'var(--bg-hover)' },
    ghost:     { bg: 'transparent',           fg: 'var(--fg-muted)',  br: 'transparent',           bgH: 'var(--bg-hover)' },
    outline:   { bg: 'transparent',           fg: 'var(--fg)',        br: 'var(--border-strong)',  bgH: 'var(--bg-hover)' },
    danger:    { bg: 'var(--risk-critical)',  fg: 'white',            br: 'var(--risk-critical)',  bgH: 'oklch(0.65 0.22 25)' },
  };
  const sizes: Record<BtnSize, { h: number; px: number; fs: string }> = {
    sm: { h: 26, px: 10, fs: 'var(--text-xs)' },
    md: { h: 32, px: 12, fs: 'var(--text-sm)' },
    lg: { h: 38, px: 16, fs: 'var(--text-base)' },
  };
  const v = variants[variant];
  const s = sizes[size];
  const [h, setH] = useState<boolean>(false);
  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      onMouseEnter={() => setH(true)}
      onMouseLeave={() => setH(false)}
      style={{
        height: s.h, padding: `0 ${s.px}px`,
        background: disabled ? 'var(--bg-elev-3)' : (h ? v.bgH : v.bg),
        color: disabled ? 'var(--fg-faint)' : v.fg,
        border: `1px solid ${v.br}`,
        borderRadius: 'var(--r-sm)',
        fontSize: s.fs, fontWeight: 500,
        display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 6,
        cursor: disabled ? 'not-allowed' : 'pointer',
        transition: 'var(--transition)',
        width: fullWidth ? '100%' : undefined,
        ...style,
      }}
    >
      {icon}{children}
    </button>
  );
};

export interface KBDProps {
  children?: ReactNode;
}

export const KBD: React.FC<KBDProps> = ({ children }) => (
  <kbd style={{
    fontFamily: 'var(--font-mono)', fontSize: 'var(--text-2xs)',
    padding: '1px 5px', borderRadius: 'var(--r-xs)',
    background: 'var(--bg-elev-3)', border: '1px solid var(--border)',
    color: 'var(--fg-muted)', boxShadow: 'inset 0 -1px 0 var(--border)',
    minWidth: 16, display: 'inline-flex', justifyContent: 'center',
  }}>{children}</kbd>
);

export interface CardProps {
  children?: ReactNode;
  pad?: number;
  style?: CSSProperties;
  title?: ReactNode;
  subtitle?: ReactNode;
  action?: ReactNode;
  dense?: boolean;
}

export const Card: React.FC<CardProps> = ({ children, pad = 16, style, title, subtitle, action, dense }) => (
  <div style={{
    background: 'var(--bg-elev-1)',
    border: '1px solid var(--border)',
    borderRadius: 'var(--r-lg)',
    overflow: 'hidden',
    ...style,
  }}>
    {(title || action) && (
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: dense ? '10px 14px' : '12px 16px',
        borderBottom: '1px solid var(--border)',
      }}>
        <div>
          {title && <div style={{ fontSize: 'var(--text-sm)', fontWeight: 600, color: 'var(--fg)' }}>{title}</div>}
          {subtitle && <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginTop: 2 }}>{subtitle}</div>}
        </div>
        {action}
      </div>
    )}
    <div style={{ padding: title ? pad : pad }}>{children}</div>
  </div>
);

export type StatTone = 'critical' | 'high' | 'medium' | 'low' | 'info' | 'accent';

export interface StatCardProps {
  label: ReactNode;
  value: ReactNode;
  delta?: number;
  sub?: ReactNode;
  tone?: StatTone;
  sparkline?: ReactNode;
  mono?: boolean;
}

export const StatCard: React.FC<StatCardProps> = ({ label, value, delta, sub, tone, sparkline, mono }) => {
  const toneMap: Record<StatTone, string> = {
    critical: 'var(--risk-critical)',
    high:     'var(--risk-high)',
    medium:   'var(--risk-medium)',
    low:      'var(--risk-low)',
    info:     'var(--risk-info)',
    accent:   'var(--accent)',
  };
  const toneColor = tone ? toneMap[tone] : undefined;
  return (
    <div style={{
      background: 'var(--bg-elev-1)', border: '1px solid var(--border)',
      borderRadius: 'var(--r-lg)', padding: 16, position: 'relative', overflow: 'hidden',
    }}>
      {tone && <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 2, background: toneColor, opacity: 0.9 }} />}
      <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 500 }}>{label}</div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginTop: 6 }}>
        <div className={mono ? 'num' : ''} style={{ fontSize: 'var(--text-3xl)', fontWeight: 600, color: toneColor || 'var(--fg)', letterSpacing: '-0.02em', lineHeight: 1 }}>{value}</div>
        {delta !== undefined && delta !== null && (
          <div style={{ fontSize: 'var(--text-xs)', color: delta >= 0 ? 'var(--risk-low)' : 'var(--risk-critical)', fontFamily: 'var(--font-mono)' }}>
            {delta >= 0 ? '↑' : '↓'} {Math.abs(delta)}
          </div>
        )}
      </div>
      {sub && <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginTop: 6 }}>{sub}</div>}
      {sparkline}
    </div>
  );
};

export interface SparklineProps {
  data: number[];
  color?: string;
  height?: number;
}

export const Sparkline: React.FC<SparklineProps> = ({ data, color = 'var(--accent)', height = 28 }) => {
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = max - min || 1;
  const W = 100;
  const H = height;
  const pts = data
    .map((v, i) => [(i / (data.length - 1)) * W, H - ((v - min) / range) * H])
    .map((p) => p.join(','))
    .join(' ');
  return (
    <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" style={{ width: '100%', height, marginTop: 8, display: 'block' }}>
      <polyline fill="none" stroke={color} strokeWidth="1.5" points={pts} />
      <polyline fill={color} opacity="0.08" stroke="none" points={`0,${H} ${pts} ${W},${H}`} />
    </svg>
  );
};
