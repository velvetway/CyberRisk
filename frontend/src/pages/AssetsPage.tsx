import React, { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { authFetch } from "../api/client";
import { Asset, RiskOverviewPoint } from "../types";
import { Btn, Card, Chip, Icon, IconBtn, RiskBadge } from "../components/design";

type Level = 'critical' | 'high' | 'medium' | 'low';

interface EnrichedAsset {
  asset: Asset;
  maxScore: number;
  level: Level | null;
  threatCount: number;
  vulnCount: number;
}

function criticalityTone(bc: number): { tone: 'danger' | 'warn' | 'ghost'; short: string } {
  if (bc >= 5) return { tone: 'danger', short: 'CRIT' };
  if (bc >= 4) return { tone: 'warn', short: 'HIGH' };
  if (bc >= 3) return { tone: 'ghost', short: 'MED' };
  return { tone: 'ghost', short: 'LOW' };
}

function segmentLabel(a: Asset): string {
  if (a.is_isolated) return 'Изолированный';
  if (a.has_internet_access) return 'DMZ';
  return 'Внутренний';
}

function kiiLabel(k: string | undefined | null): string {
  if (!k || k === 'none') return '—';
  if (k === 'cat1') return '1 категория';
  if (k === 'cat2') return '2 категория';
  if (k === 'cat3') return '3 категория';
  return k;
}

function pdnLabel(p: string | undefined | null): string {
  if (!p) return '—';
  if (/^uz\d$/.test(p)) return 'УЗ-' + p.slice(2);
  return p;
}

function regulatoryTags(a: Asset): string[] {
  const tags: string[] = [];
  if (a.kii_category && a.kii_category !== 'none') tags.push('КИИ');
  if (a.has_personal_data) tags.push('152-ФЗ');
  if (a.data_category === 'state_secret') tags.push('Гостайна');
  if (a.data_category === 'banking_secret') tags.push('КТайна');
  return tags;
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

  // Enrich assets with risk aggregates from overview
  const enriched = useMemo((): EnrichedAsset[] => {
    return assets.map(asset => {
      const ap = points.filter(p => p.asset_id === asset.id);
      const maxScore = ap.reduce((m, p) => Math.max(m, p.score ?? 0), 0);
      const topPoint = ap.reduce((best: RiskOverviewPoint | null, p) => (!best || (p.score ?? 0) > (best.score ?? 0)) ? p : best, null);
      return {
        asset,
        maxScore,
        level: topPoint ? (topPoint.level as Level) : null,
        threatCount: ap.length,
        vulnCount: 0, // no vuln-count endpoint yet
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
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Реестр ИТ-активов</div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Активы организации</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            {filtered.length} из {assets.length} · CIA, категория КИИ, уровень защищённости ПДн
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
          {/* Filter bar */}
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
                      {['ID', 'Актив', 'Тип', 'CIA', 'Критич.', 'КИИ', 'ПДн', 'Сегмент', 'Угроз', 'Риск', 'Скор', ''].map((h, i) => (
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
                      const { asset, level, maxScore, threatCount } = e;
                      const crit = criticalityTone(asset.business_criticality);
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
                          <td style={{ padding: '10px', color: 'var(--fg-muted)' }}>{asset.type ?? '—'}</td>
                          <td style={{ padding: '10px' }}>
                            <div style={{ display: 'flex', gap: 2, fontFamily: 'var(--font-mono)', fontSize: 11 }}>
                              {[
                                { k: 'C', v: asset.confidentiality },
                                { k: 'I', v: asset.integrity },
                                { k: 'A', v: asset.availability }
                              ].map(x => (
                                <span key={x.k} style={{
                                  width: 18, height: 18, display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
                                  background: x.v >= 4 ? 'var(--risk-critical-bg)' : x.v === 3 ? 'var(--risk-medium-bg)' : 'var(--bg-elev-3)',
                                  color: x.v >= 4 ? 'var(--risk-critical)' : x.v === 3 ? 'var(--risk-medium)' : 'var(--fg-dim)',
                                  borderRadius: 3, fontWeight: 600,
                                }}>{x.v}</span>
                              ))}
                            </div>
                          </td>
                          <td style={{ padding: '10px' }}>
                            <Chip tone={crit.tone} mono>{crit.short}</Chip>
                          </td>
                          <td style={{ padding: '10px', fontFamily: 'var(--font-mono)', fontSize: 11, color: asset.kii_category && asset.kii_category !== 'none' ? 'var(--risk-high)' : 'var(--fg-faint)' }}>{kiiLabel(asset.kii_category)}</td>
                          <td style={{ padding: '10px', fontFamily: 'var(--font-mono)', fontSize: 11, color: asset.protection_level ? 'var(--fg-muted)' : 'var(--fg-faint)' }}>{pdnLabel(asset.protection_level)}</td>
                          <td style={{ padding: '10px', fontSize: 11, color: 'var(--fg-muted)' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                              {asset.has_internet_access && <div title="Интернет" style={{ width: 5, height: 5, borderRadius: 999, background: 'var(--risk-high)' }} />}
                              {segmentLabel(asset)}
                            </div>
                          </td>
                          <td style={{ padding: '10px', fontFamily: 'var(--font-mono)', fontSize: 12, textAlign: 'center', color: 'var(--fg-muted)' }}>{threatCount}</td>
                          <td style={{ padding: '10px' }}>
                            {level ? (
                              <div style={{ width: 80, height: 4, background: 'var(--bg-elev-3)', borderRadius: 2, overflow: 'hidden' }}>
                                <div style={{ width: `${maxScore / 25 * 100}%`, height: '100%', background: `var(--risk-${level})` }} />
                              </div>
                            ) : <span style={{ color: 'var(--fg-faint)', fontSize: 11 }}>—</span>}
                          </td>
                          <td style={{ padding: '10px' }}>
                            {level ? <RiskBadge level={level} score={maxScore} compact /> : <span style={{ color: 'var(--fg-faint)', fontSize: 11 }}>—</span>}
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
                const { asset, level, maxScore } = e;
                return (
                  <div key={asset.id} onClick={() => setSel(e)} style={{
                    padding: 14, background: 'var(--bg-elev-1)', border: '1px solid var(--border)',
                    borderRadius: 'var(--r-lg)', cursor: 'pointer', position: 'relative', overflow: 'hidden'
                  }}>
                    {level && <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 2, background: `var(--risk-${level})` }} />}
                    <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 8 }}>
                      <span className="mono" style={{ fontSize: 10, color: 'var(--fg-dim)' }}>A-{String(asset.id).padStart(3, '0')}</span>
                      {level && <RiskBadge level={level} score={maxScore} compact />}
                    </div>
                    <div style={{ fontSize: 'var(--text-md)', fontWeight: 500, marginBottom: 4 }}>{asset.name}</div>
                    <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginBottom: 12 }}>{asset.type ?? '—'} · {asset.owner ?? '—'}</div>
                    <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                      {regulatoryTags(asset).map(t => <Chip key={t} tone="ghost" mono>{t}</Chip>)}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </>
      )}

      {/* Detail drawer */}
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
                <div><div style={{ color: 'var(--fg-dim)' }}>Тип</div><div>{sel.asset.type ?? '—'}</div></div>
                <div><div style={{ color: 'var(--fg-dim)' }}>Владелец</div><div>{sel.asset.owner ?? '—'}</div></div>
                <div><div style={{ color: 'var(--fg-dim)' }}>Сегмент</div><div>{segmentLabel(sel.asset)}</div></div>
                <div><div style={{ color: 'var(--fg-dim)' }}>Интернет</div><div>{sel.asset.has_internet_access ? 'Да' : 'Нет'}</div></div>
                <div><div style={{ color: 'var(--fg-dim)' }}>Категория КИИ</div><div className="mono">{kiiLabel(sel.asset.kii_category)}</div></div>
                <div><div style={{ color: 'var(--fg-dim)' }}>УЗ ПДн</div><div className="mono">{pdnLabel(sel.asset.protection_level)}</div></div>
              </div>
              {sel.level && (
                <div>
                  <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 500, marginBottom: 8 }}>Текущий риск</div>
                  <div style={{ padding: 14, background: 'var(--bg-elev-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-md)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <RiskBadge level={sel.level} />
                    <div className="num" style={{ fontSize: 30, fontWeight: 600, color: `var(--risk-${sel.level})` }}>{sel.maxScore.toFixed(1)}</div>
                  </div>
                </div>
              )}
              <Btn variant="primary" fullWidth icon={<Icon name="zap" size={13} />} onClick={() => navigate(`/assets/${sel.asset.id}/risks`)}>Профиль рисков</Btn>
              <Btn variant="outline" fullWidth icon={<Icon name="edit" size={13} />} onClick={() => navigate(`/assets/edit/${sel.asset.id}`)}>Редактировать</Btn>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};
