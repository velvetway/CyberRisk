import React from 'react';

// PTSZI formula component with proper typographic fraction layout:
// W = ( Q^threat + q^threat + (1 − Q^reaction) )  /  3  ·  Z
// Rendered as a real stacked fraction (numerator / line / denominator).

interface VarProps {
  color: string;
  children: React.ReactNode;
  sup?: string;
  bold?: boolean;
  baseSize: number;
}

const Var: React.FC<VarProps> = ({ color, children, sup, bold, baseSize }) => (
  <span style={{ display: 'inline-flex', alignItems: 'baseline' }}>
    <span style={{
      color,
      fontStyle: 'italic',
      fontWeight: bold ? 700 : 600,
      fontSize: baseSize,
      letterSpacing: '-0.01em',
      fontFamily: "'Inter', 'Inter Display', -apple-system, sans-serif",
      lineHeight: 1,
    }}>
      {children}
    </span>
    {sup && (
      <span style={{
        fontSize: Math.max(10, Math.round(baseSize * 0.42)),
        color,
        opacity: 0.85,
        fontWeight: 500,
        letterSpacing: '0.02em',
        alignSelf: 'flex-start',
        marginTop: `-${Math.round(baseSize * 0.22)}px`,
        marginLeft: 1,
        lineHeight: 1,
      }}>
        {sup}
      </span>
    )}
  </span>
);

const Op: React.FC<{ children: React.ReactNode; size: number; muted?: boolean; gap?: number }> = ({ children, size, muted, gap = 0.35 }) => (
  <span style={{
    color: muted ? 'var(--fg-dim)' : 'var(--fg-muted)',
    fontSize: size,
    margin: `0 ${gap}em`,
    fontWeight: 400,
    lineHeight: 1,
  }}>
    {children}
  </span>
);

const Numeric: React.FC<{ children: React.ReactNode; size: number; bold?: boolean }> = ({ children, size, bold }) => (
  <span style={{
    color: 'var(--fg)',
    fontFamily: 'var(--font-mono)',
    fontSize: Math.round(size * 0.9),
    fontWeight: bold ? 600 : 500,
    fontVariantNumeric: 'tabular-nums',
    lineHeight: 1,
  }}>
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
  align = 'center',
  withBackground = true,
}) => {
  const baseSize = size === 'xl' ? 32 : size === 'lg' ? 24 : 19;
  const fractionLineThickness = Math.max(1.5, baseSize * 0.06);

  return (
    <div style={{
      padding: withBackground ? '24px 32px' : 0,
      background: withBackground ? 'var(--bg-elev-2)' : 'transparent',
      border: withBackground ? '1px solid var(--border)' : 'none',
      borderRadius: 'var(--r-md)',
      color: 'var(--fg)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: align === 'center' ? 'center' : 'flex-start',
      overflowX: 'auto',
      lineHeight: 1,
      minHeight: baseSize * 3.2,
    }}>
      {/* W */}
      <Var color="var(--accent)" bold baseSize={Math.round(baseSize * 1.15)}>W</Var>

      {/* = */}
      <Op size={baseSize} gap={0.55}>=</Op>

      {/* Fraction */}
      <span style={{
        display: 'inline-flex',
        flexDirection: 'column',
        alignItems: 'center',
        margin: `0 ${baseSize * 0.02}em`,
      }}>
        {/* Numerator */}
        <span style={{
          display: 'inline-flex',
          alignItems: 'baseline',
          whiteSpace: 'nowrap',
          padding: `0 ${baseSize * 0.02}em ${Math.round(baseSize * 0.3)}px`,
        }}>
          <Var color="var(--risk-critical)" sup="threat" baseSize={baseSize}>Q</Var>
          <Op size={baseSize}>+</Op>
          <Var color="var(--risk-high)" sup="threat" baseSize={baseSize}>q</Var>
          <Op size={baseSize}>+</Op>
          <Op size={baseSize} muted gap={0.12}>(</Op>
          <Numeric size={baseSize}>1</Numeric>
          <Op size={baseSize}>−</Op>
          <Var color="var(--risk-info)" sup="reaction" baseSize={baseSize}>Q</Var>
          <Op size={baseSize} muted gap={0.12}>)</Op>
        </span>

        {/* Fraction line */}
        <span style={{
          width: '100%',
          height: fractionLineThickness,
          background: 'var(--fg-muted)',
          opacity: 0.75,
          borderRadius: fractionLineThickness,
        }} />

        {/* Denominator */}
        <span style={{ padding: `${Math.round(baseSize * 0.3)}px 0 0` }}>
          <Numeric size={baseSize} bold>3</Numeric>
        </span>
      </span>

      {/* · Z */}
      <Op size={baseSize} gap={0.4}>·</Op>
      <Var color="var(--risk-medium)" bold baseSize={Math.round(baseSize * 1.15)}>Z</Var>
    </div>
  );
};
