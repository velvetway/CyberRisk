-- Перестраиваем asset_vulnerabilities из «связь с пустой legacy-таблицей
-- vulnerabilities» в честный инвентарь CVE/БДУ-записей, обнаруженных на
-- активе. Каждая строка — конкретная БДУ-уязвимость, привязанная к VL-
-- категории (для формулы W) и опционально к ПО, через которое её нашли
-- (для каскадного удаления при отвязке ПО).
--
-- Источники записи (поле source):
--   'auto:asset_software' — автодетекция при добавлении ПО на актив.
--   'manual'             — ручное добавление оператором через UI.
--
-- Уникальность: один и тот же БДУ-id может быть привязан к активу
-- несколько раз — но только если идёт от разных software_id (либо из
-- разных «ручных» источников). Поэтому UNIQUE на (asset_id, bdu_id, source, software_id).

-- Таблица сейчас пустая (truncate в 030). Безопасно перестраивать.
ALTER TABLE asset_vulnerabilities
    DROP CONSTRAINT IF EXISTS asset_vulnerabilities_vulnerability_id_fkey;
ALTER TABLE asset_vulnerabilities
    DROP CONSTRAINT IF EXISTS uniq_asset_vulnerability;
ALTER TABLE asset_vulnerabilities
    DROP COLUMN vulnerability_id;

ALTER TABLE asset_vulnerabilities
    ADD COLUMN bdu_id          VARCHAR(32),
    ADD COLUMN cve             VARCHAR(64),
    ADD COLUMN cwe             VARCHAR(16),
    ADD COLUMN vl_category_id  SMALLINT REFERENCES vl_categories(id) ON DELETE SET NULL,
    ADD COLUMN cvss_score      NUMERIC(3,1),
    ADD COLUMN severity_level  SMALLINT,
    ADD COLUMN title           TEXT,
    ADD COLUMN source          VARCHAR(32) NOT NULL DEFAULT 'manual',
    ADD COLUMN software_id     BIGINT REFERENCES software_catalog(id) ON DELETE CASCADE,
    ADD COLUMN discovered_at   TIMESTAMPTZ NOT NULL DEFAULT now();

-- bdu_id обязателен для всех новых записей. Делаем NOT NULL отдельно
-- от ADD COLUMN, чтобы старые (если бы были) пустые строки прошли.
ALTER TABLE asset_vulnerabilities
    ALTER COLUMN bdu_id SET NOT NULL;

CREATE UNIQUE INDEX uniq_asset_vuln_bdu_source ON asset_vulnerabilities
    (asset_id, bdu_id, source, COALESCE(software_id, 0));

CREATE INDEX idx_asset_vuln_bdu       ON asset_vulnerabilities(bdu_id);
CREATE INDEX idx_asset_vuln_vl        ON asset_vulnerabilities(vl_category_id);
CREATE INDEX idx_asset_vuln_software  ON asset_vulnerabilities(software_id);
