import React, { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import toast from "react-hot-toast";
import { authFetch } from "../api/client";
import { RiskOverviewPoint } from "../types";
import { Btn, Card, Chip, Icon, RiskBadge } from "../components/design";

const IMPACT_LABELS = ['Минимальное', 'Низкое', 'Умеренное', 'Высокое', 'Критическое'];
const LIKELIHOOD_LABELS = ['Маловероятно', 'Редко', 'Возможно', 'Вероятно', 'Почти наверняка'];

type Level = 'critical' | 'high' | 'medium' | 'low';

function riskLevelForCell(i: number, j: number): Level {
  const score = (i + 1) * (j + 1);
  if (score >= 16) return 'critical';
  if (score >= 11) return 'high';
  if (score >= 6) return 'medium';
  return 'low';
}

export const RiskMapPage: React.FC = () => {
  const navigate = useNavigate();
  const [points, setPoints] = useState<RiskOverviewPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hover, setHover] = useState<{ i: number; j: number } | null>(null);
  const [selected, setSelected] = useState<{ i: number; j: number } | null>(null);

  useEffect(() => {
    authFetch('/api/risk/overview')
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
      .then(data => {
        setPoints(Array.isArray(data) ? data : []);
        setLoading(false);
      })
      .catch(e => { setError(e.message); setLoading(false); });
  }, []);

  // Aggregate by (impact-1, likelihood-1) cell
  const matrix = useMemo(() => {
    const m: number[][] = Array.from({ length: 5 }, () => Array(5).fill(0));
    for (const p of points) {
      const i = Math.max(0, Math.min(4, Math.round(p.impact) - 1));
      const j = Math.max(0, Math.min(4, Math.round(p.likelihood) - 1));
      m[i][j] += 1;
    }
    return m;
  }, [points]);

  const total = points.length;
  const counts = useMemo(() => {
    const c: Record<Level, number> = { critical: 0, high: 0, medium: 0, low: 0 };
    for (let i = 0; i < 5; i++) for (let j = 0; j < 5; j++) {
      const lvl = riskLevelForCell(i, j);
      c[lvl] += matrix[i][j];
    }
    return c;
  }, [matrix]);

  const cellSize = 68;
  const current = hover || selected;

  const pointsInCell = (i: number, j: number): RiskOverviewPoint[] =>
    points.filter(p => {
      const pi = Math.round(p.impact) - 1;
      const pj = Math.round(p.likelihood) - 1;
      return pi === i && pj === j;
    });

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Карта рисков · 5×5</div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Влияние × Вероятность</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            {total} рисковых сценариев · ФСТЭК: score = impact × likelihood (1–25)
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <Btn variant="outline" icon={<Icon name="download" size={13} />} onClick={() => toast('PDF-экспорт: в разработке')}>PDF-отчёт</Btn>
          <Btn variant="primary" icon={<Icon name="plus" size={13} />} onClick={() => navigate('/risk/preview')}>Новый сценарий</Btn>
        </div>
      </div>

      {error && (
        <div style={{ padding: 14, background: 'var(--risk-critical-bg)', border: '1px solid var(--risk-critical-br)', borderRadius: 'var(--r-md)', color: 'var(--risk-critical)', fontSize: 'var(--text-sm)' }}>
          ⚠ Не удалось загрузить данные: {error}
        </div>
      )}

      {loading && !error && (
        <div style={{ textAlign: 'center', padding: 60, color: 'var(--fg-dim)' }}>Загрузка…</div>
      )}

      {!loading && !error && (
        <>
          {/* Legend + filters */}
          <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
            {(['critical', 'high', 'medium', 'low'] as Level[]).map(lvl => (
              <div key={lvl} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 10px', background: 'var(--bg-elev-1)', border: '1px solid var(--border)', borderRadius: 'var(--r-sm)' }}>
                <div style={{ width: 10, height: 10, borderRadius: 2, background: `var(--risk-${lvl})` }} />
                <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)' }}>
                  {lvl === 'critical' ? 'Критический' : lvl === 'high' ? 'Высокий' : lvl === 'medium' ? 'Средний' : 'Низкий'}
                </span>
                <span className="num" style={{ fontSize: 'var(--text-xs)', color: 'var(--fg)' }}>{counts[lvl]}</span>
              </div>
            ))}
            <div style={{ flex: 1 }} />
            <Chip tone="ghost" icon={<Icon name="compass" size={12} />}>Показано: все активы</Chip>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: 16 }}>
            {/* Matrix */}
            <Card pad={20}>
              <div style={{ display: 'flex', gap: 12 }}>
                {/* Y axis label */}
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', paddingTop: 24 }}>
                  <div style={{ writingMode: 'vertical-rl' as any, transform: 'rotate(180deg)', fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.1em', fontWeight: 500 }}>Влияние (impact)</div>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 4, flex: 1 }}>
                  <div style={{ display: 'grid', gridTemplateColumns: `36px repeat(5, 1fr)`, gap: 4 }}>
                    <div />
                    {[1, 2, 3, 4, 5].map(n => (
                      <div key={n} style={{ textAlign: 'center', fontSize: 10, color: 'var(--fg-faint)', fontFamily: 'var(--font-mono)' }}>{LIKELIHOOD_LABELS[n - 1]}</div>
                    ))}
                  </div>
                  {[4, 3, 2, 1, 0].map(i => (
                    <div key={i} style={{ display: 'grid', gridTemplateColumns: `36px repeat(5, 1fr)`, gap: 4, alignItems: 'center' }}>
                      <div style={{ fontSize: 10, color: 'var(--fg-faint)', fontFamily: 'var(--font-mono)', textAlign: 'right', paddingRight: 4 }}>{IMPACT_LABELS[i]}</div>
                      {[0, 1, 2, 3, 4].map(j => {
                        const count = matrix[i][j];
                        const lvl = riskLevelForCell(i, j);
                        const score = (i + 1) * (j + 1);
                        const intensity = count > 0 ? Math.min(1, 0.4 + count / 10) : 0.08;
                        const isHover = hover && hover.i === i && hover.j === j;
                        const isSelected = selected && selected.i === i && selected.j === j;
                        return (
                          <div
                            key={j}
                            onMouseEnter={() => setHover({ i, j })}
                            onMouseLeave={() => setHover(null)}
                            onClick={() => setSelected({ i, j })}
                            style={{
                              height: cellSize,
                              background: count > 0 ? `color-mix(in oklch, var(--risk-${lvl}) ${Math.round(intensity * 90)}%, transparent)` : 'var(--bg-elev-2)',
                              border: `1px solid ${isSelected ? `var(--risk-${lvl})` : (isHover ? 'var(--border-strong)' : 'var(--border)')}`,
                              borderRadius: 'var(--r-sm)',
                              display: 'flex', alignItems: 'center', justifyContent: 'center',
                              flexDirection: 'column',
                              cursor: 'pointer',
                              transition: 'var(--transition)',
                              position: 'relative',
                              transform: isHover ? 'scale(1.03)' : 'scale(1)',
                              zIndex: isHover ? 2 : 1,
                            }}
                          >
                            <div className="num" style={{
                              fontSize: count > 0 ? 22 : 14,
                              fontWeight: 600,
                              color: count > 0 ? 'white' : 'var(--fg-faint)',
                              lineHeight: 1,
                              textShadow: count > 0 ? '0 1px 2px oklch(0 0 0 / 0.3)' : 'none'
                            }}>{count}</div>
                            <div style={{ fontSize: 9, color: count > 0 ? 'oklch(1 0 0 / 0.8)' : 'var(--fg-faint)', fontFamily: 'var(--font-mono)', marginTop: 2 }}>{score}</div>
                          </div>
                        );
                      })}
                    </div>
                  ))}
                  <div style={{ marginTop: 8, textAlign: 'center', fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.1em', fontWeight: 500 }}>Вероятность (likelihood) →</div>
                </div>
              </div>
            </Card>

            {/* Side panel */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <Card
                title={current ? `Ячейка (${current.i + 1}×${current.j + 1})` : 'Детали ячейки'}
                subtitle={current ? `${IMPACT_LABELS[current.i]} × ${LIKELIHOOD_LABELS[current.j]}` : 'Наведи на ячейку матрицы'}
                dense
              >
                {current ? (() => {
                  const lvl = riskLevelForCell(current.i, current.j);
                  const score = (current.i + 1) * (current.j + 1);
                  const list = pointsInCell(current.i, current.j);
                  return (
                    <div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 }}>
                        <RiskBadge level={lvl} />
                        <div className="num" style={{ fontSize: 28, fontWeight: 600, color: `var(--risk-${lvl})`, lineHeight: 1 }}>{score}</div>
                        <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)' }}>score</div>
                      </div>
                      <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 8 }}>
                        Сценариев в ячейке: {list.length}
                      </div>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, maxHeight: 320, overflow: 'auto' }}>
                        {list.length === 0 ? (
                          <div style={{ padding: 12, textAlign: 'center', color: 'var(--fg-dim)', fontSize: 'var(--text-xs)' }}>Нет рисков в ячейке</div>
                        ) : (
                          list.slice(0, 8).map(p => (
                            <div key={`${p.asset_id}-${p.threat_id}`}
                              onClick={() => navigate(`/risk/graph/${p.asset_id}?threat=${p.threat_id}`)}
                              style={{ padding: 8, background: 'var(--bg-elev-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-sm)', cursor: 'pointer' }}
                              onMouseEnter={e => (e.currentTarget.style.borderColor = 'var(--border-strong)')}
                              onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}
                            >
                              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 3 }}>
                                <span className="mono" style={{ fontSize: 10, color: 'var(--fg-muted)' }}>A-{String(p.asset_id).padStart(3, '0')}</span>
                                <span style={{ fontSize: 'var(--text-sm)', fontWeight: 500 }}>{p.asset_name}</span>
                              </div>
                              <div style={{ fontSize: 11, color: 'var(--fg-dim)' }}>{p.threat_name}</div>
                            </div>
                          ))
                        )}
                      </div>
                    </div>
                  );
                })() : (
                  <div style={{ padding: '20px 0', textAlign: 'center', color: 'var(--fg-dim)' }}>
                    <Icon name="target" size={32} color="var(--fg-faint)" />
                    <div style={{ fontSize: 'var(--text-sm)', marginTop: 10 }}>Выберите ячейку для просмотра</div>
                  </div>
                )}
              </Card>

              <Card title="Регуляторный множитель" subtitle="152-ФЗ · 187-ФЗ · ПП-1119" dense>
                <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', lineHeight: 1.5 }}>
                  Учитывается в итоговой оценке как коэффициент <span className="mono" style={{ color: 'var(--accent)' }}>RegulatoryFactor</span>:
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 10 }}>
                  {[
                    { k: 'КИИ 1 категории', v: '×1.8' },
                    { k: 'КИИ 2 категории', v: '×1.5' },
                    { k: 'ПДн УЗ-1, >1M субъектов', v: '×1.6' },
                    { k: 'ПДн УЗ-2', v: '×1.3' },
                    { k: 'Гостайна', v: '×2.0' },
                  ].map(x => (
                    <div key={x.k} style={{ display: 'flex', justifyContent: 'space-between', fontSize: 'var(--text-xs)', padding: '5px 0', borderBottom: '1px dashed var(--border)' }}>
                      <span style={{ color: 'var(--fg-muted)' }}>{x.k}</span>
                      <span className="num" style={{ color: 'var(--risk-high)', fontWeight: 500 }}>{x.v}</span>
                    </div>
                  ))}
                </div>
              </Card>
            </div>
          </div>
        </>
      )}
    </motion.div>
  );
};
