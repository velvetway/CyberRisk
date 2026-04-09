import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { DataCategory, ProtectionLevel, KIICategory } from "../types";

interface AssetFormData {
    name: string;
    description: string;
    type: string;
    location: string;
    owner: string;
    environment: string;
    // Регуляторные поля
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
    // Регуляторные поля
    data_category: "internal",
    protection_level: "",
    kii_category: "none",
    has_personal_data: false,
    personal_data_volume: "",
    has_internet_access: true,
    is_isolated: false,
};

// Функция расчёта критичности (зеркалит бэкенд логику)
const calculateCriticality = (form: AssetFormData): number => {
    let score = 2.0;

    // 1. Категория данных
    switch (form.data_category) {
        case "state_secret":
            score = 5.0;
            break;
        case "personal_data_special":
        case "personal_data_biometric":
            score = 4.5;
            break;
        case "kii":
            score = 5.0;
            break;
        case "personal_data":
            score = 3.5;
            break;
        case "confidential":
        case "banking_secret":
        case "medical_secret":
        case "commercial_secret":
            score = 3.5;
            break;
        case "internal":
            score = 2.5;
            break;
        case "public":
            score = 1.5;
            break;
    }

    // 2. Категория КИИ
    switch (form.kii_category) {
        case "cat1":
            score = Math.max(score, 5.0);
            break;
        case "cat2":
            score = Math.max(score, 4.0);
            break;
        case "cat3":
            score = Math.max(score, 3.5);
            break;
    }

    // 3. Уровень защищённости ПДн
    switch (form.protection_level) {
        case "uz1":
            score = Math.max(score, 4.5);
            break;
        case "uz2":
            score = Math.max(score, 4.0);
            break;
        case "uz3":
            score = Math.max(score, 3.0);
            break;
        case "uz4":
            score = Math.max(score, 2.5);
            break;
    }

    // 4. Объём ПДн
    if (form.has_personal_data && form.personal_data_volume) {
        if (form.personal_data_volume === ">100000") {
            score += 0.5;
        } else if (form.personal_data_volume === "1000-100000") {
            score += 0.25;
        }
    }

    // 5. Среда
    switch (form.environment) {
        case "prod":
            score += 0.5;
            break;
        case "test":
            score -= 0.5;
            break;
        case "dev":
            score -= 1.0;
            break;
    }

    // 6. Сетевая доступность
    if (form.has_internet_access) {
        score += 0.5;
    }
    if (form.is_isolated) {
        score -= 0.5;
    }

    return Math.min(5, Math.max(1, Math.round(score)));
};

// Справочники для отображения
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
    const [error, setError] = useState<string | null>(null);
    const [loadingData, setLoadingData] = useState(isEditMode);

    useEffect(() => {
        if (isEditMode && id) {
            fetchAsset(id);
        }
    }, [id, isEditMode]);

    const fetchAsset = async (assetId: string) => {
        setLoadingData(true);
        try {
            const res = await fetch(`/api/assets/${assetId}`);
            if (!res.ok) {
                throw new Error(`Ошибка загрузки актива: ${res.status}`);
            }
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
            setError(e.message || "Ошибка загрузки данных актива");
        } finally {
            setLoadingData(false);
        }
    };

    const handleChange = (
        e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
    ) => {
        const { name, value, type } = e.target;
        const checked = (e.target as HTMLInputElement).checked;

        setFormData((prev) => ({
            ...prev,
            [name]: type === "checkbox" ? checked : value,
        }));
    };

    // Расчётная критичность на основе текущих данных формы
    const calculatedCriticality = calculateCriticality(formData);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);

        if (!formData.name.trim()) {
            setError("Название актива обязательно для заполнения");
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
                // Регуляторные поля (критичность рассчитывается автоматически на бэкенде)
                data_category: formData.data_category,
                protection_level: formData.protection_level || undefined,
                kii_category: formData.kii_category,
                has_personal_data: formData.has_personal_data,
                personal_data_volume: formData.personal_data_volume || undefined,
                has_internet_access: formData.has_internet_access,
                is_isolated: formData.is_isolated,
            };

            const res = await fetch(url, {
                method,
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            if (!res.ok) {
                const text = await res.text();
                throw new Error(`Ошибка ${isEditMode ? "обновления" : "создания"}: ${text}`);
            }

            navigate("/assets");
        } catch (e: any) {
            setError(e.message || `Ошибка при ${isEditMode ? "обновлении" : "создании"} актива`);
        } finally {
            setLoading(false);
        }
    };

    const handleCancel = () => {
        navigate("/assets");
    };

    // Автоматически включать ПДн при выборе соответствующей категории
    useEffect(() => {
        if (
            formData.data_category === "personal_data" ||
            formData.data_category === "personal_data_special" ||
            formData.data_category === "personal_data_biometric"
        ) {
            setFormData((prev) => ({ ...prev, has_personal_data: true }));
        }
    }, [formData.data_category]);

    if (loadingData) {
        return (
            <div className="fade-in" style={{ textAlign: "center", padding: "60px 20px" }}>
                <div className="loading-spinner" style={{ margin: "0 auto 16px" }} />
                <p style={{ color: "var(--color-text-light)" }}>Загрузка данных актива...</p>
            </div>
        );
    }

    return (
        <div className="fade-in">
            <div style={{ marginBottom: "32px" }}>
                <h1 style={{ marginBottom: "8px" }}>
                    {isEditMode ? "Редактировать актив" : "Создать новый актив"}
                </h1>
                <p style={{ color: "var(--color-text-light)", fontSize: "15px" }}>
                    {isEditMode
                        ? "Обновите информацию об ИТ-активе"
                        : "Добавьте новый ИТ-актив в систему управления рисками"}
                </p>
            </div>

            <div className="card" style={{ padding: "32px", maxWidth: "900px" }}>
                <form onSubmit={handleSubmit}>
                    {/* Основная информация */}
                    <div style={{ marginBottom: "32px" }}>
                        <h3
                            style={{
                                fontSize: "16px",
                                fontWeight: 600,
                                marginBottom: "20px",
                                color: "var(--color-primary-dark)",
                                borderBottom: "2px solid var(--color-border)",
                                paddingBottom: "8px",
                            }}
                        >
                            Основная информация
                        </h3>

                        <div style={{ marginBottom: "20px" }}>
                            <label className="form-label">
                                Название актива <span style={{ color: "var(--color-accent)" }}>*</span>
                            </label>
                            <input
                                type="text"
                                name="name"
                                value={formData.name}
                                onChange={handleChange}
                                className="form-input"
                                placeholder="Например: Web Application Server"
                                required
                            />
                        </div>

                        <div style={{ marginBottom: "20px" }}>
                            <label className="form-label">Описание</label>
                            <textarea
                                name="description"
                                value={formData.description}
                                onChange={handleChange}
                                className="form-input"
                                rows={3}
                                placeholder="Краткое описание актива и его назначения"
                                style={{ resize: "vertical" }}
                            />
                        </div>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}>
                            <div>
                                <label className="form-label">Тип актива</label>
                                <select
                                    name="type"
                                    value={formData.type}
                                    onChange={handleChange}
                                    className="form-input"
                                >
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
                                <label className="form-label">Расчётная критичность</label>
                                <div
                                    style={{
                                        padding: "12px 16px",
                                        background: "var(--color-bg)",
                                        borderRadius: "var(--radius-sm)",
                                        border: "1px solid var(--color-border)",
                                        display: "flex",
                                        alignItems: "center",
                                        justifyContent: "space-between",
                                    }}
                                >
                                    <span style={{ fontSize: "13px", color: "var(--color-text-muted)" }}>
                                        Автоматический расчёт
                                    </span>
                                    <span
                                        style={{
                                            fontSize: "24px",
                                            fontWeight: 700,
                                            color: getMetricColor(calculatedCriticality),
                                        }}
                                    >
                                        {calculatedCriticality}
                                    </span>
                                </div>
                                <div style={{ fontSize: "11px", color: "var(--color-text-muted)", marginTop: "4px" }}>
                                    На основе категории данных, КИИ, ПДн и среды
                                </div>
                            </div>
                        </div>

                        <div
                            style={{
                                display: "grid",
                                gridTemplateColumns: "repeat(3, 1fr)",
                                gap: "16px",
                                marginTop: "20px",
                            }}
                        >
                            <div>
                                <label className="form-label">Среда</label>
                                <select
                                    name="environment"
                                    value={formData.environment}
                                    onChange={handleChange}
                                    className="form-input"
                                >
                                    <option value="prod">Production</option>
                                    <option value="stage">Staging</option>
                                    <option value="test">Test</option>
                                    <option value="dev">Development</option>
                                </select>
                            </div>

                            <div>
                                <label className="form-label">Расположение</label>
                                <input
                                    type="text"
                                    name="location"
                                    value={formData.location}
                                    onChange={handleChange}
                                    className="form-input"
                                    placeholder="Datacenter A, Rack 12"
                                />
                            </div>

                            <div>
                                <label className="form-label">Владелец</label>
                                <input
                                    type="text"
                                    name="owner"
                                    value={formData.owner}
                                    onChange={handleChange}
                                    className="form-input"
                                    placeholder="IT Department"
                                />
                            </div>
                        </div>
                    </div>

                    {/* Регуляторные требования */}
                    <div style={{ marginBottom: "32px" }}>
                        <h3
                            style={{
                                fontSize: "16px",
                                fontWeight: 600,
                                marginBottom: "20px",
                                color: "var(--color-primary-dark)",
                                borderBottom: "2px solid var(--color-border)",
                                paddingBottom: "8px",
                            }}
                        >
                            Регуляторные требования
                        </h3>

                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginBottom: "20px" }}>
                            <div>
                                <label className="form-label">Категория данных</label>
                                <select
                                    name="data_category"
                                    value={formData.data_category}
                                    onChange={handleChange}
                                    className="form-input"
                                >
                                    {(Object.entries(dataCategoryLabels) as [DataCategory, string][]).map(([value, label]) => (
                                        <option key={value} value={value}>{label}</option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="form-label">Категория КИИ (187-ФЗ)</label>
                                <select
                                    name="kii_category"
                                    value={formData.kii_category}
                                    onChange={handleChange}
                                    className="form-input"
                                >
                                    {(Object.entries(kiiCategoryLabels) as [KIICategory, string][]).map(([value, label]) => (
                                        <option key={value} value={value}>{label}</option>
                                    ))}
                                </select>
                            </div>
                        </div>

                        {/* Персональные данные */}
                        <div style={{
                            padding: "16px",
                            background: formData.has_personal_data ? "rgba(231, 76, 60, 0.05)" : "var(--color-bg)",
                            borderRadius: "var(--radius-sm)",
                            border: `1px solid ${formData.has_personal_data ? "var(--color-risk-high)" : "var(--color-border)"}`,
                            marginBottom: "20px",
                        }}>
                            <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "12px" }}>
                                <input
                                    type="checkbox"
                                    id="has_personal_data"
                                    name="has_personal_data"
                                    checked={formData.has_personal_data}
                                    onChange={handleChange}
                                    style={{ width: "18px", height: "18px" }}
                                />
                                <label htmlFor="has_personal_data" style={{ fontWeight: 600, cursor: "pointer" }}>
                                    Обрабатывает персональные данные (152-ФЗ)
                                </label>
                            </div>

                            {formData.has_personal_data && (
                                <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginTop: "16px" }}>
                                    <div>
                                        <label className="form-label">Уровень защищённости ПДн (ПП-1119)</label>
                                        <select
                                            name="protection_level"
                                            value={formData.protection_level}
                                            onChange={handleChange}
                                            className="form-input"
                                        >
                                            <option value="">-- Не определён --</option>
                                            {(Object.entries(protectionLevelLabels) as [ProtectionLevel, string][]).map(([value, label]) => (
                                                <option key={value} value={value}>{label}</option>
                                            ))}
                                        </select>
                                    </div>

                                    <div>
                                        <label className="form-label">Объём ПДн (кол-во субъектов)</label>
                                        <select
                                            name="personal_data_volume"
                                            value={formData.personal_data_volume}
                                            onChange={handleChange}
                                            className="form-input"
                                        >
                                            <option value="">-- Не определён --</option>
                                            <option value="<1000">Менее 1 000</option>
                                            <option value="1000-100000">От 1 000 до 100 000</option>
                                            <option value=">100000">Более 100 000</option>
                                        </select>
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* Сетевые параметры */}
                        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}>
                            <div style={{
                                padding: "16px",
                                background: "var(--color-bg)",
                                borderRadius: "var(--radius-sm)",
                                border: "1px solid var(--color-border)",
                            }}>
                                <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
                                    <input
                                        type="checkbox"
                                        id="has_internet_access"
                                        name="has_internet_access"
                                        checked={formData.has_internet_access}
                                        onChange={handleChange}
                                        style={{ width: "18px", height: "18px" }}
                                    />
                                    <label htmlFor="has_internet_access" style={{ cursor: "pointer" }}>
                                        <strong>Доступ в интернет</strong>
                                        <div style={{ fontSize: "12px", color: "var(--color-text-muted)", marginTop: "4px" }}>
                                            Актив имеет доступ к сети Интернет
                                        </div>
                                    </label>
                                </div>
                            </div>

                            <div style={{
                                padding: "16px",
                                background: formData.is_isolated ? "rgba(46, 204, 113, 0.05)" : "var(--color-bg)",
                                borderRadius: "var(--radius-sm)",
                                border: `1px solid ${formData.is_isolated ? "var(--color-risk-low)" : "var(--color-border)"}`,
                            }}>
                                <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
                                    <input
                                        type="checkbox"
                                        id="is_isolated"
                                        name="is_isolated"
                                        checked={formData.is_isolated}
                                        onChange={handleChange}
                                        style={{ width: "18px", height: "18px" }}
                                    />
                                    <label htmlFor="is_isolated" style={{ cursor: "pointer" }}>
                                        <strong>Изолированный сегмент</strong>
                                        <div style={{ fontSize: "12px", color: "var(--color-text-muted)", marginTop: "4px" }}>
                                            Снижает вероятность сетевых атак
                                        </div>
                                    </label>
                                </div>
                            </div>
                        </div>

                        {/* Информационная панель */}
                        {(formData.kii_category !== "none" || formData.has_personal_data) && (
                            <div style={{
                                marginTop: "20px",
                                padding: "16px",
                                background: "rgba(231, 76, 60, 0.08)",
                                borderRadius: "var(--radius-sm)",
                                border: "1px solid var(--color-risk-high)",
                            }}>
                                <div style={{ fontSize: "14px", fontWeight: 600, marginBottom: "8px", color: "var(--color-risk-high)" }}>
                                    ⚠️ Внимание: повышенные требования к защите
                                </div>
                                <ul style={{ margin: 0, paddingLeft: "20px", fontSize: "13px", color: "var(--color-text)", lineHeight: 1.6 }}>
                                    {formData.kii_category !== "none" && (
                                        <li>
                                            <strong>КИИ:</strong> Требуется соответствие 187-ФЗ и приказам ФСТЭК №235, №239.
                                            Необходимо подключение к ГосСОПКА.
                                        </li>
                                    )}
                                    {formData.has_personal_data && (
                                        <li>
                                            <strong>ПДн:</strong> Требуется соответствие 152-ФЗ, ПП-1119 и приказу ФСТЭК №21.
                                            Необходимо уведомление Роскомнадзора.
                                        </li>
                                    )}
                                </ul>
                            </div>
                        )}
                    </div>

                    {/* Автоматический расчёт */}
                    <div style={{
                        marginBottom: "32px",
                        padding: "16px",
                        background: "linear-gradient(135deg, rgba(52, 152, 219, 0.05) 0%, rgba(155, 89, 182, 0.05) 100%)",
                        borderRadius: "var(--radius-sm)",
                        border: "1px solid var(--color-primary)"
                    }}>
                        <div style={{ fontSize: "13px", color: "var(--color-primary-dark)", marginBottom: "8px" }}>
                            ℹ️ <strong>Автоматический расчёт критичности и CIA:</strong>
                        </div>
                        <div style={{ fontSize: "12px", color: "var(--color-text)", lineHeight: 1.6 }}>
                            <strong>Критичность</strong> рассчитывается автоматически на основе:
                            <ul style={{ margin: "8px 0", paddingLeft: "20px" }}>
                                <li>Категории обрабатываемых данных (гостайна, КИИ, ПДн)</li>
                                <li>Категории КИИ по 187-ФЗ</li>
                                <li>Уровня защищённости ПДн по 152-ФЗ</li>
                                <li>Объёма обрабатываемых ПДн</li>
                                <li>Среды развёртывания и сетевой доступности</li>
                            </ul>
                            <strong>CIA-триада</strong> (конфиденциальность, целостность, доступность) рассчитывается на основе типа актива и критичности.
                        </div>
                    </div>

                    {/* Кнопки управления */}
                    <div
                        style={{
                            display: "flex",
                            gap: "12px",
                            justifyContent: "flex-end",
                            paddingTop: "20px",
                            borderTop: "1px solid var(--color-border)",
                        }}
                    >
                        <button
                            type="button"
                            onClick={handleCancel}
                            className="btn"
                            style={{
                                padding: "10px 24px",
                                background: "var(--color-surface)",
                                color: "var(--color-text)",
                                border: "1px solid var(--color-border)",
                            }}
                        >
                            Отмена
                        </button>
                        <button
                            type="submit"
                            className="btn btn-primary"
                            disabled={loading}
                            style={{ padding: "10px 24px" }}
                        >
                            {loading ? (
                                <>
                                    <span className="loading-spinner" />
                                    {isEditMode ? "Сохранение..." : "Создание..."}
                                </>
                            ) : (
                                <>{isEditMode ? "Сохранить изменения" : "Создать актив"}</>
                            )}
                        </button>
                    </div>

                    {error && (
                        <div
                            style={{
                                marginTop: "20px",
                                padding: "12px",
                                background: "rgba(231, 76, 60, 0.1)",
                                borderRadius: "var(--radius-sm)",
                                color: "var(--color-risk-critical)",
                                fontSize: "14px",
                                border: "1px solid var(--color-risk-critical)",
                            }}
                        >
                            ⚠️ {error}
                        </div>
                    )}
                </form>
            </div>
        </div>
    );
};

const getMetricColor = (value: number): string => {
    if (value >= 4) return "var(--color-risk-critical)";
    if (value === 3) return "var(--color-risk-medium)";
    return "var(--color-risk-low)";
};
