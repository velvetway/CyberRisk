import React, { useEffect, useState } from "react";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { authFetch } from "../api/client";
import { AttackPath } from "../types/riskGraph";
import { RiskGraphSankey } from "../components/RiskGraphSankey";

export const RiskGraphPage: React.FC = () => {
    const { assetId } = useParams();
    const [params] = useSearchParams();
    const threatId = params.get("threat");
    const nav = useNavigate();
    const [path, setPath] = useState<AttackPath | null>(null);
    const [err, setErr] = useState<string | null>(null);

    useEffect(() => {
        if (!assetId || !threatId) return;
        authFetch(`/api/risk/graph/${assetId}/${threatId}`)
            .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
            .then(setPath)
            .catch(e => setErr(e.message));
    }, [assetId, threatId]);

    if (!threatId) return <p style={{ padding: 24 }}>Укажите <code>?threat=&lt;id&gt;</code> в URL.</p>;
    if (err) return <p style={{ padding: 24, color: "var(--danger)" }}>Ошибка: {err}</p>;
    if (!path) return <p style={{ padding: 24 }}>Загрузка...</p>;

    return (
        <div style={{ padding: 24 }}>
            <button
                onClick={() => nav(-1)}
                style={{
                    marginBottom: 16,
                    background: "transparent",
                    border: "1px solid var(--perimeter)",
                    borderRadius: 6,
                    padding: "6px 12px",
                    color: "var(--ink)",
                    cursor: "pointer",
                }}
            >← Назад</button>

            <h1 style={{ margin: "0 0 4px", fontSize: 22 }}>Граф атаки</h1>
            <div style={{ marginBottom: 20, color: "var(--ink-muted)", fontSize: 14 }}>
                <strong>{path.asset.name}</strong> ← {path.threat.name}
                {path.threat.bdu_id && <span> · {path.threat.bdu_id}</span>}
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "repeat(5, 1fr)", gap: 12, marginBottom: 24 }}>
                <Stat label="W"           value={path.w.toFixed(3)}         tone={path.level} />
                <Stat label="Q^threat"    value={path.q_threat.toFixed(2)}  />
                <Stat label="q^severity"  value={path.q_severity.toFixed(2)} />
                <Stat label="Q^reaction"  value={path.q_reaction.toFixed(2)} />
                <Stat label="Z"           value={path.z.toFixed(2)} />
            </div>

            <RiskGraphSankey path={path} />
        </div>
    );
};

const Stat: React.FC<{ label: string; value: string; tone?: string }> = ({ label, value, tone }) => {
    const color = tone === "critical" ? "var(--danger)"
               : tone === "high"     ? "var(--threat-high)"
               : tone === "medium"   ? "var(--warning)"
               : tone === "low"      ? "var(--success)"
               : "var(--ink)";
    return (
        <div className="card" style={{ padding: 16 }}>
            <div style={{ fontSize: 11, color: "var(--ink-muted)", fontWeight: 600 }}>{label}</div>
            <div style={{ fontSize: 22, fontWeight: 700, color }}>{value}</div>
        </div>
    );
};
