-- Откат миграции 004

DROP TABLE IF EXISTS asset_software;
DROP TABLE IF EXISTS software_catalog;
DROP TABLE IF EXISTS software_categories;

ALTER TABLE assets
    DROP COLUMN IF EXISTS data_category,
    DROP COLUMN IF EXISTS protection_level,
    DROP COLUMN IF EXISTS kii_category,
    DROP COLUMN IF EXISTS has_personal_data,
    DROP COLUMN IF EXISTS personal_data_volume,
    DROP COLUMN IF EXISTS has_internet_access,
    DROP COLUMN IF EXISTS is_isolated;

DROP TYPE IF EXISTS kii_category;
DROP TYPE IF EXISTS protection_level;
DROP TYPE IF EXISTS data_category_type;
