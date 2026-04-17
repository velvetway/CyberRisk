import React, { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { authFetch } from "../api/client";
import { Asset, RiskOverviewPoint } from "../types";
import { Btn, Card, Chip, Icon, PtsziFormula, RiskBadge, Sparkline, StatCard } from "../components/design";
import { useNavigate } from "react-router-dom";

type Level = 'critical' | 'high' | 'medium' | 'low';

const levelOrder: Level[] = ['critical', 'high', 'medium', 'low'];
const levelLabel: Record<Level, string> = {
  critical: 'Критические',
  high: 'Высокие',
  medium: 'Средние',
  low: 'Низкие',
};

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const [assets, setAssets] = useState<Asset[]>([]);
  const [points, setPoints] = useState<RiskOverviewPoint[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      authFetch('/api/assets').then(r => r.ok ? r.json() : []),
      authFetch('/api/risk/overview').then(r => r.ok ? r.json() : []),
    ]).then(([a, p]) => {
      setAssets(Array.isArray(a) ? a : []);
      setPoints(Array.isArray(p) ? p : []);
      setLoading(false);
    }).catch(() => setLoading(false));
  }, []);

  // Derived: count by level
  const counts = useMemo(() => {
    const c: Record<Level, number> = { critical: 0, high: 0, medium: 0, low: 0 };
    for (const p of points) {
      const lvl = p.level as Level;
      if (c[lvl] !== undefined) c[lvl]++;
    }
    return c;
  }, [points]);

  const totalScore = useMemo(() => {
    if (points.length === 0) return 0;
    const avg = points.reduce((s, p) => s + (p.score ?? 0), 0) / points.length;
    return Number((avg * 4).toFixed(1)); // scale score 1..25 → 0..100-ish for display
  }, [points]);

  const topAssets = useMemo(() => {
    // aggregate max score per asset
    const byAsset: Record<number, { asset: Asset; maxScore: number; level: Level }> = {};
    for (const p of points) {
      const cur = byAsset[p.asset_id];
      const score = p.score ?? 0;
      if (!cur || cur.maxScore < score) {
        const asset = assets.find(a => a.id === p.asset_id);
        if (!asset) continue;
        byAsset[p.asset_id] = { asset, maxScore: score, level: (p.level as Level) };
      }
    }
    return Object.values(byAsset).sort((a, b) => b.maxScore - a.maxScore).slice(0, 5);
  }, [points, assets]);

  // Fake sparkline data (dashboard aesthetic)
  const spark1 = [12, 14, 11, 18, 22, 19, 24, 21, 28, 31, 27, 34, 38, 42, 39, 44];
  const spark2 = [8, 9, 11, 9, 12, 15, 13, 17, 16, 19, 22, 20, 24, 21, 23];
  const spark3 = [4, 3, 2, 4, 5, 3, 2, 1, 2, 1, 0, 1, 2, 1, 0];
  const spark4 = [60, 58, 62, 61, 65, 67, 66, 70, 69, 72, 74, 73, 76, 78, 81];

  // Count of assets that have at least one risk point
  const assetsAtRisk = new Set(points.map(p => p.asset_id)).size;
  const kiiCount = assets.filter(a => a.kii_category && a.kii_category !== 'none').length;

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 6 }}>
            <div style={{ width: 6, height: 6, borderRadius: 999, background: 'var(--risk-low)', boxShadow: '0 0 8px var(--risk-low)' }} />
            <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>
              Операционный центр · {new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} МСК
            </span>
          </div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Обзор киберрисков</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            {assets.length} активов под мониторингом · методология ФСТЭК БДУ · обновлено только что
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <Btn variant="outline" icon={<Icon name="refresh" size={13} />}>Пересчитать</Btn>
          <Btn variant="outline" icon={<Icon name="download" size={13} />}>Экспорт</Btn>
          <Btn variant="primary" icon={<Icon name="zap" size={13} />} onClick={() => navigate('/risk/preview')}>Симулятор</Btn>
        </div>
      </div>

      {/* Stat strip */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12 }}>
        <StatCard label="Общий риск-скор" value={totalScore.toFixed(1)} tone="high" mono
          sub={`Усреднён по ${points.length} сценариям`}
          sparkline={<Sparkline data={spark1} color="var(--risk-high)" />} />
        <StatCard label="Критических рисков" value={counts.critical} tone="critical" mono
          sub={`${counts.critical} требуют реакции в 24ч`}
          sparkline={<Sparkline data={spark2} color="var(--risk-critical)" />} />
        <StatCard label="Активов с риском" value={assetsAtRisk} tone="medium" mono
          sub={`Из ${assets.length} зарегистрированных`}
          sparkline={<Sparkline data={spark4} color="var(--risk-medium)" />} />
        <StatCard label="Объектов КИИ" value={kiiCount} tone="low" mono
          sub="Подпадают под 187-ФЗ"
          sparkline={<Sparkline data={spark3} color="var(--risk-low)" />} />
      </div>

      {/* Main section */}
      <div>
        {/* Risk breakdown */}
        <Card title="Распределение рисков по уровням" subtitle="Методология ФСТЭК · свод всех пар «актив × угроза»" dense
          action={<Chip tone="ghost" onClick={() => navigate('/risk/map')}>Открыть матрицу →</Chip>}>

          {/* Stacked bar */}
          <div>
            {(() => {
              const total = Math.max(1, points.length);
              return (
                <div style={{ display: 'flex', height: 28, borderRadius: 6, overflow: 'hidden', border: '1px solid var(--border)' }}>
                  {levelOrder.map(lvl => {
                    const n = counts[lvl];
                    const flex = (n / total) * 100;
                    if (flex === 0) return null;
                    return (
                      <div key={lvl} style={{
                        flex, background: `var(--risk-${lvl})`,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: 11,
                        color: lvl === 'critical' ? 'white' : 'oklch(0.2 0.01 60)',
                        fontWeight: 600, fontFamily: 'var(--font-mono)'
                      }}>{n}</div>
                    );
                  })}
                </div>
              );
            })()}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginTop: 14 }}>
              {levelOrder.map(lvl => (
                <div key={lvl} style={{ padding: 12, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-sm)', border: '1px solid var(--border)' }}>
                  <RiskBadge level={lvl} />
                  <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginTop: 10 }}>
                    <span className="num" style={{ fontSize: 'var(--text-2xl)', fontWeight: 600 }}>{counts[lvl]}</span>
                    <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)' }}>{levelLabel[lvl].toLowerCase()}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div style={{ marginTop: 20 }}>
            <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 500, marginBottom: 10 }}>
              Топ-5 активов по риску
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {topAssets.map((t, i) => (
                <div key={t.asset.id} onClick={() => navigate(`/assets/${t.asset.id}/risks`)}
                  style={{
                    display: 'flex', alignItems: 'center', gap: 12,
                    padding: '8px 10px', background: 'var(--bg-elev-2)',
                    borderRadius: 'var(--r-sm)', border: '1px solid var(--border)', cursor: 'pointer',
                  }}
                  onMouseEnter={e => (e.currentTarget.style.borderColor = 'var(--border-strong)')}
                  onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}
                >
                  <span className="mono" style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', width: 24 }}>{String(i + 1).padStart(2, '0')}</span>
                  <span className="mono" style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', width: 52 }}>A-{String(t.asset.id).padStart(3, '0')}</span>
                  <span style={{ flex: 1, fontSize: 'var(--text-sm)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.asset.name}</span>
                  <div style={{ width: 120 }}>
                    <div style={{ height: 4, background: 'var(--bg-elev-3)', borderRadius: 2, overflow: 'hidden' }}>
                      <div style={{ width: `${t.maxScore / 25 * 100}%`, height: '100%', background: `var(--risk-${t.level})` }} />
                    </div>
                  </div>
                  <RiskBadge level={t.level} score={t.maxScore} compact />
                </div>
              ))}
              {topAssets.length === 0 && !loading && (
                <div style={{ padding: 12, textAlign: 'center', color: 'var(--fg-dim)', fontSize: 'var(--text-sm)' }}>Нет данных для отображения</div>
              )}
            </div>
          </div>
        </Card>

      </div>

      {/* PTSZI formula callout */}
      <Card title="Формула ПТСЗИ · центральный алгоритм оценки" subtitle="docs/risk-model.md" dense>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <PtsziFormula size="xl" align="center" />
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 10, fontSize: 'var(--text-xs)' }}>
            <div style={{ padding: 10, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-sm)', border: '1px solid var(--border)' }}>
              <div style={{ color: 'var(--risk-critical)', fontWeight: 600, marginBottom: 4, fontSize: 13 }}><span style={{ fontStyle: 'italic' }}>Q</span><sup style={{ fontSize: 9, position: 'relative', top: -6, marginLeft: 2 }}>threat</sup></div>
              <div className="num" style={{ color: 'var(--fg-muted)' }}>степень реализации угрозы, 0…1</div>
            </div>
            <div style={{ padding: 10, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-sm)', border: '1px solid var(--border)' }}>
              <div style={{ color: 'var(--risk-high)', fontWeight: 600, marginBottom: 4, fontSize: 13 }}><span style={{ fontStyle: 'italic' }}>q</span><sup style={{ fontSize: 9, position: 'relative', top: -6, marginLeft: 2 }}>threat</sup></div>
              <div className="num" style={{ color: 'var(--fg-muted)' }}>степень опасности угрозы, 0…1</div>
            </div>
            <div style={{ padding: 10, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-sm)', border: '1px solid var(--border)' }}>
              <div style={{ color: 'var(--risk-info)', fontWeight: 600, marginBottom: 4, fontSize: 13 }}><span style={{ fontStyle: 'italic' }}>Q</span><sup style={{ fontSize: 9, position: 'relative', top: -6, marginLeft: 2 }}>reaction</sup></div>
              <div className="num" style={{ color: 'var(--fg-muted)' }}>степень предотвращения угрозы, 0…1</div>
            </div>
            <div style={{ padding: 10, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-sm)', border: '1px solid var(--border)' }}>
              <div style={{ color: 'var(--risk-medium)', fontWeight: 600, marginBottom: 4, fontSize: 13 }}><span style={{ fontStyle: 'italic' }}>Z</span></div>
              <div className="num" style={{ color: 'var(--fg-muted)' }}>критичность для контура, 0.5 или 1</div>
            </div>
          </div>
        </div>
      </Card>
    </motion.div>
  );
};
