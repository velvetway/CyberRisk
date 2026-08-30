// src/api/client.ts
import {
    Asset,
    Threat,
    RiskOverviewPoint,
    Software,
    SoftwareCategory,
    AssetSoftwareAlternative,
} from "../types";
import {
    PtsziAssetProfile,
    PtsziAttackPath,
    PtsziControl,
    PtsziThreat,
    PtsziUBIThreat,
    PtsziVulnerableLink,
    ThreatSource,
    DestructiveAction,
} from "../types/ptszi";
import { AssetScale, OptimizerPlan, Roadmap, SensitivityReport } from "../types/optimizer";

function getToken(): string | null {
    return localStorage.getItem("token");
}

/** Drop-in replacement for fetch() that adds Authorization header */
export async function authFetch(input: string, init?: RequestInit): Promise<Response> {
    const token = getToken();
    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        ...(init?.headers as Record<string, string> || {}),
    };
    if (token) {
        headers["Authorization"] = `Bearer ${token}`;
    }
    const res = await fetch(input, { ...init, headers });
    if (res.status === 401) {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        window.dispatchEvent(new Event("auth:logout"));
        throw new Error("Сессия истекла. Войдите снова.");
    }
    return res;
}

async function request<T>(input: string, init?: RequestInit): Promise<T> {
    const token = getToken();
    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        ...(init?.headers as Record<string, string> || {}),
    };
    if (token) {
        headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(input, {
        ...init,
        headers,
    });

    if (res.status === 401) {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        window.dispatchEvent(new Event("auth:logout"));
        throw new Error("Сессия истекла. Войдите снова.");
    }

    if (!res.ok) {
        let message = `HTTP ${res.status}`;
        try {
            const data = await res.json();
            if (data && typeof data.error === "string") {
                message = data.error;
            }
        } catch {
            // тело пустое или не JSON — оставляем дефолт
        }
        throw new Error(message);
    }

    if (res.status === 204) {
        return undefined as T;
    }

    // Fiber's SendStatus(201) emits body "Created" (plain text), а мы зовём
    // эту функцию для всех типов endpoint'ов — JSON и не-JSON. Если ответ
    // не объявлен как application/json, не пытаемся парсить.
    const ct = res.headers.get("content-type") || "";
    if (!ct.includes("application/json")) {
        return undefined as T;
    }

    const text = await res.text();
    if (!text) {
        return undefined as T;
    }

    return JSON.parse(text) as T;
}

// Auth API (no token needed)
export interface LoginResponse {
    token: string;
}

export interface UserResponse {
    id: number;
    username: string;
    role: string;
    is_active: boolean;
}

export async function loginAPI(username: string, password: string): Promise<LoginResponse> {
    const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
    });
    if (!res.ok) {
        let message = "Ошибка входа";
        try {
            const data = await res.json();
            if (data?.error) message = data.error;
        } catch {}
        throw new Error(message);
    }
    return res.json();
}

export async function registerAPI(username: string, password: string, role: string): Promise<UserResponse> {
    const res = await fetch("/api/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password, role }),
    });
    if (!res.ok) {
        let message = "Ошибка регистрации";
        try {
            const data = await res.json();
            if (data?.error) message = data.error;
        } catch {}
        throw new Error(message);
    }
    return res.json();
}

export const api = {
    // Assets
    getAssets(): Promise<Asset[]> {
        return request<Asset[]>("/api/assets");
    },

    getAsset(id: number): Promise<Asset> {
        return request<Asset>(`/api/assets/${id}`);
    },

    createAsset(asset: Partial<Asset>): Promise<Asset> {
        return request<Asset>("/api/assets", {
            method: "POST",
            body: JSON.stringify(asset),
        });
    },

    updateAsset(id: number, asset: Partial<Asset>): Promise<Asset> {
        return request<Asset>(`/api/assets/${id}`, {
            method: "PUT",
            body: JSON.stringify(asset),
        });
    },

    deleteAsset(id: number): Promise<void> {
        return request<void>(`/api/assets/${id}`, {
            method: "DELETE",
        });
    },

    async getAssetSoftwareAlternatives(id: number): Promise<AssetSoftwareAlternative[]> {
        try {
            return await request<AssetSoftwareAlternative[]>(`/api/assets/${id}/software/alternatives`);
        } catch (err) {
            return request<AssetSoftwareAlternative[]>(`/api/software/asset/${id}/alternatives`);
        }
    },

    // Threats
    getThreats(): Promise<Threat[]> {
        return request<Threat[]>("/api/threats");
    },

    // Risk (PTSZI W-model)
    getRiskOverview(): Promise<RiskOverviewPoint[]> {
        return request<RiskOverviewPoint[]>("/api/risk/overview");
    },

    // Software catalog
    getSoftware(): Promise<Software[]> {
        return request<Software[]>("/api/software");
    },

    getSoftwareCategories(): Promise<SoftwareCategory[]> {
        return request<SoftwareCategory[]>("/api/software/categories");
    },

    getRussianSoftware(): Promise<Software[]> {
        return request<Software[]>("/api/software/russian");
    },

    getCertifiedSoftware(): Promise<Software[]> {
        return request<Software[]>("/api/software/certified");
    },

    getSoftwareAlternatives(id: number): Promise<Software[]> {
        return request<Software[]>(`/api/software/${id}/alternatives`);
    },

    searchSoftware(params: { category?: string; russian?: boolean; certified?: boolean }): Promise<Software[]> {
        const query = new URLSearchParams();
        if (params.category) query.set("category", params.category);
        if (params.russian) query.set("russian", "true");
        if (params.certified) query.set("certified", "true");
        return request<Software[]>(`/api/software?${query.toString()}`);
    },

    // Canonical PTSZI model
    getPtsziSources(): Promise<ThreatSource[]> {
        return request<ThreatSource[]>("/api/ptszi/sources");
    },

    getPtsziThreats(): Promise<PtsziThreat[]> {
        return request<PtsziThreat[]>("/api/ptszi/threats");
    },

    getPtsziVulnerableLinks(): Promise<PtsziVulnerableLink[]> {
        return request<PtsziVulnerableLink[]>("/api/ptszi/vulnerable-links");
    },

    /**
     * Подбор комплекса средств в рамках единого бюджета.
     *
     * maxClass ограничивает выбор классом защиты ФСТЭК: средства слабее
     * недопустимы для систем высокого класса защищённости.
     */
    optimizeAsset(
        assetId: number,
        budget: number,
        scale: AssetScale,
        maxClass?: number,
    ): Promise<OptimizerPlan> {
        const params = new URLSearchParams({
            budget: String(budget),
            workstations: String(scale.workstations),
            servers: String(scale.servers),
            appliances: String(scale.appliances),
        });
        if (maxClass) params.set("max_class", String(maxClass));
        return request<OptimizerPlan>(`/api/ptszi/assets/${assetId}/optimize?${params}`);
    },

    /** План внедрения на несколько лет при годовом бюджете. */
    getAssetRoadmap(
        assetId: number,
        budgetPerYear: number,
        years: number,
        scale: AssetScale,
        maxClass?: number,
        discountRate = 0,
        degradationRate = 0,
    ): Promise<Roadmap> {
        const params = new URLSearchParams({
            budget_per_year: String(budgetPerYear),
            years: String(years),
            workstations: String(scale.workstations),
            servers: String(scale.servers),
            appliances: String(scale.appliances),
        });
        if (maxClass) params.set("max_class", String(maxClass));
        if (discountRate > 0) params.set("discount_rate", String(discountRate));
        if (degradationRate > 0) params.set("degradation_rate", String(degradationRate));
        return request<Roadmap>(`/api/ptszi/assets/${assetId}/roadmap?${params}`);
    },

    /**
     * Проверка устойчивости плана к неточности экспертных коэффициентов.
     *
     * Коэффициенты покрытия и эффективности задаются оценочно, поэтому
     * важно знать, меняется ли состав закупки при их разумном сдвиге.
     */
    getAssetSensitivity(
        assetId: number,
        budget: number,
        scale: AssetScale,
        runs = 300,
        variation = 0.2,
        maxClass?: number,
    ): Promise<SensitivityReport> {
        const params = new URLSearchParams({
            budget: String(budget),
            workstations: String(scale.workstations),
            servers: String(scale.servers),
            appliances: String(scale.appliances),
            runs: String(runs),
            variation: String(variation),
        });
        if (maxClass) params.set("max_class", String(maxClass));
        return request<SensitivityReport>(`/api/ptszi/assets/${assetId}/sensitivity?${params}`);
    },

    getPtsziControls(): Promise<PtsziControl[]> {
        return request<PtsziControl[]>("/api/ptszi/controls");
    },

    getPtsziDestructiveActions(): Promise<DestructiveAction[]> {
        return request<DestructiveAction[]>("/api/ptszi/destructive-actions");
    },

    getPtsziUBI(params?: { limit?: number; offset?: number; q?: string }): Promise<PtsziUBIThreat[]> {
        const query = new URLSearchParams();
        if (params?.limit) query.set("limit", String(params.limit));
        if (params?.offset) query.set("offset", String(params.offset));
        if (params?.q) query.set("q", params.q);
        return request<PtsziUBIThreat[]>(`/api/ptszi/ubi?${query.toString()}`);
    },

    getAssetPtsziProfile(assetId: number): Promise<PtsziAssetProfile> {
        return request<PtsziAssetProfile>(`/api/assets/${assetId}/ptszi/profile`);
    },

    updateAssetPtsziVulnerableLinks(assetId: number, ids: number[]): Promise<void> {
        return request<void>(`/api/assets/${assetId}/ptszi/vulnerable-links`, {
            method: "PUT",
            body: JSON.stringify({ vulnerable_link_ids: ids }),
        });
    },

    updateAssetPtsziControls(assetId: number, controls: Array<{ control_id: number; effectiveness: number }>): Promise<void> {
        return request<void>(`/api/assets/${assetId}/ptszi/controls`, {
            method: "PUT",
            body: JSON.stringify({ controls }),
        });
    },

    getApplicablePtsziThreats(assetId: number): Promise<PtsziAttackPath[]> {
        return request<PtsziAttackPath[]>(`/api/ptszi/assets/${assetId}/threats`);
    },

    getPtsziGraph(assetId: number, threatId: number): Promise<PtsziAttackPath> {
        return request<PtsziAttackPath>(`/api/ptszi/graph/${assetId}/${threatId}`);
    },

    // P8: Asset detail page — установленное ПО, инвентарь CVE, контроли.

    getAssetSoftware(assetID: number): Promise<AssetSoftwareLink[]> {
        return request<AssetSoftwareLink[]>(`/api/assets/${assetID}/software`);
    },

    attachSoftware(assetID: number, softwareID: number, version?: string): Promise<{ detected_vulnerabilities: number }> {
        return request<{ detected_vulnerabilities: number }>(`/api/assets/${assetID}/software`, {
            method: "POST",
            body: JSON.stringify({ software_id: softwareID, version }),
        });
    },

    detachSoftware(assetID: number, softwareID: number): Promise<void> {
        return request<void>(`/api/assets/${assetID}/software/${softwareID}`, { method: "DELETE" });
    },

    getAssetVulnerabilities(assetID: number): Promise<AssetVulnerability[]> {
        return request<AssetVulnerability[]>(`/api/assets/${assetID}/vulnerabilities`);
    },

    addManualVulnerability(assetID: number, bduID: string, vlCategoryID?: number): Promise<void> {
        return request<void>(`/api/assets/${assetID}/vulnerabilities`, {
            method: "POST",
            body: JSON.stringify({ bdu_id: bduID, vl_category_id: vlCategoryID }),
        });
    },

    deleteAssetVulnerability(assetID: number, vulnID: number): Promise<void> {
        return request<void>(`/api/assets/${assetID}/vulnerabilities/${vulnID}`, { method: "DELETE" });
    },

    getControls(): Promise<Control[]> {
        return request<Control[]>("/api/controls");
    },

    getAssetControls(assetID: number): Promise<Control[]> {
        return request<Control[]>(`/api/assets/${assetID}/controls`);
    },

    attachControl(assetID: number, controlID: number): Promise<void> {
        return request<void>(`/api/assets/${assetID}/controls`, {
            method: "POST",
            body: JSON.stringify({ control_id: controlID }),
        });
    },

    detachControl(assetID: number, controlID: number): Promise<void> {
        return request<void>(`/api/assets/${assetID}/controls/${controlID}`, { method: "DELETE" });
    },

    getAssetAttackPaths(assetID: number): Promise<AssetAttackPathsResponse> {
        return request<AssetAttackPathsResponse>(`/api/risk/asset/${assetID}/attack-paths`);
    },

    // P9: каталог угроз ФСТЭК и справочники.
    getThreatsAll(limit = 500, offset = 0): Promise<ThreatFull[]> {
        return request<ThreatFull[]>(`/api/threats?limit=${limit}&offset=${offset}`);
    },
    updateThreat(id: number, payload: ThreatUpdatePayload): Promise<ThreatFull> {
        return request<ThreatFull>(`/api/threats/${id}`, {
            method: "PUT",
            body: JSON.stringify(payload),
        });
    },
    getAssetTypes(): Promise<AssetTypeRef[]> {
        return request<AssetTypeRef[]>("/api/asset-types");
    },
    getVLCategories(): Promise<VLCategoryRef[]> {
        return request<VLCategoryRef[]>("/api/vl-categories");
    },

    // Compliance / ОСЗ — оценка состояния защищённости
    getComplianceStandards(): Promise<ComplianceStandard[]> {
        return request<ComplianceStandard[]>("/api/compliance/standards");
    },
    getAssetCompliance(assetID: number): Promise<AssetComplianceOverview[]> {
        return request<AssetComplianceOverview[]>(`/api/compliance/asset/${assetID}`);
    },
    getAssetComplianceDetail(assetID: number, standardCode: string): Promise<AssetStandardCompliance> {
        return request<AssetStandardCompliance>(`/api/compliance/asset/${assetID}/standard/${standardCode}`);
    },

    // Organization-level
    getOrganizationOverview(): Promise<OrganizationOverview> {
        return request<OrganizationOverview>("/api/organization/overview");
    },
    getOrganizationAssetMatrix(): Promise<OrganizationAssetRow[]> {
        return request<OrganizationAssetRow[]>("/api/organization/asset-matrix");
    },
    getOrganizationCriticalRisks(limit = 20): Promise<OrganizationCriticalRisk[]> {
        return request<OrganizationCriticalRisk[]>(`/api/organization/critical-risks?limit=${limit}`);
    },
};

// ---------- P8 типы ----------

export interface AssetSoftwareLink {
    link: {
        id: number;
        asset_id: number;
        software_id: number;
        version?: string;
        install_date?: string;
        license_type?: string;
        license_expires?: string;
        notes?: string;
        created_at: string;
        updated_at: string;
    };
    software: Software;
}

export interface AssetVulnerability {
    id: number;
    asset_id: number;
    bdu_id: string;
    cve?: string;
    cwe?: string;
    vl_category_id?: number;
    cvss_score?: number;
    severity_level?: number;
    title?: string;
    source: string; // "auto:asset_software" | "manual"
    software_id?: number;
    status: string;
    discovered_at: string;
    created_at: string;
    updated_at: string;
}

export interface Control {
    id: number;
    name: string;
    control_type_id?: number;
    description?: string;
    created_at: string;
    updated_at: string;
}

export interface AssetAttackPathsResponse {
    asset: { id: number; name: string };
    aggregate: {
        w_max: number;
        level: string;
        threat_count: number;
        uncovered_count: number;
    };
    paths: unknown[]; // подробные тип в types/riskGraph.ts; здесь не нужны
}

// ---------- P9 типы ----------

export interface ThreatFull {
    id: number;
    name: string;
    threat_category_id?: number;
    source_type: string; // "external" | "internal" | "third_party"
    description?: string;
    q_threat: number;
    q_severity: number;
    bdu_id?: string;
    applies_to_targets?: string;
    applies_to_asset_types?: number[];
    impact_c: boolean;
    impact_i: boolean;
    impact_a: boolean;
    status?: string;
    created_at: string;
    updated_at: string;
}

export interface ThreatUpdatePayload {
    name: string;
    threat_category_id?: number | null;
    source_type: string;
    description?: string | null;
    q_threat: number;
    q_severity: number;
    bdu_id?: string | null;
    applies_to_targets?: string | null;
    applies_to_asset_types?: number[];
    impact_c: boolean;
    impact_i: boolean;
    impact_a: boolean;
}

export interface AssetTypeRef {
    id: number;
    name: string;
    description?: string;
}

export interface VLCategoryRef {
    id: number;
    code: string;
    name: string;
    description?: string;
}

// ---------- Compliance ----------

export interface ComplianceStandard {
    id: number;
    code: string;
    name: string;
    full_name: string;
    jurisdiction: "RU" | "INT";
    description?: string;
    sort_order: number;
}

export interface ComplianceRequirement {
    id: number;
    standard_id: number;
    code: string;
    category: string;
    title: string;
    description?: string;
    priority: 1 | 2 | 3;
    sort_order: number;
}

export interface RequirementStatus {
    requirement: ComplianceRequirement;
    coverage: number;
    covering_controls?: Control[];
    missing_controls?: Control[];
}

export interface AssetComplianceOverview {
    standard: ComplianceStandard;
    overall_score: number;
    covered_count: number;
    partial_count: number;
    uncovered_count: number;
    total_count: number;
}

export interface AssetStandardCompliance extends AssetComplianceOverview {
    requirements: RequirementStatus[];
}

// ---------- Organization-level ----------

export interface OrganizationAssetTypeBucket {
    type_id?: number;
    type_name: string;
    count: number;
}

export interface OrganizationComplianceSummary {
    standard: ComplianceStandard;
    avg_score: number;
    min_score: number;
    max_score: number;
    assets_count: number;
}

export interface OrganizationOverview {
    total_assets: number;
    isolated_assets: number;
    assets_by_environment: Record<string, number>;
    assets_by_type: OrganizationAssetTypeBucket[];
    risk_distribution: Record<string, number>;
    w_max: number;
    w_max_asset?: string;
    w_max_threat?: string;
    avg_w_per_asset: number;
    total_controls: number;
    uncovered_vls: number;
    compliance_by_standard: OrganizationComplianceSummary[];
}

export interface OrganizationAssetRow {
    asset_id: number;
    name: string;
    type_name?: string;
    environment?: string;
    is_isolated: boolean;
    w_max: number;
    level: string;
    threat_count: number;
    control_count: number;
    compliance_by_standard: AssetComplianceOverview[];
}

export interface OrganizationCriticalRisk {
    asset_id: number;
    asset_name: string;
    threat_id: number;
    threat_name: string;
    bdu_id?: string;
    w: number;
    level: string;
}
