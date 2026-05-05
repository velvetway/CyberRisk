-- VL-категории из диплома (раздел «Уязвимые звенья»). Шесть штук: VL1..VL6.
--
-- До этой миграции `threat_vulnerable_links` и `vulnerability_controls`
-- ссылались на таблицу `vulnerabilities`, которая хранила одновременно и
-- семантические VL, и конкретные CVE. Это противоречит модели ПТСЗИ:
-- в дипломе VL — это категория уязвимого звена (всего 6 штук), а
-- конкретные CVE/БДУ-записи — отдельный «инвентарь свидетельств».
--
-- В этой миграции:
--   1. создаём `vl_categories` (6 строк).
--   2. переименовываем `vulnerability_controls` → `vl_category_controls`
--      и переключаем FK на `vl_categories`.
--   3. `threat_vulnerable_links` тоже переключаем на `vl_categories`.
--   4. таблицы `vulnerabilities` и `asset_vulnerabilities` НЕ трогаем —
--      их назначение меняется в P6 на «инвентарь CVE/БДУ» и тогда же
--      они получат поле `vl_category_id`.
--
-- Все три таблицы пусты после 030_clear_legacy_seeds, так что
-- DROP/ADD колонок безопасен.

CREATE TABLE vl_categories (
    id          SMALLSERIAL PRIMARY KEY,
    code        VARCHAR(8)  NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    description TEXT
);

INSERT INTO vl_categories (code, name, description) VALUES
    ('VL1', 'Нештатное дополнительное ПО',
        'Драйверы, утилиты, неавторизованные расширения, скрипты'),
    ('VL2', 'Устаревшие версии ПО или версии, имеющие уязвимости',
        'CVE/БДУ-записи на установленном ПО, отсутствие патчей'),
    ('VL3', 'Допустимость установки не декларируемого ПО',
        'Возможность установки ПО вне корпоративного списка, недекларированные функции'),
    ('VL4', 'Наличие процедуры обхода администратором правил безопасности',
        'Привилегированные обходы политик, повышение привилегий, misconfiguration'),
    ('VL5', 'Носители информации',
        'Флеш-накопители, жёсткие диски, съёмные носители'),
    ('VL6', 'Открытые ОС / отсутствие средств защиты ЛВС',
        'Слабая периметровая защита, открытые порты, отсутствие сегментации');

-- ---------------------------------------------------------------
-- threat_vulnerable_links: vulnerability_id (BIGINT) → vl_category_id (SMALLINT)
-- ---------------------------------------------------------------

ALTER TABLE threat_vulnerable_links
    DROP CONSTRAINT IF EXISTS threat_vulnerable_links_vulnerability_id_fkey;
DROP INDEX IF EXISTS idx_tvl_vuln;
ALTER TABLE threat_vulnerable_links
    DROP COLUMN vulnerability_id,
    ADD  COLUMN vl_category_id SMALLINT NOT NULL REFERENCES vl_categories(id) ON DELETE CASCADE;
ALTER TABLE threat_vulnerable_links
    ADD PRIMARY KEY (threat_id, vl_category_id);
CREATE INDEX idx_tvl_vl_category ON threat_vulnerable_links(vl_category_id);

-- ---------------------------------------------------------------
-- vulnerability_controls → vl_category_controls
-- ---------------------------------------------------------------

ALTER TABLE vulnerability_controls
    DROP CONSTRAINT IF EXISTS vulnerability_controls_vulnerability_id_fkey;
DROP INDEX IF EXISTS idx_vc_control;
ALTER TABLE vulnerability_controls
    DROP COLUMN vulnerability_id,
    ADD  COLUMN vl_category_id SMALLINT NOT NULL REFERENCES vl_categories(id) ON DELETE CASCADE;
ALTER TABLE vulnerability_controls
    ADD PRIMARY KEY (vl_category_id, control_id);
ALTER TABLE vulnerability_controls
    RENAME TO vl_category_controls;
CREATE INDEX idx_vlc_control ON vl_category_controls(control_id);
