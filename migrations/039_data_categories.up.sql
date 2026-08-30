-- Категория обрабатываемой информации (по 7.png диплома —
-- блок «Определение вида обрабатываемой информации»).
--
-- Идемпотентно: ON CONFLICT DO NOTHING / IF NOT EXISTS.

-- ---------------------------------------------------------------
-- 1. Справочник
-- ---------------------------------------------------------------

CREATE TABLE IF NOT EXISTS data_categories (
    id          smallserial PRIMARY KEY,
    code        text        NOT NULL UNIQUE,
    name        text        NOT NULL,
    description text,
    sort_order  smallint    NOT NULL DEFAULT 0
);

INSERT INTO data_categories (code, name, description, sort_order) VALUES
    ('public',   'Общедоступная информация',
     'Сведения, не требующие защиты от несанкционированного доступа; защищается только от модификации/недоступности.', 10),
    ('internal', 'Служебная информация ограниченного распространения',
     'Сведения, доступ к которым ограничен внутренним регламентом организации.', 20),
    ('pdn',      'Персональные данные (ПДн)',
     'Любая информация, относящаяся к идентифицированному физическому лицу (152-ФЗ).', 30),
    ('confidential', 'Конфиденциальная информация (коммерческая тайна)',
     'Сведения, составляющие коммерческую тайну, профессиональную тайну, тайну переписки и т. п.', 40),
    ('dsp',      'Для служебного пользования (ДСП)',
     'Сведения ограниченного распространения, не отнесённые к гостайне.', 50),
    ('kii',      'Информация в составе объекта КИИ',
     'Сведения, обрабатываемые на значимых объектах критической информационной инфраструктуры (187-ФЗ).', 60)
ON CONFLICT (code) DO UPDATE
   SET name = EXCLUDED.name,
       description = EXCLUDED.description,
       sort_order = EXCLUDED.sort_order;

-- ---------------------------------------------------------------
-- 2. Поле data_category_id в активе
-- ---------------------------------------------------------------

ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS data_category_id smallint REFERENCES data_categories(id);

CREATE INDEX IF NOT EXISTS idx_assets_data_category ON assets(data_category_id);
