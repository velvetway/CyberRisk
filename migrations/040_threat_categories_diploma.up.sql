-- Дополнить threat_categories до 4 групп из 8.png диплома (Модель угроз ИБ):
--   • Угрозы НСД                                  — уже есть (id=8)
--   • Угрозы от утечки по техническим каналам     — уже есть как «Утечка информации» (id=9)
--   • Угрозы хищения, искажения, блокировки       — добавить (новая)
--   • Угрозы непреднамеренных воздействий         — добавить (новая)
--
-- Затем backfill всех threats без категории по эвристике на основе name.
-- Старые английские категории остаются (для совместимости со старыми
-- импортами), но новые угрозы маппятся в дипломные группы.

-- ---------------------------------------------------------------
-- 1. Новые категории
-- ---------------------------------------------------------------

INSERT INTO threat_categories (name, description) VALUES
    ('Хищение, искажение, блокировка',
     'Угрозы хищения, искажения, блокировки информации (DA1, DA3, DA4, DA5, DA6 по диплому ПТСЗИ).'),
    ('Непреднамеренные воздействия',
     'Угрозы непреднамеренных воздействий: ошибки персонала, сбои оборудования, отключения питания.')
ON CONFLICT (name) DO UPDATE
   SET description = EXCLUDED.description;

-- ---------------------------------------------------------------
-- 2. Переименуем старую «Утечка информации» в более точное название
-- из 8.png — «Утечка по техническим каналам». Это безопасно через
-- UPDATE — id остаётся прежним.
-- ---------------------------------------------------------------

UPDATE threat_categories
SET name = 'Утечка по техническим каналам',
    description = 'Угрозы от утечки информации по техническим каналам (акустические, оптические, ПЭМИН и др.).'
WHERE name = 'Утечка информации';

-- ---------------------------------------------------------------
-- 3. Backfill threats.threat_category_id по name
-- ---------------------------------------------------------------

-- Хищение, искажение, блокировка — все слова-маркеры
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Хищение, искажение, блокировка')
WHERE threat_category_id IS NULL
  AND (
        name ILIKE '%хищени%'
     OR name ILIKE '%кражи%'
     OR name ILIKE '%искажени%'
     OR name ILIKE '%блокирован%'
     OR name ILIKE '%уничтожени%'
     OR name ILIKE '%модификации%'
     OR name ILIKE '%подмен%'
  );

-- Непреднамеренные воздействия
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Непреднамеренные воздействия')
WHERE threat_category_id IS NULL
  AND (
        name ILIKE '%непреднам%'
     OR name ILIKE '%неправильн%'
     OR name ILIKE '%ошибк%'
     OR name ILIKE '%сбо%'
     OR name ILIKE '%отказ оборудован%'
     OR name ILIKE '%отказ питан%'
     OR name ILIKE '%некомпетентн%'
  );

-- НСД — добиваем оставшиеся
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Несанкционированный доступ')
WHERE threat_category_id IS NULL
  AND (
        name ILIKE '%несанкционирован%'
     OR name ILIKE '% нсд%'
     OR name ILIKE '%неправомерн%'
     OR name ILIKE '%несанкц%'
  );

-- Сетевые атаки / DoS
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Сетевые атаки')
WHERE threat_category_id IS NULL
  AND (
        name ILIKE '%перехват%'
     OR name ILIKE '%сканирован%'
     OR name ILIKE '%проникновен%'
     OR name ILIKE '%атак%'
  );

UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Отказ в обслуживании')
WHERE threat_category_id IS NULL
  AND (name ILIKE '%dos%' OR name ILIKE '%отказ%');

-- Вредоносное ПО — оставшиеся «вирус», «вредонос», «троян»…
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Вредоносное ПО')
WHERE threat_category_id IS NULL
  AND (
        name ILIKE '%вредонос%'
     OR name ILIKE '%вирус%'
     OR name ILIKE '%троян%'
     OR name ILIKE '%бэкдор%'
  );

-- Утечка по техническим каналам
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Утечка по техническим каналам')
WHERE threat_category_id IS NULL
  AND (
        name ILIKE '%утечк%'
     OR name ILIKE '%разглашен%'
     OR name ILIKE '%раскрыти%'
  );

-- Финальный fallback — всё что осталось NULL → «Несанкционированный доступ»
-- (самая широкая категория из дипломной модели, ст2).
UPDATE threats
SET threat_category_id = (SELECT id FROM threat_categories WHERE name = 'Несанкционированный доступ')
WHERE threat_category_id IS NULL;
