import React, { useMemo, useState } from "react";
import { Card, RiskBadge } from "../design";
import type { AttackPath } from "../../types/riskGraph";

type SortKey = 'w' | 'uncovered' | 'name';
type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface ThreatListProps {
  paths: AttackPath[];
  selectedThreatId: number | null;
  onSelect: (threatId: number) => void;
}

export const ThreatList: React.FC<ThreatListProps> = ({ paths, selectedThreatId, onSelect }) => {
  const [sortKey, setSortKey] = useState<SortKey>('w');

  const sorted = useMemo(() => {
    const copy = [...paths];
    copy.sort((a, b) => {
      switch (sortKey) {
        case 'w':
          return b.w - a.w;
        case 'uncovered': {
          const au = a.vulnerable_links.filter(v => v.uncovered).length;
          const bu = b.vulnerable_links.filter(v => v.uncovered).length;
          return bu - au;
        }
        case 'name':
          return a.threat.name.localeCompare(b.threat.name, 'ru');
        default:
          return 0;
      }
    });
    return copy;
  }, [paths, sortKey]);

  if (paths.length === 0) {
    return (
      <Card title="Угрозы актива" dense>
        <div style={{ padding: 12, color: 'var(--fg-muted)' }}>
          У этого актива нет связанных угроз с непустыми путями.
        </div>
      </Card>
    );
  }

  return (
    <Card pad={0}>
      <div style={{
        padding: '12px 16px',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}>
        <div style={{
          fontSize: 'var(--text-xs)', color: 'var(--fg-dim)',
          textTransform: 'uppercase', letterSpacing: '0.08em', fontWeight: 500,
        }}>
          Угрозы актива · {paths.length}
        </div>
        <div style={{ display: 'flex', gap: 6, fontSize: 'var(--text-xs)' }}>
          {(['w', 'uncovered', 'name'] as SortKey[]).map(k => (
            <button
              key={k}
              onClick={() => setSortKey(k)}
              style={{
                padding: '4px 8px',
                background: sortKey === k ? 'var(--bg-elev-2)' : 'transparent',
                border: '1px solid var(--border)',
                borderRadius: 'var(--r-sm)',
                color: sortKey === k ? 'var(--fg)' : 'var(--fg-muted)',
                cursor: 'pointer',
                fontFamily: 'var(--font-mono)',
              }}
            >
              {k === 'w' ? 'W ↓' : k === 'uncovered' ? 'непокрыт.' : 'имя'}
            </button>
          ))}
        </div>
      </div>

      <div role="listbox" aria-label="Угрозы актива">
        {sorted.map(p => {
          const isSelected = p.threat.id === selectedThreatId;
          const uncoveredCount = p.vulnerable_links.filter(v => v.uncovered).length;
          const level = p.level as Level;

          return (
            <div
              key={p.threat.id}
              role="option"
              aria-selected={isSelected}
              tabIndex={0}
              onClick={() => onSelect(p.threat.id)}
              onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(p.threat.id); } }}
              style={{
                display: 'grid',
                gridTemplateColumns: '24px 110px 1fr 60px 80px 64px',
                gap: 12,
                alignItems: 'center',
                padding: '10px 16px',
                borderBottom: '1px solid var(--border)',
                background: isSelected ? 'var(--bg-elev-2)' : 'transparent',
                cursor: 'pointer',
                borderLeft: `3px solid ${isSelected ? 'var(--accent)' : 'transparent'}`,
              }}
            >
              <span style={{ color: isSelected ? 'var(--accent)' : 'var(--fg-faint)' }}>{isSelected ? '▶' : ''}</span>
              <span className="mono" style={{
                fontSize: 'var(--text-xs)',
                color: 'var(--fg-muted)',
                fontFamily: 'var(--font-mono)',
              }}>
                {p.threat.bdu_id || `T-${p.threat.id}`}
              </span>
              <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {p.threat.name}
              </span>
              <div style={{
                height: 6, background: 'var(--bg-elev-3)', borderRadius: 999, overflow: 'hidden',
              }}>
                <div style={{
                  width: `${Math.round(p.w * 100)}%`,
                  height: '100%',
                  background: `var(--risk-${level})`,
                  transition: 'width 200ms',
                }} />
              </div>
              <span className="num" style={{
                fontSize: 'var(--text-sm)',
                fontFamily: 'var(--font-mono)',
                color: 'var(--fg)',
                textAlign: 'right',
              }}>
                {p.w.toFixed(2)}
              </span>
              <span style={{ textAlign: 'right' }}>
                {uncoveredCount > 0 ? (
                  <span className="mono" style={{
                    fontSize: 'var(--text-xs)',
                    color: 'var(--risk-critical)',
                    fontFamily: 'var(--font-mono)',
                  }}>
                    {uncoveredCount} откр.
                  </span>
                ) : (
                  <RiskBadge level={level} />
                )}
              </span>
            </div>
          );
        })}
      </div>
    </Card>
  );
};
