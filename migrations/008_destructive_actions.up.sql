-- Справочник деструктивных действий (DA1..DA7)
-- Каждое действие затрагивает одну или несколько граней CIA
CREATE TABLE destructive_actions (
    id                        SMALLSERIAL PRIMARY KEY,
    code                      VARCHAR(8)  NOT NULL UNIQUE,
    name                      TEXT        NOT NULL,
    affects_confidentiality   BOOLEAN     NOT NULL DEFAULT FALSE,
    affects_integrity         BOOLEAN     NOT NULL DEFAULT FALSE,
    affects_availability      BOOLEAN     NOT NULL DEFAULT FALSE,
    description               TEXT,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO destructive_actions (code, name, affects_confidentiality, affects_integrity, affects_availability, description) VALUES
    ('DA1', 'Копирование (чтение) информации', TRUE,  FALSE, FALSE, 'Несанкционированное чтение/снятие копии защищаемых данных'),
    ('DA2', 'Перехват информации',             TRUE,  FALSE, FALSE, 'Перехват данных в канале связи'),
    ('DA3', 'Уничтожение информации',          FALSE, FALSE, TRUE,  'Полное удаление данных с носителя'),
    ('DA4', 'Модификация информации',          FALSE, TRUE,  FALSE, 'Несанкционированное изменение данных'),
    ('DA5', 'Блокирование информации',         FALSE, FALSE, TRUE,  'Сокрытие или временный отказ в доступе'),
    ('DA6', 'Хищение информации',              TRUE,  FALSE, TRUE,  'Изъятие защищаемых данных (носителя)'),
    ('DA7', 'Нарушение работоспособности системы', FALSE, TRUE, TRUE, 'Deface, DoS, компрометация целостности среды');
