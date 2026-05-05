-- Stage 2 of the legacy purge: bring the schema down to exactly what the
-- PTSZI W formula needs. See docs/risk-model.md.

-- 1. Drop legacy risk-scenario tables (the old impact×likelihood engine).
DROP TABLE IF EXISTS risk_scenario_recommendations;
DROP TABLE IF EXISTS recommendation_templates;
DROP TABLE IF EXISTS risk_scenarios;

-- 2. assets — strip every column not used by the W formula or basic UI labelling.
DROP INDEX IF EXISTS idx_assets_data_category;
DROP INDEX IF EXISTS idx_assets_kii;
DROP INDEX IF EXISTS idx_assets_protection_level;
DROP INDEX IF EXISTS idx_assets_type;

ALTER TABLE assets
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS business_criticality,
    DROP COLUMN IF EXISTS confidentiality,
    DROP COLUMN IF EXISTS integrity,
    DROP COLUMN IF EXISTS availability,
    DROP COLUMN IF EXISTS data_category,
    DROP COLUMN IF EXISTS protection_level,
    DROP COLUMN IF EXISTS kii_category,
    DROP COLUMN IF EXISTS has_personal_data,
    DROP COLUMN IF EXISTS personal_data_volume,
    DROP COLUMN IF EXISTS has_internet_access;

-- ENUMs that no column references any more.
DROP TYPE IF EXISTS data_category_type;
DROP TYPE IF EXISTS protection_level;
DROP TYPE IF EXISTS kii_category;
DROP TYPE IF EXISTS risk_status;
DROP TYPE IF EXISTS risk_level;
DROP TYPE IF EXISTS recommendation_status;

-- 3. threats — q_threat / q_severity already cover everything the W formula needs.
ALTER TABLE threats
    DROP COLUMN IF EXISTS base_likelihood,
    DROP COLUMN IF EXISTS attack_vector,
    DROP COLUMN IF EXISTS impact_confidentiality,
    DROP COLUMN IF EXISTS impact_integrity,
    DROP COLUMN IF EXISTS impact_availability;

-- 4. controls — coverage now lives entirely in vulnerability_controls.coverage.
ALTER TABLE controls
    DROP COLUMN IF EXISTS reduces_likelihood_by,
    DROP COLUMN IF EXISTS reduces_impact_by;

-- 5. asset_controls.effectiveness was a legacy weight; Q^reaction uses VL coverage.
ALTER TABLE asset_controls
    DROP COLUMN IF EXISTS effectiveness;
