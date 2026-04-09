import React, { useEffect, useState, useMemo } from "react";
import * as d3 from "d3";
import { api } from "../api/client";
import { RiskOverviewPoint } from "../types";
import { RiskLevelKey, getRiskLevelLabel, normalizeRiskLevel } from "../utils/i18n";

export const RiskMapPage: React.FC = () => {
    const [points, setPoints] = useState<RiskOverviewPoint[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [hover, setHover] = useState<RiskOverviewPoint | null>(null);

    useEffect(() => {
        setLoading(true);
        api
            .getRiskOverview()
            .then((data) => {
                setPoints(data);
                setError(null);
            })
            .catch((e) => setError(e.message || "Не удалось загрузить карту рисков"))
            .finally(() => setLoading(false));
    }, []);

    const width = 600;
    const height = 600;
    const padding = { top: 50, right: 50, bottom: 70, left: 70 };

    const innerWidth = width - padding.left - padding.right;
    const innerHeight = height - padding.top - padding.bottom;

    const xScale = useMemo(
        () =>
            d3
                .scaleBand<number>()
                .domain([1, 2, 3, 4, 5])
                .range([0, innerWidth])
                .paddingInner(0.05),
        [innerWidth]
    );

    const yScale = useMemo(
        () =>
            d3
                .scaleBand<number>()
                .domain([5, 4, 3, 2, 1])
                .range([0, innerHeight])
                .paddingInner(0.05),
        [innerHeight]
    );

    const levelColors: Record<RiskLevelKey, string> = {
        low: "#27ae60",
        medium: "#f39c12",
        high: "#e67e22",
        critical: "#e74c3c",
    };

    const cells: { x: number; y: number; i: number; j: number }[] = [];
    Array.from({ length: 5 }, (_, i) => i + 1).forEach((lik) => {
        Array.from({ length: 5 }, (_, i) => i + 1).forEach((imp) => {
            const x = xScale(lik) ?? 0;
            const y = yScale(imp) ?? 0;
            cells.push({ x, y, i: imp, j: lik });
        });
    });

    // Статистика
    const stats = useMemo(() => {
        const total = points.length;
        const byLevel: Record<RiskLevelKey, number> = {
            low: 0,
            medium: 0,
            high: 0,
            critical: 0,
        };

        points.forEach((p) => {
            const key = normalizeRiskLevel(p.level);
            if (key) {
                byLevel[key] += 1;
            }
        });

        return { total, byLevel };
    }, [points]);

    return (
        <div className="fade-in">
            <div style={{ marginBottom: "32px" }}>
                <h1 style={{ marginBottom: "8px" }}>Карта рисков</h1>
                <p style={{ color: "var(--color-text-light)", fontSize: "15px" }}>
                    Визуализация всех сценариев риска на матрице влияния и вероятности
                </p>
            </div>

            {loading && (
                <div style={{ textAlign: "center", padding: "60px 20px" }}>
                    <div
                        className="loading-spinner"
                        style={{ width: "40px", height: "40px", margin: "0 auto" }}
                    />
                    <p style={{ marginTop: "16px", color: "var(--color-text-light)" }}>
                        Загрузка карты рисков...
                    </p>
                </div>
            )}

            {error && (
                <div
                    className="card"
                    style={{
                        padding: "20px",
                        background: "rgba(231, 76, 60, 0.05)",
                        borderColor: "var(--color-risk-critical)",
                    }}
                >
                    <p style={{ color: "var(--color-risk-critical)", margin: 0 }}>
                        ⚠️ Ошибка: {error}
                    </p>
                </div>
            )}

            {!loading && !error && (
                <>
                    {/* Статистика */}
                    <div
                        style={{
                            display: "grid",
                            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
                            gap: "16px",
                            marginBottom: "24px",
                        }}
                    >
                        <StatCard
                            label="Всего сценариев"
                            value={stats.total}
                            color="var(--color-info)"
                            icon={
                                <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                                    <rect x="3" y="3" width="8" height="8" />
                                    <rect x="13" y="3" width="8" height="8" />
                                    <rect x="3" y="13" width="8" height="8" />
                                    <rect x="13" y="13" width="8" height="8" />
                                </svg>
                            }
                        />
                        <StatCard
                            label={getRiskLevelLabel("critical")}
                            value={stats.byLevel.critical}
                            color="var(--color-risk-critical)"
                            icon="⚠️"
                        />
                        <StatCard
                            label={getRiskLevelLabel("high")}
                            value={stats.byLevel.high}
                            color="var(--color-risk-high)"
                            icon="⬆️"
                        />
                        <StatCard
                            label={getRiskLevelLabel("medium")}
                            value={stats.byLevel.medium}
                            color="var(--color-risk-medium)"
                            icon="➡️"
                        />
                        <StatCard
                            label={getRiskLevelLabel("low")}
                            value={stats.byLevel.low}
                            color="var(--color-risk-low)"
                            icon="⬇️"
                        />
                    </div>

                    <div
                        style={{
                            display: "grid",
                            gap: "24px",
                            gridTemplateColumns: "auto 1fr",
                            alignItems: "start",
                        }}
                    >
                        {/* Матрица рисков */}
                        <div className="card" style={{ padding: "24px" }}>
                            <h2
                                style={{
                                    fontSize: "18px",
                                    marginBottom: "20px",
                                    display: "flex",
                                    alignItems: "center",
                                    gap: "8px",
                                }}
                            >
                                <svg
                                    width="20"
                                    height="20"
                                    viewBox="0 0 20 20"
                                    fill="var(--color-accent)"
                                >
                                    <rect x="2" y="2" width="6" height="6" />
                                    <rect x="12" y="2" width="6" height="6" />
                                    <rect x="2" y="12" width="6" height="6" />
                                    <rect x="12" y="12" width="6" height="6" />
                                </svg>
                                Матрица влияния × вероятность
                            </h2>

                            <svg
                                width={width}
                                height={height}
                                style={{
                                    border: "1px solid var(--color-border)",
                                    background: "var(--color-surface)",
                                    borderRadius: "var(--radius-md)",
                                }}
                            >
                                <g transform={`translate(${padding.left},${padding.top})`}>
                                    {/* Фоновые клетки */}
                                    {cells.map((cell) => {
                                        const score = cell.i * cell.j;
                                        let fillColor = "#f0f0f0";
                                        if (score >= 16) fillColor = "rgba(231, 76, 60, 0.08)";
                                        else if (score >= 11) fillColor = "rgba(230, 126, 34, 0.08)";
                                        else if (score >= 6) fillColor = "rgba(243, 156, 18, 0.08)";
                                        else fillColor = "rgba(39, 174, 96, 0.08)";

                                        return (
                                            <rect
                                                key={`${cell.i}-${cell.j}`}
                                                x={cell.x}
                                                y={cell.y}
                                                width={xScale.bandwidth()}
                                                height={yScale.bandwidth()}
                                                fill={fillColor}
                                                stroke="var(--color-border)"
                                                strokeWidth={0.5}
                                            />
                                        );
                                    })}

                                    {/* Точки сценариев */}
                                    {points.map((p, idx) => {
                                        const cx = (xScale(p.likelihood) ?? 0) + xScale.bandwidth() / 2;
                                        const cy = (yScale(p.impact) ?? 0) + yScale.bandwidth() / 2;
                                        const levelKey = normalizeRiskLevel(p.level);
                                        const levelColor = levelKey ? levelColors[levelKey] : "#95a5a6";

                                        return (
                                            <circle
                                                key={idx}
                                                cx={cx}
                                                cy={cy}
                                                r={hover === p ? 8 : 6}
                                                fill={levelColor}
                                                fillOpacity={hover === p ? 1 : 0.85}
                                                stroke="white"
                                                strokeWidth={hover === p ? 2.5 : 1.5}
                                                onMouseEnter={() => setHover(p)}
                                                onMouseLeave={() => setHover(null)}
                                                style={{
                                                    cursor: "pointer",
                                                    transition: "all 0.2s ease",
                                                }}
                                            />
                                        );
                                    })}

                                    {/* Подписи по X */}
                                    {[1, 2, 3, 4, 5].map((val) => (
                                        <text
                                            key={`x-${val}`}
                                            x={(xScale(val) ?? 0) + xScale.bandwidth() / 2}
                                            y={innerHeight + 20}
                                            textAnchor="middle"
                                            fontSize={13}
                                            fontWeight={600}
                                            fill="var(--color-text)"
                                        >
                                            {val}
                                        </text>
                                    ))}

                                    {/* Подписи по Y */}
                                    {[1, 2, 3, 4, 5].map((val) => (
                                        <text
                                            key={`y-${val}`}
                                            x={-12}
                                            y={(yScale(val) ?? 0) + yScale.bandwidth() / 2 + 4}
                                            textAnchor="end"
                                            fontSize={13}
                                            fontWeight={600}
                                            fill="var(--color-text)"
                                        >
                                            {val}
                                        </text>
                                    ))}

                                    {/* Подписи осей */}
                                    <text
                                        x={innerWidth / 2}
                                        y={innerHeight + 45}
                                        textAnchor="middle"
                                        fontSize={14}
                                        fontWeight={600}
                                        fill="var(--color-primary-dark)"
                                    >
                                        Вероятность →
                                    </text>

                                    <text
                                        x={-42}
                                        y={innerHeight / 2}
                                        textAnchor="middle"
                                        fontSize={14}
                                        fontWeight={600}
                                        fill="var(--color-primary-dark)"
                                        transform={`rotate(-90, -42, ${innerHeight / 2})`}
                                    >
                                        Влияние →
                                    </text>
                                </g>
                            </svg>
                        </div>

                        {/* Детали и легенда */}
                        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
                            <div className="card" style={{ padding: "20px" }}>
                                <h3 style={{ fontSize: "16px", marginBottom: "16px" }}>
                                    Детали сценария
                                </h3>
                                {!hover && (
                                    <p
                                        style={{
                                            fontSize: "13px",
                                            color: "var(--color-text-muted)",
                                            margin: 0,
                                        }}
                                    >
                                        Наведите курсор на точку, чтобы увидеть подробности.
                                    </p>
                                )}
                                {hover && (() => {
                                    const hoverLevelKey = normalizeRiskLevel(hover.level);
                                    const hoverColor = hoverLevelKey ? levelColors[hoverLevelKey] : "#95a5a6";

                                    return (
                                        <div style={{ fontSize: "13px", color: "var(--color-text)" }}>
                                            <div
                                                style={{
                                                    marginBottom: "12px",
                                                    paddingBottom: "12px",
                                                    borderBottom: "1px solid var(--color-border)",
                                                }}
                                            >
                                                <div
                                                    style={{
                                                        fontWeight: 700,
                                                        fontSize: "15px",
                                                        marginBottom: "4px",
                                                        color: "var(--color-primary-dark)",
                                                    }}
                                                >
                                                    {hover.asset_name}
                                                </div>
                                                <div
                                                    style={{
                                                        fontSize: "11px",
                                                        color: "var(--color-text-muted)",
                                                    }}
                                                >
                                                    Актив ID: {hover.asset_id}
                                                </div>
                                            </div>

                                            <div style={{ marginBottom: "12px" }}>
                                                <div
                                                    style={{
                                                        fontWeight: 600,
                                                        color: "var(--color-text-light)",
                                                        marginBottom: "2px",
                                                    }}
                                                >
                                                    Угроза:
                                                </div>
                                                <div>{hover.threat_name}</div>
                                                <div
                                                    style={{
                                                        fontSize: "11px",
                                                        color: "var(--color-text-muted)",
                                                    }}
                                                >
                                                    Угроза ID: {hover.threat_id}
                                                </div>
                                            </div>

                                            <div
                                                style={{
                                                    display: "grid",
                                                    gridTemplateColumns: "1fr 1fr",
                                                    gap: "8px",
                                                    marginBottom: "12px",
                                                }}
                                            >
                                                <div>
                                                    <div
                                                        style={{
                                                            fontSize: "11px",
                                                            color: "var(--color-text-muted)",
                                                        }}
                                                    >
                                                        Влияние
                                                    </div>
                                                    <div
                                                        style={{
                                                            fontSize: "20px",
                                                            fontWeight: 700,
                                                            color: "var(--color-primary-dark)",
                                                        }}
                                                    >
                                                        {hover.impact}
                                                    </div>
                                                </div>
                                                <div>
                                                    <div
                                                        style={{
                                                            fontSize: "11px",
                                                            color: "var(--color-text-muted)",
                                                        }}
                                                    >
                                                        Вероятность
                                                    </div>
                                                    <div
                                                        style={{
                                                            fontSize: "20px",
                                                            fontWeight: 700,
                                                            color: "var(--color-primary-dark)",
                                                        }}
                                                    >
                                                        {hover.likelihood}
                                                    </div>
                                                </div>
                                            </div>

                                            <div
                                                style={{
                                                    padding: "12px",
                                                    background: `linear-gradient(135deg, ${hoverColor}15 0%, ${hoverColor}05 100%)`,
                                                    borderRadius: "var(--radius-sm)",
                                                    border: `2px solid ${hoverColor}`,
                                                }}
                                            >
                                                <div
                                                    style={{
                                                        fontSize: "11px",
                                                        color: "var(--color-text-light)",
                                                        marginBottom: "4px",
                                                    }}
                                                >
                                                    Итоговая оценка
                                                </div>
                                                <div
                                                    style={{
                                                        fontSize: "24px",
                                                        fontWeight: 700,
                                                        color: hoverColor,
                                                    }}
                                                >
                                                    {getRiskLevelLabel(hover.level)}
                                                </div>
                                                <div
                                                    style={{
                                                        fontSize: "13px",
                                                        color: "var(--color-text-light)",
                                                    }}
                                                >
                                                    Риск: {hover.score}/25
                                                </div>
                                            </div>
                                        </div>
                                    );
                                })()}
                            </div>

                            <div className="card" style={{ padding: "20px" }}>
                                <h3 style={{ fontSize: "16px", marginBottom: "16px" }}>Легенда</h3>
                                <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
                                    {Object.entries(levelColors).map(([level, color]) => (
                                        <div
                                            key={level}
                                            style={{
                                                display: "flex",
                                                alignItems: "center",
                                                gap: "10px",
                                            }}
                                        >
                                            <div
                                                style={{
                                                    width: "16px",
                                                    height: "16px",
                                                    borderRadius: "50%",
                                                    background: color,
                                                    border: "2px solid white",
                                                    boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                                                }}
                                            />
                                            <span
                                                style={{
                                                    fontSize: "13px",
                                                    color: "var(--color-text)",
                                                    fontWeight: 500,
                                                }}
                                            >
                                                {getRiskLevelLabel(level)}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </div>
    );
};

const StatCard: React.FC<{
    label: string;
    value: number;
    color: string;
    icon: string | React.ReactNode;
}> = ({ label, value, color, icon }) => {
    return (
        <div className="card" style={{ padding: "20px", textAlign: "center" }}>
            <div
                style={{
                    fontSize: "32px",
                    marginBottom: "8px",
                    color: color,
                }}
            >
                {typeof icon === "string" ? icon : icon}
            </div>
            <div
                style={{
                    fontSize: "28px",
                    fontWeight: 700,
                    color: "var(--color-primary-dark)",
                    marginBottom: "4px",
                }}
            >
                {value}
            </div>
            <div
                style={{
                    fontSize: "12px",
                    color: "var(--color-text-light)",
                    fontWeight: 600,
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                }}
            >
                {label}
            </div>
        </div>
    );
};
