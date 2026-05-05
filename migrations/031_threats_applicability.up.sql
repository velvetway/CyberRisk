-- Restore the C/I/A impact flags on threats (we dropped them in 020) and add
-- the applicability fields needed for the FSTEC importer.
--
-- The C/I/A flags now serve a different purpose than they did in the legacy
-- engine: they feed `q_severity` derivation (count of affected CIA axes / 3)
-- and the threat ↔ DA mapping.
--
-- `applies_to_targets` keeps the raw "Объект воздействия" text from the FSTEC
-- catalogue for human reference.
-- `applies_to_asset_types` is the derived list of asset_type IDs computed by
-- the importer from the raw target text via a regex dictionary.
-- `status` carries the FSTEC publication status verbatim ("Опубликована" /
-- "В работе" / …).

ALTER TABLE threats
    ADD COLUMN IF NOT EXISTS applies_to_targets    TEXT,
    ADD COLUMN IF NOT EXISTS applies_to_asset_types SMALLINT[],
    ADD COLUMN IF NOT EXISTS impact_c              BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS impact_i              BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS impact_a              BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS status                VARCHAR(64);

-- GIN index lets us cheaply ask "which threats apply to asset_type=N".
CREATE INDEX IF NOT EXISTS idx_threats_applies_types ON threats USING GIN (applies_to_asset_types);
CREATE INDEX IF NOT EXISTS idx_threats_status        ON threats(status);
