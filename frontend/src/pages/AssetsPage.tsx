import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { Asset, AssetSoftwareAlternative } from "../types";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Plus, Server, Database, AppWindow, Globe, Laptop, Smartphone, Cpu, Cloud, Trash2, Edit2, ShieldAlert, ChevronDown, ChevronUp } from "lucide-react";
import "./AssetsPage.css";

export const AssetsPage: React.FC = () => {
    const navigate = useNavigate();
    const [assets, setAssets] = useState<Asset[]>([]);
    const [loading, setLoading] = useState(true);
    const [deletingId, setDeletingId] = useState<number | null>(null);

    useEffect(() => {
        api.getAssets()
            .then(setAssets)
            .catch((err) => toast.error(err.message))
            .finally(() => setLoading(false));
    }, []);

    const envBadge = (env: string) => {
        switch (env?.toLowerCase()) {
            case "prod": return <span className="badge" style={{ background: '#dcfce7', color: '#059669' }}>PROD</span>;
            case "test": return <span className="badge" style={{ background: '#fef9c3', color: '#d97706' }}>TEST</span>;
            case "dev":  return <span className="badge" style={{ background: '#e0e7ff', color: '#2563eb' }}>DEV</span>;
            default:     return <span className="badge" style={{ background: 'var(--well)', color: 'var(--ink-muted)' }}>{env || '—'}</span>;
        }
    };

    const critBadge = (c: number) => {
        if (c >= 4) return <span className="badge badge-critical">Критичная</span>;
        if (c >= 3) return <span className="badge badge-high">Высокая</span>;
        return <span className="badge badge-low">Низкая</span>;
    };

    const getIcon = (type: string) => {
        switch (type) {
            case 'server': return <Server size={20} />;
            case 'database': return <Database size={20} />;
            case 'application': return <AppWindow size={20} />;
            case 'network': return <Globe size={20} />;
            case 'workstation': return <Laptop size={20} />;
            case 'mobile': return <Smartphone size={20} />;
            case 'iot': return <Cpu size={20} />;
            case 'cloud': return <Cloud size={20} />;
            default: return <Server size={20} />;
        }
    }

    const handleDelete = async (asset: Asset) => {
        if (!window.confirm(`Удалить актив «${asset.name}»?`)) return;
        try {
            setDeletingId(asset.id);
            await api.deleteAsset(asset.id);
            setAssets((prev) => prev.filter((a) => a.id !== asset.id));
            toast.success("Актив удален");
        } catch (err: any) {
            toast.error(err.message || "Не удалось удалить");
        } finally {
            setDeletingId(null);
        }
    };

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }}>
            <div className="assets-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '30px' }}>
                <div>
                    <h1>ИТ-активы организации</h1>
                    <p style={{ color: 'var(--ink-muted)' }}>
                        Управление и мониторинг критически важных активов
                    </p>
                </div>
                <button onClick={() => navigate("/assets/new")} className="btn btn-primary">
                    <Plus size={18} /> Создать актив
                </button>
            </div>

            {loading && (
                <div style={{ textAlign: 'center', padding: '60px' }}>
                    <div className="loading-spinner" style={{ width: 28, height: 28, margin: "0 auto" }} />
                    <p style={{ color: 'var(--ink-muted)', marginTop: '10px' }}>Загрузка активов...</p>
                </div>
            )}

            {!loading && assets.length === 0 && (
                <div className="card" style={{ padding: '60px', textAlign: 'center' }}>
                    <Server size={48} color="var(--perimeter-emphasis)" style={{ margin: '0 auto 16px' }} />
                    <h3 style={{ fontSize: '18px' }}>Активы отсутствуют</h3>
                    <p style={{ color: 'var(--ink-muted)' }}>Добавьте первый актив для начала работы</p>
                    <button onClick={() => navigate("/assets/new")} className="btn btn-primary" style={{ marginTop: '16px' }}>
                        Создать актив
                    </button>
                </div>
            )}

            {!loading && assets.length > 0 && (
                <>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px', marginBottom: '24px' }}>
                    <div className="card" style={{ padding: '20px' }}>
                        <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--ink)' }}>{assets.length}</div>
                        <div style={{ fontSize: '13px', color: 'var(--ink-muted)', fontWeight: 500 }}>Всего активов</div>
                    </div>
                    <div className="card" style={{ padding: '20px' }}>
                        <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--danger)' }}>{assets.filter(a => a.business_criticality >= 4).length}</div>
                        <div style={{ fontSize: '13px', color: 'var(--ink-muted)', fontWeight: 500 }}>Критичных</div>
                    </div>
                    <div className="card" style={{ padding: '20px' }}>
                        <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--success)' }}>{assets.filter(a => a.environment === "prod").length}</div>
                        <div style={{ fontSize: '13px', color: 'var(--ink-muted)', fontWeight: 500 }}>В продакшене</div>
                    </div>
                    <div className="card" style={{ padding: '20px' }}>
                        <div style={{ fontSize: '28px', fontWeight: 700, color: 'var(--command)' }}>{(assets.reduce((s, a) => s + a.business_criticality, 0) / assets.length).toFixed(1)}</div>
                        <div style={{ fontSize: '13px', color: 'var(--ink-muted)', fontWeight: 500 }}>Средняя критичность</div>
                    </div>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {assets.map((asset, index) => (
                        <motion.div 
                            key={asset.id} 
                            className="card"
                            initial={{ opacity: 0, y: 10 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ duration: 0.3, delay: index * 0.05 }}
                            style={{ overflow: 'hidden' }}
                        >
                            <div style={{ padding: '20px', display: 'flex', alignItems: 'center', gap: '20px' }}>
                                <div style={{ 
                                    width: '48px', height: '48px', 
                                    background: 'var(--well)', 
                                    borderRadius: 'var(--r-md)',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    color: 'var(--ink-secondary)'
                                }}>
                                    {getIcon(asset.type || '')}
                                </div>
                                <div style={{ flex: 1 }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '4px' }}>
                                        <h3 style={{ margin: 0, fontSize: '16px' }}>{asset.name}</h3>
                                        <span style={{ fontSize: '12px', color: 'var(--ink-muted)', fontFamily: 'var(--font-mono)' }}>#{asset.id}</span>
                                        {envBadge(asset.environment)}
                                        {critBadge(asset.business_criticality)}
                                    </div>
                                    <div style={{ fontSize: '13px', color: 'var(--ink-secondary)' }}>
                                        {asset.description || 'Нет описания'}
                                    </div>
                                    <div style={{ display: 'flex', gap: '16px', marginTop: '8px', fontSize: '12px', color: 'var(--ink-muted)' }}>
                                        {asset.owner && <span><strong>Владелец:</strong> {asset.owner}</span>}
                                        {asset.location && <span><strong>Локация:</strong> {asset.location}</span>}
                                    </div>
                                </div>

                                <div style={{ display: 'flex', gap: '16px', alignItems: 'center', padding: '0 20px', borderLeft: '1px solid var(--perimeter)', borderRight: '1px solid var(--perimeter)' }}>
                                    <div style={{ textAlign: 'center' }}>
                                        <div style={{ fontSize: '11px', color: 'var(--ink-muted)', fontWeight: 600 }}>C</div>
                                        <div style={{ fontWeight: 700, color: asset.confidentiality >= 4 ? 'var(--danger)' : 'var(--ink)' }}>{asset.confidentiality}</div>
                                    </div>
                                    <div style={{ textAlign: 'center' }}>
                                        <div style={{ fontSize: '11px', color: 'var(--ink-muted)', fontWeight: 600 }}>I</div>
                                        <div style={{ fontWeight: 700, color: asset.integrity >= 4 ? 'var(--danger)' : 'var(--ink)' }}>{asset.integrity}</div>
                                    </div>
                                    <div style={{ textAlign: 'center' }}>
                                        <div style={{ fontSize: '11px', color: 'var(--ink-muted)', fontWeight: 600 }}>A</div>
                                        <div style={{ fontWeight: 700, color: asset.availability >= 4 ? 'var(--danger)' : 'var(--ink)' }}>{asset.availability}</div>
                                    </div>
                                </div>

                                <div style={{ display: 'flex', gap: '8px' }}>
                                    <button onClick={() => navigate(`/assets/${asset.id}/risks`)} className="btn" title="Риски">
                                        <ShieldAlert size={16} />
                                    </button>
                                    <button onClick={() => navigate(`/assets/edit/${asset.id}`)} className="btn" title="Изменить">
                                        <Edit2 size={16} />
                                    </button>
                                    <button
                                        onClick={() => handleDelete(asset)}
                                        className="btn btn-danger"
                                        disabled={deletingId === asset.id}
                                        title="Удалить"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            </div>
                            <AltsPanel assetId={asset.id} />
                        </motion.div>
                    ))}
                </div>
                </>
            )}
        </motion.div>
    );
};

const AltsPanel: React.FC<{ assetId: number }> = ({ assetId }) => {
    const [expanded, setExpanded] = useState(false);
    const [loading, setLoading] = useState(false);
    const [items, setItems] = useState<AssetSoftwareAlternative[]>([]);

    const toggle = () => {
        const next = !expanded;
        setExpanded(next);
        if (next && items.length === 0 && !loading) {
            setLoading(true);
            api.getAssetSoftwareAlternatives(assetId)
                .then((d) => setItems(Array.isArray(d) ? d : []))
                .catch((e) => toast.error(e?.message || "Ошибка загрузки аналогов"))
                .finally(() => setLoading(false));
        }
    };

    return (
        <div style={{ background: 'var(--well)', borderTop: '1px solid var(--perimeter)' }}>
            <button 
                onClick={toggle}
                style={{ 
                    width: '100%', padding: '10px 20px', 
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                    background: 'none', border: 'none', cursor: 'pointer',
                    fontSize: '13px', fontWeight: 600, color: 'var(--ink-secondary)'
                }}
            >
                <span>Импортозамещение (Российские аналоги ПО)</span>
                {expanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            </button>
            {expanded && (
                <div style={{ padding: '0 20px 20px' }}>
                    {loading && <span style={{ fontSize: '13px', color: 'var(--ink-muted)' }}>Загрузка...</span>}
                    {!loading && items.length === 0 && (
                        <span style={{ fontSize: '13px', color: 'var(--ink-muted)' }}>Нет ПО, требующего замены или аналоги не найдены.</span>
                    )}
                    {!loading && items.map((item) => (
                        <div key={item.asset_software_id} style={{ marginTop: '12px', background: 'var(--raised)', padding: '12px', borderRadius: 'var(--r-sm)', border: '1px solid var(--perimeter)' }}>
                            <div style={{ fontWeight: 600, fontSize: '14px', marginBottom: '8px' }}>
                                {item.software.name} <span style={{ color: 'var(--ink-muted)', fontWeight: 400 }}>— {item.software.vendor}</span>
                                {item.version && <span style={{ marginLeft: '8px', color: 'var(--ink-muted)', fontWeight: 400, fontSize: '12px' }}>v{item.version}</span>}
                            </div>
                            <ul style={{ margin: 0, paddingLeft: '20px', fontSize: '13px' }}>
                                {item.alternatives.slice(0, 5).map((alt) => (
                                    <li key={alt.id} style={{ marginBottom: '4px' }}>
                                        <strong>{alt.name}</strong> ({alt.vendor})
                                        {alt.fstec_certified && <span className="badge badge-medium" style={{ marginLeft: '8px', padding: '2px 4px', fontSize: '10px' }}>ФСТЭК</span>}
                                        {alt.fsb_certified && <span className="badge badge-high" style={{ marginLeft: '6px', padding: '2px 4px', fontSize: '10px' }}>ФСБ</span>}
                                        {alt.registry_number && <span className="badge badge-low" style={{ marginLeft: '6px', padding: '2px 4px', fontSize: '10px' }}>№{alt.registry_number}</span>}
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
