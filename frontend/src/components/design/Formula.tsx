import React from 'react';

// PTSZI formula component — W = (Q^th + q + (1 - Q^re)) / 3 · Z
// Typography-first: large readable variables, sup rendered at 0.7em with high vertical align.

interface VarSpanProps {
  color: string;
  children: React.ReactNode;
  sup?: string;
  bold?: boolean;
}

const V: React.FC<VarSpanProps> = ({ color, children, sup, bold }) => (
  <span style={{
    color,
    fontWeight: bold ? 700 : 500,
    letterSpacing: '0.01em',
    fontStyle: 'italic',
    fontFamily: "'Inter', 'Inter Display', -apple-system, sans-serif",
  }}>
    {children}
    {sup && (
      <sup style={{
        fontSize: '0.62em',
        verticalAlign: 'baseline',
        position: 'relative',
        top: '-0.45em',
        marginLeft: 1,
        fontStyle: 'normal',
        fontWeight: 500,
        letterSpacing: '0.02em',
      }}>{sup}</sup>
    )}
  </span>
);

const Op: React.FC<{ children: React.ReactNode; muted?: boolean }> = ({ children, muted }) => (
  <span style={{
    color: muted ? 'var(--fg-dim)' : 'var(--fg-muted)',
    margin: '0 0.45em',
    fontWeight: 400,
  }}>
    {children}
  </span>
);

const Num: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <span style={{ color: 'var(--fg)', fontFamily: 'var(--font-mono)', fontSize: '0.92em' }}>
    {children}
  </span>
);

export interface PtsziFormulaProps {
  size?: 'md' | 'lg' | 'xl';
  align?: 'left' | 'center';
  withBackground?: boolean;
}

export const PtsziFormula: React.FC<PtsziFormulaProps> = ({
  size = 'lg',
  align = 'left',
  withBackground = true,
}) => {
  const fs = size === 'xl' ? 28 : size === 'lg' ? 22 : 17;
  return (
    <div style={{
      padding: withBackground ? '18px 24px' : 0,
      background: withBackground ? 'var(--bg-elev-2)' : 'transparent',
      border: withBackground ? '1px solid var(--border)' : 'none',
      borderRadius: 'var(--r-md)',
      fontSize: fs,
      lineHeight: 1.2,
      textAlign: align,
      letterSpacing: '-0.005em',
      color: 'var(--fg)',
      whiteSpace: 'nowrap',
      overflowX: 'auto',
    }}>
      <V color="var(--accent)" bold>W</V>
      <Op>=</Op>
      <Op muted>(</Op>
      <V color="var(--risk-critical)" sup="th">Q</V>
      <Op>+</Op>
      <V color="var(--risk-high)">q</V>
      <Op>+</Op>
      <Op muted>(</Op>
      <Num>1</Num>
      <Op>−</Op>
      <V color="var(--risk-info)" sup="re">Q</V>
      <Op muted>)</Op>
      <Op muted>)</Op>
      <Op>/</Op>
      <Num>3</Num>
      <Op>·</Op>
      <V color="var(--risk-medium)" bold>Z</V>
    </div>
  );
};

// Compact inline version for use inside cards where background is controlled by parent
export const PtsziFormulaInline: React.FC<{ size?: number }> = ({ size = 17 }) => (
  <div style={{
    fontSize: size,
    lineHeight: 1.2,
    color: 'var(--fg)',
    display: 'inline-flex',
    alignItems: 'center',
    whiteSpace: 'nowrap',
  }}>
    <V color="var(--accent)" bold>W</V>
    <Op>=</Op>
    <Op muted>(</Op>
    <V color="var(--risk-critical)" sup="th">Q</V>
    <Op>+</Op>
    <V color="var(--risk-high)">q</V>
    <Op>+</Op>
    <Op muted>(</Op>
    <Num>1</Num>
    <Op>−</Op>
    <V color="var(--risk-info)" sup="re">Q</V>
    <Op muted>)</Op>
    <Op muted>)</Op>
    <Op>/</Op>
    <Num>3</Num>
    <Op>·</Op>
    <V color="var(--risk-medium)" bold>Z</V>
  </div>
);
