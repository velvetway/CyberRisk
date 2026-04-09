import React, { useState, useEffect } from "react";
import { api } from "../api/client";
import { Software, SoftwareCategory } from "../types";

const normalizeSoftware = (raw: any): Software => ({
    id: raw?.id ?? raw?.ID ?? 0,
    name: raw?.name ?? raw?.Name ?? "",
    vendor: raw?.vendor ?? raw?.Vendor ?? "",
    version: raw?.version ?? raw?.Version,
    category_id: raw?.category_id ?? raw?.CategoryID,
    category_name: raw?.category_name ?? raw?.CategoryName,
    is_russian: Boolean(raw?.is_russian ?? raw?.IsRussian),
    registry_number: raw?.registry_number ?? raw?.RegistryNumber,
    registry_date: raw?.registry_date ?? raw?.RegistryDate,
    registry_url: raw?.registry_url ?? raw?.RegistryURL,
    fstec_certified: Boolean(raw?.fstec_certified ?? raw?.FSTECCertified),
    fstec_certificate_num: raw?.fstec_certificate_num ?? raw?.FSTECCertificateNum,
    fstec_certificate_date: raw?.fstec_certificate_date ?? raw?.FSTECCertificateDate,
    fstec_protection_class: raw?.fstec_protection_class ?? raw?.FSTECProtectionClass,
    fstec_valid_until: raw?.fstec_valid_until ?? raw?.FSTECValidUntil,
    fsb_certified: Boolean(raw?.fsb_certified ?? raw?.FSBCertified),
    fsb_certificate_num: raw?.fsb_certificate_num ?? raw?.FSBCertificateNum,
    fsb_protection_class: raw?.fsb_protection_class ?? raw?.FSBProtectionClass,
    description: raw?.description ?? raw?.Description,
    website: raw?.website ?? raw?.Website,
    created_at: raw?.created_at ?? raw?.CreatedAt ?? "",
    updated_at: raw?.updated_at ?? raw?.UpdatedAt ?? "",
});

const normalizeCategory = (raw: any): SoftwareCategory => ({
    id: raw?.id ?? raw?.ID ?? 0,
    code: raw?.code ?? raw?.Code ?? "",
    name: raw?.name ?? raw?.Name ?? "",
    description: raw?.description ?? raw?.Description ?? "",
});

export const SoftwareCatalogPage: React.FC = () => {
    const [software, setSoftware] = useState<Software[]>([]);
    const [categories, setCategories] = useState<SoftwareCategory[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Фильтры
    const [searchTerm, setSearchTerm] = useState("");
    const [categoryFilter, setCategoryFilter] = useState<string>("");
    const [russianOnly, setRussianOnly] = useState(false);
    const [certifiedOnly, setCertifiedOnly] = useState(false);

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        setLoading(true);
        try {
            const [swRes, catRes] = await Promise.all([
                fetch("/api/software"),
                fetch("/api/software/categories"),
            ]);

            if (!swRes.ok || !catRes.ok) {
                throw new Error("Ошибка загрузки данных");
            }

            const swData = await swRes.json();
            const catData = await catRes.json();

            setSoftware(Array.isArray(swData) ? swData.map(normalizeSoftware) : []);
            setCategories(Array.isArray(catData) ? catData.map(normalizeCategory) : []);
        } catch (e: any) {
            setError(e.message || "Ошибка загрузки справочника ПО");
        } finally {
            setLoading(false);
        }
    };

    // Фильтрация
    const filteredSoftware = software.filter((sw) => {
        const name = sw.name || "";
        const vendor = sw.vendor || "";
        const matchesSearch =
            name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            vendor.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesCategory = !categoryFilter || sw.category_id?.toString() === categoryFilter;
        const matchesRussian = !russianOnly || sw.is_russian;
        const matchesCertified = !certifiedOnly || sw.fstec_certified || sw.fsb_certified;

        return matchesSearch && matchesCategory && matchesRussian && matchesCertified;
    });

    // Статистика
    const stats = {
        total: software.length,
        russian: software.filter((s) => s.is_russian).length,
        fstecCertified: software.filter((s) => s.fstec_certified).length,
        fsbCertified: software.filter((s) => s.fsb_certified).length,
    };

    if (loading) {
        return (
            <div className="fade-in" style={{ textAlign: "center", padding: "60px 20px" }}>
                <div className="loading-spinner" style={{ margin: "0 auto 16px" }} />
                <p style={{ color: "var(--color-text-light)" }}>Загрузка справочника ПО...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="fade-in" style={{ textAlign: "center", padding: "60px 20px" }}>
                <p style={{ color: "var(--color-risk-critical)" }}>⚠️ {error}</p>
                <button className="btn btn-primary" onClick={fetchData} style={{ marginTop: "16px" }}>
                    Повторить
                </button>
            </div>
        );
    }

    return (
        <div className="fade-in">
            <div style={{ marginBottom: "32px" }}>
                <h1 style={{ marginBottom: "8px" }}>Справочник программного обеспечения</h1>
                <p style={{ color: "var(--color-text-light)", fontSize: "15px" }}>
                    Каталог ПО с информацией о сертификации ФСТЭК/ФСБ и включении в реестр Минцифры
                </p>
            </div>

            {/* Статистика */}
            <div
                style={{
                    display: "grid",
                    gridTemplateColumns: "repeat(4, 1fr)",
                    gap: "16px",
                    marginBottom: "24px",
                }}
            >
                <div className="card" style={{ padding: "16px", textAlign: "center" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--color-primary)" }}>
                        {stats.total}
                    </div>
                    <div style={{ fontSize: "13px", color: "var(--color-text-light)" }}>Всего ПО</div>
                </div>
                <div className="card" style={{ padding: "16px", textAlign: "center" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--color-risk-low)" }}>
                        {stats.russian}
                    </div>
                    <div style={{ fontSize: "13px", color: "var(--color-text-light)" }}>
                        В реестре Минцифры
                    </div>
                </div>
                <div className="card" style={{ padding: "16px", textAlign: "center" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--color-risk-medium)" }}>
                        {stats.fstecCertified}
                    </div>
                    <div style={{ fontSize: "13px", color: "var(--color-text-light)" }}>
                        Сертификат ФСТЭК
                    </div>
                </div>
                <div className="card" style={{ padding: "16px", textAlign: "center" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--color-risk-high)" }}>
                        {stats.fsbCertified}
                    </div>
                    <div style={{ fontSize: "13px", color: "var(--color-text-light)" }}>
                        Сертификат ФСБ
                    </div>
                </div>
            </div>

            {/* Фильтры */}
            <div className="card" style={{ padding: "20px", marginBottom: "24px" }}>
                <div
                    style={{
                        display: "grid",
                        gridTemplateColumns: "2fr 1fr auto auto",
                        gap: "16px",
                        alignItems: "end",
                    }}
                >
                    <div>
                        <label className="form-label">Поиск</label>
                        <input
                            type="text"
                            className="form-input"
                            placeholder="Название или производитель..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                    </div>
                    <div>
                        <label className="form-label">Категория</label>
                        <select
                            className="form-input"
                            value={categoryFilter}
                            onChange={(e) => setCategoryFilter(e.target.value)}
                        >
                            <option value="">Все категории</option>
                            {categories.map((cat) => (
                                <option key={cat.id} value={cat.id}>
                                    {cat.name}
                                </option>
                            ))}
                        </select>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: "8px", paddingBottom: "8px" }}>
                        <input
                            type="checkbox"
                            id="russianOnly"
                            checked={russianOnly}
                            onChange={(e) => setRussianOnly(e.target.checked)}
                        />
                        <label htmlFor="russianOnly" style={{ cursor: "pointer", whiteSpace: "nowrap" }}>
                            Только российское
                        </label>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: "8px", paddingBottom: "8px" }}>
                        <input
                            type="checkbox"
                            id="certifiedOnly"
                            checked={certifiedOnly}
                            onChange={(e) => setCertifiedOnly(e.target.checked)}
                        />
                        <label htmlFor="certifiedOnly" style={{ cursor: "pointer", whiteSpace: "nowrap" }}>
                            Сертифицированное
                        </label>
                    </div>
                </div>
            </div>

            {/* Результаты */}
            <div style={{ fontSize: "14px", color: "var(--color-text-light)", marginBottom: "12px" }}>
                Найдено: {filteredSoftware.length}
            </div>

            {/* Таблица ПО */}
            <div className="card" style={{ overflow: "hidden" }}>
                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                    <thead>
                        <tr style={{ background: "var(--color-bg)", borderBottom: "2px solid var(--color-border)" }}>
                            <th style={thStyle}>Название</th>
                            <th style={thStyle}>Производитель</th>
                            <th style={thStyle}>Категория</th>
                            <th style={{ ...thStyle, textAlign: "center" }}>Реестр РФ</th>
                            <th style={{ ...thStyle, textAlign: "center" }}>ФСТЭК</th>
                            <th style={{ ...thStyle, textAlign: "center" }}>ФСБ</th>
                            <th style={thStyle}>Российские аналоги</th>
                        </tr>
                    </thead>
                    <tbody>
                        {filteredSoftware.length === 0 ? (
                            <tr>
                                <td colSpan={7} style={{ padding: "40px", textAlign: "center", color: "var(--color-text-light)" }}>
                                    Ничего не найдено
                                </td>
                            </tr>
                        ) : (
                            filteredSoftware.map((sw) => (
                                <tr key={sw.id} style={{ borderBottom: "1px solid var(--color-border)" }}>
                                    <td style={tdStyle}>
                                        <div style={{ fontWeight: 600 }}>{sw.name}</div>
                                        {sw.version && (
                                            <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
                                                v{sw.version}
                                            </div>
                                        )}
                                    </td>
                                    <td style={tdStyle}>{sw.vendor}</td>
                                    <td style={tdStyle}>
                                        <span
                                            style={{
                                                background: "var(--color-bg)",
                                                padding: "4px 8px",
                                                borderRadius: "4px",
                                                fontSize: "12px",
                                            }}
                                        >
                                            {sw.category_name || "—"}
                                        </span>
                                    </td>
                                    <td style={{ ...tdStyle, textAlign: "center" }}>
                                        {sw.is_russian ? (
                                            <span title={`Реестр №${sw.registry_number}`} style={badgeGreen}>
                                                ✓ {sw.registry_number}
                                            </span>
                                        ) : (
                                            <span style={badgeGray}>—</span>
                                        )}
                                    </td>
                                    <td style={{ ...tdStyle, textAlign: "center" }}>
                                        {sw.fstec_certified ? (
                                            <span title={`Класс ${sw.fstec_protection_class}`} style={badgeOrange}>
                                                ✓ Кл.{sw.fstec_protection_class}
                                            </span>
                                        ) : (
                                            <span style={badgeGray}>—</span>
                                        )}
                                    </td>
                                    <td style={{ ...tdStyle, textAlign: "center" }}>
                                        {sw.fsb_certified ? (
                                            <span title={`Класс ${sw.fsb_protection_class}`} style={badgeRed}>
                                                ✓ {sw.fsb_protection_class}
                                            </span>
                                        ) : (
                                            <span style={badgeGray}>—</span>
                                        )}
                                    </td>
                                    <td style={tdStyle}>
                                        {sw.is_russian ? (
                                            <span style={{ color: "var(--color-risk-low)", fontWeight: 600 }}>
                                                Уже российское
                                            </span>
                                        ) : (
                                            <AlternativesCell software={sw} />
                                        )}
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Легенда */}
            <div
                style={{
                    marginTop: "24px",
                    padding: "16px",
                    background: "var(--color-bg)",
                    borderRadius: "var(--radius-sm)",
                    fontSize: "13px",
                    color: "var(--color-text-light)",
                }}
            >
                <strong>Легенда:</strong>
                <div style={{ display: "flex", gap: "24px", marginTop: "8px", flexWrap: "wrap" }}>
                    <span>
                        <span style={badgeGreen}>✓</span> — В реестре российского ПО (Минцифры)
                    </span>
                    <span>
                        <span style={badgeOrange}>✓</span> — Сертификат ФСТЭК России
                    </span>
                    <span>
                        <span style={badgeRed}>✓</span> — Сертификат ФСБ России (СКЗИ)
                    </span>
                </div>
            </div>
        </div>
    );
};

// Стили
const thStyle: React.CSSProperties = {
    padding: "12px 16px",
    textAlign: "left",
    fontSize: "13px",
    fontWeight: 600,
    color: "var(--color-text-light)",
};

const tdStyle: React.CSSProperties = {
    padding: "12px 16px",
    fontSize: "14px",
};

const badgeGreen: React.CSSProperties = {
    display: "inline-block",
    padding: "4px 8px",
    borderRadius: "4px",
    fontSize: "11px",
    fontWeight: 600,
    background: "rgba(46, 204, 113, 0.15)",
    color: "#27ae60",
};

const badgeOrange: React.CSSProperties = {
    display: "inline-block",
    padding: "4px 8px",
    borderRadius: "4px",
    fontSize: "11px",
    fontWeight: 600,
    background: "rgba(243, 156, 18, 0.15)",
    color: "#e67e22",
};

const badgeRed: React.CSSProperties = {
    display: "inline-block",
    padding: "4px 8px",
    borderRadius: "4px",
    fontSize: "11px",
    fontWeight: 600,
    background: "rgba(231, 76, 60, 0.15)",
    color: "#c0392b",
};

const badgeGray: React.CSSProperties = {
    display: "inline-block",
    padding: "4px 8px",
    borderRadius: "4px",
    fontSize: "11px",
    color: "var(--color-text-muted)",
};

const AlternativesCell: React.FC<{ software: Software }> = ({ software }) => {
    const [expanded, setExpanded] = useState(false);
    const [loading, setLoading] = useState(false);
    const [items, setItems] = useState<Software[]>([]);
    const [error, setError] = useState<string | null>(null);

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
            const data = await api.getSoftwareAlternatives(software.id);
            setItems(Array.isArray(data) ? data : []);
        } catch (err: any) {
            setError(err?.message || "Не удалось загрузить аналоги");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <button
                className="btn"
                onClick={toggle}
                disabled={loading}
                style={{ padding: "4px 10px", fontSize: "12px" }}
            >
                {loading ? "Поиск..." : expanded ? "Скрыть" : "Показать аналоги"}
            </button>
            {expanded && (
                <div style={{ marginTop: "8px", fontSize: "12px", color: "var(--color-text-light)" }}>
                    {loading && <div>Загрузка...</div>}
                    {!loading && error && (
                        <div style={{ color: "var(--color-risk-critical)" }}>⚠️ {error}</div>
                    )}
                    {!loading && !error && items.length === 0 && (
                        <div>Российские аналоги в этой категории не найдены.</div>
                    )}
                    {!loading && !error && items.length > 0 && (
                        <ul style={{ margin: 0, paddingLeft: "18px", color: "var(--color-text)" }}>
                            {items.slice(0, 5).map((item) => (
                                <li key={item.id}>
                                    <strong>{item.name}</strong> — {item.vendor}
                                    {item.fstec_certified && (
                                        <span style={{ marginLeft: "6px", ...badgeOrange }}>ФСТЭК</span>
                                    )}
                                    {item.fsb_certified && (
                                        <span style={{ marginLeft: "4px", ...badgeRed }}>ФСБ</span>
                                    )}
                                </li>
                            ))}
                            {items.length > 5 && (
                                <li style={{ color: "var(--color-text-muted)" }}>
                                    И ещё {items.length - 5} вариантов...
                                </li>
                            )}
                        </ul>
                    )}
                </div>
            )}
        </div>
    );
};
