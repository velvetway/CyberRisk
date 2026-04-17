import React, { useState, useEffect } from "react";
import { authFetch } from "../api/client";
import { Software, SoftwareCategory } from "../types";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Search } from "lucide-react";

export const SoftwareCatalogPage: React.FC = () => {
    // ... [existing normalization logic from SoftwareCatalogPage, copying it...]
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

    const [software, setSoftware] = useState<Software[]>([]);
    const [categories, setCategories] = useState<SoftwareCategory[]>([]);
    const [loading, setLoading] = useState(true);

    const [searchTerm, setSearchTerm] = useState("");
    const [categoryFilter, setCategoryFilter] = useState<string>("");
    const [russianOnly, setRussianOnly] = useState(false);
    const [certifiedOnly, setCertifiedOnly] = useState(false);

    useEffect(() => {
        fetchData();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const fetchData = async () => {
        setLoading(true);
        try {
            const [swRes, catRes] = await Promise.all([
                authFetch("/api/software"),
                authFetch("/api/software/categories"),
            ]);

            if (!swRes.ok || !catRes.ok) throw new Error("Ошибка загрузки данных");

            const swData = await swRes.json();
            const catData = await catRes.json();

            setSoftware(Array.isArray(swData) ? swData.map(normalizeSoftware) : []);
            setCategories(Array.isArray(catData) ? catData.map(normalizeCategory) : []);
        } catch (e: any) {
            toast.error(e.message || "Ошибка загрузки справочника ПО");
        } finally {
            setLoading(false);
        }
    };

    const filteredSoftware = software.filter((sw) => {
        const name = sw.name || "";
        const vendor = sw.vendor || "";
        const matchesSearch = name.toLowerCase().includes(searchTerm.toLowerCase()) || vendor.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesCategory = !categoryFilter || sw.category_id?.toString() === categoryFilter;
        const matchesRussian = !russianOnly || sw.is_russian;
        const matchesCertified = !certifiedOnly || sw.fstec_certified || sw.fsb_certified;
        return matchesSearch && matchesCategory && matchesRussian && matchesCertified;
    });

    const stats = {
        total: software.length,
        russian: software.filter((s) => s.is_russian).length,
        fstecCertified: software.filter((s) => s.fstec_certified).length,
        fsbCertified: software.filter((s) => s.fsb_certified).length,
    };

    return (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }}>
            <div style={{ marginBottom: "32px" }}>
                <h1>Справочник программного обеспечения</h1>
                <p style={{ color: "var(--ink-muted)" }}>
                    Корпоративный каталог ПО с информацией о сертификации и включении в реестр
                </p>
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: "16px", marginBottom: "24px" }}>
                <div className="card" style={{ padding: "16px" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--ink)" }}>{stats.total}</div>
                    <div style={{ fontSize: "13px", color: "var(--ink-muted)", fontWeight: 500 }}>Всего ПО</div>
                </div>
                <div className="card" style={{ padding: "16px" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--success)" }}>{stats.russian}</div>
                    <div style={{ fontSize: "13px", color: "var(--ink-muted)", fontWeight: 500 }}>В реестре РФ</div>
                </div>
                <div className="card" style={{ padding: "16px" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--warning)" }}>{stats.fstecCertified}</div>
                    <div style={{ fontSize: "13px", color: "var(--ink-muted)", fontWeight: 500 }}>Сертификат ФСТЭК</div>
                </div>
                <div className="card" style={{ padding: "16px" }}>
                    <div style={{ fontSize: "28px", fontWeight: 700, color: "var(--danger)" }}>{stats.fsbCertified}</div>
                    <div style={{ fontSize: "13px", color: "var(--ink-muted)", fontWeight: 500 }}>Сертификат ФСБ</div>
                </div>
            </div>

            <div className="card" style={{ padding: "20px", marginBottom: "24px" }}>
                <div style={{ display: "grid", gridTemplateColumns: "2fr 1fr auto auto", gap: "16px", alignItems: "end" }}>
                    <div>
                        <label className="form-label">Поиск</label>
                        <div style={{ position: 'relative' }}>
                            <Search size={16} style={{ position: 'absolute', left: 12, top: 12, color: 'var(--ink-muted)' }} />
                            <input
                                type="text"
                                className="form-input"
                                style={{ paddingLeft: 36 }}
                                placeholder="Название или производитель..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                            />
                        </div>
                    </div>
                    <div>
                        <label className="form-label">Категория</label>
                        <select className="form-input" value={categoryFilter} onChange={(e) => setCategoryFilter(e.target.value)}>
                            <option value="">Все категории</option>
                            {categories.map((cat) => (
                                <option key={cat.id} value={cat.id}>{cat.name}</option>
                            ))}
                        </select>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: "8px", paddingBottom: "10px" }}>
                        <input type="checkbox" id="russianOnly" checked={russianOnly} onChange={(e) => setRussianOnly(e.target.checked)} style={{ width: 16, height: 16 }} />
                        <label htmlFor="russianOnly" style={{ cursor: "pointer", fontWeight: 500 }}>Реестр РФ</label>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: "8px", paddingBottom: "10px" }}>
                        <input type="checkbox" id="certifiedOnly" checked={certifiedOnly} onChange={(e) => setCertifiedOnly(e.target.checked)} style={{ width: 16, height: 16 }} />
                        <label htmlFor="certifiedOnly" style={{ cursor: "pointer", fontWeight: 500 }}>Сертифицировано</label>
                    </div>
                </div>
            </div>

            <div className="card" style={{ overflow: "hidden" }}>
                {loading ? (
                    <div style={{ padding: '60px', textAlign: 'center' }}>
                        <div className="loading-spinner" />
                    </div>
                ) : (
                    <table style={{ width: "100%", borderCollapse: "collapse", textAlign: "left" }}>
                        <thead style={{ background: "var(--well)", borderBottom: "1px solid var(--perimeter)" }}>
                            <tr>
                                <th style={{ padding: "12px 16px", fontWeight: 600, fontSize: "13px", color: "var(--ink-secondary)" }}>ПО</th>
                                <th style={{ padding: "12px 16px", fontWeight: 600, fontSize: "13px", color: "var(--ink-secondary)" }}>Производитель</th>
                                <th style={{ padding: "12px 16px", fontWeight: 600, fontSize: "13px", color: "var(--ink-secondary)" }}>Категория</th>
                                <th style={{ padding: "12px 16px", fontWeight: 600, fontSize: "13px", color: "var(--ink-secondary)", textAlign: "center" }}>Сертификация</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredSoftware.length === 0 ? (
                                <tr><td colSpan={4} style={{ padding: '40px', textAlign: 'center', color: 'var(--ink-muted)' }}>Ничего не найдено</td></tr>
                            ) : filteredSoftware.map(sw => (
                                <tr key={sw.id} style={{ borderBottom: "1px solid var(--perimeter)", background: 'var(--raised)' }}>
                                    <td style={{ padding: "12px 16px" }}>
                                        <div style={{ fontWeight: 600, color: 'var(--ink)' }}>{sw.name}</div>
                                        {sw.version && <div style={{ fontSize: "12px", color: "var(--ink-muted)" }}>v{sw.version}</div>}
                                    </td>
                                    <td style={{ padding: "12px 16px", fontSize: "14px", color: "var(--ink-secondary)" }}>{sw.vendor}</td>
                                    <td style={{ padding: "12px 16px" }}>
                                        <span className="badge" style={{ background: 'var(--well)', color: 'var(--ink-tertiary)' }}>{sw.category_name || '—'}</span>
                                    </td>
                                    <td style={{ padding: "12px 16px", display: 'flex', gap: '6px', justifyContent: 'center' }}>
                                        {sw.is_russian && <span className="badge badge-low" title={`Реестр №${sw.registry_number}`}>РФ</span>}
                                        {sw.fstec_certified && <span className="badge badge-medium" title={`Класс ${sw.fstec_protection_class}`}>ФСТЭК</span>}
                                        {sw.fsb_certified && <span className="badge badge-critical" title={`Класс ${sw.fsb_protection_class}`}>ФСБ</span>}
                                        {!sw.is_russian && !sw.fstec_certified && !sw.fsb_certified && <span style={{ color: 'var(--ink-muted)' }}>—</span>}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </motion.div>
    );
};
