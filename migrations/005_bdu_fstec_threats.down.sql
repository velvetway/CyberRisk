-- Откат миграции 005

DELETE FROM threats WHERE bdu_id IS NOT NULL;

ALTER TABLE threats
    DROP COLUMN IF EXISTS bdu_id,
    DROP COLUMN IF EXISTS attack_vector,
    DROP COLUMN IF EXISTS impact_confidentiality,
    DROP COLUMN IF EXISTS impact_integrity,
    DROP COLUMN IF EXISTS impact_availability;
