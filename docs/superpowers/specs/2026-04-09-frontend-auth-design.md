# Фронтенд-авторизация — Спецификация

## Обзор

Добавить авторизацию в React-фронтенд: логин, регистрация, защита роутов, токен в API-запросах, отображение роли пользователя.

## Новые файлы

| Файл | Назначение |
|------|-----------|
| `src/context/AuthContext.tsx` | AuthProvider + useAuth хук: токен, юзер, login/register/logout |
| `src/components/ProtectedRoute.tsx` | Редирект на /login если нет токена |
| `src/pages/LoginPage.tsx` | Форма логина |
| `src/pages/RegisterPage.tsx` | Форма регистрации с выбором роли |

## Изменения

| Файл | Изменения |
|------|-----------|
| `src/api/client.ts` | Authorization header + login/register API + auto-logout на 401 |
| `src/App.tsx` | AuthProvider, ProtectedRoute на все роуты кроме /login и /register, юзер в хедере |

## Детали

### AuthContext
- Хранит: token (string|null), user ({id, username, role}|null)
- Токен в localStorage (ключ "token")
- При инициализации проверяет localStorage
- Функции: login(username, password), register(username, password, role), logout()
- logout() чистит localStorage и стейт

### API Client
- Функция getToken() читает из localStorage
- Каждый request() добавляет Authorization: Bearer <token> если токен есть
- При 401 — вызывает window.dispatchEvent(new Event('auth:logout'))
- AuthContext слушает это событие и делает logout

### ProtectedRoute
- Если !token — Navigate to="/login"
- Иначе — рендерит children

### LoginPage
- Поля: username, password
- Кнопка "Войти"
- Ссылка "Нет аккаунта? Зарегистрироваться"
- При успехе — редирект на /

### RegisterPage
- Поля: username, password, role (select: viewer/auditor/admin)
- Кнопка "Зарегистрироваться"
- Ссылка "Уже есть аккаунт? Войти"
- При успехе — редирект на /login

### App.tsx хедер
- Показывает: имя пользователя, бейдж роли, кнопка "Выход"
- Роль viewer: скрыть кнопки создания/удаления на страницах
- Роль auditor: скрыть кнопки удаления
- Роль admin: всё доступно

## Что НЕ входит
- Refresh-токены
- Смена пароля
- Управление пользователями
