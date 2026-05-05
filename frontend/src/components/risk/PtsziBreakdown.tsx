import React from "react";
import { Card, PtsziFormula, RiskBadge } from "../design";
import type { AttackPath } from "../../types/riskGraph";

type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

const levelFromW = (w: number): Level => {
  if (w >= 0.75) return 'critical';
  if (w >= 0.50) return 'high';
  if (w >= 0.25) return 'medium';
  return 'low';
};

export interface PtsziBreakdownProps {
  path: AttackPath;
  /** Optional what-if simulation overrides for q_reaction and w. */
  simulated?: { qReaction: number; w: number; delta: number };
}

export const PtsziBreakdown: React.FC<PtsziBreakdownProps> = ({ path, simulated }) => {
  const effectiveQR = simulated?.qReaction ?? path.q_reaction;
  const effectiveW = simulated?.w ?? path.w;
  const level = levelFromW(effectiveW);

  const rows = [
    { label: 'Q^th · потенциал угрозы', v: path.q_threat, info: path.threat.name, color: 'var(--risk-critical)' },
    { label: 'q · опасность уязвимостей', v: path.q_severity, info: `${path.vulnerable_links.length} VL`, color: 'var(--risk-high)' },
    { label: 'Q^re · покрытие СЗИ', v: effectiveQR, info: 'Ниже = риск выше (1−Q^re)', color: 'var(--risk-info)' },
    { label: 'Z · вес контура', v: path.z, info: 'По окружению актива', color: 'var(--risk-medium)' },
  ];

  return (
    <Card title="Формула ПТСЗИ" subtitle="Выбранный сценарий" dense>
      <div style={{ marginBottom: 14 }}>
        <PtsziFormula size="lg" align="center" />
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {rows.map(p => (
          <div key={p.label} style={{
            padding: '8px 10px', background: 'var(--bg-elev-2)',
            border: '1px solid var(--border)', borderRadius: 'var(--r-sm)',
          }}>
            <div style={{
              display: 'flex', justifyContent: 'space-between', alignItems: 'baseline',
              marginBottom: 4,
            }}>
              <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)' }}>{p.label}</span>
              <span className="num" style={{ fontSize: 'var(--text-sm)', color: p.color, fontWeight: 600 }}>
                {p.v.toFixed(2)}
              </span>
            </div>
            <div style={{
              height: 3, background: 'var(--bg-elev-3)', borderRadius: 999, overflow: 'hidden',
            }}>
              <div style={{
                width: `${Math.max(0, Math.min(1, p.v)) * 100}%`,
                height: '100%', background: p.color, transition: 'width 300ms',
              }} />
            </div>
            <div style={{ fontSize: 10, color: 'var(--fg-faint)', marginTop: 4 }}>{p.info}</div>
          </div>
        ))}
      </div>

      <div style={{
        marginTop: 14, padding: '12px 14px',
        background: 'var(--accent-ghost)',
        border: '1px solid oklch(0.62 var(--accent-c) var(--accent-h) / 0.35)',
        borderRadius: 'var(--r-md)',
      }}>
        <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', marginBottom: 4 }}>
          {simulated ? 'W (оценка симуляции, клиент-сайд)' : 'Итоговый вес сценария W'}
        </div>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
          <span className="num" style={{
            fontSize: 28, fontWeight: 600, color: 'var(--accent)',
            letterSpacing: '-0.02em',
          }}>
            {effectiveW.toFixed(3)}
          </span>
          <RiskBadge level={level} />
          {simulated && simulated.delta !== 0 && (
            <span className="num" style={{
              fontSize: 'var(--text-sm)',
              color: simulated.delta > 0 ? 'var(--risk-critical)' : 'var(--risk-low)',
              fontFamily: 'var(--font-mono)',
            }}>
              {simulated.delta > 0 ? '+' : ''}{simulated.delta.toFixed(2)}
            </span>
          )}
        </div>
        <div style={{
          fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginTop: 6,
          fontFamily: 'var(--font-mono)',
        }}>
          ({path.q_threat.toFixed(2)} + {path.q_severity.toFixed(2)} + {(1 - effectiveQR).toFixed(2)}) / 3 × {path.z.toFixed(2)}
        </div>
      </div>
    </Card>
  );
};
