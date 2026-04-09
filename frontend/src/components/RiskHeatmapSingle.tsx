import React, { useMemo } from "react";
import * as d3 from "d3";
import { getRiskLevelLabel } from "../utils/i18n";

interface Props {
    impact: number;      // 1..5
    likelihood: number;  // 1..5
    score: number;
    level: string;
}

export const RiskHeatmapSingle: React.FC<Props> = ({
    impact,
    likelihood,
    score,
    level,
}) => {
    const width = 400;
    const height = 400;
    const padding = { top: 40, right: 30, bottom: 60, left: 60 };

    const innerWidth = width - padding.left - padding.right;
    const innerHeight = height - padding.top - padding.bottom;

    const xScale = useMemo(
        () =>
            d3
                .scaleBand<number>()
                .domain([1, 2, 3, 4, 5])
                .range([0, innerWidth])
                .paddingInner(0.08),
        [innerWidth]
    );

    const yScale = useMemo(
        () =>
            d3
                .scaleBand<number>()
                .domain([5, 4, 3, 2, 1])
                .range([0, innerHeight])
                .paddingInner(0.08),
        [innerHeight]
    );

    const getRiskColor = (cellScore: number) => {
        if (cellScore >= 16) return "#e74c3c";
        if (cellScore >= 11) return "#e67e22";
        if (cellScore >= 6) return "#f39c12";
        return "#27ae60";
    };

    const getLevelColor = (level: string) => {
        switch (level.toLowerCase()) {
            case "critical": return "#e74c3c";
            case "high": return "#e67e22";
            case "medium": return "#f39c12";
            case "low": return "#27ae60";
            default: return "#95a5a6";
        }
    };

    const cells: { x: number; y: number; i: number; j: number }[] = [];
    [1, 2, 3, 4, 5].forEach((lik) => {
        [1, 2, 3, 4, 5].forEach((imp) => {
            const x = xScale(lik) ?? 0;
            const y = yScale(imp) ?? 0;
            cells.push({ x, y, i: imp, j: lik });
        });
    });

    const activeColor = getLevelColor(level);
    const activeCell = cells.find((c) => c.i === impact && c.j === likelihood);

    return (
        <div style={{ marginTop: "20px" }}>
            <h3
                style={{
                    fontSize: "16px",
                    marginBottom: "16px",
                    color: "var(--color-primary-dark)",
                    display: "flex",
                    alignItems: "center",
                    gap: "8px",
                }}
            >
                <svg width="18" height="18" viewBox="0 0 18 18" fill="var(--color-info)">
                    <rect x="2" y="2" width="5" height="5" />
                    <rect x="11" y="2" width="5" height="5" />
                    <rect x="2" y="11" width="5" height="5" />
                    <rect x="11" y="11" width="5" height="5" />
                </svg>
                Позиция на матрице рисков
            </h3>

            <div style={{ display: "flex", justifyContent: "center" }}>
                <svg
                    width={width}
                    height={height}
                    style={{
                        background: "linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%)",
                        borderRadius: "var(--radius-md)",
                        border: "1px solid var(--color-border)",
                        boxShadow: "var(--shadow-sm)",
                    }}
                >
                    {/* Определяем градиенты и фильтры */}
                    <defs>
                        <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
                            <feGaussianBlur stdDeviation="4" result="coloredBlur" />
                            <feMerge>
                                <feMergeNode in="coloredBlur" />
                                <feMergeNode in="SourceGraphic" />
                            </feMerge>
                        </filter>
                        <linearGradient id="activeGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                            <stop offset="0%" stopColor={activeColor} stopOpacity="1" />
                            <stop offset="100%" stopColor={activeColor} stopOpacity="0.7" />
                        </linearGradient>
                    </defs>

                    <g transform={`translate(${padding.left},${padding.top})`}>
                        {/* Фоновые клетки с градиентами */}
                        {cells.map((cell) => {
                            const cellScore = cell.i * cell.j;
                            const isActive = cell.i === impact && cell.j === likelihood;
                            const baseColor = getRiskColor(cellScore);

                            return (
                                <g key={`${cell.i}-${cell.j}`}>
                                    <rect
                                        x={cell.x}
                                        y={cell.y}
                                        width={xScale.bandwidth()}
                                        height={yScale.bandwidth()}
                                        fill={
                                            isActive
                                                ? "url(#activeGradient)"
                                                : `${baseColor}15`
                                        }
                                        stroke={isActive ? activeColor : "var(--color-border)"}
                                        strokeWidth={isActive ? 3 : 1}
                                        rx={4}
                                        style={{
                                            transition: "all 0.3s ease",
                                        }}
                                    />
                                    {/* Показываем score в каждой клетке */}
                                    {!isActive && (
                                        <text
                                            x={cell.x + xScale.bandwidth() / 2}
                                            y={cell.y + yScale.bandwidth() / 2}
                                            textAnchor="middle"
                                            dominantBaseline="middle"
                                            fontSize={11}
                                            fill={`${baseColor}80`}
                                            fontWeight={600}
                                        >
                                            {cellScore}
                                        </text>
                                    )}
                                </g>
                            );
                        })}

                        {/* Активная точка с эффектом свечения */}
                        {activeCell && (
                            <>
                                {/* Круглый пульсирующий индикатор */}
                                <circle
                                    cx={activeCell.x + xScale.bandwidth() / 2}
                                    cy={activeCell.y + yScale.bandwidth() / 2}
                                    r={18}
                                    fill={activeColor}
                                    opacity={0.2}
                                    filter="url(#glow)"
                                >
                                    <animate
                                        attributeName="r"
                                        values="18;22;18"
                                        dur="2s"
                                        repeatCount="indefinite"
                                    />
                                    <animate
                                        attributeName="opacity"
                                        values="0.2;0.4;0.2"
                                        dur="2s"
                                        repeatCount="indefinite"
                                    />
                                </circle>

                                {/* Центральная точка */}
                                <circle
                                    cx={activeCell.x + xScale.bandwidth() / 2}
                                    cy={activeCell.y + yScale.bandwidth() / 2}
                                    r={12}
                                    fill={activeColor}
                                    stroke="white"
                                    strokeWidth={3}
                                    filter="url(#glow)"
                                />

                                {/* Score в центре */}
                                <text
                                    x={activeCell.x + xScale.bandwidth() / 2}
                                    y={activeCell.y + yScale.bandwidth() / 2}
                                    textAnchor="middle"
                                    dominantBaseline="middle"
                                    fontSize={14}
                                    fill="white"
                                    fontWeight={700}
                                >
                                    {score}
                                </text>
                            </>
                        )}

                        {/* Подписи по X с улучшенным стилем */}
                        {[1, 2, 3, 4, 5].map((val) => (
                            <text
                                key={`x-${val}`}
                                x={(xScale(val) ?? 0) + xScale.bandwidth() / 2}
                                y={innerHeight + 20}
                                textAnchor="middle"
                                fontSize={13}
                                fontWeight={val === likelihood ? 700 : 500}
                                fill={
                                    val === likelihood
                                        ? activeColor
                                        : "var(--color-text)"
                                }
                            >
                                {val}
                            </text>
                        ))}

                        {/* Подписи по Y с улучшенным стилем */}
                        {[1, 2, 3, 4, 5].map((val) => (
                            <text
                                key={`y-${val}`}
                                x={-12}
                                y={(yScale(val) ?? 0) + yScale.bandwidth() / 2 + 4}
                                textAnchor="end"
                                fontSize={13}
                                fontWeight={val === impact ? 700 : 500}
                                fill={
                                    val === impact
                                        ? activeColor
                                        : "var(--color-text)"
                                }
                            >
                                {val}
                            </text>
                        ))}

                        {/* Подписи осей с иконками */}
                        <text
                            x={innerWidth / 2}
                            y={innerHeight + 45}
                            textAnchor="middle"
                            fontSize={13}
                            fontWeight={600}
                            fill="var(--color-primary-dark)"
                        >
                            Вероятность →
                        </text>

                        <text
                            x={-38}
                            y={innerHeight / 2}
                            textAnchor="middle"
                            fontSize={13}
                            fontWeight={600}
                            fill="var(--color-primary-dark)"
                            transform={`rotate(-90, -38, ${innerHeight / 2})`}
                        >
                            Влияние →
                        </text>
                    </g>
                </svg>
            </div>

            {/* Информационная панель под графиком */}
            <div
                style={{
                    marginTop: "16px",
                    padding: "16px",
                    background: "var(--color-bg)",
                    borderRadius: "var(--radius-md)",
                    border: "1px solid var(--color-border)",
                }}
            >
                <div
                    style={{
                        display: "grid",
                        gridTemplateColumns: "1fr 1fr 1fr",
                        gap: "16px",
                        textAlign: "center",
                    }}
                >
                    <div>
                        <div
                            style={{
                                fontSize: "11px",
                                color: "var(--color-text-light)",
                                marginBottom: "4px",
                                textTransform: "uppercase",
                                letterSpacing: "0.5px",
                                fontWeight: 600,
                            }}
                        >
                            Влияние
                        </div>
                        <div
                            style={{
                                fontSize: "24px",
                                fontWeight: 700,
                                color: activeColor,
                            }}
                        >
                            {impact}
                        </div>
                    </div>

                    <div>
                        <div
                            style={{
                                fontSize: "11px",
                                color: "var(--color-text-light)",
                                marginBottom: "4px",
                                textTransform: "uppercase",
                                letterSpacing: "0.5px",
                                fontWeight: 600,
                            }}
                        >
                            Вероятность
                        </div>
                        <div
                            style={{
                                fontSize: "24px",
                                fontWeight: 700,
                                color: activeColor,
                            }}
                        >
                            {likelihood}
                        </div>
                    </div>

                    <div>
                        <div
                            style={{
                                fontSize: "11px",
                                color: "var(--color-text-light)",
                                marginBottom: "4px",
                                textTransform: "uppercase",
                                letterSpacing: "0.5px",
                                fontWeight: 600,
                            }}
                        >
                            Итого
                        </div>
                        <div
                            style={{
                                fontSize: "24px",
                                fontWeight: 700,
                                color: activeColor,
                            }}
                        >
                            {score}
                            <span
                                style={{
                                    fontSize: "14px",
                                    color: "var(--color-text-light)",
                                    fontWeight: 500,
                                }}
                            >
                                /25
                            </span>
                        </div>
                    </div>
                </div>

                {/* Легенда уровней риска */}
                <div
                    style={{
                        marginTop: "16px",
                        paddingTop: "16px",
                        borderTop: "1px solid var(--color-border)",
                        display: "flex",
                        justifyContent: "space-around",
                        fontSize: "11px",
                    }}
                >
                    {[
                        { level: "low", color: "#27ae60", range: "1-5" },
                        { level: "medium", color: "#f39c12", range: "6-10" },
                        { level: "high", color: "#e67e22", range: "11-15" },
                        { level: "critical", color: "#e74c3c", range: "16-25" },
                    ].map((item) => (
                        <div
                            key={item.level}
                            style={{
                                display: "flex",
                                alignItems: "center",
                                gap: "6px",
                            }}
                        >
                            <div
                                style={{
                                    width: "12px",
                                    height: "12px",
                                    borderRadius: "3px",
                                    background: item.color,
                                    boxShadow: `0 2px 4px ${item.color}40`,
                                }}
                            />
                            <div>
                                <div
                                    style={{
                                        fontWeight: 600,
                                        color: "var(--color-text)",
                                    }}
                                >
                                    {getRiskLevelLabel(item.level)}
                                </div>
                                <div style={{ color: "var(--color-text-muted)" }}>
                                    {item.range}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};
