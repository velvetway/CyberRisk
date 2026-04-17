import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { authFetch } from "../api/client";
import { getPriorityLabel, getRiskLevelLabel, getThreatTypeLabel } from "../utils/i18n";
import { motion } from "framer-motion";
import { ArrowLeft, ChevronDown, ChevronUp, CheckCircle } from "lucide-react";

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
                const res = await authFetch(`/api/risk/asset/${id}`);
                if (!res.ok) throw new Error(`Ошибка загрузки рисков: ${res.status}`);
                const data: AssetRisk[] = await res.json();
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

    const stats = {
        critical: risks.filter((r) => r.level.toLowerCase() === "critical").length,
        high: risks.filter((r) => r.level.toLowerCase() === "high").length,
        medium: risks.filter((r) => r.level.toLowerCase() === "medium").length,
        low: risks.filter((r) => r.level.toLowerCase() === "low").length,
        total: risks.length
    };

    if (loading) return (
        <div style={{ textAlign: "center", padding: "60px 20px" }}>
            <div className="loading-spinner" style={{ width: "40px", height: "40px", margin: "0 auto" }} />
            <p style={{ marginTop: "16px", color: "var(--ink-muted)" }}>Анализ рисков для актива...</p>
        </div>
    );

    if (error) return (
        <div className="card" style={{ padding: "20px", background: "var(--threat-critical-dim)", border: "1px solid var(--danger)" }}>
            <p style={{ color: "var(--danger)", margin: 0, fontWeight: 500 }}>⚠️ {error}</p>
        </div>
    );

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }}>
            <div style={{ marginBottom: "32px", display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <div>
                    <h1>Профиль рисков актива #{id}</h1>
                    <p style={{ color: "var(--ink-muted)" }}>Автоматически рассчитанные риски от всех известных угроз</p>
                </div>
                <button onClick={() => navigate("/assets")} className="btn">
                    <ArrowLeft size={16} /> Назад к активам
                </button>
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: "16px", marginBottom: "32px" }}>
                <StatCard label="Критические" value={stats.critical} color="var(--danger)" />
                <StatCard label="Высокие" value={stats.high} color="var(--threat-high)" />
                <StatCard label="Средние" value={stats.medium} color="var(--warning)" />
                <StatCard label="Низкие" value={stats.low} color="var(--success)" />
                <StatCard label="Всего угроз" value={stats.total} color="var(--ink)" />
            </div>

            {risks.length === 0 ? (
                <div className="card" style={{ padding: "60px 20px", textAlign: "center" }}>
                    <CheckCircle size={48} color="var(--success)" style={{ margin: "0 auto 16px" }} />
                    <h3 style={{ marginBottom: "8px" }}>Риски не обнаружены</h3>
                    <p style={{ color: "var(--ink-muted)", margin: 0 }}>Для данного актива не найдено применимых угроз или уязвимостей</p>
                </div>
            ) : (
                <div style={{ display: "grid", gap: "16px" }}>
                    {risks.map((risk) => <RiskCard key={risk.threat_id} risk={risk} />)}
                </div>
            )}
        </motion.div>
    );
};

const StatCard: React.FC<{ label: string; value: number; color: string }> = ({ label, value, color }) => (
    <div className="card" style={{ padding: "20px" }}>
        <div style={{ fontSize: "32px", fontWeight: 700, color, marginBottom: "4px" }}>{value}</div>
        <div style={{ fontSize: "13px", color: "var(--ink-muted)", fontWeight: 600 }}>{label}</div>
    </div>
);

const RiskCard: React.FC<{ risk: AssetRisk }> = ({ risk }) => {
    const [expanded, setExpanded] = useState(false);

    const getRiskColor = (level: string) => {
        switch (level.toLowerCase()) {
            case "critical": return "var(--danger)";
            case "high": return "var(--threat-high)";
            case "medium": return "var(--warning)";
            case "low": return "var(--success)";
            default: return "var(--ink)";
        }
    };

    const getRiskBg = (level: string) => {
        switch (level.toLowerCase()) {
            case "critical": return "var(--threat-critical-dim)";
            case "high": return "var(--threat-high-dim)";
            case "medium": return "var(--threat-medium-dim)";
            case "low": return "var(--threat-low-dim)";
            default: return "var(--well)";
        }
    };

    const color = getRiskColor(risk.level);
    const bg = getRiskBg(risk.level);

    return (
        <div className="card" style={{ overflow: "hidden", borderLeft: `4px solid ${color}` }}>
            <div style={{ padding: "20px", display: "flex", gap: "24px", alignItems: "flex-start" }}>
                <div style={{ flex: 1 }}>
                    <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "8px" }}>
                        <h3 style={{ margin: 0, fontSize: "16px" }}>{risk.threat_name}</h3>
                        <span className="badge" style={{ background: "var(--well)", color: "var(--ink-muted)" }}>
                            {getThreatTypeLabel(risk.threat_type)}
                        </span>
                    </div>
                    {risk.threat_description && (
                        <p style={{ margin: "0", color: "var(--ink-secondary)", fontSize: "14px" }}>{risk.threat_description}</p>
                    )}
                </div>

                <div style={{ display: "flex", gap: "16px" }}>
                    <div style={{ textAlign: "center", padding: "10px 16px", background: "var(--well)", borderRadius: "var(--r-sm)" }}>
                        <div style={{ fontSize: "11px", color: "var(--ink-muted)", fontWeight: 600, marginBottom: "4px" }}>I / L</div>
                        <div style={{ fontSize: "16px", fontWeight: 700, color: "var(--ink)" }}>{risk.impact} / {risk.likelihood}</div>
                    </div>
                    <div style={{ textAlign: "center", padding: "10px 24px", background: bg, borderRadius: "var(--r-sm)", minWidth: "120px" }}>
                        <div style={{ fontSize: "11px", color: color, fontWeight: 600, marginBottom: "4px" }}>УРОВЕНЬ</div>
                        <div style={{ fontSize: "20px", fontWeight: 700, color }}>{getRiskLevelLabel(risk.level)}</div>
                    </div>
                </div>
            </div>

            {risk.recommendations && risk.recommendations.length > 0 && (
                <div style={{ borderTop: "1px solid var(--perimeter)", background: "var(--ground)" }}>
                    <button
                        onClick={() => setExpanded(!expanded)}
                        style={{ width: "100%", padding: "12px 20px", background: "none", border: "none", display: "flex", justifyContent: "space-between", alignItems: "center", cursor: "pointer", fontSize: "13px", fontWeight: 600, color: "var(--ink-secondary)" }}
                    >
                        <span>Рекомендации по снижению риска ({risk.recommendations.length})</span>
                        {expanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                    </button>
                    {expanded && (
                        <div style={{ padding: "0 20px 20px 20px", display: "grid", gap: "12px" }}>
                            {risk.recommendations.map((rec, idx) => (
                                <div key={idx} style={{ padding: "16px", background: "var(--raised)", border: "1px solid var(--perimeter)", borderRadius: "var(--r-sm)" }}>
                                    <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
                                        <h4 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>{rec.title}</h4>
                                        {rec.priority && <span className={`badge badge-${rec.priority.toLowerCase()}`}>{getPriorityLabel(rec.priority)}</span>}
                                    </div>
                                    <p style={{ margin: 0, fontSize: "13px", color: "var(--ink-muted)" }}>{rec.description}</p>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};
