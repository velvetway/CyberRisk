import React, { useMemo } from "react";
import { sankey, sankeyLinkHorizontal, SankeyGraph, SankeyLink, SankeyNode } from "d3-sankey";
import { AttackPath } from "../types/riskGraph";

type NodeDatum = { name: string; kind: "S" | "ST" | "VL" | "DA"; uncovered?: boolean };
type LinkDatum = { source: number; target: number; value: number };

export const RiskGraphSankey: React.FC<{ path: AttackPath }> = ({ path }) => {
    const { nodes, links } = useMemo(() => buildGraph(path), [path]);
    const width = 960;
    const height = 360;

    const graph: SankeyGraph<NodeDatum, LinkDatum> = useMemo(() => {
        const layout = sankey<NodeDatum, LinkDatum>()
            .nodeWidth(18)
            .nodePadding(12)
            .extent([[24, 24], [width - 24, height - 24]]);
        return layout({
            nodes: nodes.map(n => ({ ...n })),
            links: links.map(l => ({ ...l })),
        });
    }, [nodes, links]);

    if (graph.nodes.length === 0) {
        return (
            <div className="card" style={{ padding: 24, textAlign: "center", color: "var(--ink-muted)" }}>
                Недостаточно данных для построения графа.
            </div>
        );
    }

    return (
        <svg width={width} height={height} style={{ background: "var(--well)", borderRadius: 12 }}>
            {(graph.links as SankeyLink<NodeDatum, LinkDatum>[]).map((l, i) => {
                const target = l.target as SankeyNode<NodeDatum, LinkDatum>;
                return (
                    <path
                        key={`l-${i}`}
                        d={sankeyLinkHorizontal()(l) ?? ""}
                        fill="none"
                        stroke={target.uncovered ? "var(--danger)" : "var(--perimeter)"}
                        strokeOpacity={0.4}
                        strokeWidth={Math.max(1, l.width ?? 1)}
                    />
                );
            })}
            {(graph.nodes as SankeyNode<NodeDatum, LinkDatum>[]).map((n, i) => {
                const nodeWidth = (n.x1 ?? 0) - (n.x0 ?? 0);
                const nodeHeight = (n.y1 ?? 0) - (n.y0 ?? 0);
                const isLeftHalf = (n.x0 ?? 0) < width / 2;
                return (
                    <g key={`n-${i}`} transform={`translate(${n.x0 ?? 0},${n.y0 ?? 0})`}>
                        <rect
                            width={nodeWidth}
                            height={nodeHeight}
                            fill={colorFor(n.kind, n.uncovered)}
                            rx={3}
                        />
                        <text
                            x={isLeftHalf ? nodeWidth + 6 : -6}
                            y={nodeHeight / 2}
                            dy="0.35em"
                            textAnchor={isLeftHalf ? "start" : "end"}
                            fontSize={12}
                            fill="var(--ink)"
                        >{n.name}</text>
                    </g>
                );
            })}
        </svg>
    );
};

function colorFor(kind: NodeDatum["kind"], uncovered?: boolean): string {
    if (kind === "VL" && uncovered) return "var(--danger)";
    switch (kind) {
        case "S":  return "var(--command)";
        case "ST": return "var(--warning)";
        case "VL": return "var(--threat-medium)";
        case "DA": return "var(--danger)";
    }
}

function buildGraph(p: AttackPath): { nodes: NodeDatum[]; links: LinkDatum[] } {
    const nodes: NodeDatum[] = [];
    const links: LinkDatum[] = [];

    const sIdx: Record<number, number> = {};
    p.sources.forEach(s => {
        sIdx[s.id] = nodes.length;
        nodes.push({ name: `${s.code}: ${s.name}`, kind: "S" });
    });

    const stIdx = nodes.length;
    nodes.push({ name: p.threat.name, kind: "ST" });

    const vlIdx: Record<number, number> = {};
    p.vulnerable_links.forEach(v => {
        vlIdx[v.vulnerability_id] = nodes.length;
        nodes.push({ name: v.name, kind: "VL", uncovered: v.uncovered });
    });

    const daIdx: Record<number, number> = {};
    p.destructive_actions.forEach(d => {
        daIdx[d.id] = nodes.length;
        nodes.push({ name: `${d.code}: ${d.name}`, kind: "DA" });
    });

    // S → ST
    p.sources.forEach(s => links.push({ source: sIdx[s.id], target: stIdx, value: 1 }));

    // ST → VL
    p.vulnerable_links.forEach(v =>
        links.push({ source: stIdx, target: vlIdx[v.vulnerability_id], value: v.severity || 1 })
    );

    // ST → DA (direct edge if there are no VLs; otherwise VL → DA fan-out)
    if (p.vulnerable_links.length === 0) {
        p.destructive_actions.forEach(d =>
            links.push({ source: stIdx, target: daIdx[d.id], value: 1 })
        );
    } else {
        p.vulnerable_links.forEach(v => {
            p.destructive_actions.forEach(d => {
                links.push({ source: vlIdx[v.vulnerability_id], target: daIdx[d.id], value: 1 });
            });
        });
    }

    return { nodes, links };
}
