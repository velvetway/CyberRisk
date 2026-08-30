# Risk Graph Visualization — Design Spec

**Дата:** 2026-05-02
**Скоуп:** переработка `RiskGraphPage` для PTSZI-модели (`W_i = (Q^th + q + (1 − Q^re)) / 3 × Z`).
**Что НЕ меняется:** формула, бэкенд-калькулятор `internal/service/risk`, сущности `S/ST/VL/DA` и junction-таблицы.

## Цели

Текущая страница имеет четыре фундаментальные проблемы, которые этот дизайн решает:

1. **Толщины рёбер «из ниоткуда»** — `0.15` fan-out, `severity/10` без обоснования, веса не складываются ни в какую цельную метрику.
2. **Контроли (СЗИ) невидимы** — `coverage_controls` есть в `AttackPath`, но не отрисованы. Связь между контролем и `Q^reaction` не прослеживается.
3. **Нет drill-down** — узлы не кликабельны, нет popover'ов с описанием уязвимостей и контролей.
4. **Сравнивать нечего** — страница `1 актив × 1 угроза`. Невозможно увидеть, какая угроза опаснее остальных.

## Решение в одном предложении

Single-asset страница `Stacked` layout: hero-вердикт сверху, sortable list всех угроз актива в середине, развёрнутый Sankey-граф `S → ST → VL → C → DA` снизу с честной flow-семантикой и client-side what-if-симуляцией контролей.

## Аудитория и сценарии

Прогрессивное раскрытие для трёх ролей одновременно:

- **CISO** видит hero (W, level, число неприкрытых угроз) — достаточно для верхнеуровневого вердикта.
- **SOC-аналитик** скроллит список угроз, кликает в опасную, изучает граф, играет в what-if чтобы понять, какой контроль даёт наибольший выигрыш.
- **Аудитор** жмёт «PDF сценария» в hero или в правой панели — выгружает текущее состояние (включая активную what-if-симуляцию).

## Маршруты

- `/risk/graph/:assetId` — основная страница (без `?threat=...`): автоматически разворачивает угрозу с max W.
- `/risk/graph/:assetId?threat=<threatId>` — deep-link на конкретную угрозу. Параметр сохраняется в URL при выборе из списка (replaceState, не push), чтобы можно было поделиться ссылкой без засорения истории.

## Архитектура страницы

```
┌─ AssetRiskHero ─────────────────────────────────────────┐
│  A-001 «БД CRM»     W=0.84 [CRITICAL]   7 угроз / 3 неприкр.│
│  Кнопки: PDF сценария · Параметры                       │
└─────────────────────────────────────────────────────────┘

┌─ ThreatList ────────────────────────────────────────────┐
│  Sortable: W ↓ / непокрытие / название                  │
│  [▶] УБИ.094  ████████████░░  0.84  3 VL · 1 неприкр.   │
│  [ ] УБИ.044  ███████████░░░  0.78  2 VL · 0 неприкр.   │
│  …                                                       │
└─────────────────────────────────────────────────────────┘

┌─ AttackFlowDetail (одна угроза) ─────────────────────────────┐
│ ┌─ AttackFlowSankey (5 колонок) ─────┐ ┌─ PtsziBreakdown ─┐ │
│ │  S → ST → VL → C(СЗИ)               │ │  Q^th  ▓▓▓▓▓▓░  │ │
│ │              ↘ DA                    │ │  q     ▓▓▓▓░░░  │ │
│ └────────────────────────────────────┘ │  1−Q^re ▓▓▓▓▓▓▓ │ │
│ ┌─ WhatIfBar (sticky) ───────────────┐ │  Z      ▓▓▓▓▓▓▓ │ │
│ │ Симуляция: WAF выкл  ΔW +0.18     │ │  W = 0.84 [CRIT]│ │
│ │                  [Сбросить] […]   │ └─────────────────┘ │
│ └────────────────────────────────────┘                     │
└────────────────────────────────────────────────────────────┘
```

## Компоненты

| Компонент | Файл | Роль |
|---|---|---|
| `RiskGraphPage` | `frontend/src/pages/RiskGraphPage.tsx` | Оркестратор: загружает summary, держит `selectedThreatId`, `disabledControlIds`, рендерит секции |
| `AssetRiskHero` | `frontend/src/components/risk/AssetRiskHero.tsx` | Шапка: имя актива, агрегатный W, level, метрики |
| `ThreatList` | `frontend/src/components/risk/ThreatList.tsx` | Sortable список угроз с риск-барами; emits `onSelectThreat(id)` |
| `AttackFlowSankey` | `frontend/src/components/risk/AttackFlowSankey.tsx` | d3-sankey раскладка + кастомные React-узлы с hover/click |
| `ControlPopover` | `frontend/src/components/risk/ControlPopover.tsx` | Popover контроля (имя, coverage, описание, toggle «выкл в симуляции») |
| `WhatIfBar` | `frontend/src/components/risk/WhatIfBar.tsx` | Sticky-баннер при активной симуляции: список выключенных, ΔW, Reset, Save Note |
| `PtsziBreakdown` | `frontend/src/components/risk/PtsziBreakdown.tsx` | Перенос блока с формулой и слайдерами в отдельный модуль |

**Удаляется:** `frontend/src/components/RiskGraphSankey.tsx` — устаревший черновик, не используется на странице.

**Бизнес-логика (отдельный модуль для тестируемости):** `frontend/src/lib/riskFlow.ts` — функции `buildSankeyGraph(path, disabled)` и `recomputeW(path, disabled)`. Эти функции — чистые, без зависимостей от React, тестируются юнит-тестами.

## Sankey flow-семантика

Поток нормирован на `1.0` на входе. Каждое ребро — доля массы. Контроли — «дренаж».

| Ребро | Значение `value` |
|---|---|
| `S_i → ST` | `1 / N_sources` |
| `ST → VL_j` | `severity_j / Σ severity` (сумма по всем VL = 1.0) |
| `VL_j → C_k` | `(c_k / Σ c) × cov × vl_inflow`, где `cov = min(1, Σ c_k)` |
| `VL_j → DA_m` | `vl_inflow × (1 − cov) / N_DA` |

**Важное замечание: визуальный flow и формула — две разные модели.**

Бэкенд считает `Q^reaction = count(VL with ≥1 control with coverage>0) / count(VL)` — это бинарная метрика «прикрыта ли VL хотя бы одним контролем». Значения `coverage` (0..1 в БД) для этой формулы не используются.

Sankey же использует значения `coverage` для **визуальной** ширины VL→C / VL→DA рёбер — это даёт информативную картинку (видно, какой контроль реально много покрывает). Но **сумма в DA ≠ `1 − Q^reaction`** в общем случае.

В UI это явно помечено подзаголовком над графом: `Толщина VL→C — coverage контроля. q_reaction в формуле — доля VL, прикрытых хотя бы одним контролем.`

**Edge cases:**
- VL без контролей → `cov = 0` → весь поток в DA.
- Один контроль с `coverage = 1.0` → весь поток в C, 0 в DA (полный блок).
- Перекрывающиеся контроли (Σ > 1) → клампаем `cov` до 1.0, нормализуем доли через `c_k / Σ c`.
- VL с `severity = 0` (если такое попадётся) — заменяется на `1` чтобы не делить на ноль и не терять связь.

**Визуальные правила:**
- Цвет потока: `--risk-critical` для VL→DA, `--risk-info` или зелёный для VL→C (показывает блокировку).
- Узлы: те же стили, что в текущей `RiskGraphPage` (используем design-tokens после PR #2).
- Hover на узле → подсветка всех связанных рёбер (текущее поведение).
- Hover на ребре → tooltip `0.42 (42%) поток` + абс. значение.
- Click на C → `ControlPopover` (id, имя, coverage value, кнопка «Выкл в симуляции» / «Включить»).

## What-if симуляция

State в `RiskGraphPage`:
```ts
const [disabledControls, setDisabledControls] = useState<Set<number>>(new Set());
```

При изменении set вызывается `recomputeW(path, disabledControls)` из `riskFlow.ts` — **зеркалит бэкенд-логику `QReactionFromVLs` из `internal/service/risk/calculator_v2.go`**:

1. Per-VL: `is_covered_vl = ∃ enabled control with coverage > 0` (бинарно)
2. `q_reaction' = count(covered VLs) / count(VLs)` (если VL пусто — `q_reaction' = 0`)
3. `W' = (Q^th + q + (1 − q_reaction')) / 3 × Z` — `Q^th`, `q`, `Z` берутся из исходного path и не меняются
4. `ΔW = W' − W_baseline`, где `W_baseline` — оригинальный `path.w` от бэкенда

Sankey пересчитывает рёбра через `buildSankeyGraph(path, disabled)` — те же `coverage_controls`, но фильтрованные.

WhatIfBar появляется когда `disabledControls.size > 0`:
- Список выключенных контролей (chips с ✕)
- ΔW и стрелка `0.78 → 0.96`
- Кнопка `Сбросить` (`setDisabledControls(new Set())`)
- Кнопка `Сохранить заметку` — пишет JSON `{assetId, threatId, disabledIds, w, w_baseline, ts}` в `localStorage["risk:notes"]` (массив). Просмотр — V2.

**Точность:** клиент использует ту же формулу, что и бэкенд (`calculator_v2.QReactionFromVLs` + `(Q^th + q + (1 − Q^re)) / 3 × Z`). Расхождений быть не должно; если бэкенд формулу поменяет, нужно синхронно обновить `riskFlow.ts`. UI помечает значение `W'` как «оценка симуляции (клиент-сайд)» мелким шрифтом для прозрачности.

## API изменения

**Новый endpoint:**
```
GET /api/risk/asset/:asset_id/attack-paths
→ {
    asset:      { id, name },
    aggregate:  {
      w_max:            float,    // max(path.w) среди всех paths
      level:            string,   // level соответствующий w_max
      threat_count:     int,      // len(paths)
      uncovered_count:  int       // count(paths where exists VL with no covering controls)
    },
    paths:      [ AttackPath, ... ]
  }
```

Возвращает все `AttackPath` актива одним запросом. Реализация — bulk-метод в `risk_graph_repository.go`, который грузит список threat_ids для актива и батчем заполняет `AttackPath` для каждого.

Если у актива нет ни одной угрозы — возвращается `paths: []`, `aggregate: { w_max: 0, level: "info", threat_count: 0, uncovered_count: 0 }` и 200 OK. 404 — только если `asset_id` не существует.

**Существующий endpoint** `GET /api/risk/graph/:asset_id/:threat_id` — оставляем для deep-links и обратной совместимости. Возвращает single `AttackPath`.

**Авторизация:** новый endpoint попадает в ту же `readOnly` группу, что и существующий.

## Бэкенд: реализация

Файлы:
- `internal/transport/http/risk_handlers.go` — добавить handler `assetAttackPaths`.
- `internal/transport/http/server.go` — зарегистрировать `readOnly.Get("/risk/asset/:asset_id/attack-paths", riskHandler.assetAttackPaths)`.
- `internal/service/risk/aggregate.go` — pure helper `ComputeAssetAggregate`.
- `internal/service/risk/service.go` — bulk-метод `AssembleAssetAttackPaths(ctx, assetID)`.
- `internal/domain/risk_graph.go` — новый тип `AssetAttackPathsResponse` (asset, aggregate, paths).

**V1 — намеренное допущение по производительности:** `AssembleAssetAttackPaths` в V1 вызывает существующий `AssembleAttackPath(asset, threat)` в цикле по всем угрозам. Это `N+1` (≈4 round-trip × N угроз). Для типичных активов (10–30 угроз) это ≈40–120 запросов — приемлемо для V1. **V2:** добавить `LoadAssetAttackPaths` bulk-метод на `RiskGraphRepository` (один запрос для списка threat_ids, один для всех VL+controls по batch, один для всех DA по batch) + `TestLoadAssetAttackPaths_NoNPlusOne` контракт-тест.

## Фронтенд: реализация

Файлы:
- `frontend/src/lib/riskFlow.ts` — новый, pure-функции для расчёта Sankey и W.
- `frontend/src/types/riskGraph.ts` — расширить типами `AssetAttackPathsResponse`.
- `frontend/src/pages/RiskGraphPage.tsx` — переписать (orchestrator).
- `frontend/src/components/risk/*.tsx` — новые компоненты (см. таблицу).
- `frontend/src/components/RiskGraphSankey.tsx` — удалить.
- `frontend/package.json` — `d3-sankey` уже подключён, доп. зависимостей не нужно.

Стили — через существующие design-tokens (CSS-переменные `--risk-*`, `--bg-elev-*`, `--font-mono`). Никаких новых.

## Тесты

**Бэкенд (Go):**
- `aggregate_test.go` → 5 unit-тестов на `ComputeAssetAggregate`: empty, two paths (один с uncovered VL), несколько uncovered VL в одном path, none-uncovered, path с пустыми VL.
- `risk_handlers_test.go` → `TestAssetAttackPaths_HappyPath` — 200, корректный shape, aggregate.w_max совпадает с max из paths.
- `risk_handlers_test.go` → `TestAssetAttackPaths_InvalidID` — 400 для нечислового asset_id.
- *V2:* `TestLoadAssetAttackPaths_NoNPlusOne` появится одновременно с bulk repository-методом.

**Фронтенд (vitest):**
- `riskFlow.test.ts` → `flow conservation`: на синтетическом path `Σ(VL→C) + Σ(VL→DA) ≈ vl_inflow ± 1e-9` для каждого VL.
- `riskFlow.test.ts` → `dropping last control increases W`: выключение всех контролей VL → `W' > W`, и `q_reaction' = 0` для этого VL.
- `riskFlow.test.ts` → `coverage clamping`: VL с тремя контролями `0.6/0.5/0.4` → агрегат `1.0`, рёбра нормализованы.
- `RiskGraphPage.test.tsx` → snapshot базового состояния и состояния after disabling control.
- `ThreatList.test.tsx` → сортировка стабильна, выбор обновляет URL.

**Доступность:**
- Все интерактивные узлы имеют `tabIndex`, `role`, `aria-label`.
- Tab-навигация по списку угроз и узлам графа; Enter открывает popover; Escape закрывает.

## Состояния страницы

| Состояние | Что рендерим |
|---|---|
| Loading (initial) | Скелетон: серая полоса hero, 3 серых строки threat-list, серый прямоугольник графа |
| Empty (актив без угроз) | Hero показывает `aggregate.threat_count = 0`, ниже card-сообщение «У актива нет связанных угроз» |
| Error (network / 404 / 5xx) | Card-сообщение с текстом ошибки и кнопкой «Назад» |
| Loaded, нет `?threat=` | Авто-выбор угрозы с max W, URL обновляется `replaceState` |
| Loaded, `?threat=<id>` валиден | Угроза подсвечена и развёрнута |
| Loaded, `?threat=<id>` невалиден | Падаем в авто-выбор + предупреждение «Угрозу id=N не найдено у актива» (toast или inline notice) |

## Out of scope (явно)

- Сохранение what-if-сценариев в БД (V1 — только `localStorage`).
- Кросс-актив сравнение (это `RiskMapPage`, не трогаем здесь).
- Перенос/изменение формулы PTSZI.
- ATT&CK / D3FEND-маппинг.
- Анимации перерасчёта (могут добавиться в V2 polish).

## Critical files (для импл.-плана)

```
NEW:
  frontend/src/lib/riskFlow.ts
  frontend/src/lib/riskFlow.test.ts
  frontend/src/components/risk/AssetRiskHero.tsx
  frontend/src/components/risk/ThreatList.tsx
  frontend/src/components/risk/AttackFlowSankey.tsx
  frontend/src/components/risk/ControlPopover.tsx
  frontend/src/components/risk/WhatIfBar.tsx
  frontend/src/components/risk/PtsziBreakdown.tsx

REWRITTEN:
  frontend/src/pages/RiskGraphPage.tsx

EXTENDED:
  internal/transport/http/risk_handlers.go
  internal/transport/http/server.go
  internal/repository/risk_graph_repository.go
  internal/domain/risk_graph.go
  frontend/src/types/riskGraph.ts

DELETED:
  frontend/src/components/RiskGraphSankey.tsx
```
