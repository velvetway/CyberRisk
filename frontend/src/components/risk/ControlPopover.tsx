import React, { useEffect, useRef } from "react";
import { Btn } from "../design";
import type { ControlCoverage } from "../../types/riskGraph";

export interface ControlPopoverProps {
  control: ControlCoverage;
  disabled: boolean;
  anchor: { x: number; y: number };
  onToggle: (id: number) => void;
  onClose: () => void;
}

export const ControlPopover: React.FC<ControlPopoverProps> = ({
  control, disabled, anchor, onToggle, onClose,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  const toggleRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    const previouslyFocused = document.activeElement as HTMLElement | null;
    toggleRef.current?.focus();

    const onDocClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('mousedown', onDocClick);
    document.addEventListener('keydown', onKey);

    return () => {
      document.removeEventListener('mousedown', onDocClick);
      document.removeEventListener('keydown', onKey);
      previouslyFocused?.focus?.();
    };
  }, [onClose]);

  return (
    <div
      ref={ref}
      role="dialog"
      aria-label={`Контроль ${control.name}`}
      style={{
        position: 'fixed',
        left: anchor.x - 130,
        top: anchor.y - 140,
        width: 260,
        background: 'var(--bg-elev-2)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--r-md)',
        padding: 14,
        boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
        zIndex: 1000,
        fontSize: 'var(--text-sm)',
      }}
    >
      <div style={{
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-dim)',
        fontFamily: 'var(--font-mono)',
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
      }}>
        СЗИ · C-{control.id}
      </div>
      <div style={{
        marginTop: 6,
        fontWeight: 600,
        fontSize: 'var(--text-md)',
        lineHeight: 1.3,
      }}>
        {control.name}
      </div>
      <div style={{ marginTop: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)' }}>Coverage</span>
        <span className="num" style={{
          fontFamily: 'var(--font-mono)',
          fontSize: 'var(--text-md)',
          color: 'var(--risk-low, #58a463)',
        }}>
          {control.coverage.toFixed(2)}
        </span>
      </div>
      <div style={{
        marginTop: 4,
        height: 4,
        background: 'var(--bg-elev-3)',
        borderRadius: 999,
        overflow: 'hidden',
      }}>
        <div style={{
          width: `${Math.max(0, Math.min(1, control.coverage)) * 100}%`,
          height: '100%',
          background: 'var(--risk-low, #58a463)',
        }} />
      </div>
      <div style={{ marginTop: 14, display: 'flex', gap: 8 }}>
        <span ref={(el) => {
          toggleRef.current = el?.querySelector('button') ?? null;
        }}>
          <Btn variant={disabled ? 'primary' : 'outline'} onClick={() => onToggle(control.id)}>
            {disabled ? 'Включить' : 'Выкл в симуляции'}
          </Btn>
        </span>
        <Btn variant="outline" onClick={onClose}>Закрыть</Btn>
      </div>
    </div>
  );
};
