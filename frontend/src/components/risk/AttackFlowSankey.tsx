import React, { useMemo, useState } from "react";
import { sankey, sankeyLinkHorizontal, SankeyNode, SankeyLink } from "d3-sankey";
import { buildSankeyGraph, SankeyNodeData, SankeyLinkData, SankeyNodeKind } from "../../lib/riskFlow";
import type { AttackPath } from "../../types/riskGraph";

const WIDTH = 960;
const HEIGHT = 460;
const NODE_WIDTH = 18;
const NODE_PAD = 12;

type LayoutNode = SankeyNode<SankeyNodeData, SankeyLinkData>;
type LayoutLink = SankeyLink<SankeyNodeData, SankeyLinkData>;

const colorFor = (kind: SankeyNodeKind, dimmed: boolean): string => {
  if (dimmed) return 'var(--bg-elev-3)';
  switch (kind) {
    case 'S': return 'var(--risk-info)';
    case 'ST': return 'var(--risk-critical)';
    case 'VL': return 'var(--risk-high)';
    case 'C': return 'var(--risk-low, #58a463)';
    case 'DA': return 'var(--risk-critical)';
  }
};

const linkColor = (l: LayoutLink): string => {
  const k = (l as unknown as SankeyLinkData).kind;
  switch (k) {
    case 'S->ST': return 'var(--risk-info)';
    case 'ST->VL': return 'var(--risk-high)';
    case 'VL->C': return 'var(--risk-low, #58a463)';
    case 'VL->DA': return 'var(--risk-critical)';
  }
};

export interface AttackFlowSankeyProps {
  path: AttackPath;
  disabledControlIds: Set<number>;
  onControlClick: (controlId: number, anchor: { x: number; y: number }) => void;
}

export const AttackFlowSankey: React.FC<AttackFlowSankeyProps> = ({
  path, disabledControlIds, onControlClick,
}) => {
  const [hoverNode, setHoverNode] = useState<string | null>(null);

  // d3-sankey can't lay out a graph with no edges from VL — short-circuit
  // when the path has zero vulnerable links. The useMemo below still runs
  // unconditionally (rules of hooks), but we skip it when VL list is empty.
  const noVL = path.vulnerable_links.length === 0;

  const layout = useMemo(() => {
    if (noVL) return null;
    const graph = buildSankeyGraph(path, disabledControlIds);

    const sk = sankey<SankeyNodeData, SankeyLinkData>()
      .nodeWidth(NODE_WIDTH)
      .nodePadding(NODE_PAD)
      .extent([[12, 28], [WIDTH - 12, HEIGHT - 12]])
      .nodeId(d => d.id);

    return sk({
      nodes: graph.nodes.map(n => ({ ...n })),
      links: graph.links.map(l => ({ ...l })),
    });
  }, [path, disabledControlIds, noVL]);

  if (noVL) {
    return (
      <div style={{
        padding: '24px',
        textAlign: 'center',
        color: 'var(--fg-muted)',
        background: 'var(--bg-elev-1)',
        fontSize: 'var(--text-sm)',
      }}>
        У этого пути нет уязвимых звеньев — граф недоступен.
      </div>
    );
  }

  const isHighlighted = (l: LayoutLink): boolean => {
    if (!hoverNode) return false;
    const s = typeof l.source === 'object' ? (l.source as LayoutNode).id : l.source;
    const t = typeof l.target === 'object' ? (l.target as LayoutNode).id : l.target;
    return s === hoverNode || t === hoverNode;
  };

  return (
    <div style={{ position: 'relative' }}>
      <div style={{
        padding: '8px 12px',
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-muted)',
        background: 'var(--bg-elev-2)',
        borderBottom: '1px solid var(--border)',
      }}>
        Толщина <code>VL→C</code> — coverage контроля. <code>q_reaction</code> в формуле — доля VL, прикрытых хотя бы одним контролем.
      </div>

      <svg
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        style={{ width: '100%', height: 'auto', display: 'block', background: 'var(--bg-elev-1)' }}
      >
        {/* Column headers — approximate positions matching d3-sankey's even spread */}
        {([
          { label: 'Источник',         x: 24,  anchor: 'start'  as const },
          { label: 'Угроза (БДУ)',     x: 288, anchor: 'middle' as const },
          { label: 'Уязвимость',       x: 480, anchor: 'middle' as const },
          { label: 'СЗИ (контроль)',   x: 672, anchor: 'middle' as const },
          { label: 'Действие',         x: 936, anchor: 'end'    as const },
        ]).map(h => (
          <text
            key={h.label}
            x={h.x}
            y={18}
            fontSize={10}
            fill="var(--fg-faint)"
            textAnchor={h.anchor}
            fontFamily="var(--font-mono)"
            style={{ textTransform: 'uppercase', letterSpacing: '0.08em' }}
          >
            {h.label}
          </text>
        ))}

        {/* Links */}
        {(layout!.links as LayoutLink[]).map((l, i) => {
          const linkData = l as unknown as SankeyLinkData;
          const pathD = sankeyLinkHorizontal()(l) ?? '';
          const highlighted = isHighlighted(l);
          const op = hoverNode ? (highlighted ? 0.85 : 0.15) : 0.5;
          return (
            <path
              key={`l-${i}`}
              d={pathD}
              fill="none"
              stroke={linkColor(l)}
              strokeOpacity={op}
              strokeWidth={Math.max(1, l.width ?? 1)}
              style={{ transition: 'stroke-opacity 200ms' }}
            >
              <title>{`${linkData.kind}: ${(l.value ?? 0).toFixed(3)} (${((l.value ?? 0) * 100).toFixed(1)}%)`}</title>
            </path>
          );
        })}

        {/* Nodes */}
        {(layout!.nodes as LayoutNode[]).map(n => {
          const x0 = n.x0 ?? 0, x1 = n.x1 ?? 0, y0 = n.y0 ?? 0, y1 = n.y1 ?? 0;
          const w = x1 - x0, h = Math.max(2, y1 - y0);
          const dimmed = !!(n.meta?.disabled);
          const fill = colorFor(n.kind, dimmed);
          const labelLeft = x0 < WIDTH / 2;
          const isHover = hoverNode === n.id;
          const clickable = n.kind === 'C';

          return (
            <g
              key={n.id}
              tabIndex={clickable ? 0 : -1}
              role={clickable ? 'button' : undefined}
              aria-label={clickable ? `Контроль ${n.label}` : undefined}
              onMouseEnter={() => setHoverNode(n.id)}
              onMouseLeave={() => setHoverNode(null)}
              onClick={(e) => {
                if (clickable) {
                  const rect = (e.currentTarget.ownerSVGElement as SVGSVGElement).getBoundingClientRect();
                  const id = parseInt(n.id.replace('C', ''), 10);
                  onControlClick(id, { x: rect.left + (x0 + x1) / 2, y: rect.top + y0 });
                }
              }}
              onKeyDown={(e) => {
                if (clickable && (e.key === 'Enter' || e.key === ' ')) {
                  e.preventDefault();
                  const rect = (e.currentTarget.ownerSVGElement as SVGSVGElement).getBoundingClientRect();
                  const id = parseInt(n.id.replace('C', ''), 10);
                  onControlClick(id, { x: rect.left + (x0 + x1) / 2, y: rect.top + y0 });
                }
              }}
              style={{ cursor: clickable ? 'pointer' : 'default' }}
            >
              <rect
                x={x0} y={y0} width={w} height={h}
                fill={fill}
                opacity={dimmed ? 0.4 : 1}
                stroke={isHover ? 'var(--fg)' : 'transparent'}
                strokeWidth={1}
                rx={3}
              />
              <text
                x={labelLeft ? x1 + 6 : x0 - 6}
                y={(y0 + y1) / 2}
                dy="0.35em"
                textAnchor={labelLeft ? 'start' : 'end'}
                fontSize={11}
                fill={dimmed ? 'var(--fg-faint)' : 'var(--fg)'}
                style={{ pointerEvents: 'none' }}
              >
                {n.label.length > 28 ? n.label.slice(0, 27) + '…' : n.label}
              </text>
              {dimmed && (
                <text
                  x={labelLeft ? x1 + 6 : x0 - 6}
                  y={(y0 + y1) / 2 + 12}
                  textAnchor={labelLeft ? 'start' : 'end'}
                  fontSize={9}
                  fill="var(--fg-faint)"
                  style={{ pointerEvents: 'none' }}
                >
                  выкл в симуляции
                </text>
              )}
              <title>{`${n.kind}: ${n.label}`}</title>
            </g>
          );
        })}
      </svg>
    </div>
  );
};
