DELETE FROM vulnerability_controls;
DELETE FROM threat_vulnerable_links;
DELETE FROM threat_destructive_actions;
DELETE FROM source_threats;

ALTER TABLE threats
    DROP COLUMN IF EXISTS q_threat,
    DROP COLUMN IF EXISTS q_severity;
