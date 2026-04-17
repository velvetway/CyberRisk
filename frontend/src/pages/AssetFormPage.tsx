import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { authFetch } from "../api/client";
import { DataCategory, ProtectionLevel, KIICategory } from "../types";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Save, X, Info, ShieldAlert, Server } from "lucide-react";

interface AssetFormData {
    name: string;
    description: string;
    type: string;
    location: string;
    owner: string;
    environment: string;
    data_category: DataCategory;
    protection_level: ProtectionLevel | "";
    kii_category: KIICategory;
    has_personal_data: boolean;
    personal_data_volume: string;
    has_internet_access: boolean;
    is_isolated: boolean;
}

const initialFormData: AssetFormData = {
    name: "",
    description: "",
    type: "server",
    location: "",
    owner: "",
    environment: "prod",
    data_category: "internal",
    protection_level: "",
    kii_category: "none",
    has_personal_data: false,
    personal_data_volume: "",
    has_internet_access: true,
    is_isolated: false,
};

const calculateCriticality = (form: AssetFormData): number => {
    let score = 2.0;

    switch (form.data_category) {
        case "state_secret": score = 5.0; break;
        case "personal_data_special":
        case "personal_data_biometric": score = 4.5; break;
        case "kii": score = 5.0; break;
        case "personal_data": score = 3.5; break;
        case "confidential":
        case "banking_secret":
        case "medical_secret":
        case "commercial_secret": score = 3.5; break;
        case "internal": score = 2.5; break;
        case "public": score = 1.5; break;
    }

    switch (form.kii_category) {
        case "cat1": score = Math.max(score, 5.0); break;
        case "cat2": score = Math.max(score, 4.0); break;
        case "cat3": score = Math.max(score, 3.5); break;
    }

    switch (form.protection_level) {
        case "uz1": score = Math.max(score, 4.5); break;
        case "uz2": score = Math.max(score, 4.0); break;
        case "uz3": score = Math.max(score, 3.0); break;
        case "uz4": score = Math.max(score, 2.5); break;
    }

    if (form.has_personal_data && form.personal_data_volume) {
        if (form.personal_data_volume === ">100000") score += 0.5;
        else if (form.personal_data_volume === "1000-100000") score += 0.25;
    }

    switch (form.environment) {
        case "prod": score += 0.5; break;
        case "test": score -= 0.5; break;
        case "dev": score -= 1.0; break;
    }

    if (form.has_internet_access) score += 0.5;
    if (form.is_isolated) score -= 0.5;

    return Math.min(5, Math.max(1, Math.round(score)));
};

const dataCategoryLabels: Record<DataCategory, string> = {
    public: "Общедоступные",
    internal: "Внутренние",
    confidential: "Конфиденциальные",
    personal_data: "Персональные данные (152-ФЗ)",
    personal_data_special: "Специальные ПДн",
    personal_data_biometric: "Биометрические ПДн",
    kii: "КИИ (187-ФЗ)",
    state_secret: "Государственная тайна",
    banking_secret: "Банковская тайна",
    medical_secret: "Врачебная тайна",
    commercial_secret: "Коммерческая тайна",
};

const protectionLevelLabels: Record<ProtectionLevel, string> = {
    uz1: "УЗ-1 (максимальный)",
    uz2: "УЗ-2",
    uz3: "УЗ-3",
    uz4: "УЗ-4 (минимальный)",
};

const kiiCategoryLabels: Record<KIICategory, string> = {
    none: "Не является КИИ",
    cat3: "3 категория",
    cat2: "2 категория",
    cat1: "1 категория (максимальная)",
};

export const AssetFormPage: React.FC = () => {
    const { id } = useParams<{ id?: string }>();
    const navigate = useNavigate();
    const isEditMode = !!id;

    const [formData, setFormData] = useState<AssetFormData>(initialFormData);
    const [loading, setLoading] = useState(false);
    const [loadingData, setLoadingData] = useState(isEditMode);

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
                type: asset.type || "server",
                location: asset.location || "",
                owner: asset.owner || "",
                environment: asset.environment || "prod",
                data_category: asset.data_category || "internal",
                protection_level: asset.protection_level || "",
                kii_category: asset.kii_category || "none",
                has_personal_data: asset.has_personal_data || false,
                personal_data_volume: asset.personal_data_volume || "",
                has_internet_access: asset.has_internet_access ?? true,
                is_isolated: asset.is_isolated || false,
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
        setFormData((prev) => ({ ...prev, [name]: type === "checkbox" ? checked : value }));
    };

    const calculatedCriticality = calculateCriticality(formData);

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
                type: formData.type,
                environment: formData.environment,
                location: formData.location || undefined,
                owner: formData.owner || undefined,
                business_criticality: calculatedCriticality,
                data_category: formData.data_category,
                protection_level: formData.protection_level || undefined,
                kii_category: formData.kii_category,
                has_personal_data: formData.has_personal_data,
                personal_data_volume: formData.personal_data_volume || undefined,
                has_internet_access: formData.has_internet_access,
                is_isolated: formData.is_isolated,
            };

            const res = await authFetch(url, {
                method,
                body: JSON.stringify(payload),
            });

            if (!res.ok) throw new Error(`Ошибка ${isEditMode ? "обновления" : "создания"}`);
            toast.success(isEditMode ? "Актив обновлен" : "Актив создан");
            navigate("/assets");
        } catch (e: any) {
            toast.error(e.message || `Ошибка при сохранении`);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (formData.data_category === "personal_data" || formData.data_category === "personal_data_special" || formData.data_category === "personal_data_biometric") {
            setFormData((prev) => ({ ...prev, has_personal_data: true }));
        }
    }, [formData.data_category]);

    if (loadingData) {
        return (
            <div style={{ textAlign: "center", padding: "60px 20px" }}>
                <div className="loading-spinner" style={{ margin: "0 auto 16px" }} />
                <p style={{ color: "var(--ink-muted)" }}>Загрузка данных актива...</p>
            </div>
        );
    }

    const getMetricColor = (value: number) => {
        if (value >= 4) return "var(--danger)";
        if (value === 3) return "var(--warning)";
        return "var(--success)";
    };

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }}>
            <div style={{ marginBottom: "32px" }}>
                <h1>{isEditMode ? "Редактировать актив" : "Новый актив"}</h1>
                <p style={{ color: "var(--ink-muted)" }}>
                    {isEditMode ? "Обновите информацию об ИТ-активе" : "Добавьте новый ИТ-актив в систему"}
                </p>
            </div>

            <div className="card" style={{ padding: "32px", maxWidth: "900px" }}>
                <form onSubmit={handleSubmit}>
                    {/* Основная информация */}
                    <div style={{ marginBottom: "32px" }}>
                        <h3 style={{ fontSize: "16px", marginBottom: "20px", paddingBottom: "10px", borderBottom: "1px solid var(--perimeter)", display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <Server size={18} color="var(--command)" /> Основная информация
                        </h3>

                        <div style={{ marginBottom: "20px" }}>
                            <label className="form-label">Название актива <span style={{ color: "var(--danger)" }}>*</span></label>
                            <input type="text" name="name" value={formData.name} onChange={handleChange} className="form-input" placeholder="Например: CRM Database Server" required />
                        </div>

                        <div style={{ marginBottom: "20px" }}>
                            <label className="form-label">Описание</label>
                            <textarea name="description" value={formData.description} onChange={handleChange} className="form-input" rows={3} placeholder="Краткое описание" style={{ resize: "vertical" }} />
                        </div>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px", marginBottom: "20px" }}>
                            <div>
                                <label className="form-label">Тип актива</label>
                                <select name="type" value={formData.type} onChange={handleChange} className="form-input">
                                    <option value="server">Сервер</option>
                                    <option value="database">База данных</option>
                                    <option value="application">Приложение</option>
                                    <option value="network">Сеть</option>
                                    <option value="workstation">Рабочая станция</option>
                                    <option value="mobile">Мобильное устройство</option>
                                    <option value="iot">IoT</option>
                                    <option value="cloud">Облако</option>
                                </select>
                            </div>
                            <div>
                                <label className="form-label">Среда</label>
                                <select name="environment" value={formData.environment} onChange={handleChange} className="form-input">
                                    <option value="prod">Production</option>
                                    <option value="stage">Staging</option>
                                    <option value="test">Test</option>
                                    <option value="dev">Development</option>
                                </select>
                            </div>
                        </div>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px" }}>
                            <div>
                                <label className="form-label">Расположение</label>
                                <input type="text" name="location" value={formData.location} onChange={handleChange} className="form-input" placeholder="Datacenter A, Rack 12" />
                            </div>
                            <div>
                                <label className="form-label">Владелец</label>
                                <input type="text" name="owner" value={formData.owner} onChange={handleChange} className="form-input" placeholder="IT Department" />
                            </div>
                        </div>
                    </div>

                    {/* Регуляторные требования */}
                    <div style={{ marginBottom: "32px" }}>
                        <h3 style={{ fontSize: "16px", marginBottom: "20px", paddingBottom: "10px", borderBottom: "1px solid var(--perimeter)", display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <ShieldAlert size={18} color="var(--warning)" /> Регуляторные требования
                        </h3>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px", marginBottom: "20px" }}>
                            <div>
                                <label className="form-label">Категория данных</label>
                                <select name="data_category" value={formData.data_category} onChange={handleChange} className="form-input">
                                    {(Object.entries(dataCategoryLabels) as [DataCategory, string][]).map(([value, label]) => (
                                        <option key={value} value={value}>{label}</option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="form-label">Категория КИИ (187-ФЗ)</label>
                                <select name="kii_category" value={formData.kii_category} onChange={handleChange} className="form-input">
                                    {(Object.entries(kiiCategoryLabels) as [KIICategory, string][]).map(([value, label]) => (
                                        <option key={value} value={value}>{label}</option>
                                    ))}
                                </select>
                            </div>
                        </div>

                        <div style={{ padding: "20px", background: formData.has_personal_data ? "var(--threat-critical-dim)" : "var(--well)", borderRadius: "var(--r-md)", border: `1px solid ${formData.has_personal_data ? "var(--danger)" : "var(--perimeter)"}`, marginBottom: "20px" }}>
                            <label style={{ display: "flex", alignItems: "center", gap: "10px", cursor: "pointer", fontWeight: 600 }}>
                                <input type="checkbox" name="has_personal_data" checked={formData.has_personal_data} onChange={handleChange} style={{ width: 16, height: 16 }} />
                                Обрабатывает персональные данные (152-ФЗ)
                            </label>
                            {formData.has_personal_data && (
                                <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px", marginTop: "16px" }}>
                                    <div>
                                        <label className="form-label">Уровень защищённости ПДн</label>
                                        <select name="protection_level" value={formData.protection_level} onChange={handleChange} className="form-input">
                                            <option value="">-- Не определён --</option>
                                            {(Object.entries(protectionLevelLabels) as [ProtectionLevel, string][]).map(([value, label]) => (
                                                <option key={value} value={value}>{label}</option>
                                            ))}
                                        </select>
                                    </div>
                                    <div>
                                        <label className="form-label">Объём ПДн (субъектов)</label>
                                        <select name="personal_data_volume" value={formData.personal_data_volume} onChange={handleChange} className="form-input">
                                            <option value="">-- Не определён --</option>
                                            <option value="<1000">Менее 1 000</option>
                                            <option value="1000-100000">От 1 000 до 100 000</option>
                                            <option value=">100000">Более 100 000</option>
                                        </select>
                                    </div>
                                </motion.div>
                            )}
                        </div>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px" }}>
                            <div style={{ padding: "16px", background: "var(--well)", borderRadius: "var(--r-md)", border: "1px solid var(--perimeter)" }}>
                                <label style={{ display: "flex", gap: "12px", cursor: "pointer" }}>
                                    <input type="checkbox" name="has_internet_access" checked={formData.has_internet_access} onChange={handleChange} style={{ width: 16, height: 16, marginTop: 4 }} />
                                    <div>
                                        <div style={{ fontWeight: 600 }}>Доступ в интернет</div>
                                        <div style={{ fontSize: "12px", color: "var(--ink-muted)", marginTop: 2 }}>Актив имеет доступ к сети Интернет</div>
                                    </div>
                                </label>
                            </div>
                            <div style={{ padding: "16px", background: formData.is_isolated ? "var(--threat-low-dim)" : "var(--well)", borderRadius: "var(--r-md)", border: `1px solid ${formData.is_isolated ? "var(--success)" : "var(--perimeter)"}` }}>
                                <label style={{ display: "flex", gap: "12px", cursor: "pointer" }}>
                                    <input type="checkbox" name="is_isolated" checked={formData.is_isolated} onChange={handleChange} style={{ width: 16, height: 16, marginTop: 4 }} />
                                    <div>
                                        <div style={{ fontWeight: 600 }}>Изолированный сегмент</div>
                                        <div style={{ fontSize: "12px", color: "var(--ink-muted)", marginTop: 2 }}>Снижает вероятность сетевых атак</div>
                                    </div>
                                </label>
                            </div>
                        </div>
                    </div>

                    {/* Расчет */}
                    <div style={{ marginBottom: "32px", padding: "20px", background: "var(--command-dim)", borderRadius: "var(--r-md)", border: "1px solid var(--command)", display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <div>
                            <div style={{ fontSize: "14px", fontWeight: 600, color: "var(--command-text)", display: 'flex', alignItems: 'center', gap: '6px' }}>
                                <Info size={16} /> Автоматический расчёт критичности
                            </div>
                            <div style={{ fontSize: "13px", color: "var(--ink-muted)", marginTop: "4px" }}>
                                На основе категории данных, КИИ, ПДн и среды
                            </div>
                        </div>
                        <div style={{ fontSize: "32px", fontWeight: 700, color: getMetricColor(calculatedCriticality) }}>
                            {calculatedCriticality}<span style={{ fontSize: '16px', color: 'var(--ink-muted)', fontWeight: 500 }}>/5</span>
                        </div>
                    </div>

                    <div style={{ display: "flex", gap: "12px", justifyContent: "flex-end", paddingTop: "20px", borderTop: "1px solid var(--perimeter)" }}>
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
