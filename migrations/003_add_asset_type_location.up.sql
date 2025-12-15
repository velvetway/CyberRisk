-- Добавляем колонку type для типа актива (используется в автоматическом расчёте CIA)
ALTER TABLE assets
    ADD COLUMN type VARCHAR(64);

-- Добавляем колонку location для расположения актива
ALTER TABLE assets
    ADD COLUMN location TEXT;

-- Создаем индекс для быстрого поиска по типу
CREATE INDEX idx_assets_type ON assets(type);
