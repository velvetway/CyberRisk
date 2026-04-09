import React, { useEffect, useState } from "react";
import { Asset, Threat, RiskPreviewResponse, RiskRecommendation } from "../types";
import { RiskHeatmapSingle } from "../components/RiskHeatmapSingle";
import { getPriorityLabel, getRecommendationCategoryLabel, getRiskLevelLabel } from "../utils/i18n";

export const RiskPreviewPage: React.FC = () => {
    const [assets, setAssets] = useState<Asset[]>([]);
    const [threats, setThreats] = useState<Threat[]>([]);
    const [selectedAssetId, setSelectedAssetId] = useState<number | "">("");
    const [selectedThreatId, setSelectedThreatId] = useState<number | "">("");
    const [result, setResult] = useState<RiskPreviewResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [assetsRes, threatsRes] = await Promise.all([
                    fetch("/api/assets"),
                    fetch("/api/threats"),
                ]);

                if (!assetsRes.ok || !threatsRes.ok) {
                    throw new Error("Ошибка загрузки данных");
                }

                const assetsJson: Asset[] = await assetsRes.json();
                const threatsJson: Threat[] = await threatsRes.json();

                setAssets(assetsJson);
                setThreats(threatsJson);

                if (assetsJson.length > 0) setSelectedAssetId(assetsJson[0].id);
                if (threatsJson.length > 0) setSelectedThreatId(threatsJson[0].id);
            } catch (e: any) {
                setError(e.message || "Ошибка загрузки данных");
            }
        };
        fetchData();
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        setResult(null);

        if (!selectedAssetId || !selectedThreatId) {
            setError("Выберите ИТ-актив и угрозу.");
            return;
        }

        setLoading(true);
        try {
            const res = await fetch("/api/risk/preview", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    asset_id: selectedAssetId,
                    threat_id: selectedThreatId,
                }),
            });

            if (!res.ok) {
                const text = await res.text();
                throw new Error(`Ошибка HTTP ${res.status}${text ? `: ${text}` : ""}`);
            }

            const json: RiskPreviewResponse = await res.json();
            setResult(json);
        } catch (e: any) {
            setError(e.message || "Ошибка при расчёте риска");
        } finally {
            setLoading(false);
        }
    };

    const handleDownloadPDF = async () => {
        if (!selectedAssetId || !selectedThreatId) {
            return;
        }

        try {
            const res = await fetch("/api/risk/report/pdf", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    asset_id: selectedAssetId,
                    threat_id: selectedThreatId,
                }),
            });

            if (!res.ok) {
                throw new Error(`Ошибка при генерации PDF: ${res.status}`);
            }

            const blob = await res.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = `risk_report_${selectedAssetId}_${selectedThreatId}_${new Date().toISOString().split('T')[0]}.pdf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (e: any) {
            setError(e.message || "Ошибка при скачивании PDF");
        }
    };

    const selectedAsset = assets.find((a) => a.id === selectedAssetId);
    const selectedThreat = threats.find((t) => t.id === selectedThreatId);

    return (
        <div className="fade-in">
            <div style={{ marginBottom: "32px" }}>
                <h1 style={{ marginBottom: "8px" }}>Симулятор риска "Что если?"</h1>
                <p style={{ color: "var(--color-text-light)", fontSize: "15px" }}>
                    Быстрый инструмент для моделирования сценариев: выберите актив и угрозу для оценки потенциального риска. Для полного автоматического анализа используйте кнопку "Риски" на карточке актива.
                </p>
            </div>

            <div style={{ display: "grid", gap: "24px", gridTemplateColumns: "repeat(auto-fit, minmax(450px, 1fr))" }}>
                {/* Левая панель - форма */}
                <div className="card" style={{ padding: "24px" }}>
                    <h2 style={{ fontSize: "18px", marginBottom: "20px", display: "flex", alignItems: "center", gap: "8px" }}>
                        <svg width="20" height="20" viewBox="0 0 20 20" fill="var(--color-accent)">
                            <path d="M10 2l6 12H4L10 2z" />
                        </svg>
                        Параметры расчёта
                    </h2>

                    <form onSubmit={handleSubmit}>
                        <div style={{ marginBottom: "20px" }}>
                            <label style={{
                                display: "block",
                                marginBottom: "8px",
                                fontWeight: 600,
                                fontSize: "14px",
                                color: "var(--color-text)",
                            }}>
                                ИТ-актив
                            </label>
                            <select
                                value={selectedAssetId}
                                onChange={(e) => setSelectedAssetId(e.target.value ? parseInt(e.target.value, 10) : "")}
                                style={{
                                    width: "100%",
                                    padding: "12px",
                                    border: "1px solid var(--color-border)",
                                    borderRadius: "var(--radius-sm)",
                                    fontSize: "14px",
                                    background: "var(--color-surface)",
                                    color: "var(--color-text)",
                                    cursor: "pointer",
                                    transition: "var(--transition)",
                                }}
                                onFocus={(e) => e.target.style.borderColor = "var(--color-accent)"}
                                onBlur={(e) => e.target.style.borderColor = "var(--color-border)"}
                            >
                                <option value="">— Выберите актив —</option>
                                {assets.map((a) => (
                                    <option key={a.id} value={a.id}>
                                        {a.name} (ID: {a.id})
                                    </option>
                                ))}
                            </select>

                            {selectedAsset && (
                                <div style={{
                                    marginTop: "12px",
                                    padding: "12px",
                                    background: "var(--color-bg)",
                                    borderRadius: "var(--radius-sm)",
                                    fontSize: "13px",
                                    color: "var(--color-text-light)",
                                }}>
                                    <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "4px" }}>
                                        <span><strong>Среда:</strong> {selectedAsset.environment?.toUpperCase()}</span>
                                        <span><strong>Критичность:</strong> {selectedAsset.business_criticality}/5</span>
                                    </div>
                                    {selectedAsset.description && (
                                        <div style={{ marginTop: "8px", fontSize: "12px", fontStyle: "italic" }}>
                                            {selectedAsset.description}
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>

                        <div style={{ marginBottom: "24px" }}>
                            <label style={{
                                display: "block",
                                marginBottom: "8px",
                                fontWeight: 600,
                                fontSize: "14px",
                                color: "var(--color-text)",
                            }}>
                                Угроза
                            </label>
                            <select
                                value={selectedThreatId}
                                onChange={(e) => setSelectedThreatId(e.target.value ? parseInt(e.target.value, 10) : "")}
                                style={{
                                    width: "100%",
                                    padding: "12px",
                                    border: "1px solid var(--color-border)",
                                    borderRadius: "var(--radius-sm)",
                                    fontSize: "14px",
                                    background: "var(--color-surface)",
                                    color: "var(--color-text)",
                                    cursor: "pointer",
                                    transition: "var(--transition)",
                                }}
                                onFocus={(e) => e.target.style.borderColor = "var(--color-accent)"}
                                onBlur={(e) => e.target.style.borderColor = "var(--color-border)"}
                            >
                                <option value="">— Выберите угрозу —</option>
                                {threats.map((t) => (
                                    <option key={t.id} value={t.id}>
                                        {t.name} (ID: {t.id})
                                    </option>
                                ))}
                            </select>

                            {selectedThreat && selectedThreat.description && (
                                <div style={{
                                    marginTop: "12px",
                                    padding: "12px",
                                    background: "var(--color-bg)",
                                    borderRadius: "var(--radius-sm)",
                                    fontSize: "13px",
                                    color: "var(--color-text-light)",
                                }}>
                                    {selectedThreat.description}
                                </div>
                            )}
                        </div>

                        <button type="submit" className="btn btn-primary" disabled={loading} style={{ width: "100%" }}>
                            {loading ? (
                                <>
                                    <span className="loading-spinner" />
                                    Расчёт...
                                </>
                            ) : (
                                <>
                                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                        <path d="M8 2l6 12H2L8 2z" />
                                    </svg>
                                    Рассчитать риск
                                </>
                            )}
                        </button>

                        {error && (
                            <div style={{
                                marginTop: "16px",
                                padding: "12px",
                                background: "rgba(231, 76, 60, 0.1)",
                                borderRadius: "var(--radius-sm)",
                                color: "var(--color-risk-critical)",
                                fontSize: "14px",
                                border: "1px solid var(--color-risk-critical)",
                            }}>
                                ⚠️ {error}
                            </div>
                        )}
                    </form>
                </div>

                {/* Правая панель - результаты */}
                <div className="card" style={{ padding: "24px" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
                        <h2 style={{ fontSize: "18px", margin: 0, display: "flex", alignItems: "center", gap: "8px" }}>
                            <svg width="20" height="20" viewBox="0 0 20 20" fill="var(--color-info)">
                                <rect x="2" y="2" width="6" height="6" />
                                <rect x="12" y="2" width="6" height="6" />
                                <rect x="2" y="12" width="6" height="6" />
                                <rect x="12" y="12" width="6" height="6" />
                            </svg>
                            Результат оценки
                        </h2>
                        {result && (
                            <button
                                type="button"
                                onClick={handleDownloadPDF}
                                className="btn"
                                style={{
                                    padding: "8px 16px",
                                    background: "var(--color-success)",
                                    color: "white",
                                    border: "none",
                                    borderRadius: "var(--radius-sm)",
                                    fontSize: "13px",
                                    fontWeight: 600,
                                    cursor: "pointer",
                                    display: "flex",
                                    alignItems: "center",
                                    gap: "6px",
                                    transition: "var(--transition)",
                                }}
                                onMouseEnter={(e) => e.currentTarget.style.transform = "translateY(-2px)"}
                                onMouseLeave={(e) => e.currentTarget.style.transform = "translateY(0)"}
                            >
                                <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                                    <path d="M8 12L3 7h3V2h4v5h3l-5 5zM2 14h12v2H2z" />
                                </svg>
                                Скачать PDF
                            </button>
                        )}
                    </div>

                    {!result && (
                        <div style={{
                            textAlign: "center",
                            padding: "60px 20px",
                            color: "var(--color-text-muted)",
                        }}>
                            <svg width="64" height="64" viewBox="0 0 64 64" fill="none" style={{ margin: "0 auto 16px", opacity: 0.3 }}>
                                <rect width="64" height="64" rx="32" fill="var(--color-bg)" />
                                <path d="M32 20v24M20 32h24" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
                            </svg>
                            <p style={{ margin: 0, fontSize: "14px" }}>
                                Выберите актив и угрозу, затем нажмите &quot;Рассчитать риск&quot;
                            </p>
                        </div>
                    )}

                    {result && (
                        <div className="fade-in">
                            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginBottom: "20px" }}>
                                <RiskMetric label="Влияние" value={result.impact} max={5} />
                                <RiskMetric label="Вероятность" value={result.likelihood} max={5} />
                            </div>

                            <div style={{
                                padding: "20px",
                                background: getRiskLevelGradient(result.level),
                                borderRadius: "var(--radius-md)",
                                textAlign: "center",
                                marginBottom: "20px",
                                border: `2px solid ${getRiskLevelBorderColor(result.level)}`,
                            }}>
                                <div style={{
                                    fontSize: "13px",
                                    fontWeight: 600,
                                    color: "var(--color-text-light)",
                                    marginBottom: "8px",
                                    textTransform: "uppercase",
                                    letterSpacing: "0.5px",
                                }}>
                                    Уровень риска
                                </div>
                                <div
                                    style={{
                                        fontSize: "32px",
                                        fontWeight: 700,
                                        color: getRiskLevelTextColor(result.level),
                                    }}
                                >
                                    {getRiskLevelLabel(result.level)}
                                </div>
                                <div style={{
                                    fontSize: "18px",
                                    fontWeight: 600,
                                    color: "var(--color-text-light)",
                                    marginTop: "8px",
                                }}>
                                    Оценка: {result.score}/25
                                </div>
                            </div>

                            <RiskHeatmapSingle
                                impact={result.impact}
                                likelihood={result.likelihood}
                                score={result.score}
                                level={result.level}
                            />
                        </div>
                    )}
                </div>
            </div>

            {/* Рекомендации */}
            {result && result.recommendations && result.recommendations.length > 0 && (
                <div className="card fade-in" style={{ marginTop: "24px", padding: "24px" }}>
                    <h2 style={{ fontSize: "18px", marginBottom: "20px", display: "flex", alignItems: "center", gap: "8px" }}>
                        <svg width="20" height="20" viewBox="0 0 20 20" fill="var(--color-success)">
                            <path d="M10 2l2 6h6l-5 4 2 6-5-4-5 4 2-6-5-4h6z" />
                        </svg>
                        Рекомендации по снижению риска ({result.recommendations.length})
                    </h2>

                    <div style={{ display: "grid", gap: "12px" }}>
                        {result.recommendations.map((rec, idx) => (
                            <RecommendationCard key={idx} rec={rec} />
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

const RiskMetric: React.FC<{ label: string; value: number; max: number }> = ({ label, value, max }) => {
    return (
        <div style={{
            padding: "16px",
            background: "var(--color-bg)",
            borderRadius: "var(--radius-sm)",
            textAlign: "center",
        }}>
            <div style={{
                fontSize: "12px",
                color: "var(--color-text-light)",
                marginBottom: "8px",
                fontWeight: 600,
                textTransform: "uppercase",
                letterSpacing: "0.5px",
            }}>
                {label}
            </div>
            <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--color-primary-dark)" }}>
                {value}
                <span style={{ fontSize: "16px", color: "var(--color-text-light)", fontWeight: 500 }}>
                    /{max}
                </span>
            </div>
        </div>
    );
};

const RecommendationCard: React.FC<{ rec: RiskRecommendation }> = ({ rec }) => {
    const getPriorityBadge = (priority: string) => {
        const label = getPriorityLabel(priority);
        if (priority === "high") return <span className="badge badge-critical">{label}</span>;
        if (priority === "medium") return <span className="badge badge-medium">{label}</span>;
        return <span className="badge badge-low">{label}</span>;
    };

    return (
        <div className="card" style={{
            padding: "16px",
            border: "1px solid var(--color-border)",
            transition: "var(--transition)",
        }}>
            <div style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "flex-start",
                marginBottom: "8px",
            }}>
                <h3 style={{ fontSize: "15px", margin: 0, flex: 1, color: "var(--color-primary-dark)" }}>
                    {rec.title}
                </h3>
                {getPriorityBadge(rec.priority)}
            </div>
            <p style={{ margin: "8px 0 0 0", fontSize: "13px", color: "var(--color-text-light)", lineHeight: 1.6 }}>
                {rec.description}
            </p>
            <div
                style={{
                    marginTop: "8px",
                    fontSize: "11px",
                    color: "var(--color-text-muted)",
                    textTransform: "uppercase",
                    letterSpacing: "0.5px",
                }}
            >
                Категория: {getRecommendationCategoryLabel(rec.category)}
            </div>
        </div>
    );
};

const getRiskLevelGradient = (level: string): string => {
    switch (level.toLowerCase()) {
        case "low": return "linear-gradient(135deg, rgba(39, 174, 96, 0.1) 0%, rgba(39, 174, 96, 0.05) 100%)";
        case "medium": return "linear-gradient(135deg, rgba(243, 156, 18, 0.1) 0%, rgba(243, 156, 18, 0.05) 100%)";
        case "high": return "linear-gradient(135deg, rgba(230, 126, 34, 0.1) 0%, rgba(230, 126, 34, 0.05) 100%)";
        case "critical": return "linear-gradient(135deg, rgba(231, 76, 60, 0.1) 0%, rgba(231, 76, 60, 0.05) 100%)";
        default: return "var(--color-bg)";
    }
};

const getRiskLevelBorderColor = (level: string): string => {
    switch (level.toLowerCase()) {
        case "low": return "var(--color-risk-low)";
        case "medium": return "var(--color-risk-medium)";
        case "high": return "var(--color-risk-high)";
        case "critical": return "var(--color-risk-critical)";
        default: return "var(--color-border)";
    }
};

const getRiskLevelTextColor = (level: string): string => {
    switch (level.toLowerCase()) {
        case "low": return "var(--color-risk-low)";
        case "medium": return "var(--color-risk-medium)";
        case "high": return "var(--color-risk-high)";
        case "critical": return "var(--color-risk-critical)";
        default: return "var(--color-text)";
    }
};
