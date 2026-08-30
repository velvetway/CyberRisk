-- Compliance: справочник стандартов ИБ + каталог требований +
-- маппинг requirement ↔ control. Используется на странице
-- «Состояние защищённости» (ОСЗ из 7.png диплома).
--
-- Назначение:
--   • compliance_standards   — какие стандарты мы умеем мерить (ФСТЭК-17, ISO 27001:2022)
--   • compliance_requirements — отдельные требования внутри стандарта
--   • requirement_controls   — какой control из 11 закрывает требование
--
-- Покрытие требования рассчитывается так:
--   coverage = max(weight) среди тех control_id, что внедрены на активе
--   через asset_controls. Если ни один — requirement считается невыполненным.
--   Сумма coverage / count(requirements) → % соответствия активa стандарту.
--
-- Идемпотентно: ON CONFLICT DO NOTHING / DO UPDATE по натуральным ключам.

-- ---------------------------------------------------------------
-- 1. Таблицы
-- ---------------------------------------------------------------

CREATE TABLE IF NOT EXISTS compliance_standards (
    id           smallserial PRIMARY KEY,
    code         text        NOT NULL UNIQUE,
    name         text        NOT NULL,
    full_name    text        NOT NULL,
    jurisdiction text        NOT NULL CHECK (jurisdiction IN ('RU', 'INT')),
    description  text,
    sort_order   smallint    NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS compliance_requirements (
    id           serial      PRIMARY KEY,
    standard_id  smallint    NOT NULL REFERENCES compliance_standards(id) ON DELETE CASCADE,
    code         text        NOT NULL,
    category     text        NOT NULL,
    title        text        NOT NULL,
    description  text,
    priority     smallint    NOT NULL DEFAULT 2 CHECK (priority IN (1,2,3)),  -- 1=high
    sort_order   smallint    NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (standard_id, code)
);

CREATE INDEX IF NOT EXISTS idx_compliance_req_standard
    ON compliance_requirements(standard_id, sort_order);

CREATE TABLE IF NOT EXISTS requirement_controls (
    requirement_id  integer  NOT NULL REFERENCES compliance_requirements(id) ON DELETE CASCADE,
    control_id      bigint   NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    coverage_weight numeric(3,2) NOT NULL DEFAULT 1.0
        CHECK (coverage_weight > 0 AND coverage_weight <= 1.0),
    PRIMARY KEY (requirement_id, control_id)
);

CREATE INDEX IF NOT EXISTS idx_req_controls_control
    ON requirement_controls(control_id);

-- ---------------------------------------------------------------
-- 2. Сидим стандарты
-- ---------------------------------------------------------------

INSERT INTO compliance_standards (code, name, full_name, jurisdiction, description, sort_order) VALUES
    ('FSTEC_17',   'ФСТЭК №17',
     'Приказ ФСТЭК России от 11.02.2013 № 17 «Требования о защите информации, не составляющей государственную тайну, содержащейся в государственных информационных системах»',
     'RU',
     'Базовый набор организационных и технических мер защиты ГИС.',
     1),
    ('ISO_27001',  'ISO/IEC 27001:2022',
     'ISO/IEC 27001:2022 — Information security, cybersecurity and privacy protection — Information security management systems — Requirements (Annex A)',
     'INT',
     'Международный стандарт управления ИБ. Annex A: 93 контрольные меры в 4 группах (Organizational/People/Physical/Technological).',
     2)
ON CONFLICT (code) DO UPDATE
   SET name = EXCLUDED.name,
       full_name = EXCLUDED.full_name,
       description = EXCLUDED.description,
       sort_order = EXCLUDED.sort_order;

-- ---------------------------------------------------------------
-- 3. ФСТЭК №17 — выборка требований, которые покрываются нашими 11 контролями
--    Коды соответствуют официальным группам мер из Приказа №17:
--      ИАФ — идентификация/аутентификация
--      УПД — управление доступом
--      РСБ — регистрация событий безопасности
--      АВЗ — антивирусная защита
--      СОВ — обнаружение вторжений
--      ОЦЛ — обеспечение целостности
--      ОДТ — обеспечение доступности
--      ЗИС — защита информационной системы (сегментация, периметр)
--      ЗТС — защита тех. средств
-- ---------------------------------------------------------------

WITH s AS (SELECT id FROM compliance_standards WHERE code = 'FSTEC_17')
INSERT INTO compliance_requirements (standard_id, code, category, title, description, priority, sort_order)
VALUES
    -- ИАФ — Идентификация и аутентификация
    ((SELECT id FROM s), 'ИАФ.1',  'Идентификация и аутентификация',
     'Идентификация и аутентификация пользователей',
     'Все пользователи системы должны быть однозначно идентифицированы и аутентифицированы до получения доступа к ресурсам.', 1, 10),
    ((SELECT id FROM s), 'ИАФ.3',  'Идентификация и аутентификация',
     'Управление идентификаторами',
     'Централизованное создание, изменение и блокирование учётных записей.', 1, 11),
    ((SELECT id FROM s), 'ИАФ.5',  'Идентификация и аутентификация',
     'Защита обратной связи при вводе аутентификационной информации',
     'Маскирование пароля при вводе, защита от подбора, блокировка после N неудачных попыток.', 2, 12),

    -- УПД — Управление доступом
    ((SELECT id FROM s), 'УПД.1',  'Управление доступом',
     'Управление учётными записями',
     'Разграничение прав доступа на основе ролей и принципа минимальных привилегий.', 1, 20),
    ((SELECT id FROM s), 'УПД.2',  'Управление доступом',
     'Реализация необходимых методов управления доступом',
     'Дискреционное и/или мандатное управление доступом, контроль доступа к ресурсам.', 1, 21),
    ((SELECT id FROM s), 'УПД.6',  'Управление доступом',
     'Ограничение неуспешных попыток входа',
     'Блокирование сессии/учётной записи после нескольких неудачных попыток аутентификации.', 2, 22),
    ((SELECT id FROM s), 'УПД.13', 'Управление доступом',
     'Реализация защищённого удалённого доступа',
     'Шифрование канала, многофакторная аутентификация для удалённого администрирования.', 1, 23),

    -- РСБ — Регистрация событий безопасности
    ((SELECT id FROM s), 'РСБ.1',  'Регистрация событий',
     'Определение событий безопасности, подлежащих регистрации',
     'Перечень регистрируемых событий: вход, попытки доступа, изменения конфигурации, ошибки СЗИ.', 2, 30),
    ((SELECT id FROM s), 'РСБ.2',  'Регистрация событий',
     'Регистрация событий безопасности',
     'Запись событий с минимальным составом полей: дата, время, идентификатор пользователя, тип, результат.', 2, 31),
    ((SELECT id FROM s), 'РСБ.3',  'Регистрация событий',
     'Защита информации о событиях безопасности',
     'Защита журналов от модификации/удаления, ограниченный доступ к ним.', 2, 32),

    -- АВЗ — Антивирусная защита
    ((SELECT id FROM s), 'АВЗ.1',  'Антивирусная защита',
     'Реализация антивирусной защиты',
     'Установка и эксплуатация средств антивирусной защиты на всех АРМ и серверах.', 1, 40),
    ((SELECT id FROM s), 'АВЗ.2',  'Антивирусная защита',
     'Обновление баз данных признаков вредоносных программ',
     'Автоматическое обновление антивирусных баз с приемлемой периодичностью.', 1, 41),

    -- СОВ — Обнаружение вторжений
    ((SELECT id FROM s), 'СОВ.1',  'Обнаружение вторжений',
     'Обнаружение и предотвращение вторжений',
     'Применение IDS/IPS на периметре и в критичных сегментах сети.', 1, 50),
    ((SELECT id FROM s), 'СОВ.2',  'Обнаружение вторжений',
     'Обновление базы решающих правил',
     'Регулярное обновление сигнатур и эвристик системы обнаружения вторжений.', 2, 51),

    -- ОЦЛ — Обеспечение целостности
    ((SELECT id FROM s), 'ОЦЛ.1',  'Обеспечение целостности',
     'Контроль целостности программного обеспечения',
     'Контроль ПО на этапе загрузки и в ходе работы (хэш-суммы, ЭЦП, замкнутая среда).', 1, 60),
    ((SELECT id FROM s), 'ОЦЛ.4',  'Обеспечение целостности',
     'Обнаружение и реагирование на нарушения целостности',
     'Алертинг при модификации защищаемых файлов/данных.', 2, 61),

    -- ОДТ — Обеспечение доступности
    ((SELECT id FROM s), 'ОДТ.1',  'Обеспечение доступности',
     'Резервное копирование информации',
     'Регулярное резервное копирование данных и конфигураций по утверждённому регламенту.', 1, 70),
    ((SELECT id FROM s), 'ОДТ.4',  'Обеспечение доступности',
     'Защита от отказа в обслуживании',
     'Защита от DDoS-атак на сетевом периметре и/или у внешних провайдеров.', 1, 71),
    ((SELECT id FROM s), 'ОДТ.5',  'Обеспечение доступности',
     'Кластеризация и резервирование критичных сервисов',
     'Отказоустойчивая конфигурация ключевых сервисов; не зависит от прикладных контролей системы.', 3, 72),

    -- ЗИС — Защита информационной системы
    ((SELECT id FROM s), 'ЗИС.1',  'Защита периметра',
     'Сегментирование информационной системы',
     'Разделение ИС на сегменты с разграничением доступа между ними.', 1, 80),
    ((SELECT id FROM s), 'ЗИС.5',  'Защита периметра',
     'Использование защищённого периметра',
     'МСЭ на границе периметра ИС, ДМЗ для публично-доступных сервисов.', 1, 81),
    ((SELECT id FROM s), 'ЗИС.10', 'Защита периметра',
     'Защита целостности передаваемой информации',
     'Применение электронной подписи и/или хэширования для контроля целостности при передаче.', 2, 82),
    ((SELECT id FROM s), 'ЗИС.18', 'Защита периметра',
     'Защита от компьютерных атак на ресурсы ИС',
     'Защита публичных сервисов от автоматизированных атак, в т.ч. DDoS.', 1, 83),
    ((SELECT id FROM s), 'ЗИС.20', 'Защита периметра',
     'Защита беспроводных соединений',
     'Шифрование Wi-Fi, аутентификация устройств, изоляция гостевых сетей.', 2, 84),
    ((SELECT id FROM s), 'ЗИС.27', 'Защита периметра',
     'Защита удалённого доступа',
     'Шифрование туннелей VPN, аутентификация по сертификатам/MFA.', 1, 85)
ON CONFLICT (standard_id, code) DO UPDATE
   SET title = EXCLUDED.title,
       description = EXCLUDED.description,
       category = EXCLUDED.category,
       priority = EXCLUDED.priority,
       sort_order = EXCLUDED.sort_order;

-- ---------------------------------------------------------------
-- 4. ISO 27001:2022 Annex A — выборка из 4 групп
-- ---------------------------------------------------------------

WITH s AS (SELECT id FROM compliance_standards WHERE code = 'ISO_27001')
INSERT INTO compliance_requirements (standard_id, code, category, title, description, priority, sort_order)
VALUES
    -- A.5 Organizational controls
    ((SELECT id FROM s), 'A.5.10',  'Organizational',
     'Acceptable use of information and other associated assets',
     'Правила приемлемого использования активов; технически — контроль установки ПО (СЗИ от НСД).', 2, 10),
    ((SELECT id FROM s), 'A.5.15',  'Organizational',
     'Access control',
     'Документированная политика управления доступом + её техническое обеспечение.', 1, 11),
    ((SELECT id FROM s), 'A.5.16',  'Organizational',
     'Identity management',
     'Жизненный цикл идентификаторов: создание, изменение, удаление.', 1, 12),
    ((SELECT id FROM s), 'A.5.17',  'Organizational',
     'Authentication information',
     'Защита аутентификационных данных (пароли, ключи, токены) от компрометации.', 1, 13),
    ((SELECT id FROM s), 'A.5.23',  'Organizational',
     'Information security for use of cloud services',
     'Безопасность облачных сервисов — частично через шифрование/MFA.', 2, 14),

    -- A.6 People controls — в целом организационные, не покрываются техническими контролями.
    --     Включаем как «требует ручной аттестации».

    -- A.7 Physical controls — физические меры; включим один тех. пункт.
    ((SELECT id FROM s), 'A.7.4',   'Physical',
     'Physical security monitoring',
     'Мониторинг физических зон + интеграция с IDS/SIEM.', 2, 30),

    -- A.8 Technological controls
    ((SELECT id FROM s), 'A.8.5',   'Technological',
     'Secure authentication',
     'Применение надёжных механизмов аутентификации (MFA, парольные политики).', 1, 40),
    ((SELECT id FROM s), 'A.8.7',   'Technological',
     'Protection against malware',
     'Антивирусная защита всех конечных точек и серверов.', 1, 41),
    ((SELECT id FROM s), 'A.8.8',   'Technological',
     'Management of technical vulnerabilities',
     'Идентификация и устранение технических уязвимостей (БДУ ФСТЭК автоматизирует).', 1, 42),
    ((SELECT id FROM s), 'A.8.9',   'Technological',
     'Configuration management',
     'Управление конфигурацией ПО и СЗИ (AD-домен, GPO/Ansible/Puppet).', 2, 43),
    ((SELECT id FROM s), 'A.8.13',  'Technological',
     'Information backup',
     'Резервное копирование данных и систем по утверждённому регламенту.', 1, 44),
    ((SELECT id FROM s), 'A.8.16',  'Technological',
     'Monitoring activities',
     'Сбор и анализ логов, обнаружение аномалий (SIEM/IDS).', 1, 45),
    ((SELECT id FROM s), 'A.8.20',  'Technological',
     'Networks security',
     'Безопасность сетей: МСЭ, сегментация, мониторинг.', 1, 46),
    ((SELECT id FROM s), 'A.8.22',  'Technological',
     'Segregation of networks',
     'Сегментация сети на изолированные зоны (ДМЗ, prod, dev).', 1, 47),
    ((SELECT id FROM s), 'A.8.23',  'Technological',
     'Web filtering',
     'Контроль исходящего веб-трафика (МСЭ + IPS).', 2, 48),
    ((SELECT id FROM s), 'A.8.24',  'Technological',
     'Use of cryptography',
     'Шифрование данных в покое и при передаче.', 1, 49),
    ((SELECT id FROM s), 'A.8.26',  'Technological',
     'Application security requirements',
     'Безопасность приложений — частично через WAF/IDS/защиту от DDoS.', 2, 50)
ON CONFLICT (standard_id, code) DO UPDATE
   SET title = EXCLUDED.title,
       description = EXCLUDED.description,
       category = EXCLUDED.category,
       priority = EXCLUDED.priority,
       sort_order = EXCLUDED.sort_order;

-- ---------------------------------------------------------------
-- 5. requirement_controls — соответствие требований и наших 11 контролей
--    Контроли: 1=A, 2=FW, 3=HP, 4=DZ, 5=IDS, 6=AD, 7=R, 8=L, 9=TE, 10=DS, 11=DD
-- ---------------------------------------------------------------

-- хелпер: id требования по standard_code + req_code
-- (в Postgres `WITH` достаточно, но удобнее inline-подзапросом)

INSERT INTO requirement_controls (requirement_id, control_id, coverage_weight)
SELECT r.id, c.id, w
FROM (
    VALUES
        -- ФСТЭК ИАФ.1 — закрывается AD (управление учёткой) и L (защита от НСД)
        ('FSTEC_17','ИАФ.1',  6, 0.7),  ('FSTEC_17','ИАФ.1',  8, 0.5),
        ('FSTEC_17','ИАФ.3',  6, 1.0),
        ('FSTEC_17','ИАФ.5',  6, 0.5),  ('FSTEC_17','ИАФ.5',  8, 0.5),

        -- УПД — AD + L
        ('FSTEC_17','УПД.1',  6, 1.0),
        ('FSTEC_17','УПД.2',  6, 0.6),  ('FSTEC_17','УПД.2',  8, 0.6),
        ('FSTEC_17','УПД.6',  6, 0.5),  ('FSTEC_17','УПД.6',  8, 0.5),
        ('FSTEC_17','УПД.13', 6, 0.4),  ('FSTEC_17','УПД.13', 9, 0.7),
                                        ('FSTEC_17','УПД.13', 8, 0.4),

        -- РСБ — IDS (события атак) + AD (учётный аудит)
        ('FSTEC_17','РСБ.1',  5, 0.6),  ('FSTEC_17','РСБ.1',  6, 0.5),
        ('FSTEC_17','РСБ.2',  5, 0.6),  ('FSTEC_17','РСБ.2',  6, 0.5),
        ('FSTEC_17','РСБ.3',  6, 0.7),  ('FSTEC_17','РСБ.3',  8, 0.5),

        -- АВЗ — A
        ('FSTEC_17','АВЗ.1',  1, 1.0),
        ('FSTEC_17','АВЗ.2',  1, 1.0),

        -- СОВ — IDS + HP
        ('FSTEC_17','СОВ.1',  5, 0.8),  ('FSTEC_17','СОВ.1',  3, 0.4),
        ('FSTEC_17','СОВ.2',  5, 1.0),

        -- ОЦЛ — DS (ЭЦП), L (контроль ПО), A (антивирус как контроль целостности тоже)
        ('FSTEC_17','ОЦЛ.1', 10, 0.6),  ('FSTEC_17','ОЦЛ.1',  8, 0.6),
                                         ('FSTEC_17','ОЦЛ.1',  1, 0.3),
        ('FSTEC_17','ОЦЛ.4',  5, 0.6),  ('FSTEC_17','ОЦЛ.4',  1, 0.4),

        -- ОДТ — R (бэкапы), DD (DDoS)
        ('FSTEC_17','ОДТ.1',  7, 1.0),
        ('FSTEC_17','ОДТ.4', 11, 1.0),
        -- ОДТ.5 — кластеризация: технических контролей нет, оставляем без покрытия

        -- ЗИС — FW, DZ, DS, TE, DD
        ('FSTEC_17','ЗИС.1',  2, 0.6),  ('FSTEC_17','ЗИС.1',  4, 0.7),
        ('FSTEC_17','ЗИС.5',  2, 0.7),  ('FSTEC_17','ЗИС.5',  4, 0.8),
                                         ('FSTEC_17','ЗИС.5',  5, 0.4),
        ('FSTEC_17','ЗИС.10',10, 0.8),  ('FSTEC_17','ЗИС.10', 9, 0.4),
        ('FSTEC_17','ЗИС.18',11, 0.8),  ('FSTEC_17','ЗИС.18', 5, 0.4),
                                         ('FSTEC_17','ЗИС.18', 2, 0.4),
        ('FSTEC_17','ЗИС.20', 9, 1.0),
        ('FSTEC_17','ЗИС.27', 9, 0.7),  ('FSTEC_17','ЗИС.27', 6, 0.4),

        -- ISO 27001
        ('ISO_27001','A.5.10', 8, 0.7),  ('ISO_27001','A.5.10', 6, 0.4),
        ('ISO_27001','A.5.15', 6, 0.6),  ('ISO_27001','A.5.15', 8, 0.5),
        ('ISO_27001','A.5.16', 6, 1.0),
        ('ISO_27001','A.5.17', 6, 0.6),  ('ISO_27001','A.5.17', 9, 0.4),
        ('ISO_27001','A.5.23', 9, 0.5),  ('ISO_27001','A.5.23', 6, 0.4),
        ('ISO_27001','A.7.4',  5, 0.5),  ('ISO_27001','A.7.4',  3, 0.3),
        ('ISO_27001','A.8.5',  6, 0.7),  ('ISO_27001','A.8.5',  9, 0.3),
        ('ISO_27001','A.8.7',  1, 1.0),
        ('ISO_27001','A.8.8',  1, 0.4),  ('ISO_27001','A.8.8',  6, 0.4),
        ('ISO_27001','A.8.9',  6, 0.7),  ('ISO_27001','A.8.9',  8, 0.4),
        ('ISO_27001','A.8.13', 7, 1.0),
        ('ISO_27001','A.8.16', 5, 0.7),  ('ISO_27001','A.8.16', 3, 0.4),
        ('ISO_27001','A.8.20', 2, 0.7),  ('ISO_27001','A.8.20', 5, 0.5),
                                          ('ISO_27001','A.8.20', 4, 0.5),
        ('ISO_27001','A.8.22', 4, 0.8),  ('ISO_27001','A.8.22', 2, 0.5),
        ('ISO_27001','A.8.23', 2, 0.7),  ('ISO_27001','A.8.23', 5, 0.4),
        ('ISO_27001','A.8.24', 9, 0.8),  ('ISO_27001','A.8.24',10, 0.4),
        ('ISO_27001','A.8.26', 2, 0.5),  ('ISO_27001','A.8.26',11, 0.4),
                                          ('ISO_27001','A.8.26', 5, 0.4)
) AS m(std_code, req_code, ctrl_id, w)
JOIN compliance_standards    s ON s.code = m.std_code
JOIN compliance_requirements r ON r.standard_id = s.id AND r.code = m.req_code
JOIN controls                c ON c.id = m.ctrl_id
ON CONFLICT (requirement_id, control_id) DO UPDATE
   SET coverage_weight = EXCLUDED.coverage_weight;
