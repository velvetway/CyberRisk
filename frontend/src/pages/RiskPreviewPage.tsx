import React, { useEffect, useState } from "react";
import { authFetch } from "../api/client";
import { Asset, Threat, RiskPreviewResponse } from "../types";
import { getPriorityLabel, getRiskLevelLabel } from "../utils/i18n";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Calculator, Download, Layers, ShieldAlert, Activity } from "lucide-react";

export const RiskPreviewPage: React.FC = () => {
    const [assets, setAssets] = useState<Asset[]>([]);
    const [threats, setThreats] = useState<Threat[]>([]);
    const [selectedAssetId, setSelectedAssetId] = useState<number | "">("");
    const [selectedThreatId, setSelectedThreatId] = useState<number | "">("");
    const [result, setResult] = useState<RiskPreviewResponse | null>(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [assetsRes, threatsRes] = await Promise.all([ authFetch("/api/assets"), authFetch("/api/threats") ]);
                if (!assetsRes.ok || !threatsRes.ok) throw new Error("Ошибка загрузки данных");
                
                const assetsJson: Asset[] = await assetsRes.json();
                const threatsJson: Threat[] = await threatsRes.json();

                setAssets(assetsJson);
                setThreats(threatsJson);

                if (assetsJson.length > 0) setSelectedAssetId(assetsJson[0].id);
                if (threatsJson.length > 0) setSelectedThreatId(threatsJson[0].id);
            } catch (e: any) {
                toast.error(e.message || "Ошибка загрузки данных");
            }
        };
        fetchData();
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setResult(null);

        if (!selectedAssetId || !selectedThreatId) {
            toast.error("Выберите ИТ-актив и угрозу");
            return;
        }

        setLoading(true);
        try {
            const res = await authFetch("/api/risk/preview", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ asset_id: selectedAssetId, threat_id: selectedThreatId }),
            });

            if (!res.ok) throw new Error(`Ошибка HTTP ${res.status}`);
            const json: RiskPreviewResponse = await res.json();
            setResult(json);
            toast.success("Риск рассчитан");
        } catch (e: any) {
            toast.error(e.message || "Ошибка при расчёте риска");
        } finally {
            setLoading(false);
        }
    };

    const handleDownloadPDF = async () => {
        if (!selectedAssetId || !selectedThreatId) return;
        try {
            const res = await authFetch("/api/risk/report/pdf", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ asset_id: selectedAssetId, threat_id: selectedThreatId }),
            });

            if (!res.ok) throw new Error(`Ошибка при генерации PDF: ${res.status}`);
            
            const blob = await res.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = `risk_report_${selectedAssetId}_${selectedThreatId}.pdf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
            toast.success("Отчет скачан");
        } catch (e: any) {
            toast.error(e.message || "Ошибка при скачивании PDF");
        }
    };

    const selectedAsset = assets.find((a) => a.id === selectedAssetId);
    const selectedThreat = threats.find((t) => t.id === selectedThreatId);

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }}>
            <div style={{ marginBottom: "32px" }}>
                <h1>Симулятор риска</h1>
                <p style={{ color: "var(--ink-muted)" }}>Инструмент быстрого моделирования "Что если?" для оценки потенциального риска</p>
            </div>

            <div style={{ display: "grid", gap: "24px", gridTemplateColumns: "repeat(auto-fit, minmax(450px, 1fr))" }}>
                <div className="card" style={{ padding: "24px" }}>
                    <h2 style={{ marginBottom: "20px", display: "flex", alignItems: "center", gap: "8px" }}>
                        <Calculator size={20} color="var(--command)" /> Параметры расчёта
                    </h2>

                    <form onSubmit={handleSubmit}>
                        <div style={{ marginBottom: "24px" }}>
                            <label className="form-label">ИТ-актив</label>
                            <select className="form-input" value={selectedAssetId} onChange={(e) => setSelectedAssetId(e.target.value ? parseInt(e.target.value, 10) : "")}>
                                <option value="">— Выберите актив —</option>
                                {assets.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}
                            </select>
                            {selectedAsset && (
                                <div style={{ marginTop: "12px", padding: "12px", background: "var(--well)", borderRadius: "var(--r-sm)", fontSize: "13px", color: "var(--ink-muted)" }}>
                                    <strong>Среда:</strong> {selectedAsset.environment?.toUpperCase()} | <strong>Критичность:</strong> {selectedAsset.business_criticality}/5
                                </div>
                            )}
                        </div>

                        <div style={{ marginBottom: "24px" }}>
                            <label className="form-label">Угроза</label>
                            <select className="form-input" value={selectedThreatId} onChange={(e) => setSelectedThreatId(e.target.value ? parseInt(e.target.value, 10) : "")}>
                                <option value="">— Выберите угрозу —</option>
                                {threats.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                            </select>
                            {selectedThreat && selectedThreat.description && (
                                <div style={{ marginTop: "12px", padding: "12px", background: "var(--well)", borderRadius: "var(--r-sm)", fontSize: "13px", color: "var(--ink-muted)" }}>
                                    {selectedThreat.description}
                                </div>
                            )}
                        </div>

                        <button type="submit" className="btn btn-primary" disabled={loading} style={{ width: "100%" }}>
                            {loading ? <span className="loading-spinner" /> : <><Activity size={18} /> Рассчитать риск</>}
                        </button>
                    </form>
                </div>

                <div className="card" style={{ padding: "24px" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
                        <h2 style={{ margin: 0, display: "flex", alignItems: "center", gap: "8px" }}>
                            <Layers size={20} color="var(--command)" /> Результат оценки
                        </h2>
                        {result && (
                            <button type="button" onClick={handleDownloadPDF} className="btn" style={{ background: "var(--well)", border: "1px solid var(--perimeter)", color: "var(--ink)" }}>
                                <Download size={16} /> Скачать PDF
                            </button>
                        )}
                    </div>

                    {!result && (
                        <div style={{ textAlign: "center", padding: "60px 20px", color: "var(--ink-muted)" }}>
                            <Activity size={48} style={{ opacity: 0.2, margin: "0 auto 16px" }} />
                            <p style={{ margin: 0, fontSize: "14px" }}>Выберите актив и угрозу для расчета</p>
                        </div>
                    )}

                    {result && (
                        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
                            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginBottom: "20px" }}>
                                <div style={{ padding: "16px", background: "var(--well)", borderRadius: "var(--r-sm)", textAlign: "center" }}>
                                    <div style={{ fontSize: "11px", color: "var(--ink-muted)", fontWeight: 600 }}>ВЛИЯНИЕ</div>
                                    <div style={{ fontSize: "32px", fontWeight: 700, color: "var(--ink)" }}>{result.impact}</div>
                                </div>
                                <div style={{ padding: "16px", background: "var(--well)", borderRadius: "var(--r-sm)", textAlign: "center" }}>
                                    <div style={{ fontSize: "11px", color: "var(--ink-muted)", fontWeight: 600 }}>ВЕРОЯТНОСТЬ</div>
                                    <div style={{ fontSize: "32px", fontWeight: 700, color: "var(--ink)" }}>{result.likelihood}</div>
                                </div>
                            </div>

                            <div style={{ padding: "24px", background: getRiskLevelBg(result.level), borderRadius: "var(--r-md)", textAlign: "center", border: `1px solid ${getRiskLevelColor(result.level)}` }}>
                                <div style={{ fontSize: "12px", fontWeight: 600, color: getRiskLevelColor(result.level), marginBottom: "4px" }}>ИТОГОВЫЙ РИСК</div>
                                <div style={{ fontSize: "36px", fontWeight: 800, color: getRiskLevelColor(result.level) }}>{getRiskLevelLabel(result.level)}</div>
                                <div style={{ fontSize: "16px", fontWeight: 600, color: getRiskLevelColor(result.level), marginTop: "8px" }}>Оценка: {result.score}/25</div>
                            </div>
                        </motion.div>
                    )}
                </div>
            </div>

            {result && result.recommendations && result.recommendations.length > 0 && (
                <motion.div className="card" initial={{ opacity: 0 }} animate={{ opacity: 1 }} style={{ marginTop: "24px", padding: "24px" }}>
                    <h2 style={{ marginBottom: "20px", display: "flex", alignItems: "center", gap: "8px" }}>
                        <ShieldAlert size={20} color="var(--success)" /> Рекомендации по снижению риска ({result.recommendations.length})
                    </h2>
                    <div style={{ display: "grid", gap: "12px" }}>
                        {result.recommendations.map((rec, idx) => (
                            <div key={idx} style={{ padding: "16px", background: "var(--raised)", border: "1px solid var(--perimeter)", borderRadius: "var(--r-sm)" }}>
                                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "8px" }}>
                                    <h3 style={{ fontSize: "15px", margin: 0, fontWeight: 600 }}>{rec.title}</h3>
                                    <span className={`badge badge-${rec.priority.toLowerCase()}`}>{getPriorityLabel(rec.priority)}</span>
                                </div>
                                <p style={{ margin: "0", fontSize: "13px", color: "var(--ink-secondary)", lineHeight: 1.6 }}>{rec.description}</p>
                            </div>
                        ))}
                    </div>
                </motion.div>
            )}
        </motion.div>
    );
};

const getRiskLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
        case "critical": return "var(--danger)";
        case "high": return "var(--threat-high)";
        case "medium": return "var(--warning)";
        case "low": return "var(--success)";
        default: return "var(--ink)";
    }
};

const getRiskLevelBg = (level: string) => {
    switch (level.toLowerCase()) {
        case "critical": return "var(--threat-critical-dim)";
        case "high": return "var(--threat-high-dim)";
        case "medium": return "var(--threat-medium-dim)";
        case "low": return "var(--threat-low-dim)";
        default: return "var(--well)";
    }
};
