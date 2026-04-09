// Категории данных (152-ФЗ, 187-ФЗ)
export type DataCategory =
    | "public"
    | "internal"
    | "confidential"
    | "personal_data"
    | "personal_data_special"
    | "personal_data_biometric"
    | "kii"
    | "state_secret"
    | "banking_secret"
    | "medical_secret"
    | "commercial_secret";

// Уровень защищённости ПДн (ПП-1119)
export type ProtectionLevel = "uz1" | "uz2" | "uz3" | "uz4";

// Категория КИИ (187-ФЗ)
export type KIICategory = "none" | "cat3" | "cat2" | "cat1";

export interface Asset {
    id: number;
    name: string;
    type: string;
    asset_type_id: number;
    owner: string;
    description: string;
    location: string;
    business_criticality: number;
    confidentiality: number;
    integrity: number;
    availability: number;
    environment: string;
    tags: Record<string, string>;
    // Регуляторные поля
    data_category?: DataCategory;
    protection_level?: ProtectionLevel;
    kii_category?: KIICategory;
    has_personal_data?: boolean;
    personal_data_volume?: string;
    has_internet_access?: boolean;
    is_isolated?: boolean;
    created_at: string;
    updated_at: string;
}

export interface Threat {
    id: number;
    name: string;
    description: string;
    base_likelihood: number;
    // БДУ ФСТЭК
    bdu_id?: string;
    attack_vector?: string;
    impact_confidentiality?: boolean;
    impact_integrity?: boolean;
    impact_availability?: boolean;
    source_type?: string;
    created_at: string;
    updated_at: string;
}

// Рекомендация по снижению риска
export interface RiskRecommendation {
    code: string;
    title: string;
    description: string;
    priority: "low" | "medium" | "high";
    category: string;
}

// Российское СЗИ
export interface RussianTool {
    name: string;
    vendor: string;
    description: string;
    fstec_certified: boolean;
    fsb_certified: boolean;
    registry_number: string;
    protection_class: string;
    use_case: string;
}

// Регуляторная рекомендация
export interface RegulatoryRecommendation extends RiskRecommendation {
    regulation: string;
    requirement: string;
    russian_tools: RussianTool[];
    deadline: string;
    penalty: string;
}

export interface RiskPreviewRequest {
    asset_id: number;
    threat_id: number;
}

export interface RiskPreviewResponse {
    asset_id: number;
    threat_id: number;
    impact: number;
    likelihood: number;
    score: number;
    level: string;
    regulatory_factor?: number;
    adjusted_score?: number;
    recommendations: RiskRecommendation[];
}

export interface RiskOverviewPoint {
    asset_id: number;
    asset_name: string;
    threat_id: number;
    threat_name: string;
    impact: number;
    likelihood: number;
    score: number;
    level: string;
    regulatory_factor?: number;
    adjusted_score?: number;
}

// Справочник ПО
export interface SoftwareCategory {
    id: number;
    code: string;
    name: string;
    description: string;
}

export interface Software {
    id: number;
    name: string;
    vendor: string;
    version?: string;
    category_id?: number;
    category_name?: string;
    // Реестр Минцифры
    is_russian: boolean;
    registry_number?: string;
    registry_date?: string;
    registry_url?: string;
    // ФСТЭК
    fstec_certified: boolean;
    fstec_certificate_num?: string;
    fstec_certificate_date?: string;
    fstec_protection_class?: string;
    fstec_valid_until?: string;
    // ФСБ
    fsb_certified: boolean;
    fsb_certificate_num?: string;
    fsb_protection_class?: string;
    description?: string;
    website?: string;
    created_at: string;
    updated_at: string;
}

export interface AssetSoftware {
    id: number;
    asset_id: number;
    software_id: number;
    software?: Software;
    version?: string;
    install_date?: string;
    license_type?: string;
    license_expires?: string;
    notes?: string;
}

export interface AssetSoftwareAlternative {
    asset_software_id: number;
    asset_id: number;
    version?: string;
    notes?: string;
    software: Software;
    alternatives: Software[];
}
