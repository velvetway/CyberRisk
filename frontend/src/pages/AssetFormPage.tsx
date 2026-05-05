import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Save, X, Server, Network } from "lucide-react";

import { authFetch } from "../api/client";

// Asset form for the PTSZI W-model. The W formula only consumes is_isolated
// (drives Z) and the asset's deployed controls (drives Q^reaction); the rest
// of the fields are descriptive metadata for UI / reports.
interface AssetFormData {
    name: string;
    description: string;
    asset_type_id: string;  // SELECT value, parsed to int on submit
    owner: string;
    environment: string;
    is_isolated: boolean;
}

const initialFormData: AssetFormData = {
    name: "",
    description: "",
    asset_type_id: "",
    owner: "",
    environment: "prod",
    is_isolated: false,
};

interface AssetTypeRef {
    id: number;
    name: string;
    description?: string;
}

export const AssetFormPage: React.FC = () => {
    const { id } = useParams<{ id?: string }>();
    const navigate = useNavigate();
    const isEditMode = !!id;

    const [formData, setFormData] = useState<AssetFormData>(initialFormData);
    const [assetTypes, setAssetTypes] = useState<AssetTypeRef[]>([]);
    const [loading, setLoading] = useState(false);
    const [loadingData, setLoadingData] = useState(isEditMode);

    useEffect(() => {
        // Best-effort load of the asset_types reference list. The endpoint may
        // not be wired in every deployment — fall back to a static list so the
        // form remains usable.
        (async () => {
            try {
                const res = await authFetch("/api/asset-types");
                if (res.ok) {
                    const list = await res.json();
                    if (Array.isArray(list) && list.length > 0) {
                        setAssetTypes(list);
                        return;
                    }
                }
            } catch {/* ignore */}
            setAssetTypes([
                { id: 1, name: "Server" },
                { id: 2, name: "Database" },
                { id: 3, name: "Application" },
                { id: 4, name: "Network" },
                { id: 5, name: "Workstation" },
                { id: 6, name: "Mobile" },
                { id: 7, name: "IoT" },
                { id: 8, name: "Cloud" },
            ]);
        })();
    }, []);

    useEffect(() => {
        if (isEditMode && id) fetchAsset(id);
    }, [id, isEditMode]);

    const fetchAsset = async (assetId: string) => {
        setLoadingData(true);
        try {
            const res = await authFetch(`/api/assets/${assetId}`);
            if (!res.ok) throw new Error(`Ошибка загрузки актива: ${res.status}`);
            const asset = await res.json();
            setFormData({
                name: asset.name || "",
                description: asset.description || "",
                asset_type_id: asset.asset_type_id != null ? String(asset.asset_type_id) : "",
                owner: asset.owner || "",
                environment: asset.environment || "prod",
                is_isolated: !!asset.is_isolated,
            });
        } catch (e: any) {
            toast.error(e.message || "Ошибка загрузки данных актива");
        } finally {
            setLoadingData(false);
        }
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
        const { name, value, type } = e.target;
        const checked = (e.target as HTMLInputElement).checked;
        setFormData(prev => ({ ...prev, [name]: type === "checkbox" ? checked : value }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!formData.name.trim()) {
            toast.error("Название актива обязательно для заполнения");
            return;
        }

        setLoading(true);
        try {
            const url = isEditMode ? `/api/assets/${id}` : "/api/assets";
            const method = isEditMode ? "PUT" : "POST";

            const payload = {
                name: formData.name,
                description: formData.description || undefined,
                asset_type_id: formData.asset_type_id ? parseInt(formData.asset_type_id, 10) : undefined,
                owner: formData.owner || undefined,
                environment: formData.environment,
                is_isolated: formData.is_isolated,
            };

            const res = await authFetch(url, { method, body: JSON.stringify(payload) });
            if (!res.ok) {
                const body = await res.json().catch(() => null);
                throw new Error(body?.error || `Ошибка ${isEditMode ? "обновления" : "создания"}`);
            }
            toast.success(isEditMode ? "Актив обновлён" : "Актив создан");
            navigate("/assets");
        } catch (e: any) {
            toast.error(e.message || "Ошибка при сохранении");
        } finally {
            setLoading(false);
        }
    };

    if (loadingData) {
        return (
            <div style={{ textAlign: "center", padding: "60px 20px" }}>
                <div className="loading-spinner" style={{ margin: "0 auto 16px" }} />
                <p style={{ color: "var(--ink-muted)" }}>Загрузка данных актива...</p>
            </div>
        );
    }

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
            <div style={{ marginBottom: 32 }}>
                <h1>{isEditMode ? "Редактировать актив" : "Новый актив"}</h1>
                <p style={{ color: "var(--ink-muted)" }}>
                    Только поля, которые имеют смысл для модели ПТСЗИ. Степень покрытия (Q<sup>reaction</sup>)
                    рассчитывается из внедрённых на активе контролей в отдельном экране.
                </p>
            </div>

            <div className="card" style={{ padding: 32, maxWidth: 760 }}>
                <form onSubmit={handleSubmit}>
                    <div style={{ marginBottom: 32 }}>
                        <h3 style={{ fontSize: 16, marginBottom: 20, paddingBottom: 10, borderBottom: "1px solid var(--perimeter)", display: "flex", alignItems: "center", gap: 8 }}>
                            <Server size={18} color="var(--command)" /> Основная информация
                        </h3>

                        <div style={{ marginBottom: 20 }}>
                            <label className="form-label">Название актива <span style={{ color: "var(--danger)" }}>*</span></label>
                            <input type="text" name="name" value={formData.name} onChange={handleChange} className="form-input" placeholder="Например: CRM Database Server" required />
                        </div>

                        <div style={{ marginBottom: 20 }}>
                            <label className="form-label">Описание</label>
                            <textarea name="description" value={formData.description} onChange={handleChange} className="form-input" rows={3} placeholder="Краткое описание актива" style={{ resize: "vertical" }} />
                        </div>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginBottom: 20 }}>
                            <div>
                                <label className="form-label">Тип актива</label>
                                <select name="asset_type_id" value={formData.asset_type_id} onChange={handleChange} className="form-input">
                                    <option value="">— не указан —</option>
                                    {assetTypes.map(t => (
                                        <option key={t.id} value={String(t.id)}>{t.name}</option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="form-label">Среда</label>
                                <select name="environment" value={formData.environment} onChange={handleChange} className="form-input">
                                    <option value="prod">Production</option>
                                    <option value="test">Test</option>
                                    <option value="dev">Development</option>
                                    <option value="other">Other</option>
                                </select>
                            </div>
                        </div>

                        <div>
                            <label className="form-label">Владелец</label>
                            <input type="text" name="owner" value={formData.owner} onChange={handleChange} className="form-input" placeholder="IT Department" />
                        </div>
                    </div>

                    <div style={{ marginBottom: 32 }}>
                        <h3 style={{ fontSize: 16, marginBottom: 20, paddingBottom: 10, borderBottom: "1px solid var(--perimeter)", display: "flex", alignItems: "center", gap: 8 }}>
                            <Network size={18} color="var(--success)" /> Контур и формула W
                        </h3>

                        <div style={{ padding: 16, background: formData.is_isolated ? "var(--threat-low-dim)" : "var(--well)", borderRadius: "var(--r-md)", border: `1px solid ${formData.is_isolated ? "var(--success)" : "var(--perimeter)"}` }}>
                            <label style={{ display: "flex", gap: 12, cursor: "pointer", alignItems: "flex-start" }}>
                                <input type="checkbox" name="is_isolated" checked={formData.is_isolated} onChange={handleChange} style={{ width: 16, height: 16, marginTop: 4 }} />
                                <div>
                                    <div style={{ fontWeight: 600 }}>Изолированный сегмент</div>
                                    <div style={{ fontSize: 12, color: "var(--ink-muted)", marginTop: 2 }}>
                                        Если актив виден только в одном контуре — Z = 0.5 (формула W ослабляется вдвое).
                                        Иначе Z = 1.0.
                                    </div>
                                </div>
                            </label>
                        </div>
                    </div>

                    <div style={{ display: "flex", gap: 12, justifyContent: "flex-end", paddingTop: 20, borderTop: "1px solid var(--perimeter)" }}>
                        <button type="button" onClick={() => navigate("/assets")} className="btn">
                            <X size={16} /> Отмена
                        </button>
                        <button type="submit" className="btn btn-primary" disabled={loading}>
                            {loading ? <span className="loading-spinner" /> : <><Save size={16} /> {isEditMode ? "Сохранить" : "Создать"}</>}
                        </button>
                    </div>
                </form>
            </div>
        </motion.div>
    );
};
