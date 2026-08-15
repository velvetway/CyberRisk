-- Добавляем колонку type для типа актива (используется в автоматическом расчёте CIA)
ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS type VARCHAR(64);

-- Добавляем колонку location для расположения актива
ALTER TABLE assets
    ADD COLUMN IF NOT EXISTS location TEXT;

-- Создаем индекс для быстрого поиска по типу
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);
