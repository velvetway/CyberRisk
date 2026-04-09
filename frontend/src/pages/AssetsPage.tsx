import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { Asset, AssetSoftwareAlternative } from "../types";
import "./AssetsPage.css";

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

    const envClass = (env: string) => {
        switch (env?.toLowerCase()) {
            case "prod": return "asset-card__env--prod";
            case "test": return "asset-card__env--test";
            case "dev":  return "asset-card__env--dev";
            default:     return "asset-card__env--default";
        }
    };

    const critBadge = (c: number) => {
        if (c >= 4) return <span className="badge badge-critical">Высокая</span>;
        if (c >= 3) return <span className="badge badge-high">Средняя</span>;
        return <span className="badge badge-low">Низкая</span>;
    };

    const handleDelete = async (asset: Asset) => {
        if (!window.confirm(`Удалить актив «${asset.name}»?`)) return;
        try {
            setDeletingId(asset.id);
            await api.deleteAsset(asset.id);
            setAssets((prev) => prev.filter((a) => a.id !== asset.id));
            setError(null);
        } catch (err: any) {
            setError(err.message || "Не удалось удалить");
        } finally {
            setDeletingId(null);
        }
    };

    return (
        <div className="fade-in">
            <div className="assets-header">
                <div className="assets-header__info">
                    <h1>ИТ-активы организации</h1>
                    <p className="assets-header__desc">
                        Управление и мониторинг критически важных активов
                    </p>
                </div>
                <button onClick={() => navigate("/assets/new")} className="btn btn-primary">
                    + Создать актив
                </button>
            </div>

            {loading && (
                <div className="assets-loading">
                    <div className="loading-spinner" style={{ width: 28, height: 28, margin: "0 auto" }} />
                    <p>Загрузка активов...</p>
                </div>
            )}

            {error && <div className="assets-error">{error}</div>}

            {!loading && assets.length === 0 && (
                <div className="card assets-empty">
                    <h3>Активы отсутствуют</h3>
                    <p>Добавьте первый актив для начала работы</p>
                </div>
            )}

            {!loading && assets.length > 0 && (
                <>
                <div className="assets-summary">
                    <div className="summary-stat">
                        <span className="summary-stat__value">{assets.length}</span>
                        <span className="summary-stat__label">Всего активов</span>
                    </div>
                    <div className="summary-stat">
                        <span className="summary-stat__value summary-stat__value--critical">{assets.filter(a => a.business_criticality >= 4).length}</span>
                        <span className="summary-stat__label">Критичных</span>
                    </div>
                    <div className="summary-stat">
                        <span className="summary-stat__value summary-stat__value--prod">{assets.filter(a => a.environment === "prod").length}</span>
                        <span className="summary-stat__label">В продакшене</span>
                    </div>
                    <div className="summary-stat">
                        <span className="summary-stat__value">{(assets.reduce((s, a) => s + a.business_criticality, 0) / assets.length).toFixed(1)}</span>
                        <span className="summary-stat__label">Средняя критичность</span>
                    </div>
                </div>
                <div className="assets-grid">
                    {assets.map((asset) => {
                        const critClass = asset.business_criticality >= 4 ? "asset-card--critical" : asset.business_criticality >= 3 ? "asset-card--medium" : "asset-card--low";
                        return (
                        <div key={asset.id} className={`card asset-card ${critClass}`}>
                            <div className="asset-card__top">
                                <div className="asset-card__info">
                                    <div className="asset-card__name-row">
                                        <h3 className="asset-card__name">{asset.name}</h3>
                                        <span className="asset-card__id">#{asset.id}</span>
                                    </div>
                                    {asset.description && (
                                        <p className="asset-card__desc">{asset.description}</p>
                                    )}
                                    <div className="asset-card__meta">
                                        {asset.owner && <span><strong>Владелец:</strong> {asset.owner}</span>}
                                        <span>
                                            <strong>Среда:</strong>{" "}
                                            <span className={`asset-card__env ${envClass(asset.environment)}`}>
                                                {asset.environment?.toUpperCase() || "—"}
                                            </span>
                                        </span>
                                    </div>
                                </div>
                                <div className="asset-card__actions">
                                    {critBadge(asset.business_criticality)}
                                    <button onClick={() => navigate(`/assets/${asset.id}/risks`)} className="btn">
                                        Риски
                                    </button>
                                    <button onClick={() => navigate(`/assets/edit/${asset.id}`)} className="btn">
                                        Изменить
                                    </button>
                                    <button
                                        onClick={() => handleDelete(asset)}
                                        className="btn btn-danger"
                                        disabled={deletingId === asset.id}
                                    >
                                        {deletingId === asset.id ? "..." : "Удалить"}
                                    </button>
                                </div>
                            </div>

                            <div className="asset-card__metrics">
                                <Metric label="Критичность" value={asset.business_criticality} max={5} />
                                <Metric label="Конфиденц." value={asset.confidentiality} max={5} />
                                <Metric label="Целостность" value={asset.integrity} max={5} />
                                <Metric label="Доступность" value={asset.availability} max={5} />
                            </div>

                            <AltsPanel assetId={asset.id} />
                        </div>
                    );
                    })}
                </div>
                </>
            )}
        </div>
    );
};

const Metric: React.FC<{ label: string; value: number; max: number }> = ({ label, value, max }) => {
    const pct = (value / max) * 100;
    const color = value >= 4 ? "var(--threat-critical)" : value >= 3 ? "var(--threat-medium)" : "var(--threat-low)";

    return (
        <div className="metric">
            <div className="metric__header">
                <span className="metric__label">{label}</span>
                <span className="metric__value" style={{ color }}>{value}/{max}</span>
            </div>
            <div className="metric__bar">
                <div className="metric__fill" style={{ width: `${pct}%`, background: color }} />
            </div>
        </div>
    );
};

const AltsPanel: React.FC<{ assetId: number }> = ({ assetId }) => {
    const [expanded, setExpanded] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [items, setItems] = useState<AssetSoftwareAlternative[]>([]);

    const toggle = () => {
        const next = !expanded;
        setExpanded(next);
        if (next && items.length === 0 && !loading) {
            setLoading(true);
            setError(null);
            api.getAssetSoftwareAlternatives(assetId)
                .then((d) => setItems(Array.isArray(d) ? d : []))
                .catch((e) => setError(e?.message || "Ошибка"))
                .finally(() => setLoading(false));
        }
    };

    return (
        <div className="alts-panel">
            <div className="alts-panel__header">
                <span className="alts-panel__title">Российские аналоги ПО</span>
                <button className="btn" onClick={toggle} disabled={loading} style={{ padding: "3px 10px", fontSize: 11 }}>
                    {loading ? "..." : expanded ? "Скрыть" : "Показать"}
                </button>
            </div>
            {expanded && (
                <div className="alts-panel__content">
                    {loading && <span>Загрузка...</span>}
                    {!loading && error && <span style={{ color: "var(--danger)" }}>{error}</span>}
                    {!loading && !error && items.length === 0 && (
                        <span>Нет ПО, требующего замены.</span>
                    )}
                    {!loading && !error && items.map((item) => (
                        <div key={item.asset_software_id} className="alt-item">
                            <div className="alt-item__name">{item.software.name} — {item.software.vendor}</div>
                            {item.version && <div className="alt-item__version">v{item.version}</div>}
                            <ul className="alt-item__list">
                                {item.alternatives.slice(0, 5).map((alt) => (
                                    <li key={alt.id}>
                                        <strong>{alt.name}</strong> — {alt.vendor}
                                        {alt.fstec_certified && <span className="alt-item__cert alt-item__cert--fstec">ФСТЭК</span>}
                                        {alt.fsb_certified && <span className="alt-item__cert alt-item__cert--fsb">ФСБ</span>}
                                        {alt.registry_number && <span className="alt-item__cert alt-item__cert--reg">№{alt.registry_number}</span>}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
