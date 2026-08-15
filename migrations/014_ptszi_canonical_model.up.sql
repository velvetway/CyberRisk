ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS security_contour VARCHAR(16) NOT NULL DEFAULT 'internal'
        CHECK (security_contour IN ('external', 'internal')),
    ADD COLUMN IF NOT EXISTS purpose TEXT,
    ADD COLUMN IF NOT EXISTS system_composition TEXT,
    ADD COLUMN IF NOT EXISTS network_location TEXT,
    ADD COLUMN IF NOT EXISTS processed_information TEXT,
    ADD COLUMN IF NOT EXISTS users_description TEXT,
    ADD COLUMN IF NOT EXISTS has_remote_access BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS communication_channels TEXT,
    ADD COLUMN IF NOT EXISTS existing_security_tools TEXT,
    ADD COLUMN IF NOT EXISTS document_flow TEXT;

UPDATE assets
SET security_contour = CASE
    WHEN has_internet_access THEN 'external'
    ELSE 'internal'
END
WHERE security_contour IS NULL OR security_contour = 'internal';

CREATE TABLE IF NOT EXISTS ptszi_threats (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    q_threat NUMERIC(3,2) NOT NULL CHECK (q_threat BETWEEN 0 AND 1),
    q_severity NUMERIC(3,2) NOT NULL CHECK (q_severity BETWEEN 0 AND 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ptszi_threat_contours (
    threat_id BIGINT NOT NULL REFERENCES ptszi_threats(id) ON DELETE CASCADE,
    contour VARCHAR(16) NOT NULL CHECK (contour IN ('external', 'internal')),
    PRIMARY KEY (threat_id, contour)
);

CREATE TABLE IF NOT EXISTS ptszi_vulnerable_links (
    id SMALLSERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS asset_vulnerable_links (
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    vulnerable_link_id SMALLINT NOT NULL REFERENCES ptszi_vulnerable_links(id) ON DELETE CASCADE,
    status VARCHAR(16) NOT NULL DEFAULT 'present'
        CHECK (status IN ('present', 'mitigated', 'accepted')),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (asset_id, vulnerable_link_id)
);

CREATE TABLE IF NOT EXISTS ptszi_controls (
    id SMALLSERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS asset_ptszi_controls (
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    control_id SMALLINT NOT NULL REFERENCES ptszi_controls(id) ON DELETE CASCADE,
    effectiveness NUMERIC(3,2) NOT NULL DEFAULT 1.0
        CHECK (effectiveness BETWEEN 0 AND 1),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (asset_id, control_id)
);

CREATE TABLE IF NOT EXISTS ptszi_threat_vulnerable_links (
    threat_id BIGINT NOT NULL REFERENCES ptszi_threats(id) ON DELETE CASCADE,
    vulnerable_link_id SMALLINT NOT NULL REFERENCES ptszi_vulnerable_links(id) ON DELETE CASCADE,
    PRIMARY KEY (threat_id, vulnerable_link_id)
);

CREATE TABLE IF NOT EXISTS ptszi_vulnerable_link_controls (
    vulnerable_link_id SMALLINT NOT NULL REFERENCES ptszi_vulnerable_links(id) ON DELETE CASCADE,
    control_id SMALLINT NOT NULL REFERENCES ptszi_controls(id) ON DELETE CASCADE,
    coverage NUMERIC(3,2) NOT NULL DEFAULT 1.0
        CHECK (coverage BETWEEN 0 AND 1),
    PRIMARY KEY (vulnerable_link_id, control_id)
);

CREATE TABLE IF NOT EXISTS ptszi_threat_destructive_actions (
    threat_id BIGINT NOT NULL REFERENCES ptszi_threats(id) ON DELETE CASCADE,
    destructive_action_id SMALLINT NOT NULL REFERENCES destructive_actions(id) ON DELETE CASCADE,
    PRIMARY KEY (threat_id, destructive_action_id)
);

CREATE TABLE IF NOT EXISTS ptszi_source_threats (
    source_id SMALLINT NOT NULL REFERENCES threat_sources(id) ON DELETE CASCADE,
    threat_id BIGINT NOT NULL REFERENCES ptszi_threats(id) ON DELETE CASCADE,
    PRIMARY KEY (source_id, threat_id)
);

CREATE TABLE IF NOT EXISTS ptszi_threat_ubi_links (
    threat_id BIGINT NOT NULL REFERENCES ptszi_threats(id) ON DELETE CASCADE,
    ubi_code VARCHAR(16) NOT NULL,
    comment TEXT,
    PRIMARY KEY (threat_id, ubi_code)
);

CREATE TABLE IF NOT EXISTS ptszi_recommendation_templates (
    id BIGSERIAL PRIMARY KEY,
    missing_control_id SMALLINT REFERENCES ptszi_controls(id),
    category VARCHAR(64) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(16) NOT NULL CHECK (priority IN ('low','medium','high','critical'))
);

INSERT INTO ptszi_threats (code, name, description, q_threat, q_severity) VALUES
    ('ST1', 'Инсталляция и запуск вирусов', 'Внедрение и запуск вредоносного программного обеспечения на объекте защиты.', 0.70, 0.80),
    ('ST2', 'Несанкционированный доступ', 'Получение доступа к информации или функциям системы без установленных прав.', 0.80, 0.90),
    ('ST3', 'DDoS-атаки', 'Нарушение доступности информационной системы за счет распределенной нагрузки.', 0.60, 0.75),
    ('ST4', 'Перехват трафика', 'Получение доступа к передаваемым данным в каналах связи.', 0.55, 0.70),
    ('ST5', 'Проникновение во внутреннюю сеть', 'Попытка закрепления и перемещения нарушителя во внутреннем сегменте.', 0.65, 0.90),
    ('ST6', 'Внешнее сканирование', 'Разведка внешнего периметра и выявление доступных сервисов.', 0.70, 0.45),
    ('ST7', 'Проникновение через сервисы', 'Эксплуатация доступных сервисов и приложений для проникновения.', 0.75, 0.85),
    ('ST8', 'Спам, подмена сообщений/отправителя', 'Рассылка нежелательных сообщений, фишинг и подмена отправителя.', 0.65, 0.60)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    q_threat = EXCLUDED.q_threat,
    q_severity = EXCLUDED.q_severity,
    updated_at = now();

INSERT INTO ptszi_vulnerable_links (code, name, description) VALUES
    ('VL1', 'Нештатное дополнительное ПО', 'Драйверы, утилиты и иное дополнительное ПО вне штатного состава.'),
    ('VL2', 'Устаревшие версии ПО или версии, имеющие уязвимости', 'Наличие ПО с известными уязвимостями, включая подтвержденные BDU/CVE.'),
    ('VL3', 'Допустимость установки недекларируемого ПО', 'Возможность установки ПО, не предусмотренного моделью эксплуатации.'),
    ('VL4', 'Обход администратором правил и режимов безопасности', 'Наличие процедур или возможностей обхода установленных правил безопасности.'),
    ('VL5', 'Носители информации', 'Флеш-накопители, жесткие диски и иные переносные носители.'),
    ('VL6', 'Открытые ОС', 'Использование открытых или недостаточно защищенных операционных систем.'),
    ('VL7', 'Отсутствие средств защиты ЛВС', 'Недостаточность средств защиты локальной вычислительной сети.')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description;

INSERT INTO ptszi_controls (code, name, description) VALUES
    ('A', 'Антивирус', 'Антивирусная защита рабочих станций, серверов и файловых хранилищ.'),
    ('FW', 'Межсетевой экран', 'Фильтрация сетевого трафика и сегментация доступа.'),
    ('HP', 'Honeypot', 'Ловушка для обнаружения попыток проникновения.'),
    ('DZ', 'Демилитаризованная зона ЛВС', 'Выделенный сегмент для публичных сервисов.'),
    ('IDS', 'Система обнаружения вторжений', 'Обнаружение атак и подозрительной активности.'),
    ('AD', 'Системы / настройки администрирования', 'Управление учетными записями, политиками и привилегиями.'),
    ('R', 'Резервное копирование', 'Резервное копирование и восстановление информации.'),
    ('L', 'Программная защита информации от НСД', 'Средства защиты от несанкционированного доступа.'),
    ('TE', 'Шифрование трафика', 'Криптографическая защита каналов связи.'),
    ('DS', 'Цифровая подпись', 'Контроль целостности и авторства данных.'),
    ('DD', 'DDoS-фильтры', 'Фильтрация и подавление DDoS-трафика.')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description;

INSERT INTO ptszi_threat_contours (threat_id, contour)
SELECT t.id, c.contour
FROM (VALUES
    ('ST1', 'internal'),
    ('ST2', 'external'), ('ST2', 'internal'),
    ('ST3', 'external'),
    ('ST4', 'external'), ('ST4', 'internal'),
    ('ST5', 'external'), ('ST5', 'internal'),
    ('ST6', 'external'),
    ('ST7', 'external'), ('ST7', 'internal'),
    ('ST8', 'external'), ('ST8', 'internal')
) AS c(code, contour)
JOIN ptszi_threats t ON t.code = c.code
ON CONFLICT DO NOTHING;

INSERT INTO ptszi_threat_vulnerable_links (threat_id, vulnerable_link_id)
SELECT t.id, vl.id
FROM (VALUES
    ('ST1','VL1'), ('ST1','VL2'), ('ST1','VL3'), ('ST1','VL5'), ('ST1','VL6'),
    ('ST2','VL2'), ('ST2','VL4'), ('ST2','VL6'), ('ST2','VL7'),
    ('ST3','VL7'),
    ('ST4','VL6'), ('ST4','VL7'),
    ('ST5','VL4'), ('ST5','VL6'), ('ST5','VL7'),
    ('ST6','VL6'), ('ST6','VL7'),
    ('ST7','VL2'), ('ST7','VL4'), ('ST7','VL6'), ('ST7','VL7'),
    ('ST8','VL3'), ('ST8','VL4'), ('ST8','VL6')
) AS m(threat_code, vl_code)
JOIN ptszi_threats t ON t.code = m.threat_code
JOIN ptszi_vulnerable_links vl ON vl.code = m.vl_code
ON CONFLICT DO NOTHING;

INSERT INTO ptszi_vulnerable_link_controls (vulnerable_link_id, control_id, coverage)
SELECT vl.id, c.id, m.coverage
FROM (VALUES
    ('VL1','A',0.80), ('VL1','AD',0.70), ('VL1','L',0.70),
    ('VL2','A',0.70), ('VL2','AD',0.75), ('VL2','R',0.60), ('VL2','L',0.65),
    ('VL3','AD',0.80), ('VL3','L',0.70),
    ('VL4','AD',0.80), ('VL4','IDS',0.55), ('VL4','L',0.70),
    ('VL5','AD',0.55), ('VL5','L',0.70), ('VL5','DS',0.65), ('VL5','R',0.70),
    ('VL6','FW',0.75), ('VL6','IDS',0.65), ('VL6','AD',0.55), ('VL6','L',0.65),
    ('VL7','FW',0.85), ('VL7','HP',0.45), ('VL7','DZ',0.70), ('VL7','IDS',0.75), ('VL7','DD',0.85)
) AS m(vl_code, control_code, coverage)
JOIN ptszi_vulnerable_links vl ON vl.code = m.vl_code
JOIN ptszi_controls c ON c.code = m.control_code
ON CONFLICT DO NOTHING;

INSERT INTO ptszi_threat_destructive_actions (threat_id, destructive_action_id)
SELECT t.id, da.id
FROM (VALUES
    ('ST1','DA3'), ('ST1','DA4'), ('ST1','DA5'), ('ST1','DA7'),
    ('ST2','DA1'), ('ST2','DA4'), ('ST2','DA6'), ('ST2','DA7'),
    ('ST3','DA5'), ('ST3','DA7'),
    ('ST4','DA1'), ('ST4','DA2'), ('ST4','DA4'),
    ('ST5','DA1'), ('ST5','DA4'), ('ST5','DA6'), ('ST5','DA7'),
    ('ST6','DA1'), ('ST6','DA7'),
    ('ST7','DA1'), ('ST7','DA4'), ('ST7','DA5'), ('ST7','DA7'),
    ('ST8','DA1'), ('ST8','DA4'), ('ST8','DA6')
) AS m(threat_code, da_code)
JOIN ptszi_threats t ON t.code = m.threat_code
JOIN destructive_actions da ON da.code = m.da_code
ON CONFLICT DO NOTHING;

INSERT INTO ptszi_source_threats (source_id, threat_id)
SELECT s.id, t.id
FROM (VALUES
    ('S1','ST2'), ('S1','ST3'), ('S1','ST4'), ('S1','ST6'), ('S1','ST8'),
    ('S2','ST2'), ('S2','ST4'), ('S2','ST5'), ('S2','ST8'),
    ('S3','ST1'), ('S3','ST2'), ('S3','ST5'), ('S3','ST7'),
    ('S4','ST1'), ('S4','ST2'), ('S4','ST3'), ('S4','ST4'), ('S4','ST5'), ('S4','ST6'), ('S4','ST7'), ('S4','ST8')
) AS m(source_code, threat_code)
JOIN threat_sources s ON s.code = m.source_code
JOIN ptszi_threats t ON t.code = m.threat_code
ON CONFLICT DO NOTHING;

INSERT INTO ptszi_threat_ubi_links (threat_id, ubi_code, comment)
SELECT t.id, m.ubi_code, 'Справочная детализация БДУ/УБИ для канонической угрозы ПТСЗИ'
FROM (VALUES
    ('ST2','УБИ.001'), ('ST2','УБИ.009'), ('ST2','УБИ.048'),
    ('ST3','УБИ.041'), ('ST3','УБИ.042'),
    ('ST4','УБИ.036'), ('ST4','УБИ.037'),
    ('ST6','УБИ.035'),
    ('ST7','УБИ.071'), ('ST7','УБИ.100'),
    ('ST8','УБИ.175'), ('ST8','УБИ.176')
) AS m(threat_code, ubi_code)
JOIN ptszi_threats t ON t.code = m.threat_code
ON CONFLICT DO NOTHING;

INSERT INTO ptszi_recommendation_templates (missing_control_id, category, title, description, priority)
SELECT c.id, m.category, m.title, m.description, m.priority
FROM (VALUES
    ('A', 'Защита АРМ', 'Внедрить антивирусную защиту', 'Установить и настроить антивирусное средство на рабочих станциях и серверах.', 'high'),
    ('FW', 'Защита ЛВС', 'Внедрить межсетевое экранирование', 'Развернуть межсетевой экран и правила сегментации сетевого доступа.', 'high'),
    ('IDS', 'Защита ЛВС', 'Внедрить систему обнаружения вторжений', 'Настроить IDS для обнаружения атак и подозрительной сетевой активности.', 'high'),
    ('L', 'Защита конфиденциальной информации', 'Внедрить средство защиты от НСД', 'Использовать программное средство защиты информации от несанкционированного доступа.', 'critical'),
    ('R', 'Защита информации', 'Организовать резервное копирование', 'Настроить регулярное резервное копирование и проверку восстановления данных.', 'medium'),
    ('TE', 'Защита каналов связи', 'Использовать шифрование трафика', 'Защитить каналы связи криптографическими средствами.', 'medium'),
    ('DD', 'Защита доступности', 'Подключить DDoS-фильтрацию', 'Использовать средства фильтрации и подавления DDoS-трафика.', 'high')
) AS m(control_code, category, title, description, priority)
JOIN ptszi_controls c ON c.code = m.control_code
WHERE NOT EXISTS (
    SELECT 1 FROM ptszi_recommendation_templates r
    WHERE r.missing_control_id = c.id AND r.title = m.title
);

INSERT INTO asset_vulnerable_links (asset_id, vulnerable_link_id, status, comment)
SELECT a.id, vl.id, 'present', 'seeded for PTSZI canonical demo'
FROM (VALUES
    ('Web Application Server','VL2'), ('Web Application Server','VL6'), ('Web Application Server','VL7'),
    ('Customer Database','VL2'), ('Customer Database','VL4'), ('Customer Database','VL6'),
    ('Employee Workstation Network','VL1'), ('Employee Workstation Network','VL3'), ('Employee Workstation Network','VL5'),
    ('Development Server','VL2'), ('Development Server','VL3'), ('Development Server','VL6'),
    ('Mobile App Backend','VL2'), ('Mobile App Backend','VL6'), ('Mobile App Backend','VL7')
) AS m(asset_name, vl_code)
JOIN assets a ON a.name = m.asset_name
JOIN ptszi_vulnerable_links vl ON vl.code = m.vl_code
ON CONFLICT DO NOTHING;

INSERT INTO asset_ptszi_controls (asset_id, control_id, effectiveness, comment)
SELECT a.id, c.id, m.effectiveness, 'seeded for PTSZI canonical demo'
FROM (VALUES
    ('Web Application Server','FW',0.80), ('Web Application Server','IDS',0.60), ('Web Application Server','AD',0.70),
    ('Customer Database','AD',0.80), ('Customer Database','R',0.90), ('Customer Database','L',0.70),
    ('Employee Workstation Network','A',0.85), ('Employee Workstation Network','AD',0.60),
    ('Development Server','AD',0.50), ('Development Server','R',0.60),
    ('Mobile App Backend','FW',0.70), ('Mobile App Backend','IDS',0.50)
) AS m(asset_name, control_code, effectiveness)
JOIN assets a ON a.name = m.asset_name
JOIN ptszi_controls c ON c.code = m.control_code
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_asset_vulnerable_links_asset ON asset_vulnerable_links(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_ptszi_controls_asset ON asset_ptszi_controls(asset_id);
CREATE INDEX IF NOT EXISTS idx_ptszi_threat_contours_contour ON ptszi_threat_contours(contour);
