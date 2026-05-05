import React, { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { authFetch } from "../api/client";
import { Asset, RiskOverviewPoint } from "../types";
import { Btn, Card, Chip, Icon, IconBtn, RiskBadge } from "../components/design";

type Level = 'critical' | 'high' | 'medium' | 'low';

interface EnrichedAsset {
  asset: Asset;
  wMax: number;
  level: Level | null;
  threatCount: number;
}

function contourLabel(a: Asset): string {
  return a.is_isolated ? 'Изолированный (Z=0.5)' : 'Открытый (Z=1.0)';
}

function envLabel(env: string): string {
  switch (env) {
    case 'prod': return 'Production';
    case 'test': return 'Test';
    case 'dev': return 'Development';
    default: return env || '—';
  }
}

export const AssetsPage: React.FC = () => {
  const navigate = useNavigate();
  const [assets, setAssets] = useState<Asset[]>([]);
  const [points, setPoints] = useState<RiskOverviewPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  const [sel, setSel] = useState<EnrichedAsset | null>(null);
  const [view, setView] = useState<'table' | 'cards'>('table');
  const [filterRisk, setFilterRisk] = useState<'all' | Level>('all');
  const [search, setSearch] = useState('');

  useEffect(() => {
    Promise.all([
      authFetch('/api/assets').then(r => r.ok ? r.json() : Promise.reject(new Error(`assets: HTTP ${r.status}`))),
      authFetch('/api/risk/overview').then(r => r.ok ? r.json() : Promise.reject(new Error(`overview: HTTP ${r.status}`))),
    ]).then(([a, p]) => {
      setAssets(Array.isArray(a) ? a : []);
      setPoints(Array.isArray(p) ? p : []);
      setLoading(false);
    }).catch(e => { setErr(e.message); setLoading(false); });
  }, []);

  const enriched = useMemo((): EnrichedAsset[] => {
    return assets.map(asset => {
      const ap = points.filter(p => p.asset_id === asset.id);
      const wMax = ap.reduce((m, p) => Math.max(m, p.w ?? 0), 0);
      const topPoint = ap.reduce((best: RiskOverviewPoint | null, p) => (!best || (p.w ?? 0) > (best.w ?? 0)) ? p : best, null);
      return {
        asset,
        wMax,
        level: topPoint ? (topPoint.level as Level) : null,
        threatCount: ap.length,
      };
    });
  }, [assets, points]);

  const filtered = useMemo(() => {
    return enriched.filter(e => {
      if (filterRisk !== 'all' && e.level !== filterRisk) return false;
      if (search.trim()) {
        const s = search.toLowerCase();
        const match = e.asset.name.toLowerCase().includes(s)
          || (e.asset.owner ?? '').toLowerCase().includes(s)
          || String(e.asset.id).includes(s);
        if (!match) return false;
      }
      return true;
    });
  }, [enriched, filterRisk, search]);

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Реестр ИТ-активов</div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Активы организации</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            {filtered.length} из {assets.length} · риск рассчитан по формуле W (модель ПТСЗИ)
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <Btn variant="outline" icon={<Icon name="upload" size={13} />}>Импорт CSV</Btn>
          <Btn variant="outline" icon={<Icon name="download" size={13} />}>Экспорт</Btn>
          <Btn variant="primary" icon={<Icon name="plus" size={13} />} onClick={() => navigate('/assets/new')}>Новый актив</Btn>
        </div>
      </div>

      {err && (
        <div style={{ padding: 14, background: 'var(--risk-critical-bg)', border: '1px solid var(--risk-critical-br)', borderRadius: 'var(--r-md)', color: 'var(--risk-critical)', fontSize: 'var(--text-sm)' }}>⚠ {err}</div>
      )}

      {loading && !err && (
        <div style={{ textAlign: 'center', padding: 60, color: 'var(--fg-dim)' }}>Загрузка…</div>
      )}

      {!loading && !err && (
        <>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 12px', background: 'var(--bg-elev-1)', border: '1px solid var(--border)', borderRadius: 'var(--r-md)' }}>
            <Icon name="search" size={14} color="var(--fg-dim)" />
            <input
              placeholder="Поиск по названию, ID, владельцу..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              style={{ flex: 1, background: 'transparent', border: 'none', outline: 'none', fontSize: 'var(--text-sm)', color: 'var(--fg)' }} />
            <div style={{ width: 1, height: 20, background: 'var(--border)' }} />
            <div style={{ display: 'flex', gap: 4 }}>
              {([{ v: 'all', l: 'Все' }, { v: 'critical', l: 'Крит.' }, { v: 'high', l: 'Высок.' }, { v: 'medium', l: 'Средн.' }, { v: 'low', l: 'Низк.' }] as const).map(o => (
                <button key={o.v} onClick={() => setFilterRisk(o.v as any)} style={{
                  height: 24, padding: '0 10px',
                  background: filterRisk === o.v ? 'var(--bg-active)' : 'transparent',
                  color: filterRisk === o.v ? 'var(--fg)' : 'var(--fg-muted)',
                  border: `1px solid ${filterRisk === o.v ? 'var(--border-strong)' : 'transparent'}`,
                  borderRadius: 'var(--r-xs)', fontSize: 'var(--text-xs)', cursor: 'pointer',
                  fontFamily: o.v !== 'all' ? 'var(--font-mono)' : 'inherit',
                }}>{o.l}</button>
              ))}
            </div>
            <div style={{ flex: 1 }} />
            <IconBtn size={26} active={view === 'table'} onClick={() => setView('table')} title="Таблица"><Icon name="table" size={14} /></IconBtn>
            <IconBtn size={26} active={view === 'cards'} onClick={() => setView('cards')} title="Карточки"><Icon name="layoutGrid" size={14} /></IconBtn>
          </div>

          {view === 'table' ? (
            <Card pad={0}>
              <div style={{ overflow: 'auto' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 'var(--text-sm)' }}>
                  <thead>
                    <tr style={{ background: 'var(--bg-elev-2)', borderBottom: '1px solid var(--border)' }}>
                      {['ID', 'Актив', 'Среда', 'Контур', 'Угроз', 'W max', 'Уровень', ''].map((h, i) => (
                        <th key={i} style={{
                          padding: '8px 10px', textAlign: 'left',
                          fontSize: 'var(--text-xs)', fontWeight: 500, color: 'var(--fg-dim)',
                          textTransform: 'uppercase', letterSpacing: '0.05em',
                          whiteSpace: 'nowrap',
                          position: 'sticky', top: 0, background: 'var(--bg-elev-2)', zIndex: 2,
                        }}>{h}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((e, idx) => {
                      const { asset, level, wMax, threatCount } = e;
                      return (
                        <tr key={asset.id}
                          onClick={() => setSel(e)}
                          style={{
                            borderBottom: idx < filtered.length - 1 ? '1px solid var(--border)' : 'none',
                            cursor: 'pointer',
                            background: sel?.asset.id === asset.id ? 'var(--accent-ghost)' : 'transparent',
                            transition: 'background 120ms'
                          }}
                          onMouseEnter={ev => { if (sel?.asset.id !== asset.id) ev.currentTarget.style.background = 'var(--bg-hover)'; }}
                          onMouseLeave={ev => { if (sel?.asset.id !== asset.id) ev.currentTarget.style.background = 'transparent'; }}
                        >
                          <td style={{ padding: '10px', fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-muted)' }}>A-{String(asset.id).padStart(3, '0')}</td>
                          <td style={{ padding: '10px' }}>
                            <div style={{ fontWeight: 500 }}>{asset.name}</div>
                            <div style={{ fontSize: 11, color: 'var(--fg-dim)', marginTop: 2 }}>{asset.owner ?? '—'}</div>
                          </td>
                          <td style={{ padding: '10px', color: 'var(--fg-muted)' }}>{envLabel(asset.environment)}</td>
                          <td style={{ padding: '10px' }}>
                            <Chip tone={asset.is_isolated ? 'ghost' : 'warn'} mono>
                              {asset.is_isolated ? 'Z=0.5' : 'Z=1.0'}
                            </Chip>
                          </td>
                          <td style={{ padding: '10px', fontFamily: 'var(--font-mono)', fontSize: 12, textAlign: 'center', color: 'var(--fg-muted)' }}>{threatCount}</td>
                          <td style={{ padding: '10px' }}>
                            {level ? (
                              <div style={{ width: 100, height: 4, background: 'var(--bg-elev-3)', borderRadius: 2, overflow: 'hidden' }}>
                                <div style={{ width: `${Math.min(100, wMax * 100)}%`, height: '100%', background: `var(--risk-${level})` }} />
                              </div>
                            ) : <span style={{ color: 'var(--fg-faint)', fontSize: 11 }}>—</span>}
                          </td>
                          <td style={{ padding: '10px' }}>
                            {level ? <RiskBadge level={level} score={Number(wMax.toFixed(2))} compact /> : <span style={{ color: 'var(--fg-faint)', fontSize: 11 }}>—</span>}
                          </td>
                          <td style={{ padding: '10px', width: 24 }}>
                            <IconBtn size={22}><Icon name="chevronR" size={12} /></IconBtn>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
              <div style={{ padding: '10px 14px', borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 'var(--text-xs)', color: 'var(--fg-dim)' }}>
                <span>Показано {filtered.length} из {assets.length}</span>
              </div>
            </Card>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 12 }}>
              {filtered.map(e => {
                const { asset, level, wMax } = e;
                return (
                  <div key={asset.id} onClick={() => setSel(e)} style={{
                    padding: 14, background: 'var(--bg-elev-1)', border: '1px solid var(--border)',
                    borderRadius: 'var(--r-lg)', cursor: 'pointer', position: 'relative', overflow: 'hidden'
                  }}>
                    {level && <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 2, background: `var(--risk-${level})` }} />}
                    <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 8 }}>
                      <span className="mono" style={{ fontSize: 10, color: 'var(--fg-dim)' }}>A-{String(asset.id).padStart(3, '0')}</span>
                      {level && <RiskBadge level={level} score={Number(wMax.toFixed(2))} compact />}
                    </div>
                    <div style={{ fontSize: 'var(--text-md)', fontWeight: 500, marginBottom: 4 }}>{asset.name}</div>
                    <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginBottom: 12 }}>{envLabel(asset.environment)} · {asset.owner ?? '—'}</div>
                    <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                      <Chip tone={asset.is_isolated ? 'ghost' : 'warn'} mono>{asset.is_isolated ? 'Z=0.5' : 'Z=1.0'}</Chip>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </>
      )}

      {sel && (
        <div onClick={() => setSel(null)} style={{ position: 'fixed', inset: 0, background: 'oklch(0 0 0 / 0.4)', zIndex: 50, display: 'flex', justifyContent: 'flex-end' }}>
          <div onClick={e => e.stopPropagation()} style={{
            width: 480, background: 'var(--bg-elev-1)', borderLeft: '1px solid var(--border)',
            height: '100vh', overflow: 'auto', boxShadow: 'var(--sh-lg)',
            display: 'flex', flexDirection: 'column'
          }}>
            <div style={{ padding: '14px 18px', borderBottom: '1px solid var(--border)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <div className="mono" style={{ fontSize: 11, color: 'var(--fg-dim)' }}>A-{String(sel.asset.id).padStart(3, '0')}</div>
                <div style={{ fontSize: 'var(--text-md)', fontWeight: 600 }}>{sel.asset.name}</div>
              </div>
              <IconBtn onClick={() => setSel(null)}><Icon name="x" size={14} /></IconBtn>
            </div>
            <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, fontSize: 'var(--text-xs)' }}>
                <div><div style={{ color: 'var(--fg-dim)' }}>Среда</div><div>{envLabel(sel.asset.environment)}</div></div>
                <div><div style={{ color: 'var(--fg-dim)' }}>Владелец</div><div>{sel.asset.owner ?? '—'}</div></div>
                <div style={{ gridColumn: 'span 2' }}>
                  <div style={{ color: 'var(--fg-dim)' }}>Контур (Z в формуле W)</div>
                  <div>{contourLabel(sel.asset)}</div>
                </div>
              </div>
              {sel.level && (
                <div>
                  <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 500, marginBottom: 8 }}>Текущий риск (W max)</div>
                  <div style={{ padding: 14, background: 'var(--bg-elev-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-md)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <RiskBadge level={sel.level} />
                    <div className="num" style={{ fontSize: 30, fontWeight: 600, color: `var(--risk-${sel.level})` }}>{sel.wMax.toFixed(2)}</div>
                  </div>
                </div>
              )}
              <Btn variant="primary" fullWidth icon={<Icon name="flow" size={13} />} onClick={() => navigate(`/risk/graph/${sel.asset.id}`)}>Граф атаки</Btn>
              <Btn variant="outline" fullWidth icon={<Icon name="edit" size={13} />} onClick={() => navigate(`/assets/edit/${sel.asset.id}`)}>Редактировать</Btn>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};
