import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getPriorityLabel, getRiskLevelLabel, getThreatTypeLabel } from "../utils/i18n";

interface AssetRisk {
    threat_id: number;
    threat_name: string;
    threat_description: string;
    threat_type: string;
    impact: number;
    likelihood: number;
    score: number;
    level: string;
    recommendations: Recommendation[];
}

interface Recommendation {
    code: string;
    title: string;
    description: string;
    priority: string;
    category: string;
}

export const AssetRiskProfilePage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [risks, setRisks] = useState<AssetRisk[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!id) return;

        const fetchRisks = async () => {
            setLoading(true);
            try {
                const res = await fetch(`/api/risk/asset/${id}`);
                if (!res.ok) {
                    throw new Error(`Ошибка загрузки рисков: ${res.status}`);
                }
                const data: AssetRisk[] = await res.json();
                // Сортируем по убыванию риска (score)
                data.sort((a, b) => b.score - a.score);
                setRisks(data);
            } catch (e: any) {
                setError(e.message || "Ошибка загрузки данных");
            } finally {
                setLoading(false);
            }
        };

        fetchRisks();
    }, [id]);

    const getRiskStats = () => {
        const critical = risks.filter((r) => r.level.toLowerCase() === "critical").length;
        const high = risks.filter((r) => r.level.toLowerCase() === "high").length;
        const medium = risks.filter((r) => r.level.toLowerCase() === "medium").length;
        const low = risks.filter((r) => r.level.toLowerCase() === "low").length;
        return { critical, high, medium, low, total: risks.length };
    };

    const stats = getRiskStats();

    if (loading) {
        return (
            <div className="fade-in" style={{ textAlign: "center", padding: "60px 20px" }}>
                <div className="loading-spinner" style={{ width: "40px", height: "40px", margin: "0 auto" }} />
                <p style={{ marginTop: "16px", color: "var(--color-text-light)" }}>
                    Анализ рисков для актива...
                </p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="fade-in">
                <div
                    className="card"
                    style={{
                        padding: "20px",
                        background: "rgba(231, 76, 60, 0.05)",
                        borderColor: "var(--color-risk-critical)",
                    }}
                >
                    <p style={{ color: "var(--color-risk-critical)", margin: 0 }}>
                        ⚠️ {error}
                    </p>
                </div>
            </div>
        );
    }

    return (
        <div className="fade-in">
            <div style={{ marginBottom: "32px", display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <div>
                    <h1 style={{ marginBottom: "8px" }}>Профиль рисков актива #{id}</h1>
                    <p style={{ color: "var(--color-text-light)", fontSize: "15px" }}>
                        Автоматически рассчитанные риски от всех известных угроз
                    </p>
                </div>
                <button
                    onClick={() => navigate("/assets")}
                    className="btn"
                    style={{
                        padding: "8px 16px",
                        background: "var(--color-surface)",
                        border: "1px solid var(--color-border)",
                        color: "var(--color-text)",
                    }}
                >
                    ← Назад к активам
                </button>
            </div>

            {/* Статистика рисков */}
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: "16px", marginBottom: "24px" }}>
                <StatCard label="Критические" value={stats.critical} color="var(--color-risk-critical)" />
                <StatCard label="Высокие" value={stats.high} color="var(--color-risk-high)" />
                <StatCard label="Средние" value={stats.medium} color="var(--color-risk-medium)" />
                <StatCard label="Низкие" value={stats.low} color="var(--color-risk-low)" />
                <StatCard label="Всего угроз" value={stats.total} color="var(--color-text)" />
            </div>

            {/* Список рисков */}
            {risks.length === 0 ? (
                <div className="card" style={{ padding: "60px 20px", textAlign: "center" }}>
                    <h3 style={{ marginBottom: "8px", color: "var(--color-text)" }}>
                        Риски не обнаружены
                    </h3>
                    <p style={{ color: "var(--color-text-light)", margin: 0 }}>
                        Для данного актива не найдено применимых угроз или уязвимостей
                    </p>
                </div>
            ) : (
                <div style={{ display: "grid", gap: "16px" }}>
                    {risks.map((risk) => (
                        <RiskCard key={risk.threat_id} risk={risk} />
                    ))}
                </div>
            )}
        </div>
    );
};

const StatCard: React.FC<{ label: string; value: number; color: string }> = ({ label, value, color }) => {
    return (
        <div className="card" style={{ padding: "20px", textAlign: "center" }}>
            <div style={{ fontSize: "32px", fontWeight: 700, color, marginBottom: "8px" }}>
                {value}
            </div>
            <div style={{ fontSize: "13px", color: "var(--color-text-light)", fontWeight: 600, textTransform: "uppercase", letterSpacing: "0.5px" }}>
                {label}
            </div>
        </div>
    );
};

const RiskCard: React.FC<{ risk: AssetRisk }> = ({ risk }) => {
    const [expanded, setExpanded] = useState(false);

    const getRiskLevelColor = (level: string): string => {
        switch (level.toLowerCase()) {
            case "critical": return "var(--color-risk-critical)";
            case "high": return "var(--color-risk-high)";
            case "medium": return "var(--color-risk-medium)";
            case "low": return "var(--color-risk-low)";
            default: return "var(--color-text)";
        }
    };

    const getRiskLevelGradient = (level: string): string => {
        switch (level.toLowerCase()) {
            case "critical": return "linear-gradient(135deg, rgba(231, 76, 60, 0.1) 0%, rgba(231, 76, 60, 0.05) 100%)";
            case "high": return "linear-gradient(135deg, rgba(230, 126, 34, 0.1) 0%, rgba(230, 126, 34, 0.05) 100%)";
            case "medium": return "linear-gradient(135deg, rgba(243, 156, 18, 0.1) 0%, rgba(243, 156, 18, 0.05) 100%)";
            case "low": return "linear-gradient(135deg, rgba(39, 174, 96, 0.1) 0%, rgba(39, 174, 96, 0.05) 100%)";
            default: return "var(--color-bg)";
        }
    };

    const color = getRiskLevelColor(risk.level);

    return (
        <div className="card" style={{ padding: "20px", borderLeft: `4px solid ${color}` }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "16px" }}>
                <div style={{ flex: 1 }}>
                    <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "8px" }}>
                        <h3 style={{ margin: 0, fontSize: "18px", color: "var(--color-primary-dark)" }}>
                            {risk.threat_name}
                        </h3>
                        <span
                            style={{
                                fontSize: "11px",
                                color: "var(--color-text-muted)",
                                background: "var(--color-bg)",
                                padding: "2px 8px",
                                borderRadius: "4px",
                                textTransform: "uppercase",
                                letterSpacing: "0.5px",
                            }}
                        >
                            {getThreatTypeLabel(risk.threat_type)}
                        </span>
                    </div>
                    {risk.threat_description && (
                        <p style={{ margin: "0 0 12px 0", color: "var(--color-text-light)", fontSize: "14px" }}>
                            {risk.threat_description}
                        </p>
                    )}
                </div>

                <div
                    style={{
                        padding: "16px",
                        background: getRiskLevelGradient(risk.level),
                        borderRadius: "var(--radius-md)",
                        textAlign: "center",
                        minWidth: "120px",
                        border: `2px solid ${color}`,
                    }}
                >
                    <div style={{ fontSize: "13px", fontWeight: 600, color: "var(--color-text-light)", marginBottom: "4px", textTransform: "uppercase", letterSpacing: "0.5px" }}>
                        Риск
                    </div>
                    <div style={{ fontSize: "24px", fontWeight: 700, color }}>
                        {getRiskLevelLabel(risk.level)}
                    </div>
                    <div style={{ fontSize: "12px", color: "var(--color-text-light)", marginTop: "4px" }}>
                        {risk.score}/25
                    </div>
                </div>
            </div>

            {/* Метрики */}
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px", marginBottom: "16px", paddingTop: "16px", borderTop: "1px solid var(--color-border)" }}>
                <div style={{ padding: "12px", background: "var(--color-bg)", borderRadius: "var(--radius-sm)", textAlign: "center" }}>
                    <div style={{ fontSize: "11px", color: "var(--color-text-light)", marginBottom: "4px", fontWeight: 600, textTransform: "uppercase" }}>
                        Влияние
                    </div>
                    <div style={{ fontSize: "24px", fontWeight: 700, color: "var(--color-primary-dark)" }}>
                        {risk.impact}<span style={{ fontSize: "14px", color: "var(--color-text-light)" }}>/5</span>
                    </div>
                </div>
                <div style={{ padding: "12px", background: "var(--color-bg)", borderRadius: "var(--radius-sm)", textAlign: "center" }}>
                    <div style={{ fontSize: "11px", color: "var(--color-text-light)", marginBottom: "4px", fontWeight: 600, textTransform: "uppercase" }}>
                        Вероятность
                    </div>
                    <div style={{ fontSize: "24px", fontWeight: 700, color: "var(--color-primary-dark)" }}>
                        {risk.likelihood}<span style={{ fontSize: "14px", color: "var(--color-text-light)" }}>/5</span>
                    </div>
                </div>
            </div>

            {/* Рекомендации (сворачиваемые) */}
            {risk.recommendations && risk.recommendations.length > 0 && (
                <div>
                    <button
                        onClick={() => setExpanded(!expanded)}
                        style={{
                            width: "100%",
                            padding: "12px",
                            background: "var(--color-bg)",
                            border: "1px solid var(--color-border)",
                            borderRadius: "var(--radius-sm)",
                            color: "var(--color-text)",
                            fontSize: "13px",
                            fontWeight: 600,
                            cursor: "pointer",
                            display: "flex",
                            justifyContent: "space-between",
                            alignItems: "center",
                            transition: "var(--transition)",
                        }}
                    >
                        <span>📋 Рекомендации ({risk.recommendations.length})</span>
                        <span style={{ fontSize: "18px" }}>{expanded ? "−" : "+"}</span>
                    </button>

                    {expanded && (
                        <div className="fade-in" style={{ marginTop: "12px", display: "grid", gap: "8px" }}>
                            {risk.recommendations.map((rec, idx) => (
                                <div
                                    key={idx}
                                    style={{
                                        padding: "12px",
                                        background: "var(--color-surface)",
                                        border: "1px solid var(--color-border)",
                                        borderRadius: "var(--radius-sm)",
                                    }}
                                >
                                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "4px" }}>
                                        <h4 style={{ margin: 0, fontSize: "14px", color: "var(--color-primary-dark)" }}>
                                            {rec.title}
                                        </h4>
                                        {rec.priority && (
                                            <span className={`badge badge-${rec.priority.toLowerCase()}`}>
                                                {getPriorityLabel(rec.priority)}
                                            </span>
                                        )}
                                    </div>
                                    <p style={{ margin: "4px 0 0 0", fontSize: "12px", color: "var(--color-text-light)" }}>
                                        {rec.description}
                                    </p>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};
