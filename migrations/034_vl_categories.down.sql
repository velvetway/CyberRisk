-- Откат: возвращаем старую схему vulnerability_controls / threat_vulnerable_links
-- на FK к таблице vulnerabilities. Категорийные данные теряются (vulnerabilities
-- сейчас всё равно пустая после 030_clear_legacy_seeds).

ALTER TABLE vl_category_controls RENAME TO vulnerability_controls;
ALTER TABLE vulnerability_controls
    DROP CONSTRAINT vulnerability_controls_pkey;
ALTER TABLE vulnerability_controls
    DROP COLUMN vl_category_id,
    ADD  COLUMN vulnerability_id BIGINT NOT NULL REFERENCES vulnerabilities(id) ON DELETE CASCADE;
ALTER TABLE vulnerability_controls
    ADD PRIMARY KEY (vulnerability_id, control_id);
DROP INDEX IF EXISTS idx_vlc_control;
CREATE INDEX idx_vc_control ON vulnerability_controls(control_id);

ALTER TABLE threat_vulnerable_links
    DROP CONSTRAINT threat_vulnerable_links_pkey;
ALTER TABLE threat_vulnerable_links
    DROP COLUMN vl_category_id,
    ADD  COLUMN vulnerability_id BIGINT NOT NULL REFERENCES vulnerabilities(id) ON DELETE CASCADE;
ALTER TABLE threat_vulnerable_links
    ADD PRIMARY KEY (threat_id, vulnerability_id);
DROP INDEX IF EXISTS idx_tvl_vl_category;
CREATE INDEX idx_tvl_vuln ON threat_vulnerable_links(vulnerability_id);

DROP TABLE vl_categories;
