-- Rollback: возвращаем старое название и удаляем добавленные категории.
-- Backfill категорий не откатываем (это потеря не данных, а только маппинга).

UPDATE threat_categories
SET name = 'Утечка информации',
    description = 'Угрозы утечки информации по техническим каналам'
WHERE name = 'Утечка по техническим каналам';

DELETE FROM threat_categories WHERE name = 'Хищение, искажение, блокировка';
DELETE FROM threat_categories WHERE name = 'Непреднамеренные воздействия';
