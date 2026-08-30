// Domain types for the PTSZI W-model — see docs/risk-model.md.
// Legacy fields (КИИ, ПДн, УЗ-1..4, C/I/A, business_criticality, impact_*,
// base_likelihood, attack_vector, regulatory_factor, …) are intentionally
// absent — they were dropped in the schema migration that aligned the data
// model with the thesis.

export interface Asset {
    id: number;
    name: string;
    asset_type_id?: number;
    owner?: string;
    description?: string;
    environment: string;
    is_isolated: boolean;
    tags?: Record<string, unknown>;
    created_at: string;
    updated_at: string;
}

export interface Threat {
    id: number;
    name: string;
    threat_category_id?: number;
    source_type: string;
    description?: string;
    q_threat: number;   // ∈ [0, 1]
    q_severity: number; // ∈ [0, 1]
    bdu_id?: string;
    created_at: string;
    updated_at: string;
}

// One row of the global PTSZI risk overview map.
export interface RiskOverviewPoint {
    asset_id: number;
    asset_name: string;
    threat_id: number;
    threat_name: string;
    w: number;
    level: string;
    q_threat: number;
    q_severity: number;
    q_reaction: number;
    z: number;
}

// Software catalog (unchanged).

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
    is_russian: boolean;
    registry_number?: string;
    registry_date?: string;
    registry_url?: string;
    fstec_certified: boolean;
    fstec_certificate_num?: string;
    fstec_certificate_date?: string;
    fstec_protection_class?: string;
    fstec_valid_until?: string;
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
