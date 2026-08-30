import React from "react";
import { Btn } from "../design";

export interface WhatIfChip {
  id: number;
  name: string;
}

export interface WhatIfBarProps {
  disabledChips: WhatIfChip[];
  baselineW: number;
  simulatedW: number;
  delta: number;
  onReset: () => void;
  onRemoveChip: (id: number) => void;
  onSaveNote: () => void;
}

export const WhatIfBar: React.FC<WhatIfBarProps> = ({
  disabledChips, baselineW, simulatedW, delta, onReset, onRemoveChip, onSaveNote,
}) => {
  if (disabledChips.length === 0) return null;

  const sign = delta > 0 ? '+' : '';
  const deltaColor = delta > 0 ? 'var(--risk-critical)' : delta < 0 ? 'var(--risk-low, #58a463)' : 'var(--fg-muted)';

  return (
    <div
      role="status"
      aria-live="polite"
      style={{
        position: 'sticky',
        bottom: 12,
        marginTop: 12,
        padding: '12px 16px',
        background: 'var(--bg-elev-2)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--r-md)',
        boxShadow: '0 4px 16px rgba(0,0,0,0.25)',
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        flexWrap: 'wrap',
      }}
    >
      <span style={{
        fontSize: 'var(--text-xs)',
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: 'var(--fg-dim)',
        fontFamily: 'var(--font-mono)',
      }}>
        Симуляция
      </span>

      <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
        {disabledChips.map(c => (
          <button
            key={c.id}
            onClick={() => onRemoveChip(c.id)}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '4px 8px',
              fontSize: 'var(--text-xs)',
              background: 'var(--bg-elev-3)',
              border: '1px solid var(--border)',
              borderRadius: 'var(--r-sm)',
              cursor: 'pointer',
              color: 'var(--fg)',
            }}
          >
            {c.name} <span aria-hidden="true">✕</span>
          </button>
        ))}
      </div>

      <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>
        <span className="num" style={{ color: 'var(--fg-muted)' }}>{baselineW.toFixed(2)}</span>
        {' → '}
        <span className="num" style={{ color: 'var(--fg)', fontWeight: 600 }}>{simulatedW.toFixed(2)}</span>
        <span className="num" style={{ marginLeft: 8, color: deltaColor }}>
          ΔW {sign}{delta.toFixed(2)}
        </span>
      </span>

      <Btn variant="outline" onClick={onReset}>Сбросить</Btn>
      <Btn variant="outline" onClick={onSaveNote}>Сохранить заметку</Btn>
    </div>
  );
};
