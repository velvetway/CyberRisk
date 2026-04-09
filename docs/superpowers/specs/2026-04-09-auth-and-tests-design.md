# Авторизация и тестирование — Спецификация

## Обзор

Добавить JWT-авторизацию с ролевой моделью (admin/auditor/viewer) и тесты (unit + интеграционные) для проекта CyberRisk.

## 1. Авторизация

### 1.1 Стек

- Хеширование паролей: `bcrypt` (golang.org/x/crypto/bcrypt)
- JWT: `github.com/golang-jwt/jwt/v5`
- Время жизни access-токена: 15 минут
- Refresh-токен: не используется
- Конфигурация: переменная окружения `JWT_SECRET`

### 1.2 Новые файлы

| Файл | Назначение |
|------|-----------|
| `internal/service/auth/service.go` | Register, Login, ValidateToken, HashPassword |
| `internal/transport/http/auth_handlers.go` | POST /api/auth/register, POST /api/auth/login, GET /api/auth/me |
| `internal/transport/http/middleware.go` | JWT middleware + role checker |

### 1.3 Эндпоинты

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/api/auth/register` | Публичный | Регистрация пользователя |
| POST | `/api/auth/login` | Публичный | Вход, возвращает JWT |
| GET | `/api/auth/me` | Авторизованный | Информация о текущем пользователе |

### 1.4 Ролевая модель

| Роль | GET (чтение) | POST/PUT (изменение) | DELETE (удаление) |
|------|-------------|---------------------|-------------------|
| viewer | Все данные | Нет | Нет |
| auditor | Все данные | Расчёт рисков, создание/редактирование активов | Нет |
| admin | Все данные | Всё | Всё |

### 1.5 Middleware-цепочка

```
Запрос → CORS → JWT Middleware → Role Check → Handler
```

Публичные роуты (`/api/auth/*`, `/health`) обходят JWT middleware.

### 1.6 JWT Claims

```go
type Claims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

### 1.7 Хранение

Используется существующая таблица `users` (миграция 001). Поля: id, username, password_hash, role, is_active, created_at, updated_at.

Новая миграция: добавить уникальный индекс на `username` если отсутствует.

### 1.8 Конфигурация

Новая переменная в `config.Config`:
- `JWT_SECRET` (string, обязательно в проде, дефолт `dev-secret-change-me` для разработки)

## 2. Тестирование

### 2.1 Unit-тесты

| Файл | Тестов | Что покрывает |
|------|--------|---------------|
| `risk/calculator_test.go` | ~15 | Impact, likelihood, score, risk level, regulatory factors, edge cases |
| `risk/recommendations_test.go` | ~8 | Генерация рекомендаций по типам угроз, регуляторные рекомендации |
| `asset/service_test.go` | ~5 | CRUD операции, валидация входных данных |
| `asset/cia_calculator_test.go` | ~5 | Расчёт CIA метрик по разным сценариям |
| `auth/service_test.go` | ~5 | Register, Login, ValidateToken, ошибки |

Сервисы тестируются через интерфейсы репозиториев (моки не нужны — передаём фейковые реализации или тестируем чистые функции).

### 2.2 Интеграционные тесты

| Файл | Тестов | Что покрывает |
|------|--------|---------------|
| `repository/testhelper_test.go` | — | Setup PostgreSQL testcontainer + прогон миграций |
| `repository/asset_repository_test.go` | ~5 | Create, GetByID, List, Update, Delete с реальной БД |
| `repository/threat_repository_test.go` | ~3 | CRUD угроз |
| `repository/vulnerability_repository_test.go` | ~3 | CRUD уязвимостей |
| `repository/software_repository_test.go` | ~3 | CRUD ПО, фильтры |

Используется `testcontainers-go` для поднятия PostgreSQL в Docker на время тестов.

### 2.3 Зависимости для тестов

- `github.com/testcontainers/testcontainers-go` — PostgreSQL контейнер
- `github.com/stretchr/testify` — assertions (assert/require)

## 3. Порядок реализации

1. Авторизация: auth service → handlers → middleware → интеграция в роутер
2. Миграция: индекс на username
3. Unit-тесты: calculator → recommendations → CIA → asset service → auth service
4. Интеграционные тесты: testhelper → repositories

## 4. Что НЕ входит

- Refresh-токены
- OAuth / SSO
- Тесты HTTP-хендлеров
- CI/CD пайплайн
- Фронтенд-интеграция авторизации
