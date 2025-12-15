-- ============================
-- СПРАВОЧНИК ПРОГРАММНОГО ОБЕСПЕЧЕНИЯ
-- ============================

-- Категории ПО
CREATE TABLE software_categories (
    id              SMALLSERIAL PRIMARY KEY,
    code            VARCHAR(32) NOT NULL UNIQUE,
    name            VARCHAR(128) NOT NULL,
    description     TEXT
);

INSERT INTO software_categories (code, name, description) VALUES
    ('os', 'Операционная система', 'Операционные системы'),
    ('dbms', 'СУБД', 'Системы управления базами данных'),
    ('erp', 'ERP-система', 'Системы управления ресурсами предприятия'),
    ('crm', 'CRM-система', 'Системы управления взаимоотношениями с клиентами'),
    ('office', 'Офисное ПО', 'Офисные приложения и пакеты'),
    ('antivirus', 'Антивирус/СЗИ', 'Средства защиты информации'),
    ('backup', 'Резервное копирование', 'Системы резервного копирования'),
    ('monitoring', 'Мониторинг', 'Системы мониторинга и логирования'),
    ('virtualization', 'Виртуализация', 'Платформы виртуализации'),
    ('network', 'Сетевое ПО', 'Сетевое и телекоммуникационное ПО'),
    ('development', 'Средства разработки', 'IDE, компиляторы, SDK'),
    ('web', 'Веб-сервер/Приложение', 'Веб-серверы и веб-приложения'),
    ('mail', 'Почтовая система', 'Почтовые серверы и клиенты'),
    ('other', 'Прочее', 'Прочее программное обеспечение');

-- Основной справочник ПО
CREATE TABLE software_catalog (
    id                      BIGSERIAL PRIMARY KEY,
    name                    VARCHAR(256) NOT NULL,
    vendor                  VARCHAR(256) NOT NULL,
    version                 VARCHAR(64),
    category_id             SMALLINT REFERENCES software_categories(id),

    -- Российский реестр ПО (Минцифры)
    is_russian              BOOLEAN NOT NULL DEFAULT FALSE,
    registry_number         VARCHAR(32),          -- Номер в реестре Минцифры
    registry_date           DATE,                 -- Дата включения в реестр
    registry_url            TEXT,                 -- Ссылка на запись в реестре

    -- Сертификация ФСТЭК
    fstec_certified         BOOLEAN NOT NULL DEFAULT FALSE,
    fstec_certificate_num   VARCHAR(64),          -- Номер сертификата ФСТЭК
    fstec_certificate_date  DATE,                 -- Дата сертификата
    fstec_protection_class  VARCHAR(16),          -- Класс защиты (например, "4" для СЗИ)
    fstec_valid_until       DATE,                 -- Срок действия сертификата

    -- Сертификация ФСБ (для СКЗИ)
    fsb_certified           BOOLEAN NOT NULL DEFAULT FALSE,
    fsb_certificate_num     VARCHAR(64),
    fsb_protection_class    VARCHAR(16),          -- КС1, КС2, КС3, КВ, КА

    description             TEXT,
    website                 TEXT,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_software_catalog_vendor ON software_catalog(vendor);
CREATE INDEX idx_software_catalog_category ON software_catalog(category_id);
CREATE INDEX idx_software_catalog_russian ON software_catalog(is_russian);
CREATE INDEX idx_software_catalog_fstec ON software_catalog(fstec_certified);

-- ============================
-- КАТЕГОРИИ ОБРАБАТЫВАЕМЫХ ДАННЫХ
-- ============================

CREATE TYPE data_category_type AS ENUM (
    'public',               -- Общедоступные данные
    'internal',             -- Внутренние данные организации
    'confidential',         -- Конфиденциальные данные
    'personal_data',        -- Персональные данные (152-ФЗ)
    'personal_data_special', -- Специальные категории ПДн
    'personal_data_biometric', -- Биометрические ПДн
    'kii',                  -- КИИ (187-ФЗ)
    'state_secret',         -- Государственная тайна
    'banking_secret',       -- Банковская тайна
    'medical_secret',       -- Врачебная тайна
    'commercial_secret'     -- Коммерческая тайна
);

-- Уровни защищённости ПДн (Постановление №1119)
CREATE TYPE protection_level AS ENUM ('uz1', 'uz2', 'uz3', 'uz4');

-- Категории значимости КИИ (187-ФЗ)
CREATE TYPE kii_category AS ENUM ('none', 'cat3', 'cat2', 'cat1');

-- ============================
-- РАСШИРЕНИЕ ТАБЛИЦЫ АКТИВОВ
-- ============================

-- Добавляем поля для регуляторных требований
ALTER TABLE assets
    ADD COLUMN data_category          data_category_type DEFAULT 'internal',
    ADD COLUMN protection_level       protection_level,           -- УЗ для ПДн
    ADD COLUMN kii_category           kii_category DEFAULT 'none', -- Категория КИИ
    ADD COLUMN has_personal_data      BOOLEAN DEFAULT FALSE,
    ADD COLUMN personal_data_volume   VARCHAR(32),                -- Объём ПДн: <1000, 1000-100000, >100000
    ADD COLUMN has_internet_access    BOOLEAN DEFAULT TRUE,       -- Доступ в интернет
    ADD COLUMN is_isolated            BOOLEAN DEFAULT FALSE;      -- Изолированный сегмент

CREATE INDEX idx_assets_data_category ON assets(data_category);
CREATE INDEX idx_assets_kii ON assets(kii_category);
CREATE INDEX idx_assets_protection_level ON assets(protection_level);

-- ============================
-- СВЯЗЬ АКТИВ <-> ПО
-- ============================

CREATE TABLE asset_software (
    id              BIGSERIAL PRIMARY KEY,
    asset_id        BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    software_id     BIGINT NOT NULL REFERENCES software_catalog(id) ON DELETE CASCADE,
    version         VARCHAR(64),              -- Конкретная версия на активе
    install_date    DATE,
    license_type    VARCHAR(64),              -- Тип лицензии
    license_expires DATE,                     -- Срок действия лицензии
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uniq_asset_software UNIQUE (asset_id, software_id)
);

CREATE INDEX idx_asset_software_asset ON asset_software(asset_id);
CREATE INDEX idx_asset_software_software ON asset_software(software_id);

-- ============================
-- НАЧАЛЬНЫЕ ДАННЫЕ: ПОПУЛЯРНОЕ РОССИЙСКОЕ ПО
-- ============================

INSERT INTO software_catalog (name, vendor, category_id, is_russian, registry_number, fstec_certified, description) VALUES
    -- Операционные системы
    ('Astra Linux Special Edition', 'РусБИТех-Астра', 1, TRUE, '1292', TRUE, 'Защищённая ОС для обработки конфиденциальной информации'),
    ('Альт 8 СП', 'Базальт СПО', 1, TRUE, '1541', TRUE, 'Российская ОС на базе Linux'),
    ('РЕД ОС', 'РЕД СОФТ', 1, TRUE, '3751', TRUE, 'Российская операционная система'),

    -- СУБД
    ('PostgreSQL (Postgres Pro)', 'Postgres Professional', 2, TRUE, '1394', TRUE, 'Российская сборка PostgreSQL с расширениями'),
    ('ClickHouse', 'Яндекс', 2, TRUE, '5151', FALSE, 'Колоночная СУБД для аналитики'),
    ('Tarantool', 'VK (Mail.ru Group)', 2, TRUE, '3796', FALSE, 'In-memory СУБД'),

    -- ERP/CRM
    ('1С:Предприятие 8', '1С', 3, TRUE, '35', FALSE, 'Платформа для автоматизации бизнеса'),
    ('Галактика ERP', 'Корпорация Галактика', 3, TRUE, '275', FALSE, 'ERP-система'),
    ('Битрикс24', '1С-Битрикс', 4, TRUE, '107', FALSE, 'CRM и корпоративный портал'),

    -- Офисное ПО
    ('МойОфис', 'Новые Облачные Технологии', 5, TRUE, '136', TRUE, 'Офисный пакет'),
    ('Р7-Офис', 'Р7', 5, TRUE, '202', FALSE, 'Офисный пакет'),

    -- Антивирусы и СЗИ
    ('Kaspersky Endpoint Security', 'Лаборатория Касперского', 6, TRUE, '55', TRUE, 'Антивирусная защита'),
    ('Dr.Web Enterprise Security Suite', 'Доктор Веб', 6, TRUE, '66', TRUE, 'Антивирусная защита'),
    ('Secret Net Studio', 'Код Безопасности', 6, TRUE, '83', TRUE, 'СЗИ от НСД'),
    ('Dallas Lock', 'Конфидент', 6, TRUE, '104', TRUE, 'СЗИ от НСД'),
    ('VipNet', 'ИнфоТеКС', 6, TRUE, '76', TRUE, 'VPN и криптозащита'),
    ('КриптоПро CSP', 'КриптоПро', 6, TRUE, '124', TRUE, 'СКЗИ'),
    ('ViPNet SIES', 'ИнфоТеКС', 6, TRUE, '76', TRUE, 'SIEM-система'),
    ('MaxPatrol SIEM', 'Positive Technologies', 6, TRUE, '1143', TRUE, 'SIEM-система'),
    ('PT Application Firewall', 'Positive Technologies', 6, TRUE, '1143', TRUE, 'WAF'),

    -- Резервное копирование
    ('Кибер Бэкап', 'Киберпротект', 7, TRUE, '489', TRUE, 'Система резервного копирования'),
    ('RuBackup', 'Рубэкап', 7, TRUE, '6078', FALSE, 'Система резервного копирования'),

    -- Виртуализация
    ('zVirt', 'Orion soft', 9, TRUE, '5570', TRUE, 'Платформа виртуализации'),
    ('Р-Виртуализация', 'РЕД СОФТ', 9, TRUE, '5318', FALSE, 'Среда виртуализации'),

    -- Веб-серверы
    ('Angie', 'Веб-Сервер', 12, TRUE, '18029', FALSE, 'Веб-сервер (форк nginx)'),

    -- Почта
    ('CommuniGate Pro', 'СталкерСофт', 13, TRUE, '364', FALSE, 'Почтовый сервер'),
    ('RuPost', 'Рупост', 13, TRUE, '10789', FALSE, 'Почтовая система');

-- Популярное зарубежное ПО (для сравнения/миграции)
INSERT INTO software_catalog (name, vendor, category_id, is_russian, description) VALUES
    ('Windows Server', 'Microsoft', 1, FALSE, 'Серверная ОС'),
    ('Red Hat Enterprise Linux', 'Red Hat', 1, FALSE, 'Корпоративный Linux'),
    ('Oracle Database', 'Oracle', 2, FALSE, 'СУБД'),
    ('Microsoft SQL Server', 'Microsoft', 2, FALSE, 'СУБД'),
    ('SAP ERP', 'SAP', 3, FALSE, 'ERP-система'),
    ('Salesforce', 'Salesforce', 4, FALSE, 'CRM-система'),
    ('Microsoft 365', 'Microsoft', 5, FALSE, 'Офисный пакет'),
    ('VMware vSphere', 'VMware', 9, FALSE, 'Платформа виртуализации'),
    ('Veeam Backup', 'Veeam', 7, FALSE, 'Система резервного копирования'),
    ('Symantec Endpoint Protection', 'Broadcom', 6, FALSE, 'Антивирус'),
    ('Splunk', 'Splunk', 8, FALSE, 'SIEM-система');
