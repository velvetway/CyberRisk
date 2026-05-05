# План: интеграция ФСТЭК / БДУ / Минцифры

**Цель.** Заменить искусственные сидерные данные (26 угроз, 5 уязвимостей, 30 ПО) на настоящие источники: 227 угроз УБИ ФСТЭК, 86 664 уязвимости БДУ, 26 000 продуктов реестра Минцифры. Перестроить модель данных так, чтобы **VL** соответствовали диплому (6 категорий «уязвимых звеньев»), а CVE были «инвентарём свидетельств» присутствия VL на активе. Применимость угрозы к активу выводить из настоящих полей ФСТЭК («Объект воздействия», «Источник угрозы»), а не выдумывать.

**Итог.** Каталог угроз и уязвимостей становится профессиональным, в системе появляется автодетекция CVE по установленному ПО, формула W получает честное `Q^reaction`, основанное на реальных VL_категориях.

## Архитектурные решения (приняты в этом плане)

| # | Решение | Альтернатива | Почему |
|---|---|---|---|
| AR-1 | Снимки ФСТЭК и Минцифры храним **рядом с приложением** как отдельные SQLite/JSON файлы (read-only). | Реплицировать в Postgres. | 86k уязвимостей × индексы — ~500 МБ. Не загромождаем основную БД. Снимок легко заменить, легко версионировать. |
| AR-2 | ETL — отдельные Go-команды (`cmd/import-fstec-threats`, `cmd/import-bdu-snapshot`, `cmd/import-minreestr`). | SQL-миграции. | XLSX/SQLite парсинг в SQL невозможен, в Go тривиально. ETL переиспользуем для обновления снимков. |
| AR-3 | VL = **6 категорий из диплома** (новая таблица `vl_categories`). CVE/БДУ — отдельный инвентарь, привязанный к категории VL через классификатор. | Оставить vulnerabilities как «и категория и CVE одновременно». | Точное соответствие модели диплома. Граф ST↔VL остаётся компактным (227×6). Инвентарь CVE может расти независимо. |
| AR-4 | Связь «угроза → VL_категории» строим **эвристикой по «Объекту воздействия»** из УБИ ФСТЭК + ручная корректировка. | Полностью ручной маппинг. | 227 угроз руками не разметить за разумное время. Эвристика даёт 70-80% правильных связей, остаток правится через UI/админку. |
| AR-5 | Связь «CVE → VL_категория» строим через **CWE-классификатор**. Default: всё что не подошло → VL_2 (устаревшее/уязвимое ПО). | Игнорировать категоризацию CVE. | CWE есть у каждой БДУ-записи (92k связей). Простой словарь CWE→VL даёт осмысленную классификацию. |
| AR-6 | «Применимость угрозы к активу» = (asset.type ∈ derived(threat.target_text)) **И** (хотя бы одно VL угрозы есть у актива). | Только presence VL без applicability. | Без applicability «BadUSB» применится к Database Server. Объект воздействия из ФСТЭК даёт чёткий тип. |
| AR-7 | Связь «CVE → ПО актива» — через **vendor + name match** на `bdu.software` ↔ `software_catalog`. Точное совпадение, без version range на первой итерации. | Полный CPE-парсинг (с диапазонами версий). | Версионные диапазоны в `bdu.software.version` строковые, разные форматы. Точное совпадение даёт быстрый MVP. Расширим если понадобится. |

## Граф зависимостей этапов

```
P1. cleanup migrations           ─┐
P2. import-fstec-threats         ─┼─► P5. vl_categories + threat↔VL
P3. import-bdu-snapshot          ─┘    ↓
P4. import-minreestr (parallel)        P6. asset_vulnerabilities = CVE detect
                                       ↓
                                       P7. risk service: applicability + W
                                       ↓
                                       P8. UI: asset card with VL/controls
                                       ↓
                                       P9. UI: threat catalog editor
```

P1-P4 можно делать параллельно (чистка + 3 импортёра). Дальше строго последовательно.

---

## P1 — Очистка legacy seeds

**Что:** новая миграция `030_clear_legacy_seeds.up.sql`. Чистит таблицы, которые сейчас наполнены сидерами:

```sql
TRUNCATE TABLE asset_vulnerabilities CASCADE;
TRUNCATE TABLE asset_controls CASCADE;
TRUNCATE TABLE threat_vulnerable_links, threat_destructive_actions, source_threats CASCADE;
TRUNCATE TABLE vulnerability_controls CASCADE;
TRUNCATE TABLE vulnerabilities CASCADE;
TRUNCATE TABLE controls CASCADE;
DELETE FROM threats;
```

`assets` оставляем — пользователь мог их добавить руками. Справочники (`asset_types`, `threat_categories`, `vulnerability_categories`, `control_types`) тоже оставляем — они пригодятся.

После этой миграции система пуста по части модели угроз. Пока импортёры не отработали — `/api/risk/overview` пустой. Это OK для развёрток, на тестовых системах админ запускает `make seed` (см. ниже).

**Тесты:** прогон testhelper-а должен показать пустые `threats`, `vulnerabilities`. Уже есть, нужно проверить что не падает.

**Объём:** 1 файл миграции + проверка.

## P2 — Импорт 227 угроз ФСТЭК (`thrlist.xlsx`)

**Что:** новая команда `cmd/import-fstec-threats/main.go`.

```go
import-fstec-threats \
  --source https://github.com/velvetway/bdu-fstec-mirror/raw/main/data/thrlist.xlsx \
  --dsn   $DB_DSN
```

Скачивает XLSX (≈100 КБ) если URL, иначе берёт локальный путь. Парсит через `github.com/xuri/excelize/v2`. Для каждой строки:

| XLSX поле | DB поле |
|---|---|
| Идентификатор УБИ | `threats.bdu_id` = `'УБИ.' + zfill(id, 3)` |
| Наименование УБИ | `threats.name` |
| Описание | `threats.description` |
| Источник угрозы | derived → `threats.source_type` (см. ниже) + `source_threats(threat_id, threat_source_id)` |
| Объект воздействия | новое поле `threats.applies_to_targets TEXT` (хранится сырой текст, плюс производное `applies_to_asset_types SMALLINT[]`) |
| Нарушение C | `threats.impact_c BOOLEAN` |
| Нарушение I | `threats.impact_i BOOLEAN` |
| Нарушение A | `threats.impact_a BOOLEAN` |
| Дата изменения | `threats.updated_at` |
| Статус угрозы | `threats.status TEXT` (Опубликована/В работе) |

**Производные поля.**

`source_type` от «Источник угрозы»:
- содержит «Внешний нарушитель» → external + связь S4 (хакеры)
- содержит «Внутренний нарушитель» → internal + связь S3 (персонал)
- оба → external (плюс обе связи S3, S4)

`applies_to_asset_types`: разбираем «Объект воздействия» через словарь регулярок:
```
'грид.*систем' / 'облач'           → cloud, server
'BIOS|UEFI|микропрограмм'           → server, workstation
'СУБД|database|базы данных'         → database
'веб.*сервер|web.*server'           → server, application
'сетевой трафик|network|маршрут'    → network
'ARM|рабоч.*станц|workstation'      → workstation
'мобильн|смартфон'                  → mobile
'IoT'                               → iot
…
```
Если ничего не подошло — пусто = «применимо ко всем» (безопасный дефолт).

`q_threat`, `q_severity` — берём детерминированно по описанию потенциала нарушителя:
- «низкий потенциал» → `q_threat = 0.3`
- «средний потенциал» → `q_threat = 0.6`
- «высокий потенциал» → `q_threat = 0.9`
- мульти-нарушители → max
- `q_severity` → `(impact_c + impact_i + impact_a) / 3`, минимум 0.3

**Миграция:** `031_threats_applicability.up.sql` добавляет:
- `threats.bdu_id VARCHAR(16)` (уже есть с `005`)
- `threats.applies_to_targets TEXT`
- `threats.applies_to_asset_types SMALLINT[]`
- `threats.impact_c BOOLEAN NOT NULL DEFAULT FALSE`
- `threats.impact_i BOOLEAN NOT NULL DEFAULT FALSE`
- `threats.impact_a BOOLEAN NOT NULL DEFAULT FALSE`
- `threats.status VARCHAR(64)`

(`impact_*` мы только что выпиливали в `020`. Возвращаем — теперь они нужны для derivации `q_severity` и для маппинга на DA.)

**Объём:** новая Go-команда (~400 строк), миграция (1 файл), словарь маппинга (~50 правил).

## P3 — Импорт снимка БДУ как локального справочника

**Что:** команда `cmd/import-bdu-snapshot/main.go`.

```go
import-bdu-snapshot \
  --source https://github.com/velvetway/bdu-fstec-mirror/raw/main/data/bdu.sqlite.gz \
  --target ./data/bdu.sqlite
```

Скачивает gzip, разжимает, кладёт в `./data/bdu.sqlite`. Один файл, ~470 МБ. Сервер при запуске открывает его в read-only режиме через отдельный пул `*sql.DB`.

**Никакой репликации в Postgres.** Это справочник, мы его только читаем.

В Go-коде новый пакет `internal/bdu`:
```go
type Snapshot interface {
    Find(ctx, query, filters) ([]BDUVuln, error)
    Get(ctx, bduID string) (*BDUVuln, error)
    SoftwareLookup(ctx, vendor, name string) ([]BDUVuln, error)  // для P6
}
```

Конфиг: `BDU_SNAPSHOT_PATH` env-var, по умолчанию `./data/bdu.sqlite`.

**Объём:** Go-команда (~150 строк), пакет `internal/bdu` с тонкой обёрткой над SQLite (~200 строк).

## P4 — Импорт каталога Минцифры

**Что:** команда `cmd/import-minreestr/main.go`. Тянет каталог через `minreestr-mcp` или прямой парсинг xn--80aajzhsbhw.xn--p1ai. Пишет в нашу `software_catalog`.

Поскольку в `software_catalog` сейчас 30+ продуктов руками — миграция `032_clear_software_catalog.up.sql`:
```sql
TRUNCATE software_catalog CASCADE;
```

Затем команда заполняет до 26 000 строк.

**Объём:** Go-команда (~200 строк), миграция (1 файл).

## P5 — Категории VL и связь ST ↔ VL

**Что:** новая таблица `vl_categories`, наполненная 6 строками из диплома. Перестраиваем граф `threat_vulnerable_links`: теперь это связь `threats ↔ vl_categories`, а не `threats ↔ vulnerabilities`.

**Миграция `033_vl_categories.up.sql`:**
```sql
CREATE TABLE vl_categories (
    id          SMALLSERIAL PRIMARY KEY,
    code        VARCHAR(8) NOT NULL UNIQUE,    -- VL1..VL6
    name        TEXT NOT NULL,
    description TEXT
);

INSERT INTO vl_categories (code, name, description) VALUES
    ('VL1', 'Нештатное дополнительное ПО', 'Драйверы, утилиты, неавторизованные расширения'),
    ('VL2', 'Устаревшие версии ПО или версии, имеющие уязвимости', 'CVE/БДУ-записи на установленном ПО'),
    ('VL3', 'Допустимость установки не декларируемого ПО', 'Возможность установки ПО вне корпоративного списка'),
    ('VL4', 'Наличие процедуры обхода администратором правил безопасности', 'Привилегированные обходы политик'),
    ('VL5', 'Носители информации', 'Флеш-накопители, жёсткие диски, съёмные носители'),
    ('VL6', 'Открытые ОС / отсутствие средств защиты ЛВС', 'Слабая периметровая защита');

-- Перестроить threat_vulnerable_links: теперь FK на vl_categories
ALTER TABLE threat_vulnerable_links
    DROP CONSTRAINT threat_vulnerable_links_vulnerability_id_fkey,
    ALTER COLUMN vulnerability_id TYPE SMALLINT,
    ALTER COLUMN vulnerability_id SET NOT NULL;
ALTER TABLE threat_vulnerable_links RENAME COLUMN vulnerability_id TO vl_category_id;
ALTER TABLE threat_vulnerable_links
    ADD CONSTRAINT threat_vl_categories_fk
    FOREIGN KEY (vl_category_id) REFERENCES vl_categories(id) ON DELETE CASCADE;
```

`vulnerability_controls` тоже надо переосмыслить: контроль закрывает **категорию VL**, не конкретную CVE. Меняем FK аналогично.

**Импортёр-расширение к P2:** после загрузки 227 угроз дописать связи `threat_vulnerable_links` через эвристику:
- угроза содержит «вирус|вредонос|malware» → VL2 + VL1
- «недекларируемое|неустановленное ПО» → VL3
- «обход.*правил|привилегий|администратор» → VL4
- «носител|флеш|USB» → VL5
- «сетев|перехват|трафик|сканирование» → VL6
- по умолчанию (нет совпадений) → VL2 (устаревшее ПО — самый общий случай)

**Тесты:** unit-тест на `applicability(asset, threat)` для нескольких типичных пар.

**Объём:** миграция (1 файл, ~80 строк SQL), словарь связей в импортёре, тесты.

## P6 — Автодетекция CVE на активах через установленное ПО

**Что:** при добавлении ПО к активу (`POST /api/assets/:id/software`) — сервис проходит по `bdu.software` через пакет `internal/bdu`, находит все БДУ-записи, у которых `vendor` и `name` совпадают с записью в `software_catalog`, и **заводит для них `asset_vulnerabilities`**.

```go
func (s *Service) AddSoftware(ctx, assetID, softwareID int64) error {
    sw := s.swRepo.Get(softwareID)
    bduMatches := s.bdu.SoftwareLookup(ctx, sw.Vendor, sw.Name)
    for _, b := range bduMatches {
        s.assetVulnRepo.Upsert(ctx, asset_vulnerability{
            AssetID:         assetID,
            BDUID:           b.ID,
            CWE:             b.CWEs[0],
            VLCategoryID:    cweToVLCategory(b.CWEs[0]),  // see below
            Status:          "open",
            DiscoveredAt:    now(),
            Source:          "auto:asset_software",
        })
    }
}
```

**`cweToVLCategory`** — статический map в Go:
```go
var cweToVL = map[string]int16{
    "CWE-506": 3,  // Embedded malicious code → VL3
    "CWE-507": 3,
    "CWE-269": 4,  // Improper privilege management → VL4
    "CWE-264": 4,
    "CWE-200": 6,  // Information exposure → VL6
    "CWE-22":  6,  // Path traversal
    "CWE-918": 1,  // SSRF → VL1 (нештатное поведение драйверов)
    "CWE-94":  1,  // Code injection
    // …default → VL2
}
```

**Меняется схема `asset_vulnerabilities`:** добавляются поля `bdu_id VARCHAR(16)`, `cwe VARCHAR(16)`, `vl_category_id SMALLINT`, `discovered_at`, `source VARCHAR(32)`. Старая FK на `vulnerabilities(id)` дропается (после P1 таблица всё равно пустая).

**Миграция `034_asset_vulnerabilities_redesign.up.sql`.**

**Тесты:** интеграционный — добавляем актив, привязываем «Astra Linux», ожидаем N открытых записей `asset_vulnerabilities` с правильным VL-категорией.

**Объём:** изменения в `internal/service/asset_vulnerability/`, новая миграция, ~300 строк кода, 2 теста.

## P7 — Risk service: applicability + новый Q^reaction

**Что:** переписываем `LoadVulnerableLinks` в `risk_graph_repository.go` так, чтобы он возвращал **по VL-категориям**, а не по CVE:

```sql
WITH threat_vls AS (                    -- VL-категории, к которым угроза прикреплена
    SELECT vl_category_id FROM threat_vulnerable_links WHERE threat_id = $2
),
asset_vl_present AS (                   -- какие VL-категории реально присутствуют у актива
    SELECT DISTINCT vl_category_id
    FROM asset_vulnerabilities
    WHERE asset_id = $1 AND status IN ('open','in_progress','mitigated')
      AND vl_category_id IS NOT NULL
),
asset_vl_covered AS (                   -- какие VL-категории закрыты внедрёнными контролями
    SELECT DISTINCT vc.vl_category_id
    FROM vulnerability_controls vc
    JOIN asset_controls ac ON ac.control_id = vc.control_id AND ac.asset_id = $1
)
SELECT vlc.id, vlc.code, vlc.name,
       (vlc.id IN (SELECT vl_category_id FROM asset_vl_covered)) AS covered,
       (vlc.id IN (SELECT vl_category_id FROM asset_vl_present)) AS present
FROM threat_vls tv
JOIN vl_categories vlc ON vlc.id = tv.vl_category_id;
```

`Q^reaction` теперь: `|covered ∩ present| / |present|`. Если `present = ∅` — пара не показывается.

`AssembleAssetAttackPaths` добавляет предварительный фильтр через `IsApplicable(asset, threat)`:

```go
func IsApplicable(a domain.Asset, t domain.Threat) bool {
    if len(t.AppliesToAssetTypes) > 0 && !contains(t.AppliesToAssetTypes, a.AssetTypeID) {
        return false
    }
    return true
}
```

(У `domain.Threat` появляются новые поля `AppliesToAssetTypes []int16`, `ImpactC/I/A bool` и т.д. — после P2.)

**Объём:** правки `internal/service/risk/`, обновление `domain.AttackPath` (`VulnerableLinks []VLNode` теперь по категориям), фронт `RiskGraphPage` подкручивается под новую структуру VL.

## P8 — UI: карточка актива с управлением VL/контролей

**Что:** новая страница `/assets/:id` с тремя секциями:

1. **«Установленное ПО»** — список из `asset_software`, кнопка «+ Добавить ПО» (поиск по `software_catalog`). При добавлении автоматически срабатывает P6 — пользователь видит, как появляются новые записи в секции уязвимостей.

2. **«Уязвимости и VL-категории»** — список `asset_vulnerabilities` сгруппированный по VL-категории. Видно: «VL2 (устаревшее ПО): 12 CVE, статусы: 8 open / 3 in_progress / 1 mitigated». Можно менять статус, можно отметить как «accepted».

3. **«Внедрённые контроли»** — список `asset_controls` с возможностью добавить из каталога `controls`. Каждый контроль показывает «закрывает VLk, VLm».

После любого изменения — кнопка «Пересчитать W» (или live-update через React Query).

**Объём:** новый экран AssetDetailPage (~600 строк фронта), новые API эндпоинты `POST/DELETE /api/assets/:id/software`, `POST/DELETE /api/assets/:id/controls`.

## P9 — UI: каталог угроз с возможностью править applicability

**Что:** страница `/threats`. Сейчас в `NavConfig.ts` есть пункт «Каталог угроз» с бейджем «БДУ», но реальной страницы нет (404). Делаем:
- Таблица 227 угроз с полями `bdu_id`, `name`, `q_threat`, `q_severity`, `applies_to_asset_types`, `impact_c/i/a`.
- Фильтры по типу актива, по источнику, по статусу.
- Карточка угрозы (drawer) с возможностью редактировать `applies_to_asset_types` и `applies_to_targets` (если эвристика P2 ошиблась).

**Объём:** новый экран ~400 строк, расширение `threat_handlers.go` (PUT с новыми полями).

---

## Что не входит в этот план (явно вне scope)

- Версионные диапазоны в маппинге CVE → ПО. Сейчас точное совпадение vendor+name. Расширим если окажется, что слишком много ложных срабатываний.
- Связь `threat ↔ destructive_action` динамически из `impact_c/i/a` (можно добавить тривиально, отдельный коммит).
- Импорт vullist.xlsx (29 МБ) — у нас уже есть `bdu.sqlite` со всеми же данными.
- Webhook/scheduled refresh снимка БДУ. Сейчас обновление — ручной запуск `import-bdu-snapshot`.

## Риски

| Риск | Вероятность | Митигация |
|---|---|---|
| Heuristic «Объект воздействия» → asset_type даёт много ложных применимостей | Средняя | UI редактирования applicability в P9. Можно ручкой поправить. |
| `bdu.software.name` ≠ `software_catalog.name` (форматы разные) | Высокая | Нормализация: lower + удалить пробелы/дефисы. На спорных — отчёт «не нашли пары» в логе импортёра. |
| 470 МБ SQLite в Docker-образе раздувает image | Средняя | Не кладём в image. Docker volume + `import-bdu-snapshot` при первом старте. |
| ETL-команды требуют `make`-обвязки | Низкая | Добавим Makefile-таргеты `make import-all`. |

## Объём работы по этапам

| Этап | Файлы (примерно) | Строки кода | Время |
|---|---|---|---|
| P1 cleanup | 1 sql | 30 | 30 мин |
| P2 fstec-threats | 5 (cmd, sql, mapping, tests) | 700 | 1 день |
| P3 bdu-snapshot | 4 (cmd, pkg, conf, tests) | 500 | 1 день |
| P4 minreestr | 3 (cmd, sql, tests) | 400 | 0.5 дня |
| P5 vl_categories | 2 (sql, mapping) | 200 | 0.5 дня |
| P6 cve detection | 4 (svc, sql, tests) | 500 | 1 день |
| P7 risk redesign | 6 (sql, svc, tests, frontend types) | 600 | 1 день |
| P8 asset detail UI | 5 (page, api, types, css) | 800 | 1.5 дня |
| P9 threat catalog UI | 4 (page, api, types) | 600 | 1 день |
| **Итого** | **~34 файла** | **~4 300 строк** | **7-8 дней** |

## Точки контроля

После каждого этапа — отдельный коммит, прогон `go test ./...` и `npm run build`. Между P5 и P6 — пауза для ревью схемы. После P7 — обязательная пауза, потому что меняется домен AttackPath. P8/P9 уже относительно безопасны.
