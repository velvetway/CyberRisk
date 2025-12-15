-- Откат добавления колонок type и location
DROP INDEX IF EXISTS idx_assets_type;

ALTER TABLE assets
    DROP COLUMN IF EXISTS location;

ALTER TABLE assets
    DROP COLUMN IF EXISTS type;
