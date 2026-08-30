-- Откат схемы asset_vulnerabilities. Реальные данные (БДУ-инвентарь)
-- будут потеряны — но в legacy-схеме их и не было, таблица была пустой.

DROP INDEX IF EXISTS idx_asset_vuln_software;
DROP INDEX IF EXISTS idx_asset_vuln_vl;
DROP INDEX IF EXISTS idx_asset_vuln_bdu;
DROP INDEX IF EXISTS uniq_asset_vuln_bdu_source;

ALTER TABLE asset_vulnerabilities
    DROP COLUMN discovered_at,
    DROP COLUMN software_id,
    DROP COLUMN source,
    DROP COLUMN title,
    DROP COLUMN severity_level,
    DROP COLUMN cvss_score,
    DROP COLUMN vl_category_id,
    DROP COLUMN cwe,
    DROP COLUMN cve,
    DROP COLUMN bdu_id;

ALTER TABLE asset_vulnerabilities
    ADD COLUMN vulnerability_id BIGINT NOT NULL REFERENCES vulnerabilities(id) ON DELETE CASCADE;
ALTER TABLE asset_vulnerabilities
    ADD CONSTRAINT uniq_asset_vulnerability UNIQUE (asset_id, vulnerability_id);
