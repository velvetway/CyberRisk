import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { Asset, AssetSoftwareAlternative } from "../types";

export const AssetsPage: React.FC = () => {
    const navigate = useNavigate();
    const [assets, setAssets] = useState<Asset[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [deletingId, setDeletingId] = useState<number | null>(null);

    useEffect(() => {
        api.getAssets()
            .then(setAssets)
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false));
    }, []);

    const getEnvironmentColor = (env: string) => {
        switch (env?.toLowerCase()) {
            case "prod":
                return { bg: "rgba(231, 76, 60, 0.1)", color: "#e74c3c" };
            case "test":
                return { bg: "rgba(243, 156, 18, 0.1)", color: "#f39c12" };
            case "dev":
                return { bg: "rgba(52, 152, 219, 0.1)", color: "#3498db" };
            default:
                return { bg: "rgba(149, 165, 166, 0.1)", color: "#95a5a6" };
        }
    };

    const getCriticalityBadge = (criticality: number) => {
        if (criticality >= 4)
            return <span className="badge badge-critical">Высокая</span>;
        if (criticality >= 3)
            return <span className="badge badge-high">Средняя</span>;
        return <span className="badge badge-low">Низкая</span>;
    };

    const handleDelete = async (asset: Asset) => {
        const confirmed = window.confirm(`Удалить актив «${asset.name}»? Это действие необратимо.`);
        if (!confirmed) return;

        try {
            setDeletingId(asset.id);
            await api.deleteAsset(asset.id);
            setAssets((prev) => prev.filter((a) => a.id !== asset.id));
            setError(null);
        } catch (err) {
            const message = err instanceof Error ? err.message : "Не удалось удалить актив";
            setError(message);
        } finally {
            setDeletingId(null);
        }
    };

    return (
        <div className="fade-in">
            <div style={{ marginBottom: "32px", display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                <div>
                    <h1 style={{ marginBottom: "8px" }}>ИТ-активы организации</h1>
                    <p style={{ color: "var(--color-text-light)", fontSize: "15px" }}>
                        Управление и мониторинг критически важных активов информационной
                        инфраструктуры
                    </p>
                </div>
                <button
                    onClick={() => navigate("/assets/new")}
                    className="btn btn-primary"
                    style={{ marginTop: "4px" }}
                >
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                        <path d="M8 2v12M2 8h12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                    </svg>
                    Создать актив
                </button>
            </div>

            {loading && (
                <div style={{ textAlign: "center", padding: "60px 20px" }}>
                    <div className="loading-spinner" style={{ width: "40px", height: "40px", margin: "0 auto" }} />
                    <p style={{ marginTop: "16px", color: "var(--color-text-light)" }}>
                        Загрузка активов...
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

            {!loading && assets.length === 0 && (
                <div className="card" style={{ padding: "60px 20px", textAlign: "center" }}>
                    <svg
                        width="64"
                        height="64"
                        viewBox="0 0 64 64"
                        fill="none"
                        style={{ margin: "0 auto 16px" }}
                    >
                        <rect
                            width="64"
                            height="64"
                            rx="32"
                            fill="var(--color-bg)"
                        />
                        <path
                            d="M32 20v24M20 32h24"
                            stroke="var(--color-text-muted)"
                            strokeWidth="3"
                            strokeLinecap="round"
                        />
                    </svg>
                    <h3 style={{ marginBottom: "8px", color: "var(--color-text)" }}>
                        Активы отсутствуют
                    </h3>
                    <p style={{ color: "var(--color-text-light)", margin: 0 }}>
                        Добавьте первый актив для начала работы с системой
                    </p>
                </div>
            )}

            {!loading && assets.length > 0 && (
                <div style={{ display: "grid", gap: "16px" }}>
                    {assets.map((asset) => {
                        const envStyle = getEnvironmentColor(asset.environment);
                        return (
                            <div key={asset.id} className="card" style={{ padding: "20px" }}>
                                <div
                                    style={{
                                        display: "flex",
                                        justifyContent: "space-between",
                                        alignItems: "flex-start",
                                        marginBottom: "16px",
                                    }}
                                >
                                    <div style={{ flex: 1 }}>
                                        <div
                                            style={{
                                                display: "flex",
                                                alignItems: "center",
                                                gap: "12px",
                                                marginBottom: "8px",
                                            }}
                                        >
                                            <h3
                                                style={{
                                                    margin: 0,
                                                    fontSize: "18px",
                                                    color: "var(--color-primary-dark)",
                                                }}
                                            >
                                                {asset.name}
                                            </h3>
                                            <span
                                                style={{
                                                    fontSize: "12px",
                                                    color: "var(--color-text-muted)",
                                                    background: "var(--color-bg)",
                                                    padding: "2px 8px",
                                                    borderRadius: "4px",
                                                }}
                                            >
                                                ID: {asset.id}
                                            </span>
                                        </div>
                                        {asset.description && (
                                            <p
                                                style={{
                                                    margin: "0 0 12px 0",
                                                    color: "var(--color-text-light)",
                                                    fontSize: "14px",
                                                }}
                                            >
                                                {asset.description}
                                            </p>
                                        )}
                                        <div
                                            style={{
                                                display: "flex",
                                                gap: "16px",
                                                fontSize: "13px",
                                                color: "var(--color-text-light)",
                                            }}
                                        >
                                            {asset.owner && (
                                                <div>
                                                    <strong>Владелец:</strong> {asset.owner}
                                                </div>
                                            )}
                                            <div>
                                                <strong>Среда:</strong>{" "}
                                                <span
                                                    style={{
                                                        background: envStyle.bg,
                                                        color: envStyle.color,
                                                        padding: "2px 8px",
                                                        borderRadius: "4px",
                                                        fontWeight: 600,
                                                    }}
                                                >
                                                    {asset.environment?.toUpperCase() || "—"}
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                    <div style={{ display: "flex", gap: "8px", alignItems: "center", flexWrap: "wrap" }}>
                                        {getCriticalityBadge(asset.business_criticality)}
                                        <button
                                            onClick={() => navigate(`/assets/${asset.id}/risks`)}
                                            className="btn"
                                            style={{
                                                padding: "6px 12px",
                                                fontSize: "12px",
                                                background: "var(--color-info)",
                                                border: "none",
                                                color: "white",
                                            }}
                                        >
                                            <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                                                <path d="M8 2l6 12H2L8 2z" />
                                            </svg>
                                            Риски
                                        </button>
                                        <button
                                            onClick={() => navigate(`/assets/edit/${asset.id}`)}
                                            className="btn"
                                            style={{
                                                padding: "6px 12px",
                                                fontSize: "12px",
                                                background: "var(--color-surface)",
                                                border: "1px solid var(--color-border)",
                                                color: "var(--color-text)",
                                            }}
                                        >
                                            <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                                                <path d="M12.854 2.854a2 2 0 00-2.828 0L2 10.879V14h3.121l8.025-8.026a2 2 0 000-2.828l-1.292-1.292zM10 4l2 2" />
                                            </svg>
                                            Редактировать
                                        </button>
                                        <button
                                            onClick={() => handleDelete(asset)}
                                            className="btn"
                                            style={{
                                                padding: "6px 12px",
                                                fontSize: "12px",
                                                background: "rgba(231, 76, 60, 0.1)",
                                                border: "1px solid rgba(231, 76, 60, 0.2)",
                                                color: "var(--color-risk-critical)",
                                            }}
                                            disabled={deletingId === asset.id}
                                        >
                                            <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                                                <path d="M6 2h4l1 1h3v2H2V3h3l1-1zm1 5v5h2V7H7zm4 0v5h2V7h-2zM3 7v5h2V7H3z" />
                                            </svg>
                                            {deletingId === asset.id ? "Удаление..." : "Удалить"}
                                        </button>
                                    </div>
                                </div>

                                <div
                                    style={{
                                        display: "grid",
                                        gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
                                        gap: "12px",
                                        paddingTop: "16px",
                                        borderTop: "1px solid var(--color-border)",
                                    }}
                                >
                                    <MetricBox
                                        label="Критичность"
                                        value={asset.business_criticality}
                                        max={5}
                                    />
                                    <MetricBox
                                        label="Конфиденциальность"
                                        value={asset.confidentiality}
                                        max={5}
                                    />
                                    <MetricBox
                                        label="Целостность"
                                        value={asset.integrity}
                                        max={5}
                                    />
                                    <MetricBox
                                        label="Доступность"
                                        value={asset.availability}
                                        max={5}
                                    />
                                </div>
                                <div style={{ marginTop: "16px", borderTop: "1px solid var(--color-border)", paddingTop: "16px" }}>
                                    <SoftwareAlternativesPanel assetId={asset.id} />
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

const MetricBox: React.FC<{ label: string; value: number; max: number }> = ({
    label,
    value,
    max,
}) => {
    const percentage = (value / max) * 100;
    const getColor = () => {
        if (value >= 4) return "var(--color-risk-critical)";
        if (value >= 3) return "var(--color-risk-medium)";
        return "var(--color-risk-low)";
    };

    return (
        <div>
            <div
                style={{
                    display: "flex",
                    justifyContent: "space-between",
                    marginBottom: "6px",
                    fontSize: "12px",
                    color: "var(--color-text-light)",
                    fontWeight: 500,
                }}
            >
                <span>{label}</span>
                <span style={{ fontWeight: 600, color: getColor() }}>
                    {value}/{max}
                </span>
            </div>
            <div
                style={{
                    height: "6px",
                    background: "var(--color-bg)",
                    borderRadius: "3px",
                    overflow: "hidden",
                }}
            >
                <div
                    style={{
                        height: "100%",
                        width: `${percentage}%`,
                        background: getColor(),
                        transition: "width 0.5s ease",
                        borderRadius: "3px",
                    }}
                />
            </div>
        </div>
    );
};

const SoftwareAlternativesPanel: React.FC<{ assetId: number }> = ({ assetId }) => {
    const [expanded, setExpanded] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [items, setItems] = useState<AssetSoftwareAlternative[]>([]);

    const toggle = () => {
        const next = !expanded;
        setExpanded(next);
        if (next && items.length === 0 && !loading) {
            loadAlternatives();
        }
    };

    const loadAlternatives = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await api.getAssetSoftwareAlternatives(assetId);
            setItems(Array.isArray(data) ? data : []);
        } catch (err: any) {
            setError(err?.message || "Не удалось получить рекомендации");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px" }}>
                <div style={{ fontWeight: 600 }}>Российские аналоги установленного ПО</div>
                <button className="btn" onClick={toggle} disabled={loading} style={{ padding: "4px 10px", fontSize: "12px" }}>
                    {loading ? "Загрузка..." : expanded ? "Скрыть" : "Показать"}
                </button>
            </div>
            {expanded && (
                <div style={{ fontSize: "13px", color: "var(--color-text-light)" }}>
                    {loading && <div>Загрузка списка...</div>}
                    {!loading && error && <div style={{ color: "var(--color-risk-critical)" }}>⚠️ {error}</div>}
                    {!loading && !error && items.length === 0 && (
                        <div>На активе нет ПО, требующего замены на российские аналоги.</div>
                    )}
                    {!loading && !error && items.length > 0 && (
                        <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
                            {items.map((item) => (
                                <div
                                    key={item.asset_software_id}
                                    style={{
                                        padding: "12px",
                                        border: "1px solid var(--color-border)",
                                        borderRadius: "var(--radius-sm)",
                                        background: "var(--color-bg)",
                                    }}
                                >
                                    <div style={{ fontWeight: 600, color: "var(--color-text)" }}>
                                        {item.software.name} — {item.software.vendor}
                                    </div>
                                    {item.version && (
                                        <div style={{ fontSize: "12px", marginTop: "2px" }}>Версия на активе: {item.version}</div>
                                    )}
                                    <div style={{ marginTop: "6px", color: "var(--color-text)" }}>
                                        <span style={{ fontWeight: 600 }}>Рекомендуемые аналоги:</span>
                                        <ul style={{ margin: "4px 0 0 18px" }}>
                                            {item.alternatives.slice(0, 5).map((alt) => (
                                                <li key={alt.id}>
                                                    <strong>{alt.name}</strong> — {alt.vendor}
                                                    {alt.fstec_certified && (
                                                        <span style={{ marginLeft: "6px", fontSize: "11px", fontWeight: 600, color: "#e67e22" }}>
                                                            ФСТЭК
                                                        </span>
                                                    )}
                                                    {alt.fsb_certified && (
                                                        <span style={{ marginLeft: "4px", fontSize: "11px", fontWeight: 600, color: "#c0392b" }}>
                                                            ФСБ
                                                        </span>
                                                    )}
                                                    {alt.registry_number && (
                                                        <span style={{ marginLeft: "4px", fontSize: "11px", fontWeight: 600, color: "#27ae60" }}>
                                                            Реестр №{alt.registry_number}
                                                        </span>
                                                    )}
                                                </li>
                                            ))}
                                        </ul>
                                        {item.alternatives.length > 5 && (
                                            <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
                                                И ещё {item.alternatives.length - 5} вариантов...
                                            </div>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};
