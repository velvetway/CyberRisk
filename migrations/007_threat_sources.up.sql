-- Справочник источников угроз (S1..S4 из модели ПТСЗИ)
CREATE TABLE threat_sources (
    id          SMALLSERIAL PRIMARY KEY,
    code        VARCHAR(8)  NOT NULL UNIQUE,  -- S1, S2, S3, S4
    name        TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO threat_sources (code, name, description) VALUES
    ('S1', 'Конкуренты',               'Внешние организации, заинтересованные в получении ЧТ и вредоносной деятельности'),
    ('S2', 'Недобросовестные партнёры','Контрагенты, злоупотребляющие делегированным доступом'),
    ('S3', 'Персонал организации',     'Внутренние сотрудники (инсайдеры, халатность)'),
    ('S4', 'Хакеры, мошенники',        'Внешние атакующие без специфической мотивации');
