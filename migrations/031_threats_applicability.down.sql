DROP INDEX IF EXISTS idx_threats_applies_types;
DROP INDEX IF EXISTS idx_threats_status;

ALTER TABLE threats
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS impact_a,
    DROP COLUMN IF EXISTS impact_i,
    DROP COLUMN IF EXISTS impact_c,
    DROP COLUMN IF EXISTS applies_to_asset_types,
    DROP COLUMN IF EXISTS applies_to_targets;
