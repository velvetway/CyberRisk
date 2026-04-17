import React, { useEffect, useState, useMemo } from "react";
import * as d3 from "d3";
import { api } from "../api/client";
import { RiskOverviewPoint } from "../types";
import { RiskLevelKey, getRiskLevelLabel, normalizeRiskLevel } from "../utils/i18n";
import { motion } from "framer-motion";
import { ShieldAlert, AlertTriangle, ArrowUpRight, ArrowDownRight, Layers } from "lucide-react";

export const RiskMapPage: React.FC = () => {
    const [points, setPoints] = useState<RiskOverviewPoint[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [hover, setHover] = useState<RiskOverviewPoint | null>(null);

    useEffect(() => {
        api.getRiskOverview()
            .then((data) => { setPoints(data); setError(null); })
            .catch((e) => setError(e.message || "Не удалось загрузить карту рисков"))
            .finally(() => setLoading(false));
    }, []);

    const width = 600;
    const height = 600;
    const padding = { top: 40, right: 40, bottom: 60, left: 60 };
    const innerWidth = width - padding.left - padding.right;
    const innerHeight = height - padding.top - padding.bottom;

    const xScale = useMemo(() => d3.scaleBand<number>().domain([1, 2, 3, 4, 5]).range([0, innerWidth]).paddingInner(0.05), [innerWidth]);
    const yScale = useMemo(() => d3.scaleBand<number>().domain([5, 4, 3, 2, 1]).range([0, innerHeight]).paddingInner(0.05), [innerHeight]);

    const levelColors: Record<RiskLevelKey, string> = {
        low: "var(--success)",
        medium: "var(--warning)",
        high: "var(--threat-high)",
        critical: "var(--danger)",
    };

    const levelDimColors: Record<RiskLevelKey, string> = {
        low: "var(--threat-low-dim)",
        medium: "var(--threat-medium-dim)",
        high: "var(--threat-high-dim)",
        critical: "var(--threat-critical-dim)",
    };

    const cells: { x: number; y: number; i: number; j: number }[] = [];
    [1, 2, 3, 4, 5].forEach((lik) => {
        [1, 2, 3, 4, 5].forEach((imp) => {
            cells.push({ x: xScale(lik) ?? 0, y: yScale(imp) ?? 0, i: imp, j: lik });
        });
    });

    const stats = useMemo(() => {
        const total = points.length;
        const byLevel: Record<RiskLevelKey, number> = { low: 0, medium: 0, high: 0, critical: 0 };
        points.forEach((p) => {
            const key = normalizeRiskLevel(p.level);
            if (key) byLevel[key] += 1;
        });
        return { total, byLevel };
    }, [points]);

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }}>
            <div style={{ marginBottom: "32px" }}>
                <h1>Карта рисков</h1>
                <p style={{ color: "var(--ink-muted)" }}>Матрица распределения сценариев риска (Влияние × Вероятность)</p>
            </div>

            {loading && (
                <div style={{ textAlign: "center", padding: "60px 20px" }}>
                    <div className="loading-spinner" />
                    <p style={{ marginTop: "16px", color: "var(--ink-muted)" }}>Загрузка...</p>
                </div>
            )}

            {error && (
                <div className="card" style={{ padding: "20px", background: "var(--threat-critical-dim)", border: "1px solid var(--danger)" }}>
                    <p style={{ color: "var(--danger)", margin: 0, fontWeight: 500 }}>⚠️ {error}</p>
                </div>
            )}

            {!loading && !error && (
                <>
                    <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: "16px", marginBottom: "24px" }}>
                        <StatCard label="Всего сценариев" value={stats.total} color="var(--command)" icon={<Layers size={24} />} />
                        <StatCard label={getRiskLevelLabel("critical")} value={stats.byLevel.critical} color="var(--danger)" icon={<ShieldAlert size={24} />} />
                        <StatCard label={getRiskLevelLabel("high")} value={stats.byLevel.high} color="var(--threat-high)" icon={<AlertTriangle size={24} />} />
                        <StatCard label={getRiskLevelLabel("medium")} value={stats.byLevel.medium} color="var(--warning)" icon={<ArrowUpRight size={24} />} />
                        <StatCard label={getRiskLevelLabel("low")} value={stats.byLevel.low} color="var(--success)" icon={<ArrowDownRight size={24} />} />
                    </div>

                    <div style={{ display: "grid", gap: "24px", gridTemplateColumns: "auto 1fr", alignItems: "start" }}>
                        <div className="card" style={{ padding: "24px", background: "var(--raised)" }}>
                            <svg width={width} height={height} style={{ background: "var(--ground)", borderRadius: "var(--r-md)", border: "1px solid var(--perimeter)" }}>
                                <g transform={`translate(${padding.left},${padding.top})`}>
                                    {cells.map((cell) => {
                                        const score = cell.i * cell.j;
                                        let fill = "var(--well)";
                                        if (score >= 16) fill = "var(--threat-critical-dim)";
                                        else if (score >= 11) fill = "var(--threat-high-dim)";
                                        else if (score >= 6) fill = "var(--threat-medium-dim)";
                                        else fill = "var(--threat-low-dim)";

                                        return (
                                            <rect
                                                key={`${cell.i}-${cell.j}`}
                                                x={cell.x} y={cell.y}
                                                width={xScale.bandwidth()} height={yScale.bandwidth()}
                                                fill={fill} stroke="var(--perimeter)" strokeWidth={1} rx={4}
                                            />
                                        );
                                    })}

                                    {points.map((p, idx) => {
                                        const cx = (xScale(p.likelihood) ?? 0) + xScale.bandwidth() / 2;
                                        const cy = (yScale(p.impact) ?? 0) + yScale.bandwidth() / 2;
                                        const levelKey = normalizeRiskLevel(p.level);
                                        const color = levelKey ? levelColors[levelKey] : "var(--ink-muted)";

                                        return (
                                            <circle
                                                key={idx} cx={cx} cy={cy}
                                                r={hover === p ? 8 : 5}
                                                fill={color} fillOpacity={hover === p ? 1 : 0.8}
                                                stroke="white" strokeWidth={hover === p ? 2 : 1}
                                                onMouseEnter={() => setHover(p)} onMouseLeave={() => setHover(null)}
                                                style={{ cursor: "pointer", transition: "all 0.2s" }}
                                            />
                                        );
                                    })}

                                    {[1, 2, 3, 4, 5].map(val => (
                                        <text key={`x-${val}`} x={(xScale(val) ?? 0) + xScale.bandwidth() / 2} y={innerHeight + 24} textAnchor="middle" fontSize={13} fontWeight={600} fill="var(--ink-muted)">{val}</text>
                                    ))}
                                    {[1, 2, 3, 4, 5].map(val => (
                                        <text key={`y-${val}`} x={-16} y={(yScale(val) ?? 0) + yScale.bandwidth() / 2 + 4} textAnchor="end" fontSize={13} fontWeight={600} fill="var(--ink-muted)">{val}</text>
                                    ))}

                                    <text x={innerWidth / 2} y={innerHeight + 48} textAnchor="middle" fontSize={14} fontWeight={600} fill="var(--ink)">Вероятность (L)</text>
                                    <text x={-48} y={innerHeight / 2} textAnchor="middle" fontSize={14} fontWeight={600} fill="var(--ink)" transform={`rotate(-90, -48, ${innerHeight / 2})`}>Влияние (I)</text>
                                </g>
                            </svg>
                        </div>

                        <div style={{ display: "flex", flexDirection: "column", gap: "24px" }}>
                            <div className="card" style={{ padding: "24px", minHeight: "240px" }}>
                                <h3 style={{ marginBottom: "16px" }}>Анализ сценария</h3>
                                {!hover ? (
                                    <p style={{ color: "var(--ink-muted)" }}>Наведите курсор на точку в матрице для просмотра деталей.</p>
                                ) : (
                                    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
                                        <div style={{ marginBottom: "16px", paddingBottom: "16px", borderBottom: "1px solid var(--perimeter)" }}>
                                            <div style={{ fontWeight: 600, fontSize: "16px" }}>{hover.asset_name}</div>
                                            <div style={{ fontSize: "12px", color: "var(--ink-muted)" }}>Актив #{hover.asset_id}</div>
                                        </div>
                                        <div style={{ marginBottom: "20px" }}>
                                            <div style={{ fontSize: "14px", fontWeight: 500, marginBottom: "4px" }}>{hover.threat_name}</div>
                                            <div style={{ fontSize: "12px", color: "var(--ink-muted)" }}>Угроза #{hover.threat_id}</div>
                                        </div>
                                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px", marginBottom: "16px" }}>
                                            <div style={{ background: "var(--well)", padding: "12px", borderRadius: "var(--r-sm)" }}>
                                                <div style={{ fontSize: "11px", color: "var(--ink-muted)", fontWeight: 600 }}>ВЛИЯНИЕ</div>
                                                <div style={{ fontSize: "24px", fontWeight: 700 }}>{hover.impact}</div>
                                            </div>
                                            <div style={{ background: "var(--well)", padding: "12px", borderRadius: "var(--r-sm)" }}>
                                                <div style={{ fontSize: "11px", color: "var(--ink-muted)", fontWeight: 600 }}>ВЕРОЯТНОСТЬ</div>
                                                <div style={{ fontSize: "24px", fontWeight: 700 }}>{hover.likelihood}</div>
                                            </div>
                                        </div>
                                        <div style={{ background: normalizeRiskLevel(hover.level) ? levelDimColors[normalizeRiskLevel(hover.level)!] : 'var(--well)', border: `1px solid ${normalizeRiskLevel(hover.level) ? levelColors[normalizeRiskLevel(hover.level)!] : 'var(--perimeter)'}`, padding: "16px", borderRadius: "var(--r-md)", textAlign: "center" }}>
                                            <div style={{ fontSize: "12px", fontWeight: 600, color: normalizeRiskLevel(hover.level) ? levelColors[normalizeRiskLevel(hover.level)!] : 'var(--ink)' }}>РИСК: {hover.score}/25</div>
                                            <div style={{ fontSize: "20px", fontWeight: 700, color: normalizeRiskLevel(hover.level) ? levelColors[normalizeRiskLevel(hover.level)!] : 'var(--ink)' }}>{getRiskLevelLabel(hover.level)}</div>
                                        </div>
                                    </motion.div>
                                )}
                            </div>
                        </div>
                    </div>
                </>
            )}
        </motion.div>
    );
};

const StatCard: React.FC<{ label: string; value: number; color: string; icon: React.ReactNode }> = ({ label, value, color, icon }) => (
    <div className="card" style={{ padding: "20px", display: "flex", alignItems: "center", gap: "16px" }}>
        <div style={{ color }}>{icon}</div>
        <div>
            <div style={{ fontSize: "24px", fontWeight: 700, color: "var(--ink)" }}>{value}</div>
            <div style={{ fontSize: "12px", color: "var(--ink-muted)", fontWeight: 600 }}>{label}</div>
        </div>
    </div>
);
