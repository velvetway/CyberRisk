# CyberRisk Platform

Комплексная платформа по управлению киберрисками и каталогом российского ПО. Репозиторий объединяет сервер на Go/Fiber, PostgreSQL, слой расчёта рисков и SPA‑фронтенд на React/TypeScript с полностью русским интерфейсом.

## Содержание

1. [Архитектура](#архитектура)
2. [База данных и миграции](#база-данных-и-миграции)
3. [REST API](#rest-api)
4. [Расчёт рисков и рекомендации](#расчёт-рисков-и-рекомендации)
5. [Фронтенд](#фронтенд)
6. [Запуск локальной среды](#запуск-локальной-среды)
7. [PDF‑отчёты](#pdf-отчёты)
8. [Тесты и проверка качества](#тесты-и-проверка-качества)
9. [Дальнейшее развитие](#дальнейшее-развитие)

## Архитектура

```
├─ cmd/server/main.go     — точка входа Fiber‑сервера
├─ internal/
│  ├─ config              — загрузка переменных окружения
│  ├─ domain              — доменные модели и enum‑ы
│  ├─ repository          — доступ к PostgreSQL (pgx)
│  ├─ service             — бизнес‑логика (активы, угрозы, риск, ПО)
│  ├─ report              — генерация PDF с gofpdf
│  └─ transport/http      — HTTP‑хендлеры Fiber
├─ migrations/            — версия БД и тестовые данные
└─ frontend/              — SPA на React 19 + TypeScript
```

- **Backend**: Go 1.22, HTTP‑сервер на Fiber v2, подключение к PostgreSQL через `pgxpool`. Основной исполняемый файл `cmd/server/main.go` создаёт контекст с graceful shutdown, загружает конфиг (`DB_DSN`, `HTTP_PORT`), инициализирует репозитории/сервисы и регистрирует HTTP‑роуты.
- **Слой конфигурации**: `internal/config/config.go` устанавливает значения по умолчанию (`postgres://app:app@localhost:5432/cyber_risk?sslmode=disable`, порт `8081`), что облегчает запуск в dev.
- **Сервисы**:
  - `service/asset`, `service/threat`, `service/vulnerability`, `service/asset_vulnerability` реализуют CRUD‑операции и валидацию вводимых данных.
  - `service/software` управляет справочником ПО, поиском российских аналогов и фильтром по сертификации ФСТЭК/ФСБ.
  - `service/risk` содержит калькулятор, правила генерации рекомендаций и агрегирующие методы (`PreviewRisk`, `Overview`, `AssetRiskProfile`).
- **Transport layer**: `internal/transport/http/server.go` создаёт Fiber‑приложение, включает CORS (`*`), panic‑recover и health‑check `/health`. Внутри роутера `/api` регистрируются хендлеры активов, угроз, уязвимостей, связей актив↔уязвимость, сервиса риска и каталога ПО.
- **Отчётность**: `internal/report` содержит шаблоны и встроенные шрифты для генерации PDF отчёта по конкретному сценарию риска.

## База данных и миграции

PostgreSQL — единственный источник данных. Миграции (каталог `migrations`) поддерживаются утилитой [golang-migrate](https://github.com/golang-migrate/migrate). Основные фичи:

- `001_init` — базовые таблицы активов, угроз, уязвимостей, связей и справочников.
- `002_seed_data` — стартовые записи активов, угроз, уязвимостей для демонстрационного профиля.
- `004_software_catalog_and_data_categories` + `006_seed_software_catalog` — схема и сид‑данные каталога ПО с принадлежностью к реестру Минцифры и сертификатам ФСТЭК/ФСБ.
- Каждая миграция имеет `.up.sql` и `.down.sql`; запуск осуществляется через контейнер `migrate` в `docker-compose.yml`.

### Подключение к БД

```bash
export DB_DSN="postgres://app:app@localhost:5432/cyber_risk?sslmode=disable"
export HTTP_PORT=8081
go run cmd/server/main.go
```

Пул `pgxpool` создаётся один раз и переиспользуется всеми репозиториями.

## REST API

Все эндпоинты префиксированы `/api` и возвращают/принимают JSON. Ошибки возвращаются в виде `{ "error": "<текст>" }`.

| Метод | Путь | Назначение |
|-------|------|------------|
| GET `/api/assets` | Список активов с пагинацией (`limit`, `offset`) |
| POST `/api/assets` | Создать актив (обязательны `name`, CIA‑оценки, критичность) |
| GET `/api/assets/:id` | Получить актив |
| PUT `/api/assets/:id` | Обновить актив |
| DELETE `/api/assets/:id` | Удалить актив |
| GET `/api/assets/:assetID/vulnerabilities` | Связанные уязвимости |
| POST `/api/assets/:assetID/vulnerabilities` | Привязать уязвимость к активу |
| DELETE `/api/assets/:assetID/vulnerabilities/:vulnID` | Отвязать |
| GET `/api/threats` | Каталог угроз (БДУ ФСТЭК, вектор атаки, влияния на CIA) |
| GET `/api/vulnerabilities` | Каталог уязвимостей |
| POST `/api/risk/preview` | Рассчитать риск для пары `asset_id` + `threat_id`, вернуть `impact`, `likelihood`, `score`, `level` и рекомендации |
| GET `/api/risk/overview` | Матрица всех рисков (каждая точка — актив × угроза) |
| GET `/api/risk/asset/:id` | Профиль рисков для выбранного актива |
| POST `/api/risk/report/pdf` | PDF отчёт по конкретному сценарию |
| GET `/api/software` | Каталог ПО (фильтры: `limit`, `offset`, `category_id`, `is_russian`, `fstec_only`, `search`) |
| POST `/api/software` | Добавить ПО в справочник |
| GET `/api/software/:id` | Детальная карточка ПО |
| PUT `/api/software/:id` | Обновление карточки |
| DELETE `/api/software/:id` | Удаление |
| GET `/api/software/categories` | Категории ПО |
| GET `/api/software/russian` | Российские аналоги для категории |
| GET `/api/software/certified` | Сертифицированное ПО (ФСТЭК/ФСБ) |

## Расчёт рисков и рекомендации

Формулы сконцентрированы в `internal/service/risk/calculator.go`:

- `Impact = f(BusinessCriticality, Confidentiality, Integrity, Availability)` + поправки за наличие угрозы, влияющей на соответствующие CIA‑метрики, и максимальную серьёзность уязвимости, привязанной к активу.
- `Likelihood = BaseLikelihood(угрозы) + среда эксплуатации + финальные корректировки` (изоляция сегмента, наличие доступа в интернет, вектор атаки).
- `Score = Impact × Likelihood`, диапазон 1–25, уровни риска: `Critical` (≥16), `High` (≥11), `Medium` (≥6), `Low` (остальное).
- `RegulatoryFactor` усиливает оценку в зависимости от полей актива:
  - Категории КИИ (187‑ФЗ), уровни защищённости ПДн (ПП‑1119), тип обрабатываемых данных (гостайна, коммерческая тайна и т.д.).
  - Объём персональных данных > 1 000 или > 100 000 субъектов увеличивает множитель.
- `AdjustedScore = Score × RegulatoryFactor`, при необходимости применяется функция `AdjustedRiskLevel`.

`service/risk/recommendations.go` содержит rule‑engine, который на основании атрибутов актива/угрозы/уязвимости формирует список рекомендаций:

- Категории: защита периметра, управление уязвимостями, резервное копирование, регуляторные требования 152‑ФЗ/187‑ФЗ и т.д.
- Для регуляторных рекомендаций выводится блок с российскими СЗИ (`RussianTool`) с пометкой о сертификатах ФСТЭК/ФСБ и номере в реестре Минцифры.

### Производительность

Метод `Overview()` кеширует уязвимости по активу в памяти, чтобы не выполнять `N×M` запросов в таблицу связей. Репозитории используют ограничение по `LIMIT/OFFSET` по умолчанию (50 строк).

## Фронтенд

Каталог `frontend/` — Create React App (React 19, TypeScript, D3). Структура:

- `src/api/client.ts` — типобезопасный HTTP клиент, повторно используемый на страницах.
- `src/types.ts` — общие типы, синхронизированные с backend JSON полями (`snake_case`).
- `src/pages/` — основные сценарии:
  - `AssetsPage`, `AssetFormPage`, `AssetRiskProfilePage`.
  - `RiskPreviewPage` — интерактивный подбор угрозы/актива и просмотр рекомендаций.
  - `RiskMapPage` — d3‑матрица 5×5 с heatmap, подсчётом статистики по уровням риска и панелью деталей при наведении.
  - `SoftwareCatalogPage` — карточный или табличный каталог ПО с поиском, фильтром по категориям/реестру/сертификации и статистикой.
- `src/components/` — переиспользуемые карточки, таблицы, фильтры.
- `src/utils/i18n.ts` — словарь и helper `getRiskLevelLabel` для вывода русских названий уровней.

`package.json` настроен с `"proxy": "http://localhost:8081"`, поэтому `npm start` автоматически проксирует API на порт Go‑сервера.

## Запуск локальной среды

1. **Запустить PostgreSQL и миграции**:
   ```bash
   docker-compose up -d postgres
   docker-compose run --rm migrate
   ```
   Таблицы и демо‑данные загрузятся автоматически.
2. **Backend**:
   ```bash
   cd /path/to/Diplom
   export DB_DSN="postgres://app:app@localhost:5432/cyber_risk?sslmode=disable"
   export HTTP_PORT=8081
   go run cmd/server/main.go
   ```
   Health‑check доступен по `http://localhost:8081/health`.
3. **Frontend**:
   ```bash
   cd frontend
   npm install
   npm start        # dev-сервер CRA на http://localhost:3000
   npm run build    # production-сборка в frontend/build
   ```

Для боевого развёртывания можно отдавать содержимое `frontend/build` статическим веб‑сервером, а Fiber запускать отдельным сервисом (Dockerfile не входит, но конфигурация прозрачна).

## PDF-отчёты

POST `/api/risk/report/pdf` принимает `asset_id` и `threat_id`, повторно использует `PreviewRisk`, рендерит шаблон в `internal/report/templates` и возвращает файл с заголовками `Content-Type: application/pdf`, `Content-Disposition: attachment; filename=risk_report.pdf`. В отчёт включаются таблицы с исходными параметрами актива, угрозы, рассчитанные метрики, список рекомендаций и реестр российских средств защиты.

## Тесты и проверка качества

- **Backend**: на текущем этапе модульные тесты не добавлены. Рекомендуется покрыть критичную арифметику (`risk.Calculator`) и репозитории (`pgxpool`) через integration‑тесты (`go test ./...`).
- **Frontend**: доступна стандартная инфраструктура CRA (`npm test`) с `@testing-library`. В проекте присутствует базовый `App.test.tsx`; добавляйте сценарные тесты для страниц расчёта риска и каталога ПО.
- **Static checks**: используйте `golangci-lint` и `eslint`/`tsc --noEmit` при CI.

## Дальнейшее развитие

- Добавить аутентификацию и разграничение ролей (enum `domain.UserRole` подготовлен).
- Вести историю изменений оценок риска и статус рекомендаций.
- Поддержать загрузку собственных шаблонов отчётов и локализаций.
- Расширить API фильтрацией (например, по категориям угроз) и пагинацией на уровне карт рисков.

Документация должна помочь быстро разобраться в проекте, расширить модули и безопасно развернуть систему. За дополнительными вопросами обращайтесь в issues репозитория.
# CyberRisk
