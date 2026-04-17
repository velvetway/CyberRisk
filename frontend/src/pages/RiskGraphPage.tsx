import React, { useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { authFetch } from "../api/client";
import { AttackPath } from "../types/riskGraph";
import { Btn, Card, Icon, RiskBadge } from "../components/design";

type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

function levelFromW(w: number): Level {
  if (w >= 0.75) return 'critical';
  if (w >= 0.50) return 'high';
  if (w >= 0.25) return 'medium';
  return 'low';
}

// Node layout constants
const NODE_W = 160;
const NODE_H = 44;
const NODE_GAP = 12;
const COL_X = [80, 320, 560, 800];
const SVG_W = 960;
const SVG_H = 520;

const bezierPath = (x1: number, y1: number, x2: number, y2: number, w: number) => {
  const midX = (x1 + x2) / 2;
  return `M ${x1} ${y1 - w / 2}
          C ${midX} ${y1 - w / 2}, ${midX} ${y2 - w / 2}, ${x2} ${y2 - w / 2}
          L ${x2} ${y2 + w / 2}
          C ${midX} ${y2 + w / 2}, ${midX} ${y1 + w / 2}, ${x1} ${y1 + w / 2}
          Z`;
};

export const RiskGraphPage: React.FC = () => {
  const { assetId } = useParams();
  const [params] = useSearchParams();
  const threatId = params.get('threat');
  const navigate = useNavigate();
  const [path, setPath] = useState<AttackPath | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [hoverNode, setHoverNode] = useState<string | null>(null);

  useEffect(() => {
    if (!assetId || !threatId) return;
    authFetch(`/api/risk/graph/${assetId}/${threatId}`)
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
      .then(setPath)
      .catch(e => setErr(e.message));
  }, [assetId, threatId]);

  // Derived layout
  const layout = useMemo(() => {
    if (!path) return null;

    // Column data — ID-to-node maps and Y positions
    const laneY = (count: number): number[] => {
      const total = count * NODE_H + (count - 1) * NODE_GAP;
      const startY = (SVG_H - total) / 2;
      return Array.from({ length: count }, (_, i) => startY + i * (NODE_H + NODE_GAP));
    };

    const sources = path.sources || [];
    const threats = [{ id: 'T', name: path.threat.name, bdu_id: path.threat.bdu_id }];
    const vulns = path.vulnerable_links || [];
    const actions = path.destructive_actions || [];

    const srcY = laneY(Math.max(1, sources.length));
    const thrY = laneY(1);
    const vulY = laneY(Math.max(1, vulns.length));
    const actY = laneY(Math.max(1, actions.length));

    const nodes = new Map<string, { x: number; y: number; column: number }>();
    sources.forEach((s, i) => nodes.set(`S${s.id}`, { x: COL_X[0], y: srcY[i], column: 0 }));
    nodes.set('ST', { x: COL_X[1], y: thrY[0], column: 1 });
    vulns.forEach((v, i) => nodes.set(`V${v.vulnerability_id}`, { x: COL_X[2], y: vulY[i], column: 2 }));
    actions.forEach((a, i) => nodes.set(`D${a.id}`, { x: COL_X[3], y: actY[i], column: 3 }));

    // Links:
    // S -> ST (weight = 1 / sources.length)
    // ST -> V (weight by severity or uniform)
    // V -> D: uniform distribution if vulns non-empty, else ST -> D direct
    const links: Array<{ from: string; to: string; weight: number }> = [];

    if (sources.length > 0) {
      const w = 1 / sources.length;
      sources.forEach(s => links.push({ from: `S${s.id}`, to: 'ST', weight: w }));
    }

    if (vulns.length > 0) {
      vulns.forEach(v => {
        const sev = Math.max(1, v.severity) / 10;
        links.push({ from: 'ST', to: `V${v.vulnerability_id}`, weight: sev });
      });
      vulns.forEach(v => {
        actions.forEach(a => {
          links.push({ from: `V${v.vulnerability_id}`, to: `D${a.id}`, weight: 0.15 });
        });
      });
    } else {
      // No VLs — draw ST → D directly
      actions.forEach(a => links.push({ from: 'ST', to: `D${a.id}`, weight: 0.3 }));
    }

    return { nodes, links, sources, threats, vulns, actions };
  }, [path]);

  if (!threatId) {
    return <div style={{ padding: 40, textAlign: 'center', color: 'var(--fg-muted)' }}>Укажите <code>?threat=&lt;id&gt;</code> в URL.</div>;
  }
  if (err) {
    return <div style={{ padding: 40 }}>
      <Card title="Ошибка загрузки" dense>
        <div style={{ color: 'var(--risk-critical)' }}>{err}</div>
        <Btn style={{ marginTop: 14 }} variant="outline" onClick={() => navigate(-1)} icon={<Icon name="arrowL" size={13} />}>Назад</Btn>
      </Card>
    </div>;
  }
  if (!path || !layout) {
    return <div style={{ padding: 60, textAlign: 'center', color: 'var(--fg-dim)' }}>Загрузка…</div>;
  }

  const level = levelFromW(path.w);

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Граф атаки · ПТСЗИ</div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Цепочка «Источник → Угроза → Уязвимость → Действие»</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            Актив <span className="mono" style={{ color: 'var(--fg)' }}>A-{String(path.asset.id).padStart(3, '0')} «{path.asset.name}»</span> · {layout.links.length} связей · толщина = вес сценария
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <Btn variant="outline" icon={<Icon name="arrowL" size={13} />} onClick={() => navigate(-1)}>Назад</Btn>
          <Btn variant="outline" icon={<Icon name="sliders" size={13} />}>Параметры</Btn>
          <Btn variant="primary" icon={<Icon name="file" size={13} />}>PDF сценария</Btn>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: 16 }}>
        <Card pad={0}>
          <div style={{ padding: '12px 16px', borderBottom: '1px solid var(--border)', display: 'flex', gap: 20 }}>
            {['Источник угрозы', 'Угроза (БДУ)', 'Уязвимость', 'Деструктивное действие'].map((t, i) => (
              <div key={t} style={{ flex: 1, textAlign: i === 0 ? 'left' : 'center' }}>
                <div style={{ fontSize: 'var(--text-2xs)', color: 'var(--fg-faint)', textTransform: 'uppercase', letterSpacing: '0.1em', fontWeight: 500 }}>{t}</div>
              </div>
            ))}
          </div>

          <div style={{ background: 'var(--bg-elev-1)', padding: 16 }}>
            <svg viewBox={`0 0 ${SVG_W} ${SVG_H}`} style={{ width: '100%', height: 'auto', display: 'block' }}>
              <defs>
                <linearGradient id="flowGradCrit" x1="0" x2="1" y1="0" y2="0">
                  <stop offset="0%" stopColor="var(--risk-critical)" stopOpacity="0.5" />
                  <stop offset="100%" stopColor="var(--risk-critical)" stopOpacity="0.15" />
                </linearGradient>
                <linearGradient id="flowGradHigh" x1="0" x2="1" y1="0" y2="0">
                  <stop offset="0%" stopColor="var(--risk-high)" stopOpacity="0.5" />
                  <stop offset="100%" stopColor="var(--risk-high)" stopOpacity="0.15" />
                </linearGradient>
                <linearGradient id="flowGradMed" x1="0" x2="1" y1="0" y2="0">
                  <stop offset="0%" stopColor="var(--risk-medium)" stopOpacity="0.45" />
                  <stop offset="100%" stopColor="var(--risk-medium)" stopOpacity="0.12" />
                </linearGradient>
              </defs>

              {/* Links */}
              {layout.links.map((l, idx) => {
                const a = layout.nodes.get(l.from);
                const b = layout.nodes.get(l.to);
                if (!a || !b) return null;
                const w = Math.max(4, l.weight * 42);
                const grad = l.weight > 0.35 ? 'url(#flowGradCrit)' : l.weight > 0.22 ? 'url(#flowGradHigh)' : 'url(#flowGradMed)';
                const isHighlighted = hoverNode && (l.from === hoverNode || l.to === hoverNode);
                return (
                  <path key={idx}
                    d={bezierPath(a.x + NODE_W, a.y + NODE_H / 2, b.x, b.y + NODE_H / 2, w)}
                    fill={grad}
                    opacity={hoverNode ? (isHighlighted ? 1 : 0.2) : 0.75}
                    style={{ transition: 'opacity 200ms', cursor: 'pointer' }}
                  />
                );
              })}

              {/* Source nodes */}
              {layout.sources.map((s) => {
                const pos = layout.nodes.get(`S${s.id}`);
                if (!pos) return null;
                const isHover = hoverNode === `S${s.id}`;
                return (
                  <g key={`S${s.id}`}
                    onMouseEnter={() => setHoverNode(`S${s.id}`)}
                    onMouseLeave={() => setHoverNode(null)}
                    style={{ cursor: 'pointer' }}>
                    <rect x={pos.x} y={pos.y} width={NODE_W} height={NODE_H} rx="6"
                      fill="var(--risk-info-bg)"
                      stroke={isHover ? "var(--risk-info)" : "var(--risk-info-br)"}
                      strokeWidth={isHover ? 1.5 : 1}
                    />
                    <text x={pos.x + 10} y={pos.y + 18} fontSize="11" fill="var(--risk-info)" fontFamily="var(--font-mono)" fontWeight="600">{s.code}</text>
                    <text x={pos.x + 10} y={pos.y + 33} fontSize="11" fill="var(--fg)">{s.name.length > 22 ? s.name.slice(0, 21) + '…' : s.name}</text>
                  </g>
                );
              })}

              {/* Threat node (single) */}
              {(() => {
                const pos = layout.nodes.get('ST');
                if (!pos) return null;
                const isHover = hoverNode === 'ST';
                return (
                  <g
                    onMouseEnter={() => setHoverNode('ST')}
                    onMouseLeave={() => setHoverNode(null)}
                    style={{ cursor: 'pointer' }}
                  >
                    <rect x={pos.x} y={pos.y} width={NODE_W} height={NODE_H} rx="6"
                      fill="var(--risk-critical-bg)"
                      stroke={isHover ? 'var(--risk-critical)' : 'var(--risk-critical-br)'}
                      strokeWidth={isHover ? 1.5 : 1}
                    />
                    <text x={pos.x + 10} y={pos.y + 18} fontSize="11" fill="var(--risk-critical)" fontFamily="var(--font-mono)" fontWeight="600">{path.threat.bdu_id || 'ST'}</text>
                    <text x={pos.x + 10} y={pos.y + 33} fontSize="11" fill="var(--fg)">{path.threat.name.length > 22 ? path.threat.name.slice(0, 21) + '…' : path.threat.name}</text>
                    <text x={pos.x + NODE_W - 10} y={pos.y + 18} fontSize="10" fill="var(--fg-dim)" fontFamily="var(--font-mono)" textAnchor="end">q^th={path.q_threat.toFixed(2)}</text>
                  </g>
                );
              })()}

              {/* Vuln nodes */}
              {layout.vulns.map(v => {
                const pos = layout.nodes.get(`V${v.vulnerability_id}`);
                if (!pos) return null;
                const isHover = hoverNode === `V${v.vulnerability_id}`;
                return (
                  <g key={`V${v.vulnerability_id}`}
                    onMouseEnter={() => setHoverNode(`V${v.vulnerability_id}`)}
                    onMouseLeave={() => setHoverNode(null)}
                    style={{ cursor: 'pointer' }}>
                    <rect x={pos.x} y={pos.y} width={NODE_W} height={NODE_H} rx="6"
                      fill={v.uncovered ? "var(--risk-critical-bg)" : "var(--risk-high-bg)"}
                      stroke={isHover ? (v.uncovered ? "var(--risk-critical)" : "var(--risk-high)") : (v.uncovered ? "var(--risk-critical-br)" : "var(--risk-high-br)")}
                      strokeWidth={isHover ? 1.5 : 1}
                    />
                    <text x={pos.x + 10} y={pos.y + 18} fontSize="11" fill={v.uncovered ? "var(--risk-critical)" : "var(--risk-high)"} fontFamily="var(--font-mono)" fontWeight="600">V{v.vulnerability_id}</text>
                    <text x={pos.x + 10} y={pos.y + 33} fontSize="11" fill="var(--fg)">{v.name.length > 22 ? v.name.slice(0, 21) + '…' : v.name}</text>
                  </g>
                );
              })}

              {/* Action nodes */}
              {layout.actions.map(a => {
                const pos = layout.nodes.get(`D${a.id}`);
                if (!pos) return null;
                const isHover = hoverNode === `D${a.id}`;
                return (
                  <g key={`D${a.id}`}
                    onMouseEnter={() => setHoverNode(`D${a.id}`)}
                    onMouseLeave={() => setHoverNode(null)}
                    style={{ cursor: 'pointer' }}>
                    <rect x={pos.x} y={pos.y} width={NODE_W} height={NODE_H} rx="6"
                      fill="var(--risk-medium-bg)"
                      stroke={isHover ? "var(--risk-medium)" : "var(--risk-medium-br)"}
                      strokeWidth={isHover ? 1.5 : 1}
                    />
                    <text x={pos.x + 10} y={pos.y + 18} fontSize="11" fill="var(--risk-medium)" fontFamily="var(--font-mono)" fontWeight="600">{a.code}</text>
                    <text x={pos.x + 10} y={pos.y + 33} fontSize="11" fill="var(--fg)">{a.name.length > 22 ? a.name.slice(0, 21) + '…' : a.name}</text>
                  </g>
                );
              })}
            </svg>
          </div>

          <div style={{ padding: '10px 16px', borderTop: '1px solid var(--border)', display: 'flex', gap: 16, alignItems: 'center' }}>
            <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)' }}>Интенсивность потока:</span>
            <div style={{ display: 'flex', gap: 12 }}>
              {[{ c: 'var(--risk-critical)', l: '> 0.35' }, { c: 'var(--risk-high)', l: '0.22 – 0.35' }, { c: 'var(--risk-medium)', l: '< 0.22' }].map(x => (
                <div key={x.l} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <div style={{ width: 18, height: 10, borderRadius: 2, background: x.c, opacity: 0.6 }} />
                  <span className="num" style={{ fontSize: 11, color: 'var(--fg-muted)' }}>{x.l}</span>
                </div>
              ))}
            </div>
          </div>
        </Card>

        {/* PTSZI breakdown panel */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <Card title="Формула ПТСЗИ" subtitle="Выбранный сценарий" dense>
            <div style={{
              padding: '14px 16px',
              background: 'var(--bg-elev-2)',
              border: '1px solid var(--border-strong)',
              borderRadius: 'var(--r-md)',
              fontFamily: 'var(--font-mono)',
              fontSize: 15, textAlign: 'center', letterSpacing: '-0.01em',
              marginBottom: 14,
            }}>
              <span style={{ color: 'var(--accent)', fontSize: 20, fontWeight: 600 }}>W</span>
              <span style={{ color: 'var(--fg-muted)', margin: '0 6px' }}>=</span>
              <span style={{ color: 'var(--fg)' }}>(Q<sup>th</sup> + q + (1−Q<sup>re</sup>))</span>
              <span style={{ color: 'var(--fg-muted)' }}> / 3 · </span>
              <span style={{ color: 'var(--risk-medium)' }}>Z</span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {[
                { label: 'Q^th · потенциал угрозы', v: path.q_threat, info: path.threat.name, color: 'var(--risk-critical)' },
                { label: 'q · опасность уязвимостей', v: path.q_severity, info: `${path.vulnerable_links.length} VL`, color: 'var(--risk-high)' },
                { label: 'Q^re · покрытие СЗИ', v: path.q_reaction, info: 'Ниже = риск выше (1−Q^re)', color: 'var(--risk-info)' },
                { label: 'Z · вес действия', v: path.z, info: 'По контуру актива', color: 'var(--risk-medium)' },
              ].map(p => (
                <div key={p.label} style={{ padding: '8px 10px', background: 'var(--bg-elev-2)', border: '1px solid var(--border)', borderRadius: 'var(--r-sm)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 4 }}>
                    <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)' }}>{p.label}</span>
                    <span className="num" style={{ fontSize: 'var(--text-sm)', color: p.color, fontWeight: 600 }}>{p.v.toFixed(2)}</span>
                  </div>
                  <div style={{ height: 3, background: 'var(--bg-elev-3)', borderRadius: 999, overflow: 'hidden' }}>
                    <div style={{ width: `${p.v * 100}%`, height: '100%', background: p.color, transition: 'width 300ms' }} />
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
              <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', marginBottom: 4 }}>Итоговый вес сценария W</div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                <span className="num" style={{ fontSize: 28, fontWeight: 600, color: 'var(--accent)', letterSpacing: '-0.02em' }}>{path.w.toFixed(3)}</span>
                <RiskBadge level={level} />
              </div>
              <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginTop: 6, fontFamily: 'var(--font-mono)' }}>
                ({path.q_threat.toFixed(2)} + {path.q_severity.toFixed(2)} + {(1 - path.q_reaction).toFixed(2)}) / 3 × {path.z.toFixed(2)}
              </div>
            </div>
          </Card>
        </div>
      </div>
    </motion.div>
  );
};
